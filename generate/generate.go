package generate

import (
	"fmt"

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

func (generator *Generator) generateProfileFile() error {
	return utils.ApplyTemplateAndSave(utils.K8S_TEW_PROFILE_TEMPLATE, struct {
		Binary        string
		BaseDirectory string
	}{
		Binary:        generator.config.GetFullTargetAssetFilename(utils.K8S_TEW_BINARY),
		BaseDirectory: generator.config.Config.DeploymentDirectory,
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_TEW_PROFILE), true)
}

func (generator *Generator) generateServiceFile() error {
	return utils.ApplyTemplateAndSave(utils.SERVICE_CONFIG_TEMPLATE, struct {
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
	return utils.ApplyTemplateAndSave(utils.GOBETWEEN_CONFIG_TEMPLATE, struct {
		LoadBalancerPort uint16
		KubeAPIServers   []string
	}{
		LoadBalancerPort: generator.config.Config.LoadBalancerPort,
		KubeAPIServers:   generator.config.GetKubeAPIServerAddresses(),
	}, generator.config.GetFullLocalAssetFilename(utils.GOBETWEEN_CONFIG), true)
}

func (generator *Generator) generateCNIFiles() error {
	if error := utils.ApplyTemplateAndSave(utils.NET_CONFIG_TEMPLATE, struct {
		ClusterCIDR string
	}{
		ClusterCIDR: generator.config.Config.ClusterCIDR,
	}, generator.config.GetFullLocalAssetFilename(utils.NET_CONFIG), false); error != nil {
		return error
	}

	return utils.ApplyTemplateAndSave(utils.CNI_CONFIG_TEMPLATE, nil, generator.config.GetFullLocalAssetFilename(utils.CNI_CONFIG), false)
}

func (generator *Generator) generateK8SKubeletConfigFile() error {
	return utils.ApplyTemplateAndSave(utils.K8S_KUBELET_CONFIG_TEMPLATE, nil, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBELET_SETUP), true)
}

func (generator *Generator) generateK8SAdminUserConfigFile() error {
	return utils.ApplyTemplateAndSave(utils.K8S_SERVICE_ACCOUNT_CONFIG_TEMPLATE, struct {
		Name      string
		Namespace string
	}{
		Name:      "admin-user",
		Namespace: "kube-system",
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_ADMIN_USER_SETUP), true)
}

func (generator *Generator) generateK8SHelmUserConfigFile() error {
	return utils.ApplyTemplateAndSave(utils.K8S_SERVICE_ACCOUNT_CONFIG_TEMPLATE, struct {
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
		log.WithFields(log.Fields{"filename": fullEncryptionConfigFilename}).Info("skipped")

		return nil
	}

	encryptionKey, error := pki.GenerateEncryptionConfig()
	if error != nil {
		return error
	}

	return utils.ApplyTemplateAndSave(utils.ENCRYPTION_CONFIG_TEMPLATE, struct {
		EncryptionKey string
	}{
		EncryptionKey: encryptionKey,
	}, fullEncryptionConfigFilename, false)
}

func (generator *Generator) generateContainerdConfig() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if !node.IsWorker() {
			continue
		}

		if error := utils.ApplyTemplateAndSave(utils.CONTAINERD_CONFIG_TEMPLATE, struct {
			ContainerdRootDirectory  string
			ContainerdStateDirectory string
			ContainerdSock           string
			CNIConfigDirectory       string
			CNIBinariesDirectory     string
			CRIBinariesDirectory     string
			IP                       string
		}{
			ContainerdRootDirectory:  generator.config.GetFullTargetAssetDirectory(utils.CONTAINERD_DATA_DIRECTORY),
			ContainerdStateDirectory: generator.config.GetFullTargetAssetDirectory(utils.CONTAINERD_STATE_DIRECTORY),
			ContainerdSock:           generator.config.GetFullTargetAssetFilename(utils.CONTAINERD_SOCK),
			CNIConfigDirectory:       generator.config.GetFullTargetAssetDirectory(utils.CNI_CONFIG_DIRECTORY),
			CNIBinariesDirectory:     generator.config.GetFullTargetAssetDirectory(utils.CNI_BINARIES_DIRECTORY),
			CRIBinariesDirectory:     generator.config.GetFullTargetAssetDirectory(utils.CRI_BINARIES_DIRECTORY),
			IP:                       node.IP,
		}, generator.config.GetFullLocalAssetFilename(utils.CONTAINERD_CONFIG), true); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateKubeSchedulerConfig() error {
	return utils.ApplyTemplateAndSave(utils.KUBE_SCHEDULER_CONFIGURATION_TEMPLATE, struct {
		KubeConfig string
	}{
		KubeConfig: generator.config.GetFullTargetAssetFilename(utils.SCHEDULER_KUBECONFIG),
	}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBE_SCHEDULER_CONFIG), true)
}

func (generator *Generator) generateKubeletConfig() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if !node.IsWorker() {
			continue
		}

		if error := utils.ApplyTemplateAndSave(utils.KUBELET_CONFIGURATION_TEMPLATE, struct {
			CA                  string
			CertificateFilename string
			KeyFilename         string
			ClusterDNSIP        string
			PODCIDR             string
			StaticPodPath       string
		}{
			CA:                  generator.config.GetFullTargetAssetFilename(utils.CA_PEM),
			CertificateFilename: generator.config.GetFullTargetAssetFilename(utils.KUBELET_PEM),
			KeyFilename:         generator.config.GetFullTargetAssetFilename(utils.KUBELET_KEY_PEM),
			ClusterDNSIP:        generator.config.Config.ClusterDNSIP,
			PODCIDR:             generator.config.Config.ClusterCIDR,
			StaticPodPath:       generator.config.GetFullTargetAssetDirectory(utils.K8S_MANIFESTS_DIRECTORY),
		}, generator.config.GetFullLocalAssetFilename(utils.K8S_KUBELET_CONFIG), true); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateCertificates() (*pki.CertificateAndPrivateKey, error) {
	fullCAFilename := generator.config.GetFullLocalAssetFilename(utils.CA_PEM)
	fullCAKeyFilename := generator.config.GetFullLocalAssetFilename(utils.CA_KEY_PEM)

	// Generate CA if not done already
	if error := pki.GenerateCA(generator.rsaSize, generator.caValidityPeriod, "Kubernetes", "Kubernetes", fullCAFilename, fullCAKeyFilename); error != nil {
		return nil, error
	}

	// Load ca certificate and private key
	ca, error := pki.LoadCertificateAndPrivateKey(fullCAFilename, fullCAKeyFilename)
	if error != nil {
		return nil, error
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
	if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "flanneld", "flanneld", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.FLANNELD_PEM), generator.config.GetFullLocalAssetFilename(utils.FLANNELD_KEY_PEM), false); error != nil {
		return nil, error
	}

	// Generate virtual-ip certificate
	if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "virtual-ip", "virtual-ip", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.VIRTUAL_IP_PEM), generator.config.GetFullLocalAssetFilename(utils.VIRTUAL_IP_KEY_PEM), false); error != nil {
		return nil, error
	}

	// Generate admin certificate
	if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "admin", "system:masters", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.ADMIN_PEM), generator.config.GetFullLocalAssetFilename(utils.ADMIN_KEY_PEM), false); error != nil {
		return nil, error
	}

	// Generate kuberentes certificate
	if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "kubernetes", "Kubernetes", kubernetesDNSNames, kubernetesIPAddresses, generator.config.GetFullLocalAssetFilename(utils.KUBERNETES_PEM), generator.config.GetFullLocalAssetFilename(utils.KUBERNETES_KEY_PEM), true); error != nil {
		return nil, error
	}

	// Generate aggregator certificate
	if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "aggregator", "Kubernetes", kubernetesDNSNames, kubernetesIPAddresses, generator.config.GetFullLocalAssetFilename(utils.AGGREGATOR_PEM), generator.config.GetFullLocalAssetFilename(utils.AGGREGATOR_KEY_PEM), true); error != nil {
		return nil, error
	}

	// Generate service accounts certificate
	if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "service-accounts", "Kubernetes", kubernetesDNSNames, kubernetesIPAddresses, generator.config.GetFullLocalAssetFilename(utils.SERVICE_ACCOUNT_PEM), generator.config.GetFullLocalAssetFilename(utils.SERVICE_ACCOUNT_KEY_PEM), false); error != nil {
		return nil, error
	}

	// Generate controller manager certificate
	if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "system:kube-controller-manager", "system:node-controller-manager", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.CONTROLLER_MANAGER_PEM), generator.config.GetFullLocalAssetFilename(utils.CONTROLLER_MANAGER_KEY_PEM), false); error != nil {
		return nil, error
	}

	// Generate scheduler certificate
	if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "system:kube-scheduler", "system:kube-scheduler", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.SCHEDULER_PEM), generator.config.GetFullLocalAssetFilename(utils.SCHEDULER_KEY_PEM), false); error != nil {
		return nil, error
	}

	// Generate proxy certificate
	if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, "system:kube-proxy", "system:node-proxier", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.PROXY_PEM), generator.config.GetFullLocalAssetFilename(utils.PROXY_KEY_PEM), false); error != nil {
		return nil, error
	}

	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := pki.GenerateClient(ca, generator.rsaSize, generator.clientValidityPeriod, fmt.Sprintf("system:node:%s", nodeName), "system:nodes", []string{nodeName}, []string{node.IP}, generator.config.GetFullLocalAssetFilename(utils.KUBELET_PEM), generator.config.GetFullLocalAssetFilename(utils.KUBELET_KEY_PEM), false); error != nil {
			return nil, error
		}
	}

	return ca, nil
}

