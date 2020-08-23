package ceph

import (
	"fmt"
	"io/ioutil"
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

func (ceph *Ceph) RunMgr(id string) error {
	cephBinary := ceph.getCephBinary()
	cephMgrBinary := ceph.getCephMgrBinary()
	directory := ceph.getMgrDirectory(id)
	keyring := ceph.getKeyring(directory)

	if !utils.FileExists(keyring) {
		if _error := ceph.createDirectory(directory); _error != nil {
			return _error
		}

		log.WithFields(log.Fields{"keyring": keyring}).Info("Generating keyring")

		if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --cluster %s auth get-or-create mgr.\"%s\" mon 'allow profile mgr' osd 'allow *' mds 'allow *' -o \"%s\"", cephBinary, ceph.config.Config.CephClusterName, id, keyring)); _error != nil {
			return _error
		}

		if _error := ceph.updateKeyringRights(keyring); _error != nil {
			return _error
		}

		if _error := ceph.updateDirectoryOwnership(directory); _error != nil {
			return _error
		}
	}

	log.WithFields(log.Fields{"keyring": keyring, "id": id}).Info("Starting mgr")

	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --cluster %s --setuser ceph --setgroup ceph -f -i \"%s\"", cephMgrBinary, ceph.config.Config.CephClusterName, id)); _error != nil {
		return _error
	}

	return nil
}

func (ceph *Ceph) RunMon(id string) error {
	cephMonBinary := ceph.getCephMonBinary()
	directory := ceph.getMonDirectory(id)
	keyring := ceph.getKeyring(directory)
	bootstrapKeyring := path.Join(ceph.configPath, utils.CephMonitorKeyring)
	monCommand := fmt.Sprintf("%s --cluster %s --setuser ceph --setgroup ceph -i %s", cephMonBinary, ceph.config.Config.CephClusterName, id)

	if !utils.FileExists(keyring) {
		if _error := ceph.createDirectory(directory); _error != nil {
			return _error
		}

		log.WithFields(log.Fields{"directory": directory}).Info("Generating mon")

		if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --mkfs --keyring %s", monCommand, bootstrapKeyring)); _error != nil {
			return _error
		}

		if _error := ceph.updateDirectoryOwnership(directory); _error != nil {
			return _error
		}
	}

	log.WithFields(log.Fields{"id": id}).Info("Starting mon")

	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s -f", monCommand)); _error != nil {
		return _error
	}

	return nil
}

func (ceph *Ceph) RunMds(id string) error {
	cephBinary := ceph.getCephBinary()
	cephMdsBinary := ceph.getCephMdsBinary()
	directory := ceph.getMdsDirectory(id)
	keyring := ceph.getKeyring(directory)

	if !utils.FileExists(keyring) {
		if _error := ceph.createDirectory(directory); _error != nil {
			return _error
		}

		log.WithFields(log.Fields{"keyring": keyring}).Info("Generating keyring")

		if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --cluster %s auth get-or-create mds.\"%s\" osd 'allow rwx' mds 'allow' mon 'allow profile mds' -o \"%s\"", cephBinary, ceph.config.Config.CephClusterName, id, keyring)); _error != nil {
			return _error
		}

		if _error := ceph.updateKeyringRights(keyring); _error != nil {
			return _error
		}

		if _error := ceph.updateDirectoryOwnership(directory); _error != nil {
			return _error
		}
	}

	log.WithFields(log.Fields{"keyring": keyring, "id": id}).Info("Starting mds")

	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --cluster %s --setuser ceph --setgroup ceph -f -i \"%s\"", cephMdsBinary, ceph.config.Config.CephClusterName, id)); _error != nil {
		return _error
	}

	return nil
}

