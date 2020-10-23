package ceph

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"

	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/utils"
)

const CephBinariesPath = "/usr/bin"
const CephConfigPath = "/etc/ceph"
const CephDataPath = "/var/lib/ceph"

type CephData struct {
	CephClusterName                    string
	CephPoolName                       string
	MonitorKey                         string
	ClientAdminKey                     string
	ClientBootstrapMetadataServerKey   string
	ClientBootstrapObjectStorageKey    string
	ClientBootstrapRadosBlockDeviceKey string
	ClientBootstrapRadosGatewayKey     string
	ClientK8STEWKey                    string
}

type ProxyTransport struct {
	http.RoundTripper
}

func (transport *ProxyTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	var buffer bytes.Buffer

	redirectRequest := *request

	if request.Body != nil {
		buffer.ReadFrom(request.Body)

		request.Body = ioutil.NopCloser(&buffer)

		redirectRequest.Body = ioutil.NopCloser(bytes.NewReader(buffer.Bytes()))
	}

	dumpResponse := func(_request *http.Request, _response *http.Response) {
		body, _ := httputil.DumpRequest(_request, false)

		fmt.Println("==========================\nRequest", "[*]", _request.URL, "\n", string(body))

		body, _ = httputil.DumpResponse(_response, false)

		fmt.Println("--------------------------\nResponse", "[*]", _request.URL, "\n", string(body))
	}

	response, _error := transport.RoundTripper.RoundTrip(request)
	if _error != nil {
		return nil, errors.Wrapf(_error, "Round trip to '%s' failed", request.URL)
	}

	dumpResponse(request, response)

	if response.StatusCode == http.StatusSeeOther {
		location, _error := response.Location()
		if _error != nil {
			return nil, errors.Wrapf(_error, "Could not location from '%s'", request.URL)
		}

		redirectRequest.URL.Host = location.Host

		response, _error = transport.RoundTripper.RoundTrip(&redirectRequest)
		if _error != nil {
			return nil, errors.Wrapf(_error, "Redirect to '%s' failed", redirectRequest.URL)
		}

		dumpResponse(&redirectRequest, response)
	}

	return response, nil
}

type Proxy struct {
	publicAddress string
	proxy         *httputil.ReverseProxy
}

func NewProxy(scheme, publicAddress, port string) *Proxy {
	proxyAddress := &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%s", publicAddress, port),
	}

	proxy := &Proxy{
		publicAddress: publicAddress,
		proxy:         httputil.NewSingleHostReverseProxy(proxyAddress),
	}

	proxy.proxy.Transport = &ProxyTransport{http.DefaultTransport}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	return proxy
}

func RunProxy(proxyPort, sslCertificate, sslKey, scheme, publicAddress, targetPort string) {
	address := fmt.Sprintf(":%s", proxyPort)

	log.Printf("Starting proxy at '%s'", address)

	http.ListenAndServeTLS(address, sslCertificate, sslKey, NewProxy(scheme, publicAddress, targetPort))
}

func (proxy *Proxy) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	proxy.proxy.ServeHTTP(response, request)
}

type Ceph struct {
	config                             *config.InternalConfig
	configPath, binariesPath, dataPath string
}

func NewCeph(config *config.InternalConfig, binariesPath, configPath, dataPath string) *Ceph {
	return &Ceph{config: config, binariesPath: binariesPath, configPath: configPath, dataPath: dataPath}
}

