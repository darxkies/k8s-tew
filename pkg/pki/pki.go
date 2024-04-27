package pki

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/darxkies/k8s-tew/pkg/utils"
)

func GenerateEncryptionConfig() (string, error) {
	buffer := make([]byte, 32)

	_, error := rand.Read(buffer)

	if error != nil {
		return "", error
	}

	return base64.StdEncoding.EncodeToString(buffer), nil
}

func newBigInt() (*big.Int, error) {
	return rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 160))
}

func bigIntHash() ([]byte, error) {
	number, error := newBigInt()
	if error != nil {
		return nil, error
	}

	hash := sha1.New()

	_, error = hash.Write(number.Bytes())
	if error != nil {
		return nil, error
	}

	return hash.Sum(nil), nil
}

type CertificateAndPrivateKey struct {
	CertificateFilename string
	PrivateKeyFilename  string
	Certificate         *x509.Certificate
	PrivateKey          *rsa.PrivateKey
}

func loadPEMBlock(filename string) (*pem.Block, error) {
	file, error := os.Open(filename)
	if error != nil {
		return nil, error
	}

	defer file.Close()

	raw, error := io.ReadAll(file)
	if error != nil {
		return nil, error
	}

	block, _ := pem.Decode(raw)

	return block, nil
}

func LoadCertificateAndPrivateKey(certificateFilename, privateKeyFilename string) (*CertificateAndPrivateKey, error) {
	result := &CertificateAndPrivateKey{CertificateFilename: certificateFilename, PrivateKeyFilename: privateKeyFilename}

	block, error := loadPEMBlock(certificateFilename)
	if error != nil {
		return nil, error
	}

	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("wrong certificate format in '%s'", certificateFilename)
	}

	result.Certificate, error = x509.ParseCertificate(block.Bytes)
	if error != nil {
		return nil, error
	}

	block, error = loadPEMBlock(privateKeyFilename)
	if error != nil {
		return nil, error
	}

	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("wrong private key format in '%s'", privateKeyFilename)
	}

	result.PrivateKey, error = x509.ParsePKCS1PrivateKey(block.Bytes)
	if error != nil {
		return nil, error
	}

	return result, nil
}

func newTemplate(validityPeriod int, commonName, organization string) (*x509.Certificate, error) {
	serialNumber, error := newBigInt()
	if error != nil {
		return nil, error
	}

	subjectKeyId, error := bigIntHash()
	if error != nil {
		return nil, error
	}

	now := time.Now()

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		SubjectKeyId: subjectKeyId,
		Subject:      pkix.Name{},
		NotBefore:    now.Add(-5 * time.Minute),
		NotAfter:     now.AddDate(validityPeriod, 0, 0),
	}

	if len(commonName) > 0 {
		template.Subject.CommonName = commonName
	}

	if len(organization) > 0 {
		template.Subject.Organization = []string{organization}
	}

	return template, nil
}

func createAndSaveCertificate(signer *CertificateAndPrivateKey, template *x509.Certificate, rsaSize int, certificateFilename, privateKeyFilename string) error {
	var error error

	privateKey, error := rsa.GenerateKey(rand.Reader, rsaSize)
	if error != nil {
		return error
	}

	if signer == nil {
		signer = &CertificateAndPrivateKey{Certificate: template, PrivateKey: privateKey}
	}

	certificateData, error := x509.CreateCertificate(rand.Reader, template, signer.Certificate, &privateKey.PublicKey, signer.PrivateKey)
	if error != nil {
		return error
	}

	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateData})

	if error := os.WriteFile(certificateFilename, certificatePEM, 0644); error != nil {
		return error
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	if error := os.WriteFile(privateKeyFilename, privateKeyPEM, 0644); error != nil {
		return error
	}

	utils.LogFilename("Generated", certificateFilename)
	utils.LogFilename("Generated", privateKeyFilename)

	return nil
}

func GenerateCA(rsaSize uint16, validityPeriod uint, commonName, organization, certificateFilename, privateKeyFilename string) error {
	if utils.FileExists(certificateFilename) && utils.FileExists(privateKeyFilename) {
		utils.LogDebugFilename("Skipped", certificateFilename)
		utils.LogDebugFilename("Skipped", privateKeyFilename)

		return nil
	}

	template, error := newTemplate(int(validityPeriod), commonName, organization)
	if error != nil {
		return error
	}

	template.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	template.BasicConstraintsValid = true
	template.IsCA = true
	template.MaxPathLen = 2

	return createAndSaveCertificate(nil, template, int(rsaSize), certificateFilename, privateKeyFilename)
}

func GenerateClient(signer *CertificateAndPrivateKey, rsaSize uint16, validityPeriod uint, commonName, organization string, dnsNames []string, ipAddresses []string, certificateFilename, privateKeyFilename string, update bool) error {
	if utils.FileExists(certificateFilename) && utils.FileExists(privateKeyFilename) && !update {
		utils.LogDebugFilename("Skipped", certificateFilename)
		utils.LogDebugFilename("Skipped", privateKeyFilename)

		return nil
	}

	template, error := newTemplate(int(validityPeriod), commonName, organization)
	if error != nil {
		return error
	}

	template.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}

	template.IPAddresses = []net.IP{}

	for _, ipString := range ipAddresses {
		ipAddress := net.ParseIP(ipString)

		if ipAddress == nil {
			return fmt.Errorf("'%s' is not a valid IP address", ipString)
		}

		template.IPAddresses = append(template.IPAddresses, ipAddress)
	}

	template.DNSNames = dnsNames

	stringSlicesAreEqual := func(firstSlice, secondSlice []string) bool {
		if len(firstSlice) != len(secondSlice) {
			return false
		}

		for _, first := range firstSlice {
			found := false

			for _, second := range secondSlice {
				if second == first {
					found = true
					break
				}
			}

			if !found {
				return false
			}
		}

		return true
	}

	ipSlicesAreEqual := func(firstSlice, secondSlice []net.IP) bool {
		if len(firstSlice) != len(secondSlice) {
			return false
		}

		for _, first := range firstSlice {
			found := false

			for _, second := range secondSlice {
				if second.Equal(first) {
					found = true
					break
				}
			}

			if !found {
				return false
			}
		}

		return true
	}

	if utils.FileExists(certificateFilename) && utils.FileExists(privateKeyFilename) {
		oldCertificateAndPrivateKey, _error := LoadCertificateAndPrivateKey(certificateFilename, privateKeyFilename)

		if _error == nil {
			dnsMatch := stringSlicesAreEqual(oldCertificateAndPrivateKey.Certificate.DNSNames, template.DNSNames)
			ipMatch := ipSlicesAreEqual(oldCertificateAndPrivateKey.Certificate.IPAddresses, template.IPAddresses)
			organizationMatch := stringSlicesAreEqual(oldCertificateAndPrivateKey.Certificate.Subject.Organization, template.Subject.Organization)
			commonNameMatch := oldCertificateAndPrivateKey.Certificate.Subject.CommonName == template.Subject.CommonName

			if dnsMatch && ipMatch && organizationMatch && commonNameMatch {
				utils.LogDebugFilename("Skipped", certificateFilename)
				utils.LogDebugFilename("Skipped", privateKeyFilename)

				return nil
			}
		}
	}

	return createAndSaveCertificate(signer, template, int(rsaSize), certificateFilename, privateKeyFilename)
}
