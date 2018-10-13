package generate

import (
	"fmt"

	"github.com/darxkies/k8s-tew/config"

	"github.com/darxkies/k8s-tew/pki"
	"github.com/darxkies/k8s-tew/utils"
	log "github.com/sirupsen/logrus"
)

type Generator struct {
	config         *config.InternalConfig
	ca             *pki.CertificateAndPrivateKey
	generatorSteps []func() error
}

func NewGenerator(config *config.InternalConfig) *Generator {
	generator := &Generator{config: config}

	generator.generatorSteps = []func() error{
		// Generate profile file
		generator.generateProfileFile,
		// Generate systemd file
		generator.generateServiceFile,
		// Generate load balancer configuration
		generator.generateGobetweenConfig,
		// Generate calico setup
		generator.generateCalicoSetup,
		// Generate scheduler config
		generator.generateKubeSchedulerConfig,
		// Generate kubelet config
		generator.generateKubeletConfig,
		// Generate kubelet configuration
		generator.generateK8SKubeletConfigFile,
		// Generate dashboard admin user configuration
		generator.generateK8SAdminUserConfigFile,
		// Generate helm user configuration
		generator.generateK8SHelmUserConfigFile,
		// Generate containerd config
		generator.generateContainerdConfig,
		// Generate kubernetes security file
		generator.generateEncryptionFile,
		// Generate kubeconfig files
		generator.generateCertificates,
		// Generate kubeconfig files
		generator.generateKubeConfigs,
		// Generate Ceph Config
		generator.generateCephConfig,
		// Generate Ceph Config
		generator.generateCephSetup,
		// Generate Ceph CSI
		generator.generateCephCSI,
		// Generate ceph files
		generator.generateCephFiles,
		// Generate Let's Encrypt Cluster Issuer
		generator.generateLetsEncryptClusterIssuer,
		// Generate CoreDNS setup file
		generator.generateCoreDNSSetup,
		// Generate ElasticSearch Operator setup file
		generator.generateElasticSearchOperatorSetup,
		// Generate ElasticSearch/Fluent-Bit/Kibana setup file
		generator.generateEFKSetup,
		// Generate ark setup file
		generator.generateARKSetup,
		// Generate heapster setup file
		generator.generateHeapsterSetup,
		// Generate kubernetes dashboard setup file
		generator.generateKubernetesDashboardSetup,
		// Generate cert-manager setup file
		generator.generateCertManagerSetup,
		// Generate nginx ingress setup file
		generator.generateNginxIngressSetup,
		// Generate metrics server setup file
		generator.generateMetricsServerSetup,
		// Generate prometheus operator setup file
		generator.generatePrometheusOperatorSetup,
		// Generate kube prometheus setup file
		generator.generateKubePrometheusSetup,
		// Generate kube prometheus dashboard setup file
		generator.generateKubePrometheusDatasourceSetup,
		// Generate kube prometheus dashboard setup file
		generator.generateKubePrometheusKubernetesClusterStatusDashboardSetup,
		// Generate kube prometheus dashboard setup file
		generator.generateKubePrometheusPodsDashboardSetup,
		// Generate kube prometheus dashboard setup file
		generator.generateKubePrometheusDeploymentDashboardSetup,
		// Generate kube prometheus dashboard setup file
		generator.generateKubePrometheusKubernetesControlPlaneStatusDashboardSetup,
		// Generate kube prometheus dashboard setup file
		generator.generateKubePrometheusStatefulsetDashboardSetup,
		// Generate kube prometheus dashboard setup file
		generator.generateKubePrometheusKubernetesCapacityPlanningDashboardSetup,
		// Generate kube prometheus dashboard setup file
		generator.generateKubePrometheusKubernetesResourceRequestsDashboardSetup,
		// Generate kube prometheus dashboard setup file
		generator.generateKubePrometheusKubernetesClusterHealthDashboardSetup,
		// Generate kube prometheus dashboard setup file
		generator.generateKubePrometheusNodesDashboardSetup,
		// Generate wordpress setup file
		generator.generateWordpressSetup,
		// Generate Bash Completion for K8S-TEW
		generator.generateBashCompletionK8STEW,
		// Generate Bash Completion for Kubectl
		generator.generateBashCompletionKubectl,
		// Generate Bash Completion for Helm
		generator.generateBashCompletionHelm,
		// Generate Bash Completion for Ark
		generator.generateBashCompletionArk,
		// Generate Bash Completion for CriCtl
		generator.generateBashCompletionCriCtl,
	}

	return generator
}

func (generator *Generator) Steps() int {
	return len(generator.generatorSteps)
}

func (generator *Generator) generateProfileFile() error {
	return utils.ApplyTemplateAndSave("profile", utils.TEMPLATE_K8S_TEW_PROFILE, struct {
		Binary        string
		BaseDirectory string
	}{
		Binary:        generator.config.GetFullTargetAssetFilename(utils.BinaryK8sTew),
		BaseDirectory: generator.config.Config.DeploymentDirectory,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_TEW_PROFILE), true, false)
}

