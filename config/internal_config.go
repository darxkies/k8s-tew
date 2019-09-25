package config

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/darxkies/k8s-tew/utils"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type InternalConfig struct {
	BaseDirectory string
	Name          string
	Node          *Node
	Config        *Config
}

func (config *InternalConfig) GetTemplateAssetFilename(name string) string {
	return fmt.Sprintf(`{{asset_file "%s"}}`, name)
}

func (config *InternalConfig) GetTemplateAssetDirectory(name string) string {
	return fmt.Sprintf(`{{asset_directory "%s"}}`, name)
}

func (config *InternalConfig) GetFullTargetAssetFilename(name string) string {
	return config.GetFullAssetFilename(config.Config.DeploymentDirectory, name)
}

func (config *InternalConfig) GetFullLocalAssetFilename(name string) string {
	return config.GetFullAssetFilename(config.BaseDirectory, name)
}

func (config *InternalConfig) GetRelativeAssetFilename(name string) string {
	return config.GetFullAssetFilename("", name)
}

func (config *InternalConfig) GetFullAssetFilename(baseDirectory, name string) string {
	var result *AssetFile
	var ok bool
	var directory *AssetDirectory

	if result, ok = config.Config.Assets.Files[name]; !ok {
		log.WithFields(log.Fields{"name": name}).Fatal("Missing asset file")
	}

	if directory, ok = config.Config.Assets.Directories[result.Directory]; !ok {
		log.WithFields(log.Fields{"name": name, "directory": result.Directory, "file": name}).Fatal("Missing asset directory")
	}

	filename := name

	if len(result.Filename) > 0 {
		filename = result.Filename
	}

	filename = path.Join(directory.Directory, filename)

	resultFilename, error := config.ApplyTemplate("asset-file", filename)
	if error != nil {
		log.WithFields(log.Fields{"name": name, "error": error}).Fatal("Asset file expansion")
	}

	if directory.Absolute {
		return path.Join("/", resultFilename)
	}

	return path.Join(baseDirectory, resultFilename)
}
func (config *InternalConfig) IsDeploymentDirectory(name string) bool {
	for _, file := range config.Config.Assets.Files {
		if file.Directory == name && file.Labels.HasLabels(Labels{utils.NodeController, utils.NodeWorker}) {
			return true
		}
	}

	return false
}

func (config *InternalConfig) GetFullLocalAssetDirectory(name string) string {
	return config.GetFullAssetDirectory(config.BaseDirectory, name)
}

func (config *InternalConfig) GetFullTargetAssetDirectory(name string) string {
	return config.GetFullAssetDirectory(config.Config.DeploymentDirectory, name)
}

func (config *InternalConfig) GetRelativeAssetDirectory(name string) string {
	return config.GetFullAssetDirectory("", name)
}

func (config *InternalConfig) GetFullAssetDirectory(baseDirectory, name string) string {
	var result *AssetDirectory
	var ok bool

	if result, ok = config.Config.Assets.Directories[name]; !ok {
		log.WithFields(log.Fields{"name": name, "directory": name}).Fatal("Missing asset directory")
	}

	if result.Absolute {
		return path.Join("/", result.Directory)
	}

	return path.Join(baseDirectory, result.Directory)
}

func (config *InternalConfig) SetNode(nodeName string, node *Node) {
	config.Name = nodeName
	config.Node = node
}

func NewInternalConfig(baseDirectory string) *InternalConfig {
	config := &InternalConfig{}
	config.BaseDirectory = baseDirectory

	config.Config = NewConfig()

	return config
}

