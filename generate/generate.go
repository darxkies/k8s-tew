package generate

import (
	"fmt"

	"github.com/darxkies/k8s-tew/config"

	"github.com/darxkies/k8s-tew/pki"
	"github.com/darxkies/k8s-tew/utils"
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
		// Generate wordpress setup file
		generator.generateWordpressSetup,
	}

	return generator
}

func (generator *Generator) Steps() int {
	return len(generator.generatorSteps)
}

func (generator *Generator) generateProfileFile() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_K8S_TEW_PROFILE, struct {
		Binary        string
		BaseDirectory string
	}{
		Binary:        generator.config.GetFullTargetAssetFilename(utils.K8S_TEW_BINARY),
		BaseDirectory: generator.config.Config.DeploymentDirectory,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_TEW_PROFILE), true)
}

func (generator *Generator) generateServiceFile() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_K8S_TEW_SERVICE, struct {
		ProjectTitle  string
		Command       string
		BaseDirectory string
		Binary        string
	}{
		ProjectTitle:  utils.PROJECT_TITLE,
		Command:       generator.config.GetFullTargetAssetFilename(utils.K8S_TEW_BINARY),
		BaseDirectory: generator.config.Config.DeploymentDirectory,
		Binary:        utils.K8S_TEW_BINARY,
	}, generator.config.GetFullLocalAssetFilename(utils.SERVICE_CONFIG), true)
}

func (generator *Generator) generateGobetweenConfig() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_GOBETWEEN_TOML, struct {
		LoadBalancerPort uint16
		KubeAPIServers   []string
	}{
		LoadBalancerPort: generator.config.Config.LoadBalancerPort,
		KubeAPIServers:   generator.config.GetKubeAPIServerAddresses(),
	}, generator.config.GetFullLocalAssetFilename(utils.GOBETWEEN_CONFIG), true)
}

func (generator *Generator) generateCalicoSetup() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_CALICO_SETUP, struct {
		ClusterCIDR          string
		CNIConfigDirectory   string
		CNIBinariesDirectory string
		CalicoTyphaImage     string
		CalicoNodeImage      string
		CalicoCNIImage       string
	}{
		ClusterCIDR:          generator.config.Config.ClusterCIDR,
		CNIConfigDirectory:   generator.config.GetFullTargetAssetDirectory(utils.CNI_CONFIG_DIRECTORY),
		CNIBinariesDirectory: generator.config.GetFullTargetAssetDirectory(utils.CNI_BINARIES_DIRECTORY),
		CalicoTyphaImage:     utils.GetFullImageName(utils.IMAGE_CALICO_TYPHA, generator.config.Config.Versions.CalicoTypha),
		CalicoNodeImage:      utils.GetFullImageName(utils.IMAGE_CALICO_NODE, generator.config.Config.Versions.CalicoNode),
		CalicoCNIImage:       utils.GetFullImageName(utils.IMAGE_CALICO_CNI, generator.config.Config.Versions.CalicoCNI),
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_CALICO_SETUP), true)
}

func (generator *Generator) generateK8SKubeletConfigFile() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_KUBELET_SETUP, nil, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBELET_SETUP), true)
}

func (generator *Generator) generateK8SAdminUserConfigFile() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_SERVICE_ACCOUNT, struct {
		Name      string
		Namespace string
	}{
		Name:      "admin-user",
		Namespace: "kube-system",
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_ADMIN_USER_SETUP), true)
}

func (generator *Generator) generateK8SHelmUserConfigFile() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_SERVICE_ACCOUNT, struct {
		Name      string
		Namespace string
	}{
		Name:      utils.HELM_SERVICE_ACCOUNT,
		Namespace: "kube-system",
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_HELM_USER_SETUP), true)
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

	return utils.ApplyTemplateAndSave(utils.TEMPLATE_ENCRYPTION_CONFIG, struct {
		EncryptionKey string
	}{
		EncryptionKey: encryptionKey,
	}, fullEncryptionConfigFilename, false)
}