func (generator *Generator) generateServiceFile() error {
	return utils.ApplyTemplateAndSave("service", utils.TEMPLATE_K8S_TEW_SERVICE, struct {
		ProjectTitle  string
		Command       string
		BaseDirectory string
		Binary        string
	}{
		ProjectTitle:  utils.ProjectTitle,
		Command:       generator.config.GetFullTargetAssetFilename(utils.BinaryK8sTew),
		BaseDirectory: generator.config.Config.DeploymentDirectory,
		Binary:        utils.BinaryK8sTew,
	}, generator.config.GetFullLocalAssetFilename(utils.SERVICE_CONFIG), true, false)
}

func (generator *Generator) generateGobetweenConfig() error {
	return utils.ApplyTemplateAndSave("gobetween", utils.TEMPLATE_GOBETWEEN_TOML, struct {
		LoadBalancerPort uint16
		KubeAPIServers   []string
	}{
		LoadBalancerPort: generator.config.Config.LoadBalancerPort,
		KubeAPIServers:   generator.config.GetKubeAPIServerAddresses(),
	}, generator.config.GetFullLocalAssetFilename(utils.GOBETWEEN_CONFIG), true, false)
}

func (generator *Generator) generateCalicoSetup() error {
	return utils.ApplyTemplateAndSave("calico-setup", utils.TEMPLATE_CALICO_SETUP, struct {
		CalicoTyphaIP        string
		ClusterCIDR          string
		CNIConfigDirectory   string
		CNIBinariesDirectory string
		CalicoTyphaImage     string
		CalicoNodeImage      string
		CalicoCNIImage       string
	}{
		CalicoTyphaIP:        generator.config.Config.CalicoTyphaIP,
		ClusterCIDR:          generator.config.Config.ClusterCIDR,
		CNIConfigDirectory:   generator.config.GetFullTargetAssetDirectory(utils.DirectoryCniConfig),
		CNIBinariesDirectory: generator.config.GetFullTargetAssetDirectory(utils.DirectoryCniBinaries),
		CalicoTyphaImage:     generator.config.Config.Versions.CalicoTypha,
		CalicoNodeImage:      generator.config.Config.Versions.CalicoNode,
		CalicoCNIImage:       generator.config.Config.Versions.CalicoCNI,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_CALICO_SETUP), true, false)
}

func (generator *Generator) generateK8SKubeletConfigFile() error {
	return utils.ApplyTemplateAndSave("kubelet-config", utils.TEMPLATE_KUBELET_SETUP, nil, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBELET_SETUP), true, false)
}

func (generator *Generator) generateK8SAdminUserConfigFile() error {
	return utils.ApplyTemplateAndSave("admin-user-config", utils.TEMPLATE_SERVICE_ACCOUNT, struct {
		Name      string
		Namespace string
	}{
		Name:      "admin-user",
		Namespace: "kube-system",
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_ADMIN_USER_SETUP), true, false)
}

func (generator *Generator) generateK8SHelmUserConfigFile() error {
	return utils.ApplyTemplateAndSave("helm-user-config", utils.TEMPLATE_SERVICE_ACCOUNT, struct {
		Name      string
		Namespace string
	}{
		Name:      utils.HelmServiceAccount,
		Namespace: "kube-system",
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_HELM_USER_SETUP), true, false)
}

func (generator *Generator) generateEncryptionFile() error {
	fullEncryptionConfigFilename := generator.config.GetFullLocalAssetFilename(utils.ENCRYPTION_CONFIG)

	if utils.FileExists(fullEncryptionConfigFilename) {
		utils.LogFilename("skipped", fullEncryptionConfigFilename)

		return nil
	}

	encryptionKey, error := pki.GenerateEncryptionConfig()
	if error != nil {
		return error
	}

	return utils.ApplyTemplateAndSave("encryption-config", utils.TEMPLATE_ENCRYPTION_CONFIG, struct {
		EncryptionKey string
	}{
		EncryptionKey: encryptionKey,
	}, fullEncryptionConfigFilename, false, false)
}