func (config *InternalConfig) registerAssetDirectories() {
	// Config
	config.addAssetDirectory(utils.DirectoryConfig, Labels{}, config.getRelativeConfigDirectory(), false)
	config.addAssetDirectory(utils.DirectoryCertificates, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryConfig), utils.SubdirectoryCertificates), false)
	config.addAssetDirectory(utils.DirectoryCniConfig, Labels{utils.NodeController, utils.NodeWorker}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryConfig), utils.SubdirectoryCni), false)
	config.addAssetDirectory(utils.DirectoryCriConfig, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryConfig), utils.SubdirectoryCri), false)

	// K8S Config
	config.addAssetDirectory(utils.DirectoryK8sConfig, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryConfig), utils.SubdirectoryK8s), false)
	config.addAssetDirectory(utils.DirectoryK8sKubeConfig, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryK8sConfig), utils.SubdirectoryKubeconfig), false)
	config.addAssetDirectory(utils.DirectoryK8sSecurityConfig, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryK8sConfig), utils.SubdirectorySecurity), false)
	config.addAssetDirectory(utils.DirectoryK8sSetupConfig, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryK8sConfig), utils.SubdirectorySetup), false)
	config.addAssetDirectory(utils.DirectoryK8sManifests, Labels{utils.NodeController, utils.NodeWorker}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryK8sConfig), utils.SubdirectoryManifests), false)

	// Binaries
	config.addAssetDirectory(utils.DirectoryBinaries, Labels{}, path.Join(utils.SubdirectoryOptional, utils.SubdirectoryK8sTew, utils.SubdirectoryBinary), false)
	config.addAssetDirectory(utils.DirectoryK8sBinaries, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryBinaries), utils.SubdirectoryK8s), false)
	config.addAssetDirectory(utils.DirectoryEtcdBinaries, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryBinaries), utils.SubdirectoryEtcd), false)
	config.addAssetDirectory(utils.DirectoryCriBinaries, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryBinaries), utils.SubdirectoryCri), false)
	config.addAssetDirectory(utils.DirectoryCniBinaries, Labels{utils.NodeController, utils.NodeWorker}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryBinaries), utils.SubdirectoryCni), false)
	config.addAssetDirectory(utils.DirectoryGobetweenBinaries, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryBinaries), utils.SubdirectoryLoadBalancer), false)
	config.addAssetDirectory(utils.DirectoryVeleroBinaries, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryBinaries), utils.SubdirectoryVelero), false)
	config.addAssetDirectory(utils.DirectoryHostBinaries, Labels{utils.NodeController, utils.NodeWorker}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryBinaries), utils.SubdirectoryHost), false)

	// Misc
	config.addAssetDirectory(utils.DirectoryGobetweenConfig, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryConfig), utils.SubdirectoryLoadBalancer), false)
	config.addAssetDirectory(utils.DirectoryDynamicData, Labels{}, path.Join(utils.SubdirectoryVariable, utils.SubdirectoryLibrary, utils.SubdirectoryK8sTew), false)
	config.addAssetDirectory(utils.DirectoryEtcdData, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryDynamicData), utils.SubdirectoryEtcd), false)
	config.addAssetDirectory(utils.DirectoryContainerdData, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryDynamicData), utils.SubdirectoryContainerd), false)
	config.addAssetDirectory(utils.DirectoryKubeletData, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryDynamicData), utils.SubdirectoryKubelet), true)
	config.addAssetDirectory(utils.DirectoryPodsData, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryKubeletData), utils.SubdirectoryPods), false)
	config.addAssetDirectory(utils.DirectoryLogging, Labels{}, path.Join(utils.SubdirectoryVariable, utils.SubdirectoryLogging, utils.SubdirectoryK8sTew), false)
	config.addAssetDirectory(utils.DirectoryService, Labels{}, path.Join(utils.SubdirectoryConfig, utils.SubdirectorySystemd, utils.SubdirectorySystem), false)
	config.addAssetDirectory(utils.DirectoryContainerdState, Labels{}, path.Join(utils.SubdirectoryVariable, utils.SubdirectoryRun, utils.SubdirectoryK8sTew, utils.SubdirectoryContainerd), false)
	config.addAssetDirectory(utils.DirectoryAbsoluteContainerdState, Labels{}, path.Join(utils.SubdirectoryRun, utils.SubdirectoryContainerd), true)
	config.addAssetDirectory(utils.DirectoryProfile, Labels{}, path.Join(utils.SubdirectoryConfig, utils.SubdirectoryProfileD), false)
	config.addAssetDirectory(utils.DirectoryHelmData, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryDynamicData), utils.SubdirectoryHelm), false)
	config.addAssetDirectory(utils.DirectoryTemporary, Labels{}, path.Join(utils.SubdirectoryTemporary), false)
	config.addAssetDirectory(utils.DirectoryBashCompletion, Labels{}, path.Join(utils.SubdirectoryConfig, utils.SubdirectoryBashCompletion), false)
	config.addAssetDirectory(utils.DirectoryKubeletPlugins, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryKubeletData), utils.SubdirectoryPlugins), true)
	config.addAssetDirectory(utils.DirectoryKubeletPluginsRegistry, Labels{}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryKubeletData), utils.SubdirectoryPluginsRegistry), false)
	config.addAssetDirectory(utils.DirectoryImages, Labels{utils.NodeController, utils.NodeWorker}, path.Join(utils.SubdirectoryVariable, utils.SubdirectoryK8sTew, utils.SubdirectoryImages), false)
	config.addAssetDirectory(utils.DirectoryRun, Labels{utils.NodeController, utils.NodeWorker}, path.Join(utils.SubdirectoryRun, utils.SubdirectoryK8sTew), false)
	config.addAssetDirectory(utils.DirectoryVarRun, Labels{utils.NodeController, utils.NodeWorker}, path.Join(utils.SubdirectoryVariable, utils.SubdirectoryRun, utils.SubdirectoryK8sTew), false)

	// Ceph
	config.addAssetDirectory(utils.DirectoryCephConfig, Labels{utils.NodeWorker}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryConfig), utils.SubdirectoryCeph), false)
	config.addAssetDirectory(utils.DirectoryCephData, Labels{utils.NodeWorker}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryDynamicData), utils.SubdirectoryCeph), false)
	config.addAssetDirectory(utils.DirectoryCephBootstrapMds, Labels{utils.NodeWorker}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryCephData), utils.DirectoryCephBootstrapMds), false)
	config.addAssetDirectory(utils.DirectoryCephBootstrapOsd, Labels{utils.NodeWorker}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryCephData), utils.DirectoryCephBootstrapOsd), false)
	config.addAssetDirectory(utils.DirectoryCephBootstrapRbd, Labels{utils.NodeWorker}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryCephData), utils.DirectoryCephBootstrapRbd), false)
	config.addAssetDirectory(utils.DirectoryCephBootstrapRgw, Labels{utils.NodeWorker}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryCephData), utils.DirectoryCephBootstrapRgw), false)
	config.addAssetDirectory(utils.DirectoryCephBootstrapRgw, Labels{utils.NodeWorker}, path.Join(config.GetRelativeAssetDirectory(utils.DirectoryCephData), utils.DirectoryCephBootstrapRgw), false)
}