func (ceph *Ceph) Setup() (*CephData, error) {
	cephData := &CephData{}
	cephData.CephClusterName = ceph.config.Config.CephClusterName
	cephData.CephPoolName = utils.CephRbdPoolName

	cephMonitoringKeyringFilename := ceph.config.GetFullLocalAssetFilename(utils.CephMonitorKeyring)

	// Reload keys if already there
	if utils.FileExists(cephMonitoringKeyringFilename) {
		cfg, _error := ini.Load(cephMonitoringKeyringFilename)
		if _error != nil {
			return nil, fmt.Errorf("Could not load Ceph Credentials from '%s' (%s)", cephMonitoringKeyringFilename, _error.Error())
		}

		cephData.MonitorKey = cfg.Section("mon.").Key("key").String()
		cephData.ClientAdminKey = cfg.Section("client.admin").Key("key").String()
		cephData.ClientBootstrapMetadataServerKey = cfg.Section("client.bootstrap-mds").Key("key").String()
		cephData.ClientBootstrapObjectStorageKey = cfg.Section("client.bootstrap-osd").Key("key").String()
		cephData.ClientBootstrapRadosBlockDeviceKey = cfg.Section("client.bootstrap-rbd").Key("key").String()
		cephData.ClientBootstrapRadosGatewayKey = cfg.Section("client.bootstrap-rgw").Key("key").String()
		cephData.ClientK8STEWKey = cfg.Section("client.k8s-tew").Key("key").String()

	} else {
		// Generate new keys
		cephData.MonitorKey = utils.GenerateCephKey()
		cephData.ClientAdminKey = utils.GenerateCephKey()
		cephData.ClientBootstrapMetadataServerKey = utils.GenerateCephKey()
		cephData.ClientBootstrapObjectStorageKey = utils.GenerateCephKey()
		cephData.ClientBootstrapRadosBlockDeviceKey = utils.GenerateCephKey()
		cephData.ClientBootstrapRadosGatewayKey = utils.GenerateCephKey()
		cephData.ClientK8STEWKey = utils.GenerateCephKey()
	}

	if error := utils.ApplyTemplateAndSave("ceph-monitor-keyring", utils.TemplateCephMonitorKeyring, cephData, cephMonitoringKeyringFilename, true, false); error != nil {
		return nil, error
	}

	if error := utils.ApplyTemplateAndSave("ceph-client-admin", utils.TemplateCephClientAdminKeyring, struct {
		Key string
	}{
		Key: cephData.ClientAdminKey,
	}, ceph.config.GetFullLocalAssetFilename(utils.CephClientAdminKeyring), true, false); error != nil {
		return nil, error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-mds-client-keyring", utils.TemplateCephClientKeyring, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-mds",
		Key:  cephData.ClientBootstrapMetadataServerKey,
	}, ceph.config.GetFullLocalAssetFilename(utils.CephBootstrapMdsKeyring), true, false); error != nil {
		return nil, error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-osd-client-keyring", utils.TemplateCephClientKeyring, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-osd",
		Key:  cephData.ClientBootstrapObjectStorageKey,
	}, ceph.config.GetFullLocalAssetFilename(utils.CephBootstrapOsdKeyring), true, false); error != nil {
		return nil, error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-rbd-client-keyring", utils.TemplateCephClientKeyring, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-rbd",
		Key:  cephData.ClientBootstrapRadosBlockDeviceKey,
	}, ceph.config.GetFullLocalAssetFilename(utils.CephBootstrapRbdKeyring), true, false); error != nil {
		return nil, error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-rgw-client-keyring", utils.TemplateCephClientKeyring, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-rgw",
		Key:  cephData.ClientBootstrapRadosGatewayKey,
	}, ceph.config.GetFullLocalAssetFilename(utils.CephBootstrapRgwKeyring), true, false); error != nil {
		return nil, error
	}

	monDataTemplate := ceph.getMonDirectory("$id")
	monKeyringTemplate := ceph.getKeyring(monDataTemplate)
	osdDataTemplate := ceph.getOsdDirectory("$id")
	osdKeyringTemplate := ceph.getKeyring(osdDataTemplate)
	osdJournalTemplate := ceph.getJournal(osdDataTemplate)

	if error := utils.ApplyTemplateAndSave("ceph-config", utils.TemplateCephConfig, struct {
		ClusterID          string
		ClusterName        string
		PublicNetwork      string
		ClusterNetwork     string
		DataDirectory      string
		StorageControllers []config.NodeData
		StorageNodes       []config.NodeData
		MonKeyringTemplate string
		MonDataTemplate    string
		OsdKeyringTemplate string
		OsdDataTemplate    string
		OsdJournalTemplate string
	}{
		ClusterID:          ceph.config.Config.ClusterID,
		ClusterName:        ceph.config.Config.CephClusterName,
		PublicNetwork:      ceph.config.Config.PublicNetwork,
		ClusterNetwork:     ceph.config.Config.PublicNetwork,
		DataDirectory:      ceph.dataPath,
		StorageControllers: ceph.config.GetStorageControllers(),
		StorageNodes:       ceph.config.GetStorageNodes(),
		MonKeyringTemplate: monKeyringTemplate,
		MonDataTemplate:    monDataTemplate,
		OsdKeyringTemplate: osdKeyringTemplate,
		OsdDataTemplate:    osdDataTemplate,
		OsdJournalTemplate: osdJournalTemplate,
	}, ceph.config.GetFullLocalAssetFilename(utils.CephConfig), true, false); error != nil {
		return nil, error
	}

	return cephData, nil
}