func (generator *Generator) generateContainerdConfig() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := utils.ApplyTemplateAndSave("containerd-config", utils.TEMPLATE_CONTAINERD_TOML, struct {
			ContainerdRootDirectory  string
			ContainerdStateDirectory string
			ContainerdSock           string
			CNIConfigDirectory       string
			CNIBinariesDirectory     string
			CRIBinariesDirectory     string
			IP                       string
			PauseImage               string
		}{
			ContainerdRootDirectory:  generator.config.GetFullTargetAssetDirectory(utils.DirectoryContainerdData),
			ContainerdStateDirectory: generator.config.GetFullTargetAssetDirectory(utils.DirectoryContainerdState),
			ContainerdSock:           generator.config.GetFullTargetAssetFilename(utils.CONTAINERD_SOCK),
			CNIConfigDirectory:       generator.config.GetFullTargetAssetDirectory(utils.DirectoryCniConfig),
			CNIBinariesDirectory:     generator.config.GetFullTargetAssetDirectory(utils.DirectoryCniBinaries),
			CRIBinariesDirectory:     generator.config.GetFullTargetAssetDirectory(utils.DirectoryCriBinaries),
			IP:                       node.IP,
			PauseImage:               generator.config.Config.Versions.Pause,
		}, generator.config.GetFullLocalAssetFilename(utils.CONTAINERD_CONFIG), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateKubeSchedulerConfig() error {
	return utils.ApplyTemplateAndSave("kube-scheduler-config", utils.TEMPLATE_KUBE_SCHEDULER_CONFIGURATION, struct {
		KubeConfig string
	}{
		KubeConfig: generator.config.GetFullTargetAssetFilename(utils.SCHEDULER_KUBECONFIG),
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_SCHEDULER_CONFIG), true, false)
}

func (generator *Generator) generateKubeletConfig() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := utils.ApplyTemplateAndSave("kubelet-configuration", utils.TEMPLATE_KUBELET_CONFIGURATION, struct {
			CA                  string
			CertificateFilename string
			KeyFilename         string
			ClusterDomain       string
			ClusterDNSIP        string
			PODCIDR             string
			StaticPodPath       string
		}{
			CA:                  generator.config.GetFullTargetAssetFilename(utils.CA_PEM),
			CertificateFilename: generator.config.GetFullTargetAssetFilename(utils.KUBELET_PEM),
			KeyFilename:         generator.config.GetFullTargetAssetFilename(utils.KUBELET_KEY_PEM),
			ClusterDomain:       generator.config.Config.ClusterDomain,
			ClusterDNSIP:        generator.config.Config.ClusterDNSIP,
			PODCIDR:             generator.config.Config.ClusterCIDR,
			StaticPodPath:       generator.config.GetFullTargetAssetDirectory(utils.DirectoryK8sManifests),
		}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBELET_CONFIG), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateCertificates() error {
	var error error

	fullCAFilename := generator.config.GetFullLocalAssetFilename(utils.CA_PEM)
	fullCAKeyFilename := generator.config.GetFullLocalAssetFilename(utils.CA_KEY_PEM)

	// Generate CA if not done already
	if error := pki.GenerateCA(generator.config.Config.RSASize, generator.config.Config.CAValidityPeriod, "Kubernetes", "Kubernetes", fullCAFilename, fullCAKeyFilename); error != nil {
		return error
	}

	// Load ca certificate and private key
	generator.ca, error = pki.LoadCertificateAndPrivateKey(fullCAFilename, fullCAKeyFilename)
	if error != nil {
		return error
	}

	// Collect dns names and ip addresses
	kubernetesDNSNames := []string{"kubernetes", "kubenetes.default", "kubenetes.default.svc", "kubenetes.default.svc.cluster.local", "localhost"}
	kubernetesIPAddresses := []string{"127.0.0.1", "10.32.0.1"}

	if len(generator.config.Config.ControllerVirtualIP) > 0 {
		kubernetesIPAddresses = append(kubernetesIPAddresses, generator.config.Config.ControllerVirtualIP)
	}

	for nodeName, node := range generator.config.Config.Nodes {
		kubernetesDNSNames = append(kubernetesDNSNames, nodeName)
		kubernetesIPAddresses = append(kubernetesIPAddresses, node.IP)
	}

	// Generate admin certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CN_ADMIN, "system:masters", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.ADMIN_PEM), generator.config.GetFullLocalAssetFilename(utils.ADMIN_KEY_PEM), false); error != nil {
		return error
	}

	// Generate kuberentes certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, "kubernetes", "Kubernetes", kubernetesDNSNames, kubernetesIPAddresses, generator.config.GetFullLocalAssetFilename(utils.KUBERNETES_PEM), generator.config.GetFullLocalAssetFilename(utils.KUBERNETES_KEY_PEM), true); error != nil {
		return error
	}

	// Generate aggregator certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CN_AGGREGATOR, "Kubernetes", kubernetesDNSNames, kubernetesIPAddresses, generator.config.GetFullLocalAssetFilename(utils.AGGREGATOR_PEM), generator.config.GetFullLocalAssetFilename(utils.AGGREGATOR_KEY_PEM), true); error != nil {
		return error
	}

	// Generate service accounts certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, "service-accounts", "Kubernetes", kubernetesDNSNames, kubernetesIPAddresses, generator.config.GetFullLocalAssetFilename(utils.SERVICE_ACCOUNT_PEM), generator.config.GetFullLocalAssetFilename(utils.SERVICE_ACCOUNT_KEY_PEM), false); error != nil {
		return error
	}

	// Generate controller manager certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CN_SYSTEM_KUBE_CONTROLLER_MANAGER, "system:node-controller-manager", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.CONTROLLER_MANAGER_PEM), generator.config.GetFullLocalAssetFilename(utils.CONTROLLER_MANAGER_KEY_PEM), false); error != nil {
		return error
	}

	// Generate scheduler certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CN_SYSTEM_KUBE_SCHEDULER, "system:kube-scheduler", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.SCHEDULER_PEM), generator.config.GetFullLocalAssetFilename(utils.SCHEDULER_KEY_PEM), false); error != nil {
		return error
	}

	// Generate proxy certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CN_SYSTEM_KUBE_PROXY, "system:node-proxier", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.PROXY_PEM), generator.config.GetFullLocalAssetFilename(utils.PROXY_KEY_PEM), false); error != nil {
		return error
	}

	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, fmt.Sprintf(utils.CN_SYSTEM_NODE_PREFIX, nodeName), "system:nodes", []string{nodeName}, []string{node.IP}, generator.config.GetFullLocalAssetFilename(utils.KUBELET_PEM), generator.config.GetFullLocalAssetFilename(utils.KUBELET_KEY_PEM), false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateConfigKubeConfig(kubeConfigFilename, caFilename, user, apiServers, certificateFilename, keyFilename string, force bool) error {
	if utils.FileExists(kubeConfigFilename) && !force {
		utils.LogFilename("skipped", kubeConfigFilename)

		return nil
	}

	base64CA, error := utils.GetBase64OfPEM(caFilename)

	if error != nil {
		return error
	}

	base64Certificate, error := utils.GetBase64OfPEM(certificateFilename)

	if error != nil {
		return error
	}

	base64Key, error := utils.GetBase64OfPEM(keyFilename)

	if error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("kubeconfig", utils.TEMPLATE_KUBECONFIG, struct {
		Name            string
		User            string
		APIServer       string
		CAData          string
		CertificateData string
		KeyData         string
	}{
		Name:            user,
		User:            user,
		APIServer:       apiServers,
		CAData:          base64CA,
		CertificateData: base64Certificate,
		KeyData:         base64Key,
	}, kubeConfigFilename, true, false); error != nil {
		return error
	}

	return nil
}