func (config *InternalConfig) registerAssetFiles() {
	// Config
	config.addAssetFile(utils.ConfigFilename, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryConfig)

	// Binaries
	config.addAssetFile(utils.BinaryK8sTew, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryBinaries)

	// ContainerD Binaries
	config.addAssetFile(utils.BinaryContainerd, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCriBinaries)
	config.addAssetFile(utils.BinaryContainerdShim, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCriBinaries)
	config.addAssetFile(utils.BinaryCtr, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCriBinaries)
	config.addAssetFile(utils.BinaryRunc, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCriBinaries)
	config.addAssetFile(utils.BinaryCrictl, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCriBinaries)

	// Etcd Binaries
	config.addAssetFile(utils.BinaryEtcdctl, Labels{utils.NodeController}, "", utils.DirectoryEtcdBinaries)

	// K8S Binaries
	config.addAssetFile(utils.BinaryKubectl, Labels{utils.NodeController}, "", utils.DirectoryK8sBinaries)
	config.addAssetFile(utils.BinaryKubelet, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryK8sBinaries)

	// Helm Binary
	config.addAssetFile(utils.BinaryHelm, Labels{}, "", utils.DirectoryK8sBinaries)

	// Velero Binaries
	config.addAssetFile(utils.BinaryVelero, Labels{}, "", utils.DirectoryVeleroBinaries)

	// Certificates
	config.addAssetFile(utils.PemCa, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemCaKey, Labels{utils.NodeController}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemKubernetes, Labels{utils.NodeController}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemKubernetesKey, Labels{utils.NodeController}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemServiceAccount, Labels{utils.NodeController}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemServiceAccountKey, Labels{utils.NodeController}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemAdmin, Labels{}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemAdminKey, Labels{}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemControllerManager, Labels{}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemControllerManagerKey, Labels{}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemScheduler, Labels{}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemSchedulerKey, Labels{}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemProxy, Labels{}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemProxyKey, Labels{}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemKubelet, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemKubeletKey, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemAggregator, Labels{utils.NodeController}, "", utils.DirectoryCertificates)
	config.addAssetFile(utils.PemAggregatorKey, Labels{utils.NodeController}, "", utils.DirectoryCertificates)

	// Kubeconfig
	config.addAssetFile(utils.KubeconfigAdmin, Labels{}, "", utils.DirectoryK8sKubeConfig)
	config.addAssetFile(utils.KubeconfigControllerManager, Labels{utils.NodeController}, "", utils.DirectoryK8sKubeConfig)
	config.addAssetFile(utils.KubeconfigScheduler, Labels{utils.NodeController}, "", utils.DirectoryK8sKubeConfig)
	config.addAssetFile(utils.KubeconfigProxy, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryK8sKubeConfig)
	config.addAssetFile(utils.KubeconfigKubelet, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryK8sKubeConfig)

	// Security
	config.addAssetFile(utils.EncryptionConfig, Labels{utils.NodeController}, "", utils.DirectoryK8sSecurityConfig)

	// CRI
	config.addAssetFile(utils.ContainerdConfig, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCriConfig)
	config.addAssetFile(utils.ContainerdSock, Labels{}, "", utils.DirectoryAbsoluteContainerdState)

	// Service
	config.addAssetFile(utils.ServiceConfig, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryService)

	// K8S Setup
	config.addAssetFile(utils.K8sKubeletSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sAdminUserSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sHelmUserSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.CephSecrets, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.CephSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.CephCsi, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.LetsencryptClusterIssuer, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sCalicoSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sMetalLBSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sCorednsSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sElasticsearchOperatorSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sEfkSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sVeleroSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sHeapsterSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubernetesDashboardSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sHelmSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sCertManagerSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sNginxIngressSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sMetricsServerSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sPrometheusOperatorSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubePrometheusSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubePrometheusDatasourceSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubePrometheusKubernetesClusterStatusDashboardSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubePrometheusPodsDashboardSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubePrometheusDeploymentDashboardSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubePrometheusKubernetesControlPlaneStatusDashboardSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubePrometheusStatefulsetDashboardSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubePrometheusKubernetesCapacityPlanningDashboardSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubePrometheusKubernetesResourceRequestsDashboardSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubePrometheusKubernetesClusterHealthDashboardSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.K8sKubePrometheusNodesDashboardSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)
	config.addAssetFile(utils.WordpressSetup, Labels{}, "", utils.DirectoryK8sSetupConfig)

	// K8S Config
	config.addAssetFile(utils.K8sKubeSchedulerConfig, Labels{utils.NodeController}, "", utils.DirectoryK8sConfig)
	config.addAssetFile(utils.K8sKubeletConfig, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryK8sConfig)

	// Manifests
	config.addAssetFile(utils.ManifestControllerVirtualIP, Labels{utils.NodeController}, "", utils.DirectoryK8sManifests)
	config.addAssetFile(utils.ManifestWorkerVirtualIP, Labels{utils.NodeWorker}, "", utils.DirectoryK8sManifests)
	config.addAssetFile(utils.ManifestGobetween, Labels{utils.NodeController}, "", utils.DirectoryK8sManifests)
	config.addAssetFile(utils.ManifestEtcd, Labels{utils.NodeController}, "", utils.DirectoryK8sManifests)
	config.addAssetFile(utils.ManifestKubeApiserver, Labels{utils.NodeController}, "", utils.DirectoryK8sManifests)
	config.addAssetFile(utils.ManifestKubeControllerManager, Labels{utils.NodeController}, "", utils.DirectoryK8sManifests)
	config.addAssetFile(utils.ManifestKubeScheduler, Labels{utils.NodeController}, "", utils.DirectoryK8sManifests)
	config.addAssetFile(utils.ManifestKubeProxy, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryK8sManifests)
	// Profile
	config.addAssetFile(utils.K8sTewProfile, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryProfile)

	// Gobetween
	config.addAssetFile(utils.GobetweenConfig, Labels{utils.NodeController}, "", utils.DirectoryGobetweenConfig)

	// Ceph
	config.addAssetFile(utils.CephConfig, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCephConfig)
	config.addAssetFile(utils.CephClientAdminKeyring, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCephConfig)
	config.addAssetFile(utils.CephMonitorKeyring, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryCephConfig)
	config.addAssetFile(utils.CephBootstrapMdsKeyring, Labels{utils.NodeController, utils.NodeWorker}, utils.CephKeyring, utils.DirectoryCephBootstrapMds)
	config.addAssetFile(utils.CephBootstrapOsdKeyring, Labels{utils.NodeController, utils.NodeWorker}, utils.CephKeyring, utils.DirectoryCephBootstrapOsd)
	config.addAssetFile(utils.CephBootstrapRbdKeyring, Labels{utils.NodeController, utils.NodeWorker}, utils.CephKeyring, utils.DirectoryCephBootstrapRbd)
	config.addAssetFile(utils.CephBootstrapRgwKeyring, Labels{utils.NodeController, utils.NodeWorker}, utils.CephKeyring, utils.DirectoryCephBootstrapRgw)

	// Bash Completion
	config.addAssetFile(utils.BashCompletionK8sTew, Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryBashCompletion)
	config.addAssetFile(utils.BashCompletionKubectl, Labels{utils.NodeController}, "", utils.DirectoryBashCompletion)
	config.addAssetFile(utils.BashCompletionCrictl, Labels{utils.NodeController}, "", utils.DirectoryBashCompletion)
	config.addAssetFile(utils.BashCompletionHelm, Labels{}, "", utils.DirectoryBashCompletion)
	config.addAssetFile(utils.BashCompletionVelero, Labels{}, "", utils.DirectoryBashCompletion)

	// Images
	for _, image := range config.Config.Versions.GetImages() {
		config.addAssetFile(image.GetImageFilename(), Labels{utils.NodeController, utils.NodeWorker}, "", utils.DirectoryImages)
	}
}