func (ceph *Ceph) RunMgr(id, publicAddress, sslCertificate, sslKey, proxyPort string) error {
	cephBinary := ceph.getCephBinary()
	cephMgrBinary := ceph.getPublicAddressBinary(ceph.getCephMgrBinary(), publicAddress)
	directory := ceph.getMgrDirectory(id)
	keyring := ceph.getKeyring(directory)

	// Create directory
	if _error := ceph.createDirectory(directory); _error != nil {
		return _error
	}

	log.WithFields(log.Fields{"keyring": keyring}).Info("Generating keyring")

	// Create or update keyring
	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s auth get-or-create mgr.%s mon 'allow profile mgr' osd 'allow *' mds 'allow *' -o %s", cephBinary, id, keyring)); _error != nil {
		return _error
	}

	// Run proxy
	go RunProxy(proxyPort, sslCertificate, sslKey, "https", publicAddress, "8443")

	log.WithFields(log.Fields{"keyring": keyring, "id": id}).Info("Starting mgr")

	// Start mgr
	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s -f -i %s", cephMgrBinary, id)); _error != nil {
		return _error
	}

	return nil
}

func (ceph *Ceph) RunMon(id, publicAddress string) error {
	cephMonBinary := ceph.getPublicAddressBinary(ceph.getCephMonBinary(), publicAddress)
	directory := ceph.getMonDirectory(id)
	keyring := ceph.getKeyring(directory)
	bootstrapKeyring := ceph.getCephMonitoringKeyring()
	monCommand := fmt.Sprintf("%s -i %s", cephMonBinary, id)

	if !utils.FileExists(keyring) {
		if _error := ceph.createDirectory(directory); _error != nil {
			return _error
		}

		log.WithFields(log.Fields{"directory": directory}).Info("Generating mon")

		if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --mkfs --keyring %s", monCommand, bootstrapKeyring)); _error != nil {
			return _error
		}
	}

	log.WithFields(log.Fields{"id": id}).Info("Starting mon")

	// Start mon
	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s -f", monCommand)); _error != nil {
		return _error
	}

	return nil
}

func (ceph *Ceph) RunMds(id, publicAddress string) error {
	cephBinary := ceph.getCephBinary()
	cephMdsBinary := ceph.getPublicAddressBinary(ceph.getCephMdsBinary(), publicAddress)
	directory := ceph.getMdsDirectory(id)
	keyring := ceph.getKeyring(directory)

	// Create directory
	if _error := ceph.createDirectory(directory); _error != nil {
		return _error
	}

	log.WithFields(log.Fields{"keyring": keyring}).Info("Generating keyring")

	// Create or update keyring
	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s auth get-or-create mds.%s osd 'allow rwx' mds 'allow' mon 'allow profile mds' -o %s", cephBinary, id, keyring)); _error != nil {
		return _error
	}

	log.WithFields(log.Fields{"keyring": keyring, "id": id}).Info("Starting mds")

	// Start mds
	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s -f -i %s", cephMdsBinary, id)); _error != nil {
		return _error
	}

	return nil
}

func (ceph *Ceph) RunOsd(id, publicAddress string) error {
	cephBinary := ceph.getCephBinary()
	cephOsdBinary := ceph.getPublicAddressBinary(ceph.getCephOsdBinary(), publicAddress)
	cephAuthtoolBinary := ceph.getCephAuthtoolBinary()
	directory := ceph.getOsdDirectory(id)
	keyring := ceph.getKeyring(directory)
	bootstrapKeyring := ceph.getCephMonitoringKeyring()
	osdCommand := fmt.Sprintf("%s -i %s", cephOsdBinary, id)

	if !utils.FileExists(keyring) {
		if _error := ceph.createDirectory(directory); _error != nil {
			return _error
		}

		key := utils.GenerateCephKey()
		uniqueID := uuid.NewV4().String()

		log.WithFields(log.Fields{"keyring": keyring}).Info("Generating keyring")

		file, _error := ioutil.TempFile("/tmp", "osd-cephx-secret")
		if _error != nil {
			return errors.Wrap(_error, "Could not create temporary file to store osd cephx secret")
		}
		defer os.Remove(file.Name())

		secret := fmt.Sprintf("{\"cephx_secret\": \"%s\"}", key)

		if _error := ioutil.WriteFile(file.Name(), []byte(secret), 0666); _error != nil {
			return errors.Wrapf(_error, "Could not write to file '%s'", file.Name())
		}

		if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s osd new %s %s -i %s -n client.bootstrap-osd -k %s", cephBinary, uniqueID, id, file.Name(), bootstrapKeyring)); _error != nil {
			return _error
		}

		if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --create-keyring %s --name osd.%s --add-key %s", cephAuthtoolBinary, keyring, id, key)); _error != nil {
			return _error
		}

		if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --mkfs --mkjournal --osd-uuid %s", osdCommand, uniqueID)); _error != nil {
			return _error
		}
	}

	log.WithFields(log.Fields{"keyring": keyring, "id": id}).Info("Starting osd")

	// Start osd
	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s -f", osdCommand)); _error != nil {
		return _error
	}

	return nil
}