func (generator *Generator) generateKubeConfigs() error {
	apiServer, error := generator.config.GetAPIServerIP()
	if error != nil {
		return error
	}

	apiServer = fmt.Sprintf("%s:%d", apiServer, generator.config.Config.LoadBalancerPort)

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.ADMIN_KUBECONFIG), generator.ca.CertificateFilename, "admin", apiServer, generator.config.GetFullLocalAssetFilename(utils.ADMIN_PEM), generator.config.GetFullLocalAssetFilename(utils.ADMIN_KEY_PEM), true); error != nil {
		return error
	}

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.CONTROLLER_MANAGER_KUBECONFIG), generator.ca.CertificateFilename, "system:kube-controller-manager", apiServer, generator.config.GetFullLocalAssetFilename(utils.CONTROLLER_MANAGER_PEM), generator.config.GetFullLocalAssetFilename(utils.CONTROLLER_MANAGER_KEY_PEM), true); error != nil {
		return error
	}

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.SCHEDULER_KUBECONFIG), generator.ca.CertificateFilename, "system:kube-scheduler", apiServer, generator.config.GetFullLocalAssetFilename(utils.SCHEDULER_PEM), generator.config.GetFullLocalAssetFilename(utils.SCHEDULER_KEY_PEM), true); error != nil {
		return error
	}

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.PROXY_KUBECONFIG), generator.ca.CertificateFilename, "system:kube-proxy", apiServer, generator.config.GetFullLocalAssetFilename(utils.PROXY_PEM), generator.config.GetFullLocalAssetFilename(utils.PROXY_KEY_PEM), true); error != nil {
		return error
	}

	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KUBELET_KUBECONFIG), generator.ca.CertificateFilename, fmt.Sprintf("system:node:%s", nodeName), apiServer, generator.config.GetFullLocalAssetFilename(utils.KUBELET_PEM), generator.config.GetFullLocalAssetFilename(utils.KUBELET_KEY_PEM), true); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateCephConfig() error {
	return utils.ApplyTemplateAndSave("ceph-config", utils.TEMPLATE_CEPH_CONFIG, struct {
		ClusterID          string
		PublicNetwork      string
		ClusterNetwork     string
		StorageControllers []config.NodeData
		StorageNodes       []config.NodeData
	}{
		ClusterID:          generator.config.Config.ClusterID,
		PublicNetwork:      generator.config.Config.PublicNetwork,
		ClusterNetwork:     generator.config.Config.PublicNetwork,
		StorageControllers: generator.config.GetStorageControllers(),
		StorageNodes:       generator.config.GetStorageNodes(),
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_CONFIG), true, false)
}

