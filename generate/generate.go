package generate

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/darxkies/k8s-tew/config"

	"github.com/darxkies/k8s-tew/pki"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
)

type Generator struct {
	config               *config.InternalConfig
	rsaSize              int
	caValidityPeriod     int
	clientValidityPeriod int
}

func NewGenerator(config *config.InternalConfig, rsaSize, caValidityPeriod int, clientValidityPeriod int) *Generator {
	return &Generator{config: config, rsaSize: rsaSize, caValidityPeriod: caValidityPeriod, clientValidityPeriod: clientValidityPeriod}
}

func (generator *Generator) generateServiceFile() error {
	fullServiceConfigFilename := generator.config.GetFullDeploymentFilename(utils.SERVICE_CONFIG)
	command := generator.config.GetFullTargetDeploymentFilename(utils.K8S_TEW_BINARY) + " run --base-directory=/"
	serviceConfigContent := fmt.Sprintf(utils.SERVICE_CONFIG_TEMPLATE, utils.PROJECT_TITLE, command, utils.K8S_TEW_BINARY)

	if utils.FileExists(fullServiceConfigFilename) {
		log.WithFields(log.Fields{"filename": fullServiceConfigFilename}).Info("skipping")

	} else {
		if error := ioutil.WriteFile(fullServiceConfigFilename, []byte(serviceConfigContent), 0644); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": fullServiceConfigFilename}).Info("generated")
	}

	return nil
}

func (generator *Generator) generateCNIFiles() error {
	fullNetConfigFilename := generator.config.GetFullDeploymentFilename(utils.NET_CONFIG)
	fullCNIConfigFilename := generator.config.GetFullDeploymentFilename(utils.CNI_CONFIG)

	if utils.FileExists(fullNetConfigFilename) {
		log.WithFields(log.Fields{"filename": fullNetConfigFilename}).Info("skipping")

	} else {
		if error := ioutil.WriteFile(fullNetConfigFilename, []byte(utils.NET_CONFIG_TEMPLATE), 0644); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": fullNetConfigFilename}).Info("generated")
	}

	if utils.FileExists(fullCNIConfigFilename) {
		log.WithFields(log.Fields{"filename": fullCNIConfigFilename}).Info("skipping")

	} else {
		if error := ioutil.WriteFile(fullCNIConfigFilename, []byte(utils.CNI_CONFIG_TEMPLATE), 0644); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": fullCNIConfigFilename}).Info("generated")
	}

	return nil
}

func (generator *Generator) generateK8SKubeletConfigFile() error {
	fullFilename := generator.config.GetFullDeploymentFilename(utils.K8S_KUBELET_CONFIG)

	if utils.FileExists(fullFilename) {
		log.WithFields(log.Fields{"filename": fullFilename}).Info("skipping")

	} else {

		if error := ioutil.WriteFile(fullFilename, []byte(utils.K8S_KUBELET_CONFIG_TEMPLATE), 0644); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": fullFilename}).Info("generated")
	}

	return nil
}

func (generator *Generator) generateK8SAdminUserConfigFile() error {
	fullFilename := generator.config.GetFullDeploymentFilename(utils.K8S_ADMIN_USER_CONFIG)

	if utils.FileExists(fullFilename) {
		log.WithFields(log.Fields{"filename": fullFilename}).Info("skipping")

	} else {

		if error := ioutil.WriteFile(fullFilename, []byte(utils.K8S_ADMIN_USER_CONFIG_TEMPLATE), 0644); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": fullFilename}).Info("generated")
	}

	return nil
}

func (generator *Generator) generateEncryptionFile() error {
	fullEncryptionConfigFilename := generator.config.GetFullDeploymentFilename(utils.ENCRYPTION_CONFIG)

	if utils.FileExists(fullEncryptionConfigFilename) {
		log.WithFields(log.Fields{"filename": fullEncryptionConfigFilename}).Info("skipping")

	} else {

		if error := pki.GenerateEncryptionConfig(fullEncryptionConfigFilename); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": fullEncryptionConfigFilename}).Info("generated")
	}

	return nil
}