func (generator *Generator) generateConfigKubeConfig(kubeConfigFilename, caFilename, user, apiServers, certificateFilename, keyFilename string, force bool) error {
	if utils.FileExists(kubeConfigFilename) && !force {
		log.WithFields(log.Fields{"filename": kubeConfigFilename}).Info("skipped")

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

	if error := utils.ApplyTemplateAndSave(utils.KUBE_CONFIG_TEMPLATE, struct {
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

func (generator *Generator) generateKubeConfigs(ca *pki.CertificateAndPrivateKey) error {
	apiServer, error := generator.config.GetAPIServerIP()
	if error != nil {
		return error
	}

	apiServer = fmt.Sprintf("%s:%d", apiServer, generator.config.Config.LoadBalancerPort)

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.ADMIN_KUBECONFIG), ca.CertificateFilename, "admin", apiServer, generator.config.GetFullLocalAssetFilename(utils.ADMIN_PEM), generator.config.GetFullLocalAssetFilename(utils.ADMIN_KEY_PEM), true); error != nil {
		return error
	}

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.CONTROLLER_MANAGER_KUBECONFIG), ca.CertificateFilename, "system:kube-controller-manager", apiServer, generator.config.GetFullLocalAssetFilename(utils.CONTROLLER_MANAGER_PEM), generator.config.GetFullLocalAssetFilename(utils.CONTROLLER_MANAGER_KEY_PEM), true); error != nil {
		return error
	}

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.SCHEDULER_KUBECONFIG), ca.CertificateFilename, "system:kube-scheduler", apiServer, generator.config.GetFullLocalAssetFilename(utils.SCHEDULER_PEM), generator.config.GetFullLocalAssetFilename(utils.SCHEDULER_KEY_PEM), true); error != nil {
		return error
	}

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.PROXY_KUBECONFIG), ca.CertificateFilename, "system:kube-proxy", apiServer, generator.config.GetFullLocalAssetFilename(utils.PROXY_PEM), generator.config.GetFullLocalAssetFilename(utils.PROXY_KEY_PEM), true); error != nil {
		return error
	}

	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KUBELET_KUBECONFIG), ca.CertificateFilename, fmt.Sprintf("system:node:%s", nodeName), apiServer, generator.config.GetFullLocalAssetFilename(utils.KUBELET_PEM), generator.config.GetFullLocalAssetFilename(utils.KUBELET_KEY_PEM), true); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateCephConfig() error {
	return utils.ApplyTemplateAndSave(utils.CEPH_CONFIG_TEMPLATE, struct {
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
	return utils.ApplyTemplateAndSave(utils.CEPH_SETUP_TEMPLATE, struct {
		CephPoolName        string
		PublicNetwork       string
		StorageControllers  []config.NodeData
		StorageNodes        []config.NodeData
		CephConfigDirectory string
		CephDataDirectory   string
	}{
		CephPoolName:        utils.CEPH_POOL_NAME,
		PublicNetwork:       generator.config.Config.PublicNetwork,
		StorageControllers:  generator.config.GetStorageControllers(),
		StorageNodes:        generator.config.GetStorageNodes(),
		CephConfigDirectory: generator.config.GetFullTargetAssetDirectory(utils.CEPH_CONFIG_DIRECTORY),
		CephDataDirectory:   generator.config.GetFullTargetAssetDirectory(utils.CEPH_DATA_DIRECTORY),
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

	if error := utils.ApplyTemplateAndSave(utils.CEPH_MONITOR_KEYRING_TEMPLATE, struct {
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

	if error := utils.ApplyTemplateAndSave(utils.CEPH_CLIENT_ADMIN_KEYRING_TEMPLATE, struct {
		Key string
	}{
		Key: clientAdminKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_CLIENT_ADMIN_KEYRING), false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave(utils.CEPH_KEYRING_TEMPLATE, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-mds",
		Key:  clientBootstrapMetadataServerKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_MDS_KEYRING), false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave(utils.CEPH_KEYRING_TEMPLATE, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-osd",
		Key:  clientBootstrapObjectStorageKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_OSD_KEYRING), false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave(utils.CEPH_KEYRING_TEMPLATE, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-rbd",
		Key:  clientBootstrapRadosBlockDeviceKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_RBD_KEYRING), false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave(utils.CEPH_KEYRING_TEMPLATE, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-rgw",
		Key:  clientBootstrapRadosGatewayKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CEPH_BOOTSTRAP_RGW_KEYRING), false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave(utils.CEPH_SECRETS_TEMPLATE, struct {
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

func (generator *Generator) GenerateFiles() error {
	// Generate profile file
	if error := generator.generateProfileFile(); error != nil {
		return error
	}

	// Generate systemd file
	if error := generator.generateServiceFile(); error != nil {
		return error
	}

	// Generate load balancer configuration
	if error := generator.generateGobetweenConfig(); error != nil {
		return error
	}
	// Generate scheduler config
	if error := generator.generateKubeSchedulerConfig(); error != nil {
		return error
	}

	// Generate kubelet config
	if error := generator.generateKubeletConfig(); error != nil {
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

	// Generate helm user configuration
	if error := generator.generateK8SHelmUserConfigFile(); error != nil {
		return error
	}

	// Generate containerd config
	if error := generator.generateContainerdConfig(); error != nil {
		return error
	}

	// Generate container network interface files
	if error := generator.generateCNIFiles(); error != nil {
		return error
	}

	// Generate kubernetes security file
	if error := generator.generateEncryptionFile(); error != nil {
		return error
	}

	// Generate kubeconfig files
	ca, error := generator.generateCertificates()
	if error != nil {
		return error
	}

	// Generate kubeconfig files
	if error := generator.generateKubeConfigs(ca); error != nil {
		return error
	}

	// Generate ceph files
	if error := generator.generateCephFiles(); error != nil {
		return error
	}

	return nil
}