func (generator *Generator) generateCephSetup() error {
	return utils.ApplyTemplateAndSave("ceph-setup", utils.TEMPLATE_CEPH_SETUP, struct {
		CephRBDPoolName      string
		CephFSPoolName       string
		PublicNetwork        string
		StorageControllers   []config.NodeData
		StorageNodes         []config.NodeData
		CephConfigDirectory  string
		CephDataDirectory    string
		CephImage            string
		CephManagerPort      uint16
		CephRadosGatewayPort uint16
	}{
		CephRBDPoolName:      utils.CEPH_RBD_POOL_NAME,
		CephFSPoolName:       utils.CEPH_FS_POOL_NAME,
		PublicNetwork:        generator.config.Config.PublicNetwork,
		StorageControllers:   generator.config.GetStorageControllers(),
		StorageNodes:         generator.config.GetStorageNodes(),
		CephConfigDirectory:  generator.config.GetFullTargetAssetDirectory(utils.DirectoryCephConfig),
		CephDataDirectory:    generator.config.GetFullTargetAssetDirectory(utils.DirectoryCephData),
		CephImage:            generator.config.Config.Versions.Ceph,
		CephManagerPort:      utils.PortCephManager,
		CephRadosGatewayPort: utils.PortCephRadosGateway,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_SETUP), true, false)
}

func (generator *Generator) generateCephCSI() error {
	return utils.ApplyTemplateAndSave("ceph-csi", utils.TEMPLATE_CEPH_CSI, struct {
		PodsDirectory           string
		PluginsDirectory        string
		CephFSPluginDirectory   string
		CephRBDPluginDirectory  string
		CephRBDPoolName         string
		CephFSPoolName          string
		StorageControllers      []config.NodeData
		CSIAttacherImage        string
		CSIProvisionerImage     string
		CSIDriverRegistrarImage string
		CSICephRBDPluginImage   string
		CSICephFSPluginImage    string
	}{
		PodsDirectory:           generator.config.GetFullTargetAssetDirectory(utils.DirectoryPodsData),
		PluginsDirectory:        generator.config.GetFullTargetAssetDirectory(utils.DirectoryKubeletPlugins),
		CephFSPluginDirectory:   generator.config.GetFullTargetAssetDirectory(utils.DirectoryCephFsPlugin),
		CephRBDPluginDirectory:  generator.config.GetFullTargetAssetDirectory(utils.DirectoryCephRbdPlugin),
		CephRBDPoolName:         utils.CEPH_RBD_POOL_NAME,
		CephFSPoolName:          utils.CEPH_FS_POOL_NAME,
		StorageControllers:      generator.config.GetStorageControllers(),
		CSIAttacherImage:        generator.config.Config.Versions.CSIAttacher,
		CSIProvisionerImage:     generator.config.Config.Versions.CSIProvisioner,
		CSIDriverRegistrarImage: generator.config.Config.Versions.CSIDriverRegistrar,
		CSICephRBDPluginImage:   generator.config.Config.Versions.CSICephRBDPlugin,
		CSICephFSPluginImage:    generator.config.Config.Versions.CSICephFSPlugin,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_CSI), true, false)
}

func (generator *Generator) generateCephFiles() error {
	if utils.FileExists(generator.config.GetFullLocalAssetFilename(utils.CEPH_MONITOR_KEYRING)) {
		return nil
	}

	monitorKey := utils.GenerateCephKey()
	clientAdminKey := utils.GenerateCephKey()
	clientBootstrapMetadataServerKey := utils.GenerateCephKey()
	clientBootstrapObjectStorageKey := utils.GenerateCephKey()
	clientBootstrapRadosBlockDeviceKey := utils.GenerateCephKey()
	clientBootstrapRadosGatewayKey := utils.GenerateCephKey()
	clientK8STEWKey := utils.GenerateCephKey()

	if error := utils.ApplyTemplateAndSave("ceph-monitor-keyring", utils.TEMPLATE_CEPH_MONITOR_KEYRING, struct {
		MonitorKey                         string
		ClientAdminKey                     string
		ClientBootstrapMetadataServerKey   string
		ClientBootstrapObjectStorageKey    string
		ClientBootstrapRadosBlockDeviceKey string
		ClientBootstrapRadosGatewayKey     string
		ClientK8STEWKey                    string
		CephPoolName                       string
	}{
		MonitorKey:                         monitorKey,
		ClientAdminKey:                     clientAdminKey,
		ClientBootstrapMetadataServerKey:   clientBootstrapMetadataServerKey,
		ClientBootstrapObjectStorageKey:    clientBootstrapObjectStorageKey,
		ClientBootstrapRadosBlockDeviceKey: clientBootstrapRadosBlockDeviceKey,
		ClientBootstrapRadosGatewayKey:     clientBootstrapRadosGatewayKey,
		ClientK8STEWKey:                    clientK8STEWKey,
		CephPoolName:                       utils.CEPH_RBD_POOL_NAME,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_MONITOR_KEYRING), false, false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("ceph-client-admin", utils.TEMPLATE_CEPH_CLIENT_ADMIN_KEYRING, struct {
		Key string
	}{
		Key: clientAdminKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_CLIENT_ADMIN_KEYRING), false, false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-mds-client-keyring", utils.TEMPLATE_CEPH_CLIENT_KEYRING, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-mds",
		Key:  clientBootstrapMetadataServerKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_MDS_KEYRING), false, false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-osd-client-keyring", utils.TEMPLATE_CEPH_CLIENT_KEYRING, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-osd",
		Key:  clientBootstrapObjectStorageKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_OSD_KEYRING), false, false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-rbd-client-keyring", utils.TEMPLATE_CEPH_CLIENT_KEYRING, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-rbd",
		Key:  clientBootstrapRadosBlockDeviceKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_RBD_KEYRING), false, false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-rgw-client-keyring", utils.TEMPLATE_CEPH_CLIENT_KEYRING, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-rgw",
		Key:  clientBootstrapRadosGatewayKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_RGW_KEYRING), false, false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("ceph-secrets", utils.TEMPLATE_CEPH_SECRETS, struct {
		ClientAdminKey  string
		ClientK8STEWKey string
	}{
		ClientAdminKey:  clientAdminKey,
		ClientK8STEWKey: clientK8STEWKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_SECRETS), false, false); error != nil {
		return error
	}

	return nil
}