func (generator *Generator) generateCAFiles(fullCAFilename, fullCAKeyFilename string) error {
	if utils.FileExists(fullCAFilename) {
		log.WithFields(log.Fields{"filename": fullCAFilename}).Info("skipping")
		log.WithFields(log.Fields{"filename": fullCAKeyFilename}).Info("skipping")

	} else {
		if error := pki.GenerateCA(generator.rsaSize, generator.caValidityPeriod, "Kubernetes", "kubernetes", fullCAFilename, fullCAKeyFilename); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": fullCAFilename}).Info("generated")
		log.WithFields(log.Fields{"filename": fullCAKeyFilename}).Info("generated")
	}

	return nil
}

func (generator *Generator) generateAdminFiles(ca *pki.CertificateAndPrivateKey) error {
	for nodeName, node := range generator.config.Config.Nodes {
		if !node.IsController() {
			continue
		}

		generator.config.SetNode(nodeName, node)

		fullAdminCertificateFilename := generator.config.GetFullDeploymentFilename(utils.ADMIN_PEM)
		fullAdminKeyFilename := generator.config.GetFullDeploymentFilename(utils.ADMIN_KEY_PEM)
		fullAdminKubeConfigFilename := generator.config.GetFullDeploymentFilename(utils.ADMIN_KUBECONFIG)

		if utils.FileExists(fullAdminCertificateFilename) {
			log.WithFields(log.Fields{"filename": fullAdminCertificateFilename}).Info("skipping")
			log.WithFields(log.Fields{"filename": fullAdminKeyFilename}).Info("skipping")

		} else {
			if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "admin", "system:masters", []string{}, []string{}, fullAdminCertificateFilename, fullAdminKeyFilename); error != nil {
				log.Println(error)

				os.Exit(-1)
			}

			log.WithFields(log.Fields{"filename": fullAdminCertificateFilename}).Info("generated")
			log.WithFields(log.Fields{"filename": fullAdminKeyFilename}).Info("generated")
		}

		if utils.FileExists(fullAdminKubeConfigFilename) {
			log.WithFields(log.Fields{"filename": fullAdminKubeConfigFilename}).Info("skipping")

		} else {
			apiServer := fmt.Sprintf("%s:%d", node.IP, generator.config.Config.APIServerPort)

			if error := config.GenerateConfigKubeConfig(fullAdminKubeConfigFilename, ca.CertificateFilename, "admin", apiServer, fullAdminCertificateFilename, fullAdminKeyFilename); error != nil {
				log.Println(error)

				os.Exit(-1)
			}

			log.WithFields(log.Fields{"filename": fullAdminKubeConfigFilename}).Info("generated")
		}
	}

	return nil
}

func (generator *Generator) generateProxyFiles(ca *pki.CertificateAndPrivateKey) error {
	fullProxyCertificateFilename := generator.config.GetFullDeploymentFilename(utils.PROXY_PEM)
	fullProxyPrivateKeyFilename := generator.config.GetFullDeploymentFilename(utils.PROXY_KEY_PEM)
	fullProxyKubeconfigFilename := generator.config.GetFullDeploymentFilename(utils.PROXY_KUBECONFIG)

	if utils.FileExists(fullProxyCertificateFilename) {
		log.WithFields(log.Fields{"filename": fullProxyCertificateFilename}).Info("skipped")
		log.WithFields(log.Fields{"filename": fullProxyPrivateKeyFilename}).Info("skipped")

	} else {
		if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "system:kube-proxy", "system:node-proxier", []string{}, []string{}, fullProxyCertificateFilename, fullProxyPrivateKeyFilename); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": fullProxyCertificateFilename}).Info("generated")
		log.WithFields(log.Fields{"filename": fullProxyPrivateKeyFilename}).Info("generated")
	}

	if utils.FileExists(fullProxyKubeconfigFilename) {
		log.WithFields(log.Fields{"filename": fullProxyKubeconfigFilename}).Info("skipped")

	} else {
		if error := config.GenerateConfigKubeConfig(fullProxyKubeconfigFilename, ca.CertificateFilename, "kube-proxy", generator.config.GetForwarderAddress(), fullProxyCertificateFilename, fullProxyPrivateKeyFilename); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": fullProxyKubeconfigFilename}).Info("generated")
	}

	return nil
}