func (config *InternalConfig) registerServers() {
	// Servers
	config.addServer("containerd", Labels{utils.NodeController, utils.NodeWorker}, config.GetTemplateAssetFilename(utils.BinaryContainerd), map[string]string{
		"config": config.GetTemplateAssetFilename(utils.ContainerdConfig),
	})

	config.addServer("kubelet", Labels{utils.NodeController, utils.NodeWorker}, config.GetTemplateAssetFilename(utils.BinaryKubelet), map[string]string{
		"config":                       config.GetTemplateAssetFilename(utils.K8sKubeletConfig),
		"container-runtime":            "remote",
		"container-runtime-endpoint":   "unix://" + config.GetTemplateAssetFilename(utils.ContainerdSock),
		"image-pull-progress-deadline": "2m",
		"kubeconfig":                   config.GetTemplateAssetFilename(utils.KubeconfigKubelet),
		"network-plugin":               "cni",
		"register-node":                "true",
		"root-dir":                     config.GetTemplateAssetDirectory(utils.DirectoryKubeletData),
		"v":                            "0",
	})
}

func (config *InternalConfig) registerCommands() {
	kubectlCommand := fmt.Sprintf("%s --request-timeout 30s --kubeconfig %s", config.GetFullLocalAssetFilename(utils.BinaryKubectl), config.GetFullLocalAssetFilename(utils.KubeconfigAdmin))
	//helmCommand := fmt.Sprintf("KUBECONFIG=%s HELM_HOME=%s %s", config.GetFullLocalAssetFilename(utils.KubeconfigAdmin), config.GetFullLocalAssetDirectory(utils.DirectoryHelmData), config.GetFullLocalAssetFilename(utils.BinaryHelm))

	// Dependencies
	config.addCommand("setup-ubuntu", Labels{utils.NodeController, utils.NodeWorker}, Features{}, OS{utils.OsUbuntu}, "apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y apt-transport-https bash-completion socat")
	config.addCommand("setup-centos", Labels{utils.NodeController, utils.NodeWorker}, Features{}, OS{utils.OsCentos}, "systemctl disable firewalld && systemctl stop firewalld && yum install -y socat bash-completion libseccomp && sed -i --follow-symlinks 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/sysconfig/selinux && (setenforce 0 || true)")
	config.addCommand("swapoff", Labels{utils.NodeController, utils.NodeWorker}, Features{}, OS{}, "swapoff -a")
	config.addCommand("load-overlay", Labels{utils.NodeController, utils.NodeWorker}, Features{}, OS{}, "modprobe overlay")
	config.addCommand("load-btrfs", Labels{utils.NodeController, utils.NodeWorker}, Features{}, OS{}, "modprobe btrfs")
	config.addCommand("load-br_netfilter", Labels{utils.NodeController, utils.NodeWorker}, Features{}, OS{}, "modprobe br_netfilter")
	config.addCommand("enable-br_netfilter", Labels{utils.NodeController, utils.NodeWorker}, Features{}, OS{}, "echo '1' > /proc/sys/net/bridge/bridge-nf-call-iptables")
	config.addCommand("enable-net-forwarding", Labels{utils.NodeController, utils.NodeWorker}, Features{}, OS{}, "sysctl net.ipv4.conf.all.forwarding=1")
	config.addCommand("kubelet-setup", Labels{utils.NodeBootstrapper}, Features{}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubeletSetup)))
	config.addCommand("admin-user-setup", Labels{utils.NodeBootstrapper}, Features{}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sAdminUserSetup)))
	config.addCommand("calico-setup", Labels{utils.NodeBootstrapper}, Features{}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sCalicoSetup)))
	config.addCommand("metallb-setup", Labels{utils.NodeBootstrapper}, Features{}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sMetalLBSetup)))
	config.addCommand("coredns-setup", Labels{utils.NodeBootstrapper}, Features{}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sCorednsSetup)))
	config.addCommand("helm-user-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeaturePackaging}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sHelmUserSetup)))
	config.addCommand("helm-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeaturePackaging}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sHelmSetup)))
	config.addCommand("ceph-secrets", Labels{utils.NodeBootstrapper}, Features{utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.CephSecrets)))
	config.addCommand("ceph-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.CephSetup)))
	config.addCommand("ceph-csi", Labels{utils.NodeBootstrapper}, Features{utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.CephCsi)))
	//config.addCommand("helm-init", Labels{utils.NodeBootstrapper}, Features{utils.FeaturePackaging}, OS{}, fmt.Sprintf("%s init --service-account %s --upgrade", helmCommand, utils.HelmServiceAccount))
	config.addCommand("kubernetes-dashboard-setup", Labels{utils.NodeBootstrapper}, Features{}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubernetesDashboardSetup)))
	config.addCommand("cert-manager-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureIngress}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sCertManagerSetup)))
	config.addCommand("nginx-ingress-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureIngress}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sNginxIngressSetup)))
	config.addCommand("letsencrypt-cluster-issuer-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureIngress}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.LetsencryptClusterIssuer)))
	config.addCommand("heapster-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sHeapsterSetup)))
	config.addCommand("metrics-server-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sMetricsServerSetup)))
	config.addCommand("prometheus-operator-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sPrometheusOperatorSetup)))
	config.addCommand("kube-prometheus-datasource-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubePrometheusDatasourceSetup)))
	config.addCommand("kube-prometheus-kuberntes-cluster-status-dashboard-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubePrometheusKubernetesClusterStatusDashboardSetup)))
	config.addCommand("kube-prometheus-kuberntes-cluster-health-dashboard-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubePrometheusKubernetesClusterHealthDashboardSetup)))
	config.addCommand("kube-prometheus-kuberntes-control-plane-status-dashboard-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubePrometheusKubernetesControlPlaneStatusDashboardSetup)))
	config.addCommand("kube-prometheus-kuberntes-capacity-planning-dashboard-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubePrometheusKubernetesCapacityPlanningDashboardSetup)))
	config.addCommand("kube-prometheus-kuberntes-resource-requests-dashboard-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubePrometheusKubernetesResourceRequestsDashboardSetup)))
	config.addCommand("kube-prometheus-nodes-dashboard-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubePrometheusNodesDashboardSetup)))
	config.addCommand("kube-prometheus-deployment-dashboard-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubePrometheusDeploymentDashboardSetup)))
	config.addCommand("kube-prometheus-statefulset-dashboard-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubePrometheusStatefulsetDashboardSetup)))
	config.addCommand("kube-prometheus-pods-dashboard-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubePrometheusPodsDashboardSetup)))
	config.addCommand("kube-prometheus-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureMonitoring, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sKubePrometheusSetup)))
	config.addCommand("elasticsearch-operator-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureLogging, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sElasticsearchOperatorSetup)))
	config.addCommand("efk-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureLogging, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sEfkSetup)))
	config.addCommand("patch-kibana-service", Labels{utils.NodeBootstrapper}, Features{utils.FeatureLogging, utils.FeatureStorage}, OS{}, fmt.Sprintf(`%s get svc kibana-elasticsearch-cluster -n logging --output=jsonpath={.spec..nodePort} | grep %d || %s patch service kibana-elasticsearch-cluster -n logging -p '{"spec":{"type":"NodePort","ports":[{"port":80,"nodePort":%d}]}}'`, kubectlCommand, utils.PortKibana, kubectlCommand, utils.PortKibana))
	config.addCommand("patch-cerebro-service", Labels{utils.NodeBootstrapper}, Features{utils.FeatureLogging, utils.FeatureStorage}, OS{}, fmt.Sprintf(`%s get svc cerebro-elasticsearch-cluster -n logging --output=jsonpath={.spec..nodePort} | grep %d || %s patch service cerebro-elasticsearch-cluster -n logging -p '{"spec":{"type":"NodePort","ports":[{"port":80,"nodePort":%d}]}}'`, kubectlCommand, utils.PortCerebro, kubectlCommand, utils.PortCerebro))
	config.addCommand("velero-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureBackup, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.K8sVeleroSetup)))
	config.addCommand("wordpress-setup", Labels{utils.NodeBootstrapper}, Features{utils.FeatureShowcase, utils.FeatureStorage}, OS{}, fmt.Sprintf("%s apply -f %s", kubectlCommand, config.GetFullLocalAssetFilename(utils.WordpressSetup)))
}

func (config *InternalConfig) Generate() {
	config.registerAssetDirectories()
	config.registerAssetFiles()
	config.registerCommands()
	config.registerServers()
}

func (config *InternalConfig) addServer(name string, labels []string, command string, arguments map[string]string) {
	// Do not add if already in the list
	for _, server := range config.Config.Servers {
		if server.Name == name {
			return
		}
	}

	config.Config.Servers = append(config.Config.Servers, ServerConfig{Name: name, Enabled: true, Labels: labels, Command: command, Arguments: arguments, Logger: LoggerConfig{Enabled: true, Filename: path.Join(config.GetTemplateAssetDirectory(utils.DirectoryLogging), name+".log")}})
}

func (config *InternalConfig) addCommand(name string, labels Labels, features Features, os OS, command string) {
	// Do not add if already in the list
	for _, command := range config.Config.Commands {
		if command.Name == name {
			return
		}
	}

	config.Config.Commands = append(config.Config.Commands, NewCommand(name, labels, features, os, command))
}

func (config *InternalConfig) addAssetFile(name string, labels Labels, filename, directory string) {
	config.Config.Assets.Files[name] = NewAssetFile(labels, filename, directory)
}

func (config *InternalConfig) addAssetDirectory(name string, labels Labels, directory string, absolute bool) {
	config.Config.Assets.Directories[name] = NewAssetDirectory(labels, directory, absolute)
}

func (config *InternalConfig) Dump() {
	log.WithFields(log.Fields{"base-directory": config.BaseDirectory}).Info("Config")
	log.WithFields(log.Fields{"name": config.Name}).Info("Config")

	if config.Node != nil {
		log.WithFields(log.Fields{"ip": config.Node.IP}).Info("Config")
		log.WithFields(log.Fields{"labels": config.Node.Labels}).Info("Config")
		log.WithFields(log.Fields{"index": config.Node.Index}).Info("Config")
	}

	for name, assetFile := range config.Config.Assets.Files {
		log.WithFields(log.Fields{"name": name, "directory": assetFile.Directory, "labels": assetFile.Labels}).Info("Config asset file")
	}

	for name, node := range config.Config.Nodes {
		log.WithFields(log.Fields{"name": name, "index": node.Index, "labels": node.Labels, "ip": node.IP}).Info("Config node")
	}

	for name, command := range config.Config.Commands {
		log.WithFields(log.Fields{"name": name, "command": command.Command, "labels": command.Labels}).Info("Config command")
	}

	for _, serverConfig := range config.Config.Servers {
		serverConfig.Dump()
	}
}

func (config *InternalConfig) getRelativeConfigDirectory() string {
	return path.Join(utils.SubdirectoryConfig, utils.SubdirectoryK8sTew)
}

func (config *InternalConfig) getConfigDirectory() string {
	return path.Join(config.BaseDirectory, config.getRelativeConfigDirectory())
}

func (config *InternalConfig) getConfigFilename() string {
	return path.Join(config.getConfigDirectory(), utils.ConfigFilename)
}

func (config *InternalConfig) Save() error {
	if error := utils.CreateDirectoryIfMissing(config.getConfigDirectory()); error != nil {
		return error
	}

	yamlOutput, error := yaml.Marshal(config.Config)
	if error != nil {
		return error
	}

	filename := config.getConfigFilename()

	if error := ioutil.WriteFile(filename, yamlOutput, 0644); error != nil {
		return error
	}

	log.WithFields(log.Fields{"_filename": filename}).Info("Saved config")

	return nil
}

func (config *InternalConfig) Load() error {
	var error error

	filename := config.getConfigFilename()

	// Check if config file exists
	if _, error := os.Stat(filename); os.IsNotExist(error) {
		return fmt.Errorf("config '%s' not found", filename)
	}

	yamlContent, error := ioutil.ReadFile(filename)

	if error != nil {
		return error
	}

	if error := yaml.Unmarshal(yamlContent, config.Config); error != nil {
		return error
	}

	if config.Config.Version != utils.VersionConfig {
		return fmt.Errorf("Unsupported config version '%s'", config.Config.Version)
	}

	if len(config.Name) == 0 {
		config.Name, error = os.Hostname()

		if error != nil {
			return error
		}
	}

	if config.Node == nil {
		for name, node := range config.Config.Nodes {
			if name != config.Name {
				continue
			}

			config.Node = node

			break
		}
	}

	return nil
}

func (config *InternalConfig) RemoveNode(name string) error {
	if _, ok := config.Config.Nodes[name]; !ok {
		return errors.New("node not found")
	}

	delete(config.Config.Nodes, name)

	return nil
}

func (config *InternalConfig) AddNode(name string, ip string, index uint, labels []string) (*Node, error) {
	name = strings.Trim(name, " \n")

	if len(name) == 0 {
		return nil, errors.New("empty node name")
	}

	if net.ParseIP(ip) == nil {
		return nil, errors.New("invalid or wrong ip format")
	}

	config.Config.Nodes[name] = NewNode(ip, index, labels)

	return config.Config.Nodes[name], nil
}

func (config *InternalConfig) GetETCDClientEndpoints() []string {
	result := []string{}

	for _, node := range config.Config.Nodes {
		if node.IsController() {
			result = append(result, fmt.Sprintf("https://%s:2379", node.IP))
		}
	}

	return result
}

func (config *InternalConfig) GetEtcdCluster() string {
	result := ""

	for name, node := range config.Config.Nodes {
		if !node.IsController() {
			continue
		}

		if len(result) > 0 {
			result += ","
		}

		result += fmt.Sprintf("%s=https://%s:2380", name, node.IP)
	}

	return result
}

func (config *InternalConfig) GetEtcdServers() string {
	result := ""

	for _, endpoint := range config.GetETCDClientEndpoints() {
		if len(result) > 0 {
			result += ","
		}

		result += endpoint
	}

	return result
}

func (config *InternalConfig) GetControllersCount() string {
	count := 0
	for _, node := range config.Config.Nodes {
		if node.IsController() {
			count++
		}
	}

	return fmt.Sprintf("%d", count)
}

func (config *InternalConfig) ApplyTemplate(label string, value string) (string, error) {
	var functions = template.FuncMap{
		"controllers_count": func() string {
			return config.GetControllersCount()
		},
		"etcd_servers": func() string {
			return config.GetEtcdServers()
		},
		"etcd_cluster": func() string {
			return config.GetEtcdCluster()
		},
		"asset_file": func(name string) string {
			return config.GetFullTargetAssetFilename(name)
		},
		"asset_directory": func(name string) string {
			return config.GetFullTargetAssetDirectory(name)
		},
	}

	var newValue bytes.Buffer

	argumentTemplate, error := template.New(label).Funcs(functions).Parse(value)

	if error != nil {
		return "", fmt.Errorf("Could not render template: %s (%s)", label, error.Error())
	}

	if error = argumentTemplate.Execute(&newValue, config); error != nil {
		return "", fmt.Errorf("Could not render template: %s (%s)", label, error.Error())
	}

	return newValue.String(), nil
}

func (config *InternalConfig) GetAPIServerIP() (string, error) {
	if len(config.Config.ControllerVirtualIP) > 0 {
		return config.Config.ControllerVirtualIP, nil
	}

	for _, node := range config.Config.Nodes {
		if node.IsController() {
			return node.IP, nil
		}
	}

	return "", errors.New("No API Server IP found")
}

func (config *InternalConfig) GetWorkerIP() (string, error) {
	if len(config.Config.WorkerVirtualIP) > 0 {
		return config.Config.WorkerVirtualIP, nil
	}

	for _, node := range config.Config.Nodes {
		if node.IsWorker() {
			return node.IP, nil
		}
	}

	return "", errors.New("No Worker IP found")
}

func (config *InternalConfig) GetSortedNodeKeys() []string {
	result := []string{}

	for key := range config.Config.Nodes {
		result = append(result, key)
	}

	sort.Strings(result)

	return result
}

func (config *InternalConfig) GetKubeAPIServerAddresses() []string {
	result := []string{}

	for _, node := range config.Config.Nodes {
		if node.IsController() {
			result = append(result, fmt.Sprintf("%s:%d", node.IP, config.Config.APIServerPort))
		}
	}

	return result
}

type NodeData struct {
	Index uint
	Name  string
	IP    string
}

func (config *InternalConfig) getLabeledOrAllNodes(label string) []NodeData {
	result := []NodeData{}

	// Add only labeled nodes
	for nodeName, node := range config.Config.Nodes {
		if node.Labels.HasLabels(Labels{label}) && node.Labels.HasLabels(Labels{utils.NodeStorage}) {
			result = append(result, NodeData{Index: node.Index, Name: nodeName, IP: node.IP})
		}
	}

	// If no labeld nodes found get all nodes
	if len(result) == 0 {
		for nodeName, node := range config.Config.Nodes {
			if node.Labels.HasLabels(Labels{label}) {
				result = append(result, NodeData{Index: node.Index, Name: nodeName, IP: node.IP})
			}
		}
	}

	// Sort nodes by index
	sort.Slice(result, func(i, j int) bool {
		return result[i].Index < result[j].Index
	})

	return result
}

func (config *InternalConfig) GetStorageControllers() []NodeData {
	return config.getLabeledOrAllNodes(utils.NodeController)
}

func (config *InternalConfig) GetStorageNodes() []NodeData {
	return config.getLabeledOrAllNodes(utils.NodeWorker)
}

func (config *InternalConfig) GetAllowedCommonNames() string {
	result := []string{utils.CnAggregator, utils.CnAdmin, utils.CnSystemKubeControllerManager, utils.CnSystemKubeControllerManager, utils.CnSystemKubeScheduler}

	for nodeName, node := range config.Config.Nodes {
		if node.IsWorker() {
			result = append(result, fmt.Sprintf(utils.CnSystemNodePrefix, nodeName))
		}
	}

	return strings.Join(result, ",")
}