func (generator *Generator) generateLetsEncryptClusterIssuer() error {
	return utils.ApplyTemplateAndSave("lets-encrypt-cluster-issuer", utils.TEMPLATE_LETSENCRYPT_CLUSTER_ISSUER_SETUP, struct {
		Email string
	}{
		Email: generator.config.Config.Email,
	}, generator.config.GetFullLocalAssetFilename(utils.LETSENCRYPT_CLUSTER_ISSUER), true, false)
}

func (generator *Generator) generateCoreDNSSetup() error {
	return utils.ApplyTemplateAndSave("core-dns", utils.TEMPLATE_COREDNS_SETUP, struct {
		ClusterDomain string
		ClusterDNSIP  string
		CoreDNSImage  string
	}{
		ClusterDomain: generator.config.Config.ClusterDomain,
		ClusterDNSIP:  generator.config.Config.ClusterDNSIP,
		CoreDNSImage:  generator.config.Config.Versions.CoreDNS,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_COREDNS_SETUP), true, false)
}

func (generator *Generator) generateElasticSearchOperatorSetup() error {
	return utils.ApplyTemplateAndSave("elasticsearch-operator", utils.TEMPLATE_ELASTICSEARCH_OPERATOR_SETUP, struct {
		ElasticsearchOperatorImage string
	}{
		ElasticsearchOperatorImage: generator.config.Config.Versions.ElasticsearchOperator,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_ELASTICSEARCH_OPERATOR_SETUP), true, false)
}

func (generator *Generator) generateEFKSetup() error {
	return utils.ApplyTemplateAndSave("efk", utils.TEMPLATE_EFK_SETUP, struct {
		ElasticsearchImage     string
		ElasticsearchCronImage string
		KibanaImage            string
		CerebroImage           string
		FluentBitImage         string
	}{
		ElasticsearchImage:     generator.config.Config.Versions.Elasticsearch,
		ElasticsearchCronImage: generator.config.Config.Versions.ElasticsearchCron,
		KibanaImage:            generator.config.Config.Versions.Kibana,
		CerebroImage:           generator.config.Config.Versions.Cerebro,
		FluentBitImage:         generator.config.Config.Versions.FluentBit,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_EFK_SETUP), true, false)
}

func (generator *Generator) generateARKSetup() error {
	return utils.ApplyTemplateAndSave("ark-setup", utils.TEMPLATE_ARK_SETUP, struct {
		ArkImage         string
		MinioServerImage string
		MinioClientImage string
		PodsDirectory    string
		MinioPort        uint16
	}{
		ArkImage:         generator.config.Config.Versions.Ark,
		MinioServerImage: generator.config.Config.Versions.MinioServer,
		MinioClientImage: generator.config.Config.Versions.MinioClient,
		PodsDirectory:    generator.config.GetFullTargetAssetDirectory(utils.DirectoryPodsData),
		MinioPort:        utils.PortMinio,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_ARK_SETUP), true, false)
}

func (generator *Generator) generateHeapsterSetup() error {
	return utils.ApplyTemplateAndSave("heapster", utils.TEMPLATE_HEAPSTER_SETUP, struct {
		HeapsterImage     string
		AddonResizerImage string
	}{
		HeapsterImage:     generator.config.Config.Versions.Heapster,
		AddonResizerImage: generator.config.Config.Versions.AddonResizer,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_HEAPSTER_SETUP), true, false)
}

func (generator *Generator) generateKubernetesDashboardSetup() error {
	return utils.ApplyTemplateAndSave("kubernetes-dashboard", utils.TEMPLATE_KUBERNETES_DASHBOARD_SETUP, struct {
		ClusterName              string
		KubernetesDashboardPort  uint16
		KubernetesDashboardImage string
	}{
		ClusterName:              generator.config.Config.ClusterName,
		KubernetesDashboardPort:  generator.config.Config.KubernetesDashboardPort,
		KubernetesDashboardImage: generator.config.Config.Versions.KubernetesDashboard,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBERNETES_DASHBOARD_SETUP), true, false)
}