func (generator *Generator) generateNodeFiles(ca *pki.CertificateAndPrivateKey) error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		fullCertificateFilename := generator.config.GetFullDeploymentFilename(utils.KUBELET_PEM)
		fullPrivateKeyFilename := generator.config.GetFullDeploymentFilename(utils.KUBELET_KEY_PEM)
		fullKubeConfigFilename := generator.config.GetFullDeploymentFilename(utils.KUBELET_KUBECONFIG)

		if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, fmt.Sprintf("system:node:%s", nodeName), "system:nodes", []string{nodeName}, []string{node.IP}, fullCertificateFilename, fullPrivateKeyFilename); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": fullCertificateFilename}).Info("generated")
		log.WithFields(log.Fields{"filename": fullPrivateKeyFilename}).Info("generated")

		if error := config.GenerateConfigKubeConfig(fullKubeConfigFilename, ca.CertificateFilename, fmt.Sprintf("system:node:%s", nodeName), generator.config.GetForwarderAddress(), fullCertificateFilename, fullPrivateKeyFilename); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": fullKubeConfigFilename}).Info("generated")
	}

	return nil
}

func (generator *Generator) generateKubernetesFiles(ca *pki.CertificateAndPrivateKey) error {
	kubernetesDNSNames := []string{"kubenetes.default"}
	kubernetesIPAddresses := []string{"127.0.0.1", "10.32.0.1"}

	for nodeName, node := range generator.config.Config.Nodes {
		kubernetesDNSNames = append(kubernetesDNSNames, nodeName)
		kubernetesIPAddresses = append(kubernetesIPAddresses, node.IP)
	}

	fullKubernetesCertificateFilename := generator.config.GetFullDeploymentFilename(utils.KUBERNETES_PEM)
	fullKubernetesPrivateKeyFilename := generator.config.GetFullDeploymentFilename(utils.KUBERNETES_KEY_PEM)

	if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "kubernetes", "Kubernetes", kubernetesDNSNames, kubernetesIPAddresses, fullKubernetesCertificateFilename, fullKubernetesPrivateKeyFilename); error != nil {
		return error
	}

	log.WithFields(log.Fields{"filename": fullKubernetesCertificateFilename}).Info("generated")
	log.WithFields(log.Fields{"filename": fullKubernetesPrivateKeyFilename}).Info("generated")

	return nil
}

func (generator *Generator) GenerateFiles() error {
	// Generate systemd file
	if error := generator.generateServiceFile(); error != nil {
		return error
	}

	// Generate kubelet configuration
	if error := generator.generateK8SKubeletConfigFile(); error != nil {
		return error
	}

	// Generate dashboard admin user configuration
	if error := generator.generateK8SAdminUserConfigFile(); error != nil {
		return error
	}

	// Generate container network interface files
	if error := generator.generateCNIFiles(); error != nil {
		return error
	}

	fullCAFilename := generator.config.GetFullDeploymentFilename(utils.CA_PEM)
	fullCAKeyFilename := generator.config.GetFullDeploymentFilename(utils.CA_KEY_PEM)

	// Generate CA if not done already
	if error := generator.generateCAFiles(fullCAFilename, fullCAKeyFilename); error != nil {
		return error
	}

	// Load ca certificate and private key
	ca, error := pki.LoadCertificateAndPrivateKey(fullCAFilename, fullCAKeyFilename)
	if error != nil {
		return error
	}

	// Generate admin certificate and kubconfigs
	if error := generator.generateAdminFiles(ca); error != nil {
		return error
	}

	// Generate proxy certificates and kubeconfig
	if error := generator.generateProxyFiles(ca); error != nil {
		return error
	}

	// Generate kubernetes security file
	if error := generator.generateEncryptionFile(); error != nil {
		return error
	}

	// Generate kubernetes certificate
	if error := generator.generateKubernetesFiles(ca); error != nil {
		return error
	}

	// Generate certificates and kubconfigs for nodes
	if error := generator.generateNodeFiles(ca); error != nil {
		return error
	}

	return nil
}