func (ceph *Ceph) RunRgw(id, publicAddress, sslCertificate, sslKey, proxyPort string) error {
	cephBinary := ceph.getCephBinary()
	cephRgwBinary := ceph.getPublicAddressBinary(ceph.getCephRgwBinary(), publicAddress)
	directory := ceph.getRgwDirectory(id)
	keyring := ceph.getKeyring(directory)
	bootstrapKeyring := ceph.getCephMonitoringKeyring()

	// Create directory
	if _error := ceph.createDirectory(directory); _error != nil {
		return _error
	}

	log.WithFields(log.Fields{"keyring": keyring}).Info("Generating keyring")

	// Create or update keyring
	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --name client.bootstrap-rgw --keyring %s auth get-or-create client.rgw.%s osd 'allow rwx' mon 'allow rw' -o %s", cephBinary, bootstrapKeyring, id, keyring)); _error != nil {
		return _error
	}

	// Run proxy
	go RunProxy(proxyPort, sslCertificate, sslKey, "http", publicAddress, "7480")

	log.WithFields(log.Fields{"id": id}).Info("Starting rgw")

	// Start rgw
	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s -n client.rgw.%s -k %s -f", cephRgwBinary, id, keyring)); _error != nil {
		return _error
	}

	return nil
}

func (ceph *Ceph) RunSetup(dashboardUsername, dashboardPassword, radosgwUsername, radosgwPassword, sslCertificate, sslKey string) error {
	cephBinary := ceph.getCephBinary()
	radosgwAdminBinary := ceph.getRadosgwAdminBinary()

	commands := []string{
		fmt.Sprintf("%s mon enable-msgr2", cephBinary),
		fmt.Sprintf("%s mgr module enable dashboard", cephBinary),
		fmt.Sprintf("%s dashboard feature disable iscsi", cephBinary),
		fmt.Sprintf("%s dashboard feature disable mirroring", cephBinary),
		fmt.Sprintf("%s dashboard feature disable nfs", cephBinary),
		fmt.Sprintf("%s dashboard set-ssl-certificate -i %s", cephBinary, sslCertificate),
		fmt.Sprintf("%s dashboard set-ssl-certificate-key -i %s", cephBinary, sslKey),
		fmt.Sprintf("%s config set mgr mgr/dashboard/ssl true", cephBinary),
		fmt.Sprintf("%s dashboard ac-user-create %s %s administrator", cephBinary, dashboardUsername, dashboardPassword),
		fmt.Sprintf("%s user create --uid=%s --display-name=%s --system --access-key=%s --secret-key=%s", radosgwAdminBinary, utils.Username, utils.Username, radosgwUsername, radosgwPassword),
		fmt.Sprintf("%s dashboard set-rgw-api-access-key %s", cephBinary, radosgwUsername),
		fmt.Sprintf("%s dashboard set-rgw-api-secret-key %s", cephBinary, radosgwPassword),
		fmt.Sprintf("%s mgr module disable dashboard", cephBinary),
		fmt.Sprintf("%s mgr module enable dashboard", cephBinary),
		fmt.Sprintf("%s osd pool create %s 256 256", cephBinary, utils.CephRbdPoolName),
		fmt.Sprintf("%s osd pool application enable %s rbd", cephBinary, utils.CephRbdPoolName),
		fmt.Sprintf("%s osd pool create %s 8", cephBinary, utils.CephFsPoolName),
		fmt.Sprintf("%s osd pool create %s_metadata 8", cephBinary, utils.CephFsPoolName),
		fmt.Sprintf("%s fs new cephfs %s_metadata %s", cephBinary, utils.CephFsPoolName, utils.CephFsPoolName),
	}

	for _, command := range commands {
		if _error := utils.RunCommandWithConsoleOutput(command); _error != nil {
			return errors.Wrapf(_error, "Could not execute command '%s'", command)
		}
	}

	return nil
}