func (ceph *Ceph) RunOsd(id string) error {
	cephBinary := ceph.getCephBinary()
	cephOsdBinary := ceph.getCephOsdBinary()
	cephAuthtoolBinary := ceph.getCephAuthtoolBinary()
	directory := ceph.getOsdDirectory(id)
	keyring := ceph.getKeyring(directory)
	bootstrapKeyring := path.Join(ceph.configPath, utils.CephMonitorKeyring)
	osdCommand := fmt.Sprintf("%s --cluster %s --setuser ceph --setgroup ceph -i %s", cephOsdBinary, ceph.config.Config.CephClusterName, id)

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

		if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --cluster %s osd new %s %s -i %s -n client.bootstrap-osd -k %s", cephBinary, ceph.config.Config.CephClusterName, uniqueID, id, file.Name(), bootstrapKeyring)); _error != nil {
			return _error
		}

		if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --create-keyring %s --name osd.%s --add-key %s", cephAuthtoolBinary, keyring, id, key)); _error != nil {
			return _error
		}

		if _error := ceph.updateKeyringRights(keyring); _error != nil {
			return _error
		}

		if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --mkfs --mkjournal --osd-uuid %s", osdCommand, uniqueID)); _error != nil {
			return _error
		}

		if _error := ceph.updateDirectoryOwnership(directory); _error != nil {
			return _error
		}
	}

	log.WithFields(log.Fields{"keyring": keyring, "id": id}).Info("Starting osd")

	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s -f", osdCommand)); _error != nil {
		return _error
	}

	return nil
}

func (ceph *Ceph) RunRgw(id string) error {
	cephBinary := ceph.getCephBinary()
	cephRgwBinary := ceph.getCephRgwBinary()
	directory := ceph.getRgwDirectory(id)
	keyring := ceph.getKeyring(directory)
	bootstrapKeyring := path.Join(ceph.configPath, utils.CephMonitorKeyring)

	if !utils.FileExists(keyring) {
		if _error := ceph.createDirectory(directory); _error != nil {
			return _error
		}

		if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --cluster %s --name client.bootstrap-rgw --keyring %s auth get-or-create client.rgw.%s osd 'allow rwx' mon 'allow rw' -o %s", cephBinary, ceph.config.Config.CephClusterName, bootstrapKeyring, id, keyring)); _error != nil {
			return _error
		}

		if _error := ceph.updateKeyringRights(keyring); _error != nil {
			return _error
		}

		if _error := ceph.updateDirectoryOwnership(directory); _error != nil {
			return _error
		}
	}

	log.WithFields(log.Fields{"id": id}).Info("Starting rgw")

	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("%s --cluster %s --setuser ceph --setgroup ceph -n client.rgw.%s -k %s -f", cephRgwBinary, ceph.config.Config.CephClusterName, id, keyring)); _error != nil {
		return _error
	}

	return nil
}

func (ceph *Ceph) getCephBinary() string {
	return ceph.getBinary("ceph")
}

func (ceph *Ceph) getCephMgrBinary() string {
	return ceph.getBinary("ceph-mgr")
}

func (ceph *Ceph) getCephMonBinary() string {
	return ceph.getBinary("ceph-mon")
}

func (ceph *Ceph) getCephMdsBinary() string {
	return ceph.getBinary("ceph-mds")
}

func (ceph *Ceph) getCephOsdBinary() string {
	return ceph.getBinary("ceph-osd")
}

func (ceph *Ceph) getCephRgwBinary() string {
	return ceph.getBinary("radosgw")
}

func (ceph *Ceph) getCephAuthtoolBinary() string {
	return ceph.getBinary("ceph-authtool")
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
	directory := path.Join(ceph.dataPath, _type, fmt.Sprintf("%s-%s", ceph.config.Config.CephClusterName, id))

	return directory
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

func (ceph *Ceph) updateDirectoryOwnership(directory string) error {
	log.WithFields(log.Fields{"directory": ceph.dataPath}).Info("Updating ownership")

	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("/bin/chown -R ceph:ceph \"%s\"", directory)); _error != nil {
		return _error
	}

	return nil
}

func (ceph *Ceph) createDirectory(directory string) error {
	if _error := utils.CreateDirectoryIfMissing(directory); _error != nil {
		return _error
	}

	return ceph.updateDirectoryOwnership(directory)
}

func (ceph *Ceph) updateKeyringRights(keyring string) error {
	log.WithFields(log.Fields{"keyring": keyring}).Info("Changed rights")

	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("/bin/chmod 600 \"%s\"", keyring)); _error != nil {
		return _error
	}

	if _error := utils.RunCommandWithConsoleOutput(fmt.Sprintf("/bin/chown ceph:ceph \"%s\"", keyring)); _error != nil {
		return _error
	}

	return nil
}