func (generator *Generator) generateContainerdConfig() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := utils.ApplyTemplateAndSave(utils.TEMPLATE_CONTAINERD_TOML, struct {
			ContainerdRootDirectory  string
			ContainerdStateDirectory string
			ContainerdSock           string
			CNIConfigDirectory       string
			CNIBinariesDirectory     string
			CRIBinariesDirectory     string
			IP                       string
			PauseImage               string
		}{
			ContainerdRootDirectory:  generator.config.GetFullTargetAssetDirectory(utils.CONTAINERD_DATA_DIRECTORY),
			ContainerdStateDirectory: generator.config.GetFullTargetAssetDirectory(utils.CONTAINERD_STATE_DIRECTORY),
			ContainerdSock:           generator.config.GetFullTargetAssetFilename(utils.CONTAINERD_SOCK),
			CNIConfigDirectory:       generator.config.GetFullTargetAssetDirectory(utils.CNI_CONFIG_DIRECTORY),
			CNIBinariesDirectory:     generator.config.GetFullTargetAssetDirectory(utils.CNI_BINARIES_DIRECTORY),
			CRIBinariesDirectory:     generator.config.GetFullTargetAssetDirectory(utils.CRI_BINARIES_DIRECTORY),
			IP:                       node.IP,
			PauseImage:               utils.GetFullImageName(utils.IMAGE_PAUSE, generator.config.Config.Versions.Pause),
		}, generator.config.GetFullLocalAssetFilename(utils.CONTAINERD_CONFIG), true); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateKubeSchedulerConfig() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_KUBE_SCHEDULER_CONFIGURATION, struct {
		KubeConfig string
	}{
		KubeConfig: generator.config.GetFullTargetAssetFilename(utils.SCHEDULER_KUBECONFIG),
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_SCHEDULER_CONFIG), true)
}

func (generator *Generator) generateKubeletConfig() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := utils.ApplyTemplateAndSave(utils.TEMPLATE_KUBELET_CONFIGURATION, struct {
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
			StaticPodPath:       generator.config.GetFullTargetAssetDirectory(utils.K8S_MANIFESTS_DIRECTORY),
		}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBELET_CONFIG), true); error != nil {
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

	// Generate flanneld certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, "flanneld", "flanneld", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.FLANNELD_PEM), generator.config.GetFullLocalAssetFilename(utils.FLANNELD_KEY_PEM), false); error != nil {
		return error
	}

	// Generate virtual-ip certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, "virtual-ip", "virtual-ip", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.VIRTUAL_IP_PEM), generator.config.GetFullLocalAssetFilename(utils.VIRTUAL_IP_KEY_PEM), false); error != nil {
		return error
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

	if error := utils.ApplyTemplateAndSave(utils.TEMPLATE_KUBECONFIG, struct {
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
	}, kubeConfigFilename, true); error != nil {
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
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_CEPH_CONFIG, struct {
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
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_CONFIG), true)
}

func (generator *Generator) generateCephSetup() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_CEPH_SETUP, struct {
		CephPoolName        string
		PublicNetwork       string
		StorageControllers  []config.NodeData
		StorageNodes        []config.NodeData
		CephConfigDirectory string
		CephDataDirectory   string
		RBDProvisionerImage string
		CephImage           string
	}{
		CephPoolName:        utils.CEPH_POOL_NAME,
		PublicNetwork:       generator.config.Config.PublicNetwork,
		StorageControllers:  generator.config.GetStorageControllers(),
		StorageNodes:        generator.config.GetStorageNodes(),
		CephConfigDirectory: generator.config.GetFullTargetAssetDirectory(utils.CEPH_CONFIG_DIRECTORY),
		CephDataDirectory:   generator.config.GetFullTargetAssetDirectory(utils.CEPH_DATA_DIRECTORY),
		RBDProvisionerImage: utils.GetFullImageName(utils.IMAGE_RBD_PROVISIONER, generator.config.Config.Versions.RBDProvisioner),
		CephImage:           utils.GetFullImageName(utils.IMAGE_CEPH, generator.config.Config.Versions.Ceph),
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_SETUP), true)
}