func (ceph *Ceph) getPublicAddressBinary(binary, publicAddress string) string {
	if len(publicAddress) > 0 {
		binary = fmt.Sprintf("%s --public-addr %s", binary, publicAddress)
	}

	return binary
}

func (ceph *Ceph) getClusterBinary(binary string) string {
	return fmt.Sprintf("%s --cluster %s", binary, ceph.config.Config.CephClusterName)
}

func (ceph *Ceph) getConfigBinary(binary string) string {
	return fmt.Sprintf("%s --conf %s", binary, ceph.getCephConfig())
}

func (ceph *Ceph) getUserGroupBinary(binary string) string {
	return fmt.Sprintf("%s --setuser ceph --setgroup ceph", binary)
}

func (ceph *Ceph) getClientAdminBinary(binary string) string {
	return fmt.Sprintf("%s --keyring %s -n client.admin", binary, ceph.getCephClientAdminKeyring())
}

func (ceph *Ceph) getRadosgwAdminBinary() string {
	return ceph.getBinary("radosgw-admin")
}

func (ceph *Ceph) getCephBinary() string {
	return ceph.getClientAdminBinary(ceph.getUserGroupBinary(ceph.getClusterBinary(ceph.getConfigBinary(ceph.getBinary("ceph")))))
}

func (ceph *Ceph) getCephMgrBinary() string {
	return ceph.getUserGroupBinary(ceph.getClusterBinary(ceph.getConfigBinary(ceph.getBinary("ceph-mgr"))))
}

func (ceph *Ceph) getCephMonBinary() string {
	return ceph.getUserGroupBinary(ceph.getClusterBinary(ceph.getConfigBinary(ceph.getBinary("ceph-mon"))))
}

func (ceph *Ceph) getCephMdsBinary() string {
	return ceph.getUserGroupBinary(ceph.getClusterBinary(ceph.getConfigBinary(ceph.getBinary("ceph-mds"))))
}

func (ceph *Ceph) getCephOsdBinary() string {
	return ceph.getUserGroupBinary(ceph.getClusterBinary(ceph.getConfigBinary(ceph.getBinary("ceph-osd"))))
}

func (ceph *Ceph) getCephRgwBinary() string {
	return ceph.getUserGroupBinary(ceph.getClusterBinary(ceph.getConfigBinary(ceph.getBinary("radosgw"))))
}

func (ceph *Ceph) getCephAuthtoolBinary() string {
	return ceph.getUserGroupBinary(ceph.getClusterBinary(ceph.getConfigBinary(ceph.getBinary("ceph-authtool"))))
}

func (ceph *Ceph) getBinary(binary string) string {
	return path.Join(ceph.binariesPath, binary)
}

func (ceph *Ceph) getKeyring(directory string) string {
	return path.Join(directory, "keyring")
}

func (ceph *Ceph) getJournal(directory string) string {
	return path.Join(directory, "journal")
}

func (ceph *Ceph) getServiceDirectory(_type, id string) string {
	return path.Join(ceph.dataPath, _type, fmt.Sprintf("%s-%s", ceph.config.Config.CephClusterName, id))
}

func (ceph *Ceph) getMgrDirectory(id string) string {
	return ceph.getServiceDirectory("mgr", id)
}

func (ceph *Ceph) getMonDirectory(id string) string {
	return ceph.getServiceDirectory("mon", id)
}

func (ceph *Ceph) getMdsDirectory(id string) string {
	return ceph.getServiceDirectory("mds", id)
}

func (ceph *Ceph) getOsdDirectory(id string) string {
	return ceph.getServiceDirectory("osd", id)
}

func (ceph *Ceph) getRgwDirectory(id string) string {
	return ceph.getServiceDirectory("rgw", id)
}

func (ceph *Ceph) getCephMonitoringKeyring() string {
	return path.Join(ceph.configPath, utils.CephMonitorKeyring)
}

func (ceph *Ceph) getCephClientAdminKeyring() string {
	return path.Join(ceph.configPath, utils.CephClientAdminKeyring)
}

func (ceph *Ceph) getCephConfig() string {
	return path.Join(ceph.configPath, utils.CephConfig)
}

func (ceph *Ceph) createDirectory(directory string) error {
	log.WithFields(log.Fields{"directory": directory}).Info("Creating directory")

	if _error := utils.CreateDirectoryIfMissing(directory); _error != nil {
		return _error
	}

	log.WithFields(log.Fields{"directory": directory}).Info("Updating ownership")

	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("/bin/chown -R ceph:ceph %s", directory)); _error != nil {
		return _error
	}

	return nil
}