func (generator *Generator) generateCertManagerSetup() error {
	return utils.ApplyTemplateAndSave("cert-manager", utils.TEMPLATE_CERT_MANAGER_SETUP, struct {
		CertManagerControllerImage string
	}{
		CertManagerControllerImage: generator.config.Config.Versions.CertManagerController,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_CERT_MANAGER_SETUP), true, false)
}

func (generator *Generator) generateNginxIngressSetup() error {
	return utils.ApplyTemplateAndSave("nginx-ingress", utils.TEMPLATE_NGINX_INGRESS_SETUP, struct {
		NginxIngressControllerImage     string
		NginxIngressDefaultBackendImage string
	}{
		NginxIngressControllerImage:     generator.config.Config.Versions.NginxIngressController,
		NginxIngressDefaultBackendImage: generator.config.Config.Versions.NginxIngressDefaultBackend,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_NGINX_INGRESS_SETUP), true, false)
}

func (generator *Generator) generateMetricsServerSetup() error {
	return utils.ApplyTemplateAndSave("metrics-server", utils.TEMPLATE_METRICS_SERVER_SETUP, struct {
		MetricsServerImage string
	}{
		MetricsServerImage: generator.config.Config.Versions.MetricsServer,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_METRICS_SERVER_SETUP), true, false)
}

func (generator *Generator) generatePrometheusOperatorSetup() error {
	return utils.ApplyTemplateAndSave("prometheus-operator", utils.TEMPLATE_PROMETHEUS_OPERATOR_SETUP, struct {
		PrometheusOperatorImage       string
		PrometheusConfigReloaderImage string
		ConfigMapReloadImage          string
	}{
		PrometheusOperatorImage:       generator.config.Config.Versions.PrometheusOperator,
		PrometheusConfigReloaderImage: generator.config.Config.Versions.PrometheusConfigReloader,
		ConfigMapReloadImage:          generator.config.Config.Versions.ConfigMapReload,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_PROMETHEUS_OPERATOR_SETUP), true, false)
}

func (generator *Generator) generateKubePrometheusSetup() error {
	return utils.ApplyTemplateAndSave("kube-prometheus", utils.TEMPLATE_KUBE_PROMETHEUS_SETUP, struct {
		AddonResizerImage           string
		KubeStateMetricsImage       string
		GrafanaImage                string
		GrafanaWatcherImage         string
		PrometheusImage             string
		PrometheusNodeExporterImage string
		PrometheusAlertManagerImage string
		GrafanaPort                 uint16
	}{
		AddonResizerImage:           generator.config.Config.Versions.AddonResizer,
		KubeStateMetricsImage:       generator.config.Config.Versions.KubeStateMetrics,
		GrafanaImage:                generator.config.Config.Versions.Grafana,
		GrafanaWatcherImage:         generator.config.Config.Versions.GrafanaWatcher,
		PrometheusImage:             generator.config.Config.Versions.Prometheus,
		PrometheusNodeExporterImage: generator.config.Config.Versions.PrometheusNodeExporter,
		PrometheusAlertManagerImage: generator.config.Config.Versions.PrometheusAlertManager,
		GrafanaPort:                 utils.PortGrafana,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_PROMETHEUS_SETUP), true, true)
}

func (generator *Generator) generateKubePrometheusDatasourceSetup() error {
	return utils.ApplyTemplateAndSave("kube-prometheus-datasource", utils.TEMPLATE_KUBE_PROMETHEUS_DATASOURCE_SETUP, struct{}{}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_PROMETHEUS_DATASOURCE_SETUP), true, true)
}

func (generator *Generator) generateKubePrometheusKubernetesClusterStatusDashboardSetup() error {
	return utils.ApplyTemplateAndSave("kube-prometheus-kubernetes-cluster-status-dashboard", utils.TEMPLATE_KUBE_PROMETHEUS_KUBERNETES_CLUSTER_STATUS_DASHBOARD_SETUP, struct{}{}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_PROMETHEUS_KUBERNETES_CLUSTER_STATUS_DASHBOARD_SETUP), true, true)
}

func (generator *Generator) generateKubePrometheusPodsDashboardSetup() error {
	return utils.ApplyTemplateAndSave("kube-prometheus-pods-dashboard", utils.TEMPLATE_KUBE_PROMETHEUS_PODS_DASHBOARD_SETUP, struct{}{}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_PROMETHEUS_PODS_DASHBOARD_SETUP), true, true)
}

func (generator *Generator) generateKubePrometheusDeploymentDashboardSetup() error {
	return utils.ApplyTemplateAndSave("kube-prometheus-deployment-dashboard", utils.TEMPLATE_KUBE_PROMETHEUS_DEPLOYMENT_DASHBOARD_SETUP, struct{}{}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_PROMETHEUS_DEPLOYMENT_DASHBOARD_SETUP), true, true)
}