func (generator *Generator) generateCephFiles() error {
	if error := generator.generateCephConfig(); error != nil {
		return error
	}

	if error := generator.generateCephSetup(); error != nil {
		return error
	}

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

	if error := utils.ApplyTemplateAndSave(utils.TEMPLATE_CEPH_MONITOR_KEYRING, struct {
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
		CephPoolName:                       utils.CEPH_POOL_NAME,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_MONITOR_KEYRING), false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave(utils.TEMPLATE_CEPH_CLIENT_ADMIN_KEYRING, struct {
		Key string
	}{
		Key: clientAdminKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_CLIENT_ADMIN_KEYRING), false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave(utils.TEMPLATE_CEPH_CLIENT_KEYRING, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-mds",
		Key:  clientBootstrapMetadataServerKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_MDS_KEYRING), false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave(utils.TEMPLATE_CEPH_CLIENT_KEYRING, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-osd",
		Key:  clientBootstrapObjectStorageKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_OSD_KEYRING), false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave(utils.TEMPLATE_CEPH_CLIENT_KEYRING, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-rbd",
		Key:  clientBootstrapRadosBlockDeviceKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_RBD_KEYRING), false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave(utils.TEMPLATE_CEPH_CLIENT_KEYRING, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-rgw",
		Key:  clientBootstrapRadosGatewayKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_RGW_KEYRING), false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave(utils.TEMPLATE_CEPH_SECRETS, struct {
		ClientAdminKey  string
		ClientK8STEWKey string
	}{
		ClientAdminKey:  clientAdminKey,
		ClientK8STEWKey: clientK8STEWKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_SECRETS), false); error != nil {
		return error
	}

	return nil
}

func (generator *Generator) generateLetsEncryptClusterIssuer() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_LETSENCRYPT_CLUSTER_ISSUER_SETUP, struct {
		Email string
	}{
		Email: generator.config.Config.Email,
	}, generator.config.GetFullLocalAssetFilename(utils.LETSENCRYPT_CLUSTER_ISSUER), true)
}

func (generator *Generator) generateCoreDNSSetup() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_COREDNS_SETUP, struct {
		ClusterDomain string
		ClusterDNSIP  string
		CoreDNSImage  string
	}{
		ClusterDomain: generator.config.Config.ClusterDomain,
		ClusterDNSIP:  generator.config.Config.ClusterDNSIP,
		CoreDNSImage:  utils.GetFullImageName(utils.IMAGE_COREDNS, generator.config.Config.Versions.CoreDNS),
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_COREDNS_SETUP), true)
}

func (generator *Generator) generateElasticSearchOperatorSetup() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_ELASTICSEARCH_OPERATOR_SETUP, struct {
		ElasticsearchOperatorImage string
	}{
		ElasticsearchOperatorImage: utils.GetFullImageName(utils.IMAGE_ELASTICSEARCH_OPERATOR, generator.config.Config.Versions.ElasticsearchOperator),
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_ELASTICSEARCH_OPERATOR_SETUP), true)
}

func (generator *Generator) generateEFKSetup() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_EFK_SETUP, struct {
		ElasticsearchImage     string
		ElasticsearchCronImage string
		KibanaImage            string
		CerebroImage           string
		FluentBitImage         string
	}{
		ElasticsearchImage:     utils.GetFullImageName(utils.IMAGE_ELASTICSEARCH, generator.config.Config.Versions.Elasticsearch),
		ElasticsearchCronImage: utils.GetFullImageName(utils.IMAGE_ELASTICSEARCH_CRON, generator.config.Config.Versions.ElasticsearchCron),
		KibanaImage:            utils.GetFullImageName(utils.IMAGE_KIBANA, generator.config.Config.Versions.Kibana),
		CerebroImage:           utils.GetFullImageName(utils.IMAGE_CEREBRO, generator.config.Config.Versions.Cerebro),
		FluentBitImage:         utils.GetFullImageName(utils.IMAGE_FLUENT_BIT, generator.config.Config.Versions.FluentBit),
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_EFK_SETUP), true)
}

func (generator *Generator) generateARKSetup() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_ARK_SETUP, struct {
		ArkImage         string
		MinioServerImage string
		MinioClientImage string
		PodsDirectory    string
	}{
		ArkImage:         utils.GetFullImageName(utils.IMAGE_ARK, generator.config.Config.Versions.Ark),
		MinioServerImage: utils.GetFullImageName(utils.IMAGE_MINIO_SERVER, generator.config.Config.Versions.MinioServer),
		MinioClientImage: utils.GetFullImageName(utils.IMAGE_MINIO_CLIENT, generator.config.Config.Versions.MinioClient),
		PodsDirectory:    generator.config.GetFullTargetAssetDirectory(utils.PODS_DATA_DIRECTORY),
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_ARK_SETUP), true)
}

func (generator *Generator) generateWordpressSetup() error {
	return utils.ApplyTemplateAndSave(utils.TEMPLATE_WORDPRESS_SETUP, struct {
		IngressDomain string
	}{
		IngressDomain: generator.config.Config.IngressDomain,
	}, generator.config.GetFullLocalAssetFilename(utils.WORDPRESS_SETUP), true)
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