func (generator *Generator) generateKubePrometheusKubernetesControlPlaneStatusDashboardSetup() error {
	return utils.ApplyTemplateAndSave("kube-prometheus-kuberntes-control-plane-status-dashboard", utils.TEMPLATE_KUBE_PROMETHEUS_KUBERNETES_CONTROL_PLANE_STATUS_DASHBOARD_SETUP, struct{}{}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_PROMETHEUS_KUBERNETES_CONTROL_PLANE_STATUS_DASHBOARD_SETUP), true, true)
}

func (generator *Generator) generateKubePrometheusStatefulsetDashboardSetup() error {
	return utils.ApplyTemplateAndSave("kube-prometheus-stateful-dashboard", utils.TEMPLATE_KUBE_PROMETHEUS_STATEFULSET_DASHBOARD_SETUP, struct{}{}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_PROMETHEUS_STATEFULSET_DASHBOARD_SETUP), true, true)
}

func (generator *Generator) generateKubePrometheusKubernetesCapacityPlanningDashboardSetup() error {
	return utils.ApplyTemplateAndSave("kube-prometheus-kubernetes-capacity-planning-dashboard", utils.TEMPLATE_KUBE_PROMETHEUS_KUBERNETES_CAPACITY_PLANNING_DASHBOARD_SETUP, struct{}{}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_PROMETHEUS_KUBERNETES_CAPACITY_PLANNING_DASHBOARD_SETUP), true, true)
}

func (generator *Generator) generateKubePrometheusKubernetesResourceRequestsDashboardSetup() error {
	return utils.ApplyTemplateAndSave("kube-prometheus-kubernetes-resource-requests-dashboard", utils.TEMPLATE_KUBE_PROMETHEUS_KUBERNETES_RESOURCE_REQUESTS_DASHBOARD_SETUP, struct{}{}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_PROMETHEUS_KUBERNETES_RESOURCE_REQUESTS_DASHBOARD_SETUP), true, true)
}

func (generator *Generator) generateKubePrometheusKubernetesClusterHealthDashboardSetup() error {
	return utils.ApplyTemplateAndSave("kube-prometheus-kubernetes-cluster-health-dashboard", utils.TEMPLATE_KUBE_PROMETHEUS_KUBERNETES_CLUSTER_HEALTH_DASHBOARD_SETUP, struct{}{}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_PROMETHEUS_KUBERNETES_CLUSTER_HEALTH_DASHBOARD_SETUP), true, true)
}

func (generator *Generator) generateKubePrometheusNodesDashboardSetup() error {
	return utils.ApplyTemplateAndSave("kube-prometheus-nodes-dashboard", utils.TEMPLATE_KUBE_PROMETHEUS_NODES_DASHBOARD_SETUP, struct{}{}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_PROMETHEUS_NODES_DASHBOARD_SETUP), true, true)
}

func (generator *Generator) generateWordpressSetup() error {
	return utils.ApplyTemplateAndSave("wordpress", utils.TEMPLATE_WORDPRESS_SETUP, struct {
		WordPressIngressDomain string
		MySQLImage             string
		WordPressImage         string
		WordPressPort          uint16
	}{
		WordPressIngressDomain: fmt.Sprintf("%s.%s", utils.IngressSubdomainWordpress, generator.config.Config.IngressDomain),
		MySQLImage:             generator.config.Config.Versions.MySQL,
		WordPressImage:         generator.config.Config.Versions.WordPress,
		WordPressPort:          utils.PortWordpress,
	}, generator.config.GetFullLocalAssetFilename(utils.WORDPRESS_SETUP), true, false)
}

func (generator *Generator) generateBashCompletion(binaryName, bashCompletionFilename string) error {
	binary := generator.config.GetFullLocalAssetFilename(binaryName)
	bashCompletionFullFilename := generator.config.GetFullLocalAssetFilename(bashCompletionFilename)

	command := fmt.Sprintf("%s completion bash > %s", binary, bashCompletionFullFilename)

	log.WithFields(log.Fields{"name": bashCompletionFilename}).Info("Generated")

	return utils.RunCommand(command)
}

func (generator *Generator) generateBashCompletionK8STEW() error {
	return generator.generateBashCompletion(utils.BinaryK8sTew, utils.BASH_COMPLETION_K8S_TEW)
}

func (generator *Generator) generateBashCompletionKubectl() error {
	return generator.generateBashCompletion(utils.BinaryKubectl, utils.BASH_COMPLETION_KUBECTL)
}

func (generator *Generator) generateBashCompletionHelm() error {
	return generator.generateBashCompletion(utils.BinaryHelm, utils.BASH_COMPLETION_HELM)
}

func (generator *Generator) generateBashCompletionArk() error {
	return generator.generateBashCompletion(utils.BinaryArk, utils.BASH_COMPLETION_ARK)
}

func (generator *Generator) generateBashCompletionCriCtl() error {
	return generator.generateBashCompletion(utils.BinaryCrictl, utils.BASH_COMPLETION_CRICTL)
}

func (generator *Generator) GenerateFiles() error {
	for _, step := range generator.generatorSteps {
		if error := step(); error != nil {
			return error
		}

		utils.IncreaseProgressStep()
	}

	return nil
}
