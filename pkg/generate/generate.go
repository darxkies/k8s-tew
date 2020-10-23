package generate

import (
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"

	"github.com/darxkies/k8s-tew/pkg/ceph"
	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/pki"
	"github.com/darxkies/k8s-tew/pkg/utils"
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
		// Generate Systemd file
		generator.generateServiceFile,
		// Generate Load Balancer configuration
		generator.generateGobetweenConfig,
		// Generate Calico setup
		generator.generateCalicoSetup,
		// Generate MetalLB setup
		generator.generateMetalLBSetup,
		// Generate Proxy config
		generator.generateKubeProxyConfig,
		// Generate Scheduler config
		generator.generateKubeSchedulerConfig,
		// Generate Kubelet config
		generator.generateKubeletConfig,
		// Generate Kubelet configuration
		generator.generateK8SKubeletConfigFile,
		// Generate Dashboard admin user configuration
		generator.generateK8SAdminUserConfigFile,
		// Generate Containerd config
		generator.generateContainerdConfig,
		// Generate Kubernetes security file
		generator.generateEncryptionFile,
		// Generate Kubeconfig files
		generator.generateCertificates,
		// Generate Kubeconfig files
		generator.generateKubeConfigs,
		// Generate Ceph Manager secrets file
		generator.generateCephManagerCredentials,
		// Generate Ceph Rados Gateway  secrets file
		generator.generateCephRadosGatewayCredentials,
		// Generate Ceph Config
		generator.generateCephSetup,
		// Generate Ceph CSI
		generator.generateCephCSI,
		// Generate Ceph files
		generator.generateCephFiles,
		// Generate Let's Encrypt Cluster Issuer
		generator.generateLetsEncryptClusterIssuer,
		// Generate CoreDNS setup file
		generator.generateCoreDNSSetup,
		// Generate Elasticsearch config map file
		generator.generateElasticsearchConfigMap,
		// Generate ElasticSearch credentials
		generator.generateElasticsearchCredentials,
		// Generate ElasticSearch/Fluent-Bit/Kibana setup file
		generator.generateEFKSetup,
		// Generate Minio secrets file
		generator.generateMinioCredentials,
		// Generate Minio config map
		generator.generateMinioConfigMap,
		// Generate Cerebro secrets file
		generator.generateCerebroCredentials,
		// Generate Velero setup file
		generator.generateVeleroSetup,
		// Generate Kubernetes dashboard setup file
		generator.generateKubernetesDashboardSetup,
		// Generate cert-manager setup file
		generator.generateCertManagerSetup,
		// Generate Nginx ingress setup file
		generator.generateNginxIngressSetup,
		// Generate Metrics Server setup file
		generator.generateMetricsServerSetup,
		// Generate Prometheus setup file
		generator.generatePrometheusSetup,
		// Generate Kube State Metrics setup file
		generator.generateKubeStateMetricsSetup,
		// Generate Node Exporter setup file
		generator.generateNodeExporterSetup,
		// Generate Grafana config map
		generator.generateGrafanaConfigMap,
		// Generate Grafana secrets file
		generator.generateGrafanaCredentials,
		// Generate Grafana setup file
		generator.generateGrafanaSetup,
		// Generate Alert Manager setup file
		generator.generateAlertManagerSetup,
		// Generate Wordpress setup file
		generator.generateWordpressSetup,
		// Generate Gobetween manifest
		generator.generateManifestGobetween,
		// Generate Controller Virtual-IP manifest
		generator.generateManifestControllerVirtualIP,
		// Generate Worker Virtual-IP manifest
		generator.generateManifestWorkerVirtualIP,
		// Generate Etcd manifest
		generator.generateManifestEtcd,
		// Generate Kube-Apiserver manifest
		generator.generateManifestKubeApiserver,
		// Generate Kube-Controller-Manager manifest
		generator.generateManifestKubeControllerManager,
		// Generate Kube-Scheduler manifest
		generator.generateManifestKubeScheduler,
		// Generate Kube-Proxy manifest
		generator.generateManifestKubeProxy,
	}

	return generator
}

func (generator *Generator) Steps() int {
	return len(generator.generatorSteps)
}

func (generator *Generator) generateProfileFile() error {
	return utils.ApplyTemplateAndSave("profile", utils.TemplateK8sTewProfile, struct {
		Binary        string
		BaseDirectory string
	}{
		Binary:        generator.config.GetFullTargetAssetFilename(utils.BinaryK8sTew),
		BaseDirectory: generator.config.Config.DeploymentDirectory,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sTewProfile), true, false)
}

func (generator *Generator) generateServiceFile() error {
	return utils.ApplyTemplateAndSave("service", utils.TemplateK8sTewService, struct {
		ProjectTitle  string
		Command       string
		BaseDirectory string
		Binary        string
	}{
		ProjectTitle:  utils.ProjectTitle,
		Command:       generator.config.GetFullTargetAssetFilename(utils.BinaryK8sTew),
		BaseDirectory: generator.config.Config.DeploymentDirectory,
		Binary:        utils.BinaryK8sTew,
	}, generator.config.GetFullLocalAssetFilename(utils.ServiceConfig), true, false)
}

func (generator *Generator) generateGobetweenConfig() error {
	return utils.ApplyTemplateAndSave("gobetween", utils.TemplateGobetweenToml, struct {
		LoadBalancerPort uint16
		KubeAPIServers   []string
	}{
		LoadBalancerPort: generator.config.Config.LoadBalancerPort,
		KubeAPIServers:   generator.config.GetKubeAPIServerAddresses(),
	}, generator.config.GetFullLocalAssetFilename(utils.GobetweenConfig), true, false)
}

func (generator *Generator) generateCalicoSetup() error {
	return utils.ApplyTemplateAndSave("calico-setup", utils.TemplateCalicoSetup, struct {
		CalicoTyphaIP              string
		ClusterCIDR                string
		CNIConfigDirectory         string
		CNIBinariesDirectory       string
		DynamicDataDirectory       string
		VarRunDirectory            string
		KubeletPluginsDirectory    string
		CalicoPod2DaemonImage      string
		CalicoTyphaImage           string
		CalicoNodeImage            string
		CalicoCNIImage             string
		CalicoKubeControllersImage string
	}{
		CalicoTyphaIP:              generator.config.Config.CalicoTyphaIP,
		ClusterCIDR:                generator.config.Config.ClusterCIDR,
		CNIConfigDirectory:         generator.config.GetFullTargetAssetDirectory(utils.DirectoryCniConfig),
		CNIBinariesDirectory:       generator.config.GetFullTargetAssetDirectory(utils.DirectoryCniBinaries),
		DynamicDataDirectory:       generator.config.GetFullTargetAssetDirectory(utils.DirectoryDynamicData),
		VarRunDirectory:            generator.config.GetFullTargetAssetDirectory(utils.DirectoryVarRun),
		KubeletPluginsDirectory:    generator.config.GetFullTargetAssetDirectory(utils.DirectoryKubeletPlugins),
		CalicoPod2DaemonImage:      generator.config.Config.Versions.CalicoPod2Daemon,
		CalicoTyphaImage:           generator.config.Config.Versions.CalicoTypha,
		CalicoNodeImage:            generator.config.Config.Versions.CalicoNode,
		CalicoCNIImage:             generator.config.Config.Versions.CalicoCNI,
		CalicoKubeControllersImage: generator.config.Config.Versions.CalicoKubeControllers,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sCalicoSetup), true, false)
}

func (generator *Generator) generateMetalLBSetup() error {
	addresses := strings.Split(generator.config.Config.MetalLBAddresses, ",")

	return utils.ApplyTemplateAndSave("metallb-setup", utils.TemplateMetalLBSetup, struct {
		MetalLBControllerImage string
		MetalLBSpeakerImage    string
		MetalLBAddresses       []string
	}{
		MetalLBControllerImage: generator.config.Config.Versions.MetalLBController,
		MetalLBSpeakerImage:    generator.config.Config.Versions.MetalLBSpeaker,
		MetalLBAddresses:       addresses,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sMetalLBSetup), true, false)
}

func (generator *Generator) generateK8SKubeletConfigFile() error {
	return utils.ApplyTemplateAndSave("kubelet-config", utils.TemplateKubeletSetup, nil, generator.config.GetFullLocalAssetFilename(utils.K8sKubeletSetup), true, false)
}

func (generator *Generator) generateK8SAdminUserConfigFile() error {
	return utils.ApplyTemplateAndSave("admin-user-config", utils.TemplateServiceAccount, struct {
		Name      string
		Namespace string
	}{
		Name:      utils.AdminUserName,
		Namespace: utils.AdminUserNamespace,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sAdminUserSetup), true, false)
}

func (generator *Generator) generateEncryptionFile() error {
	fullEncryptionConfigFilename := generator.config.GetFullLocalAssetFilename(utils.EncryptionConfig)

	if utils.FileExists(fullEncryptionConfigFilename) {
		utils.LogDebugFilename("skipped", fullEncryptionConfigFilename)

		return nil
	}

	encryptionKey, error := pki.GenerateEncryptionConfig()
	if error != nil {
		return error
	}

	return utils.ApplyTemplateAndSave("encryption-config", utils.TemplateEncryptionConfig, struct {
		EncryptionKey string
	}{
		EncryptionKey: encryptionKey,
	}, fullEncryptionConfigFilename, false, false)
}

func (generator *Generator) generateContainerdConfig() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := utils.ApplyTemplateAndSave("containerd-config", utils.TemplateContainerdToml, struct {
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
			ContainerdSock:           generator.config.GetFullTargetAssetFilename(utils.ContainerdSock),
			CNIConfigDirectory:       generator.config.GetFullTargetAssetDirectory(utils.DirectoryCniConfig),
			CNIBinariesDirectory:     generator.config.GetFullTargetAssetDirectory(utils.DirectoryCniBinaries),
			CRIBinariesDirectory:     generator.config.GetFullTargetAssetDirectory(utils.DirectoryCriBinaries),
			IP:                       node.IP,
			PauseImage:               generator.config.Config.Versions.Pause,
		}, generator.config.GetFullLocalAssetFilename(utils.ContainerdConfig), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateKubeProxyConfig() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if _error := utils.ApplyTemplateAndSave("kube-proxy-config", utils.TemplateKubeProxyConfiguration, struct {
			KubeConfig  string
			ClusterCIDR string
		}{
			KubeConfig:  generator.config.GetFullTargetAssetFilename(utils.KubeconfigProxy),
			ClusterCIDR: generator.config.Config.ClusterCIDR,
		}, generator.config.GetFullLocalAssetFilename(utils.K8sKubeProxyConfig), true, false); _error != nil {
			return _error
		}
	}

	return nil
}

func (generator *Generator) generateKubeSchedulerConfig() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if _error := utils.ApplyTemplateAndSave("kube-scheduler-config", utils.TemplateKubeSchedulerConfiguration, struct {
			KubeConfig string
		}{
			KubeConfig: generator.config.GetFullTargetAssetFilename(utils.KubeconfigScheduler),
		}, generator.config.GetFullLocalAssetFilename(utils.K8sKubeSchedulerConfig), true, false); _error != nil {
			return _error
		}
	}

	return nil
}

func (generator *Generator) generateKubeletConfig() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := utils.ApplyTemplateAndSave("kubelet-configuration", utils.TemplateKubeletConfiguration, struct {
			CA                  string
			CertificateFilename string
			KeyFilename         string
			ClusterDomain       string
			ClusterDNSIP        string
			PODCIDR             string
			StaticPodPath       string
			ResolvConf          string
			MaxPods             uint16
		}{
			CA:                  generator.config.GetFullTargetAssetFilename(utils.PemCa),
			CertificateFilename: generator.config.GetFullTargetAssetFilename(utils.PemKubelet),
			KeyFilename:         generator.config.GetFullTargetAssetFilename(utils.PemKubeletKey),
			ClusterDomain:       generator.config.Config.ClusterDomain,
			ClusterDNSIP:        generator.config.Config.ClusterDNSIP,
			PODCIDR:             generator.config.Config.ClusterCIDR,
			StaticPodPath:       generator.config.GetFullTargetAssetDirectory(utils.DirectoryK8sManifests),
			ResolvConf:          generator.config.Config.ResolvConf,
			MaxPods:             generator.config.Config.MaxPods,
		}, generator.config.GetFullLocalAssetFilename(utils.K8sKubeletConfig), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateManifestGobetween() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := utils.ApplyTemplateAndSave("manifest-gobetween", utils.TemplateManifestGobetween, struct {
			GobetweenImage string
			Config         string
		}{
			GobetweenImage: generator.config.Config.Versions.Gobetween,
			Config:         generator.config.GetFullTargetAssetFilename(utils.GobetweenConfig),
		}, generator.config.GetFullLocalAssetFilename(utils.ManifestGobetween), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateManifestControllerVirtualIP() error {
	if len(generator.config.Config.ControllerVirtualIPInterface) == 0 || len(generator.config.Config.ControllerVirtualIP) == 0 {
		return nil
	}

	peers := ""

	for nodeName, node := range generator.config.Config.Nodes {
		if !node.IsController() {
			continue
		}
		if len(peers) > 0 {
			peers += ","
		}

		peers += fmt.Sprintf("%s=%s:%d", nodeName, node.IP, generator.config.Config.VIPRaftControllerPort)
	}

	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if !node.IsController() {
			continue
		}

		if error := utils.ApplyTemplateAndSave("manifest-controller-virtual-ip", utils.TemplateManifestVirtualIP, struct {
			VirtualIPImage string
			Type           string
			ID             string
			Bind           string
			VirtualIP      string
			Interface      string
			Peers          string
		}{
			VirtualIPImage: generator.config.Config.Versions.VirtualIP,
			Type:           "controller",
			ID:             nodeName,
			Bind:           fmt.Sprintf("%s:%d", node.IP, generator.config.Config.VIPRaftControllerPort),
			VirtualIP:      generator.config.Config.ControllerVirtualIP,
			Interface:      generator.config.Config.ControllerVirtualIPInterface,
			Peers:          peers,
		}, generator.config.GetFullLocalAssetFilename(utils.ManifestControllerVirtualIP), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateManifestWorkerVirtualIP() error {
	if len(generator.config.Config.WorkerVirtualIPInterface) == 0 || len(generator.config.Config.WorkerVirtualIP) == 0 {
		return nil
	}

	peers := ""

	for nodeName, node := range generator.config.Config.Nodes {
		if !node.IsWorker() {
			continue
		}
		if len(peers) > 0 {
			peers += ","
		}

		peers += fmt.Sprintf("%s=%s:%d", nodeName, node.IP, generator.config.Config.VIPRaftWorkerPort)
	}

	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if !node.IsWorker() {
			continue
		}

		if error := utils.ApplyTemplateAndSave("manifest-worker-virtual-ip", utils.TemplateManifestVirtualIP, struct {
			VirtualIPImage string
			Type           string
			ID             string
			Bind           string
			VirtualIP      string
			Interface      string
			Peers          string
		}{
			VirtualIPImage: generator.config.Config.Versions.VirtualIP,
			Type:           "worker",
			ID:             nodeName,
			Bind:           fmt.Sprintf("%s:%d", node.IP, generator.config.Config.VIPRaftWorkerPort),
			VirtualIP:      generator.config.Config.WorkerVirtualIP,
			Interface:      generator.config.Config.WorkerVirtualIPInterface,
			Peers:          peers,
		}, generator.config.GetFullLocalAssetFilename(utils.ManifestWorkerVirtualIP), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateManifestEtcd() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if !node.IsController() {
			continue
		}

		if error := utils.ApplyTemplateAndSave("manifest-etcd", utils.TemplateManifestEtcd, struct {
			EtcdImage         string
			Name              string
			PemCA             string
			PemKubernetes     string
			PemKubernetesKey  string
			NodeIP            string
			EtcdDataDirectory string
			EtcdCluster       string
		}{
			EtcdImage:         generator.config.Config.Versions.Etcd,
			Name:              nodeName,
			PemCA:             generator.config.GetFullTargetAssetFilename(utils.PemCa),
			PemKubernetes:     generator.config.GetFullTargetAssetFilename(utils.PemKubernetes),
			PemKubernetesKey:  generator.config.GetFullTargetAssetFilename(utils.PemKubernetesKey),
			NodeIP:            node.IP,
			EtcdDataDirectory: generator.config.GetFullTargetAssetDirectory(utils.DirectoryEtcdData),
			EtcdCluster:       generator.config.GetEtcdCluster(),
		}, generator.config.GetFullLocalAssetFilename(utils.ManifestEtcd), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateManifestKubeApiserver() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if !node.IsController() {
			continue
		}

		if error := utils.ApplyTemplateAndSave("manifest-kube-apiserver", utils.TemplateManifestKubeApiserver, struct {
			KubernetesImage   string
			ControllersCount  string
			AuditLog          string
			EtcdServers       string
			PemCA             string
			PemKubernetes     string
			PemKubernetesKey  string
			PemAggregator     string
			PemAggregatorKey  string
			PemServiceAccount string
			EncryptionConfig  string
			NodeIP            string
			APIServerPort     uint16
			ClusterIPRange    string
		}{
			KubernetesImage:   generator.config.Config.Versions.KubeAPIServer,
			ControllersCount:  generator.config.GetControllersCount(),
			AuditLog:          path.Join(generator.config.GetFullTargetAssetDirectory(utils.DirectoryLogging), utils.AuditLog),
			EtcdServers:       generator.config.GetEtcdServers(),
			PemCA:             generator.config.GetFullTargetAssetFilename(utils.PemCa),
			PemKubernetes:     generator.config.GetFullTargetAssetFilename(utils.PemKubernetes),
			PemKubernetesKey:  generator.config.GetFullTargetAssetFilename(utils.PemKubernetesKey),
			PemAggregator:     generator.config.GetFullTargetAssetFilename(utils.PemAggregator),
			PemAggregatorKey:  generator.config.GetFullTargetAssetFilename(utils.PemAggregatorKey),
			PemServiceAccount: generator.config.GetFullTargetAssetFilename(utils.PemServiceAccount),
			EncryptionConfig:  generator.config.GetFullTargetAssetFilename(utils.EncryptionConfig),
			NodeIP:            node.IP,
			APIServerPort:     generator.config.Config.APIServerPort,
			ClusterIPRange:    generator.config.Config.ClusterIPRange,
		}, generator.config.GetFullLocalAssetFilename(utils.ManifestKubeApiserver), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateManifestKubeControllerManager() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if !node.IsController() {
			continue
		}

		if error := utils.ApplyTemplateAndSave("manifest-kube-controller-manager", utils.TemplateManifestKubeControllerManager, struct {
			KubernetesImage      string
			ClusterCIDR          string
			ClusterIPRange       string
			PemCA                string
			PemCAKey             string
			Kubeconfig           string
			PemKubernetes        string
			PemKubernetesKey     string
			PemServiceAccountKey string
		}{
			KubernetesImage:      generator.config.Config.Versions.KubeControllerManager,
			ClusterCIDR:          generator.config.Config.ClusterCIDR,
			ClusterIPRange:       generator.config.Config.ClusterIPRange,
			PemCA:                generator.config.GetFullTargetAssetFilename(utils.PemCa),
			PemCAKey:             generator.config.GetFullTargetAssetFilename(utils.PemCaKey),
			Kubeconfig:           generator.config.GetFullTargetAssetFilename(utils.KubeconfigControllerManager),
			PemKubernetes:        generator.config.GetFullTargetAssetFilename(utils.PemKubernetes),
			PemKubernetesKey:     generator.config.GetFullTargetAssetFilename(utils.PemKubernetesKey),
			PemServiceAccountKey: generator.config.GetFullTargetAssetFilename(utils.PemServiceAccountKey),
		}, generator.config.GetFullLocalAssetFilename(utils.ManifestKubeControllerManager), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateManifestKubeScheduler() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if !node.IsController() {
			continue
		}

		if error := utils.ApplyTemplateAndSave("manifest-kube-scheduler", utils.TemplateManifestKubeScheduler, struct {
			KubernetesImage         string
			KubeSchedulerConfig     string
			KubeSchedulerKubeconfig string
		}{
			KubernetesImage:         generator.config.Config.Versions.KubeScheduler,
			KubeSchedulerConfig:     generator.config.GetFullTargetAssetFilename(utils.K8sKubeSchedulerConfig),
			KubeSchedulerKubeconfig: generator.config.GetFullTargetAssetFilename(utils.KubeconfigScheduler),
		}, generator.config.GetFullLocalAssetFilename(utils.ManifestKubeScheduler), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateManifestKubeProxy() error {
	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := utils.ApplyTemplateAndSave("manifest-kube-proxy", utils.TemplateManifestKubeProxy, struct {
			KubernetesImage     string
			ClusterCIDR         string
			KubeProxyKubeconfig string
			KubeProxyConfig     string
		}{
			KubernetesImage:     generator.config.Config.Versions.KubeProxy,
			ClusterCIDR:         generator.config.Config.ClusterCIDR,
			KubeProxyKubeconfig: generator.config.GetFullTargetAssetFilename(utils.KubeconfigProxy),
			KubeProxyConfig:     generator.config.GetFullTargetAssetFilename(utils.K8sKubeProxyConfig),
		}, generator.config.GetFullLocalAssetFilename(utils.ManifestKubeProxy), true, false); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateCertificates() error {
	var error error

	fullCAFilename := generator.config.GetFullLocalAssetFilename(utils.PemCa)
	fullCAKeyFilename := generator.config.GetFullLocalAssetFilename(utils.PemCaKey)

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
	kubernetesDNSNames := []string{"kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster.local", "localhost"}
	kubernetesIPAddresses := []string{"127.0.0.1", "10.32.0.1"}

	if len(generator.config.Config.ControllerVirtualIP) > 0 {
		kubernetesIPAddresses = append(kubernetesIPAddresses, generator.config.Config.ControllerVirtualIP)
	}

	for nodeName, node := range generator.config.Config.Nodes {
		kubernetesDNSNames = append(kubernetesDNSNames, nodeName)
		kubernetesIPAddresses = append(kubernetesIPAddresses, node.IP)
	}

	// Merge a string array with an array encoded as a comma separated string and return the new list
	mergeLists := func(oldList []string, values string) []string {
		newList := oldList[:]

		tokens := strings.Split(values, ",")

		for _, token := range tokens {
			token = strings.TrimSpace(token)

			if len(token) == 0 {
				continue
			}

			newList = append(newList, token)
		}

		return newList
	}

	apiServerDNSNames := mergeLists(kubernetesDNSNames[:], generator.config.Config.SANDNSNames)
	apiServerIPAddresses := mergeLists(kubernetesIPAddresses[:], generator.config.Config.SANIPAddresses)

	// Generate admin certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CnAdmin, "system:masters", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.PemAdmin), generator.config.GetFullLocalAssetFilename(utils.PemAdminKey), false); error != nil {
		return error
	}

	// Generate Kubernetes certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, "kubernetes", "Kubernetes", apiServerDNSNames, apiServerIPAddresses, generator.config.GetFullLocalAssetFilename(utils.PemKubernetes), generator.config.GetFullLocalAssetFilename(utils.PemKubernetesKey), true); error != nil {
		return error
	}

	// Generate aggregator certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CnAggregator, "Kubernetes", kubernetesDNSNames, kubernetesIPAddresses, generator.config.GetFullLocalAssetFilename(utils.PemAggregator), generator.config.GetFullLocalAssetFilename(utils.PemAggregatorKey), true); error != nil {
		return error
	}

	// Generate service accounts certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, "service-accounts", "Kubernetes", kubernetesDNSNames, kubernetesIPAddresses, generator.config.GetFullLocalAssetFilename(utils.PemServiceAccount), generator.config.GetFullLocalAssetFilename(utils.PemServiceAccountKey), false); error != nil {
		return error
	}

	// Generate controller manager certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CnSystemKubeControllerManager, "system:node-controller-manager", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.PemControllerManager), generator.config.GetFullLocalAssetFilename(utils.PemControllerManagerKey), false); error != nil {
		return error
	}

	// Generate scheduler certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CnSystemKubeScheduler, "system:kube-scheduler", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.PemScheduler), generator.config.GetFullLocalAssetFilename(utils.PemSchedulerKey), false); error != nil {
		return error
	}

	// Generate proxy certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CnSystemKubeProxy, "system:node-proxier", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.PemProxy), generator.config.GetFullLocalAssetFilename(utils.PemProxyKey), false); error != nil {
		return error
	}

	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, fmt.Sprintf(utils.CnSystemNodePrefix, nodeName), "system:nodes", []string{nodeName}, []string{node.IP}, generator.config.GetFullLocalAssetFilename(utils.PemKubelet), generator.config.GetFullLocalAssetFilename(utils.PemKubeletKey), false); error != nil {
			return error
		}
	}

	// Generate Elasticsearch certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CnElasticsearch, "elasticsearch", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.PemElasticsearch), generator.config.GetFullLocalAssetFilename(utils.PemElasticsearchKey), false); error != nil {
		return error
	}

	// Generate Minio certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CnMinio, "minio", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.PemMinio), generator.config.GetFullLocalAssetFilename(utils.PemMinioKey), false); error != nil {
		return error
	}

	// Generate Grafana certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CnGrafana, "grafana", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.PemGrafana), generator.config.GetFullLocalAssetFilename(utils.PemGrafanaKey), false); error != nil {
		return error
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

	return utils.ApplyTemplateAndSave("kubeconfig", utils.TemplateKubeconfig, struct {
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
	}, kubeConfigFilename, true, false)
}

func (generator *Generator) getAPIServerAddress(ip string) string {
	return fmt.Sprintf("%s:%d", ip, generator.config.Config.LoadBalancerPort)
}

func (generator *Generator) generateKubeConfigs() error {
	apiServer, error := generator.config.GetAPIServerIP()
	if error != nil {
		return error
	}

	apiServer = generator.getAPIServerAddress(apiServer)

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KubeconfigAdmin), generator.ca.CertificateFilename, "admin", apiServer, generator.config.GetFullLocalAssetFilename(utils.PemAdmin), generator.config.GetFullLocalAssetFilename(utils.PemAdminKey), true); error != nil {
		return error
	}

	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		apiServer = generator.getAPIServerAddress(node.IP)

		if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KubeconfigControllerManager), generator.ca.CertificateFilename, "system:kube-controller-manager", apiServer, generator.config.GetFullLocalAssetFilename(utils.PemControllerManager), generator.config.GetFullLocalAssetFilename(utils.PemControllerManagerKey), true); error != nil {
			return error
		}

		if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KubeconfigScheduler), generator.ca.CertificateFilename, "system:kube-scheduler", apiServer, generator.config.GetFullLocalAssetFilename(utils.PemScheduler), generator.config.GetFullLocalAssetFilename(utils.PemSchedulerKey), true); error != nil {
			return error
		}

		if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KubeconfigProxy), generator.ca.CertificateFilename, "system:kube-proxy", apiServer, generator.config.GetFullLocalAssetFilename(utils.PemProxy), generator.config.GetFullLocalAssetFilename(utils.PemProxyKey), true); error != nil {
			return error
		}

		if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KubeconfigKubelet), generator.ca.CertificateFilename, fmt.Sprintf("system:node:%s", nodeName), apiServer, generator.config.GetFullLocalAssetFilename(utils.PemKubelet), generator.config.GetFullLocalAssetFilename(utils.PemKubeletKey), true); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateCephSetup() error {
	return utils.ApplyTemplateAndSave("ceph-setup", utils.TemplateCephSetup, struct {
		CephRBDPoolName             string
		CephFSPoolName              string
		PublicNetwork               string
		StorageControllers          []config.NodeData
		StorageNodes                []config.NodeData
		CephConfigDirectory         string
		CephDataDirectory           string
		CephImage                   string
		CephManagerPort             uint16
		CephRadosGatewayPort        uint16
		K8sTewBinary                string
		K8sTewConfig                string
		CephManagerCredentials      string
		CephRadosGatewayCredentials string
	}{
		CephRBDPoolName:             utils.CephRbdPoolName,
		CephFSPoolName:              utils.CephFsPoolName,
		PublicNetwork:               generator.config.Config.PublicNetwork,
		StorageControllers:          generator.config.GetStorageControllers(),
		StorageNodes:                generator.config.GetStorageNodes(),
		CephConfigDirectory:         generator.config.GetFullTargetAssetDirectory(utils.DirectoryCephConfig),
		CephDataDirectory:           generator.config.GetFullTargetAssetDirectory(utils.DirectoryCephData),
		CephImage:                   generator.config.Config.Versions.Ceph,
		CephManagerPort:             utils.PortCephManager,
		CephRadosGatewayPort:        utils.PortCephRadosGateway,
		K8sTewBinary:                generator.config.GetFullTargetAssetFilename(utils.BinaryK8sTew),
		K8sTewConfig:                generator.config.GetFullTargetAssetFilename(utils.ConfigFilename),
		CephManagerCredentials:      utils.CephManagerCredentials,
		CephRadosGatewayCredentials: utils.CephRadosGatewayCredentials,
	}, generator.config.GetFullLocalAssetFilename(utils.CephSetup), true, false)
}

func (generator *Generator) generateCephCSI() error {
	return utils.ApplyTemplateAndSave("ceph-csi", utils.TemplateCephCsi, struct {
		ClusterID                  string
		KubeletDirectory           string
		PluginsDirectory           string
		PluginsRegistryDirectory   string
		PodsDirectory              string
		CephRBDPoolName            string
		CephFSPoolName             string
		StorageControllers         []config.NodeData
		CSIAttacherImage           string
		CSIProvisionerImage        string
		CSIDriverRegistrarImage    string
		CSISnapshotterImage        string
		CSISnapshotControllerImage string
		CSIResizerImage            string
		CSICephPluginImage         string
	}{
		ClusterID:                  generator.config.Config.ClusterID,
		KubeletDirectory:           generator.config.GetFullTargetAssetDirectory(utils.DirectoryKubeletData),
		PluginsDirectory:           generator.config.GetFullTargetAssetDirectory(utils.DirectoryKubeletPlugins),
		PluginsRegistryDirectory:   generator.config.GetFullTargetAssetDirectory(utils.DirectoryKubeletPluginsRegistry),
		PodsDirectory:              generator.config.GetFullTargetAssetDirectory(utils.DirectoryPodsData),
		CephRBDPoolName:            utils.CephRbdPoolName,
		CephFSPoolName:             utils.CephFsPoolName,
		StorageControllers:         generator.config.GetStorageControllers(),
		CSIAttacherImage:           generator.config.Config.Versions.CSIAttacher,
		CSIProvisionerImage:        generator.config.Config.Versions.CSIProvisioner,
		CSIDriverRegistrarImage:    generator.config.Config.Versions.CSIDriverRegistrar,
		CSISnapshotterImage:        generator.config.Config.Versions.CSISnapshotter,
		CSISnapshotControllerImage: generator.config.Config.Versions.CSISnapshotController,
		CSIResizerImage:            generator.config.Config.Versions.CSIResizer,
		CSICephPluginImage:         generator.config.Config.Versions.CSICephPlugin,
	}, generator.config.GetFullLocalAssetFilename(utils.CephCsi), true, false)
}

func (generator *Generator) generateCephFiles() error {
	ceph := ceph.NewCeph(generator.config, ceph.CephBinariesPath, ceph.CephConfigPath, ceph.CephDataPath)

	cephData, _error := ceph.Setup()
	if _error != nil {
		return _error
	}

	return utils.ApplyTemplateAndSave("ceph-secrets", utils.TemplateCephSecrets, cephData, generator.config.GetFullLocalAssetFilename(utils.CephSecrets), true, false)
}

func (generator *Generator) generateLetsEncryptClusterIssuer() error {
	return utils.ApplyTemplateAndSave("lets-encrypt-cluster-issuer", utils.TemplateLetsencryptClusterIssuerSetup, struct {
		Email string
	}{
		Email: generator.config.Config.Email,
	}, generator.config.GetFullLocalAssetFilename(utils.LetsencryptClusterIssuer), true, false)
}

func (generator *Generator) generateCoreDNSSetup() error {
	return utils.ApplyTemplateAndSave("core-dns", utils.TemplateCorednsSetup, struct {
		ClusterDomain string
		ClusterDNSIP  string
		CoreDNSImage  string
	}{
		ClusterDomain: generator.config.Config.ClusterDomain,
		ClusterDNSIP:  generator.config.Config.ClusterDNSIP,
		CoreDNSImage:  generator.config.Config.Versions.CoreDNS,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sCorednsSetup), true, false)
}

func (generator *Generator) generateElasticsearchCredentials() error {
	elasticsearchPassword, error := generator.generatePassword()
	if error != nil {
		return error
	}

	return utils.ApplyTemplateAndSave(utils.ElasticsearchCredentials, utils.TemplateCredentials, struct {
		Namespace  string
		SecretName string
		Data       map[string]string
	}{
		Namespace:  utils.FeatureLogging,
		SecretName: utils.ElasticsearchCredentials,
		Data:       map[string]string{utils.KeyUsername: utils.Username, utils.KeyPassword: elasticsearchPassword},
	}, generator.config.GetFullLocalAssetFilename(utils.K8sElasticsearchCredentials), false, false)
}

func (generator *Generator) generateEFKSetup() error {
	counts := []string{}

	for i := uint16(0); i < generator.config.Config.ElasticsearchCount; i++ {
		counts = append(counts, fmt.Sprintf("%d", i))
	}

	return utils.ApplyTemplateAndSave("efk", utils.TemplateEfkSetup, struct {
		ElasticsearchImage  string
		KibanaImage         string
		CerebroImage        string
		FluentBitImage      string
		BusyboxImage        string
		KibanaPort          string
		CerebroPort         string
		ElasticsearchSize   uint16
		ElasticsearchCount  uint16
		ElasticsearchCounts []string
	}{
		ElasticsearchImage:  generator.config.Config.Versions.Elasticsearch,
		KibanaImage:         generator.config.Config.Versions.Kibana,
		CerebroImage:        generator.config.Config.Versions.Cerebro,
		FluentBitImage:      generator.config.Config.Versions.FluentBit,
		BusyboxImage:        generator.config.Config.Versions.Busybox,
		KibanaPort:          fmt.Sprintf("%d", utils.PortKibana),
		CerebroPort:         fmt.Sprintf("%d", utils.PortCerebro),
		ElasticsearchSize:   generator.config.Config.ElasticsearchSize,
		ElasticsearchCount:  generator.config.Config.ElasticsearchCount,
		ElasticsearchCounts: counts,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sEfkSetup), true, false)
}

func (generator *Generator) generateMinioCredentials() error {
	password, error := generator.generatePassword()
	if error != nil {
		return error
	}

	return utils.ApplyTemplateAndSave(utils.MinioCredentials, utils.TemplateCredentials, struct {
		Namespace  string
		SecretName string
		Data       map[string]string
	}{
		Namespace:  utils.FeatureBackup,
		SecretName: utils.MinioCredentials,
		Data:       map[string]string{utils.KeyUsername: utils.Username, utils.KeyPassword: password},
	}, generator.config.GetFullLocalAssetFilename(utils.K8sMinioCredentials), false, false)
}

func (generator *Generator) generateCerebroCredentials() error {
	cerebroPassword, error := generator.generatePassword()
	if error != nil {
		return error
	}

	secret, error := password.Generate(32, 8, 0, false, true)
	if error != nil {
		return errors.Wrap(error, "Could not generate secret for Cerebro")
	}

	return utils.ApplyTemplateAndSave(utils.CerebroCredentials, utils.TemplateCredentials, struct {
		Namespace  string
		SecretName string
		Data       map[string]string
	}{
		Namespace:  utils.FeatureLogging,
		SecretName: utils.CerebroCredentials,
		Data:       map[string]string{utils.KeyUsername: utils.Username, utils.KeyPassword: cerebroPassword, utils.KeySecret: secret},
	}, generator.config.GetFullLocalAssetFilename(utils.K8sCerebroCredentials), false, false)
}

func (generator *Generator) generateCephManagerCredentials() error {
	password, error := generator.generatePassword()
	if error != nil {
		return error
	}

	return utils.ApplyTemplateAndSave(utils.CephManagerCredentials, utils.TemplateCredentials, struct {
		Namespace  string
		SecretName string
		Data       map[string]string
	}{
		Namespace:  utils.FeatureStorage,
		SecretName: utils.CephManagerCredentials,
		Data:       map[string]string{utils.KeyUsername: utils.Username, utils.KeyPassword: password},
	}, generator.config.GetFullLocalAssetFilename(utils.K8sCephManagerCredentials), false, false)
}

func (generator *Generator) generateCephRadosGatewayCredentials() error {
	accessKey, error := password.Generate(20, 6, 0, false, true)
	if error != nil {
		return errors.Wrap(error, "Could not generate access key")
	}

	secretKey, error := password.Generate(40, 8, 0, false, true)
	if error != nil {
		return errors.Wrap(error, "Could not generate secret key")
	}

	return utils.ApplyTemplateAndSave(utils.CephRadosGatewayCredentials, utils.TemplateCredentials, struct {
		Namespace  string
		SecretName string
		Data       map[string]string
	}{
		Namespace:  utils.FeatureStorage,
		SecretName: utils.CephRadosGatewayCredentials,
		Data:       map[string]string{utils.KeyUsername: strings.ToUpper(accessKey), utils.KeyPassword: secretKey},
	}, generator.config.GetFullLocalAssetFilename(utils.K8sCephRadosGatewayCredentials), false, false)
}

func (generator *Generator) generateVeleroSetup() error {
	return utils.ApplyTemplateAndSave("velero-setup", utils.TemplateVeleroSetup, struct {
		VeleroImage          string
		VeleroPluginAWSImage string
		MinioServerImage     string
		MinioClientImage     string
		PodsDirectory        string
		MinioPort            uint16
		MinioSize            uint16
	}{
		VeleroImage:          generator.config.Config.Versions.Velero,
		VeleroPluginAWSImage: generator.config.Config.Versions.VeleroPluginAWS,
		MinioServerImage:     generator.config.Config.Versions.MinioServer,
		MinioClientImage:     generator.config.Config.Versions.MinioClient,
		PodsDirectory:        generator.config.GetFullTargetAssetDirectory(utils.DirectoryPodsData),
		MinioPort:            utils.PortMinio,
		MinioSize:            generator.config.Config.MinioSize,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sVeleroSetup), true, false)
}

func (generator *Generator) generateKubernetesDashboardSetup() error {
	return utils.ApplyTemplateAndSave("kubernetes-dashboard", utils.TemplateKubernetesDashboardSetup, struct {
		ClusterName              string
		KubernetesDashboardPort  uint16
		KubernetesDashboardImage string
		MetricsScraperImage      string
	}{
		ClusterName:              generator.config.Config.ClusterName,
		KubernetesDashboardPort:  generator.config.Config.KubernetesDashboardPort,
		KubernetesDashboardImage: generator.config.Config.Versions.KubernetesDashboard,
		MetricsScraperImage:      generator.config.Config.Versions.MetricsScraper,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sKubernetesDashboardSetup), true, false)
}

func (generator *Generator) generateCertManagerSetup() error {
	return utils.ApplyTemplateAndSave("cert-manager", utils.TemplateCertManagerSetup, struct {
		CertManagerControllerImage string
		CertManagerCAInjectorImage string
		CertManagerWebHookImage    string
	}{
		CertManagerControllerImage: generator.config.Config.Versions.CertManagerController,
		CertManagerCAInjectorImage: generator.config.Config.Versions.CertManagerCAInjector,
		CertManagerWebHookImage:    generator.config.Config.Versions.CertManagerWebHook,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sCertManagerSetup), true, false)
}

func (generator *Generator) generateNginxIngressSetup() error {
	return utils.ApplyTemplateAndSave("nginx-ingress", utils.TemplateNginxIngressSetup, struct {
		NginxIngressControllerImage     string
		NginxIngressDefaultBackendImage string
	}{
		NginxIngressControllerImage:     generator.config.Config.Versions.NginxIngressController,
		NginxIngressDefaultBackendImage: generator.config.Config.Versions.NginxIngressDefaultBackend,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sNginxIngressSetup), true, false)
}

func (generator *Generator) generateMetricsServerSetup() error {
	return utils.ApplyTemplateAndSave("metrics-server", utils.TemplateMetricsServerSetup, struct {
		MetricsServerImage string
	}{
		MetricsServerImage: generator.config.Config.Versions.MetricsServer,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sMetricsServerSetup), true, false)
}

func (generator *Generator) generatePrometheusSetup() error {
	return utils.ApplyTemplateAndSave("prometheus", utils.TemplatePrometheusSetup, struct {
		PrometheusImage string
		PrometheusSize  uint16
		BusyboxImage    string
	}{
		PrometheusImage: generator.config.Config.Versions.Prometheus,
		PrometheusSize:  generator.config.Config.PrometheusSize,
		BusyboxImage:    generator.config.Config.Versions.Busybox,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sPrometheusSetup), true, true)
}

func (generator *Generator) generateNodeExporterSetup() error {
	return utils.ApplyTemplateAndSave("node-exporter", utils.TemplateNodeExporterSetup, struct {
		NodeExporterImage string
	}{
		NodeExporterImage: generator.config.Config.Versions.NodeExporter,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sNodeExporterSetup), true, false)
}

func (generator *Generator) generateKubeStateMetricsSetup() error {
	return utils.ApplyTemplateAndSave("kube-state-metrics", utils.TemplateKubeStateMetricsSetup, struct {
		KubeStateMetricsImage string
		KubeStateMetricsCount uint16
	}{
		KubeStateMetricsImage: generator.config.Config.Versions.KubeStateMetrics,
		KubeStateMetricsCount: generator.config.Config.KubeStateMetricsCount,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sKubeStateMetricsSetup), true, false)
}

func (generator *Generator) generateGrafanaConfigMap() error {
	ca, error := utils.ReadFile(generator.config.GetFullLocalAssetFilename(utils.PemCa))
	if error != nil {
		return error
	}

	grafana, error := utils.ReadFile(generator.config.GetFullLocalAssetFilename(utils.PemGrafana))
	if error != nil {
		return error
	}

	grafanaKey, error := utils.ReadFile(generator.config.GetFullLocalAssetFilename(utils.PemGrafanaKey))
	if error != nil {
		return error
	}

	data := map[string]string{"ca.pem": ca, "grafana.pem": grafana, "grafana-key.pem": grafanaKey}

	return utils.ApplyTemplateAndSave(utils.GrafanaCertificates, utils.TemplateConfigMap, struct {
		Namespace string
		Name      string
		Data      map[string]string
	}{
		Namespace: utils.FeatureMonitoring,
		Name:      utils.GrafanaCertificates,
		Data:      data,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sGrafanaCertificates), false, false)
}

func (generator *Generator) generateMinioConfigMap() error {
	ca, error := utils.ReadFile(generator.config.GetFullLocalAssetFilename(utils.PemCa))
	if error != nil {
		return error
	}

	minio, error := utils.ReadFile(generator.config.GetFullLocalAssetFilename(utils.PemMinio))
	if error != nil {
		return error
	}

	minioKey, error := utils.ReadFile(generator.config.GetFullLocalAssetFilename(utils.PemMinioKey))
	if error != nil {
		return error
	}

	data := map[string]string{"ca.pem": ca, "minio.pem": minio, "minio-key.pem": minioKey}

	return utils.ApplyTemplateAndSave(utils.MinioCertificates, utils.TemplateConfigMap, struct {
		Namespace string
		Name      string
		Data      map[string]string
	}{
		Namespace: utils.FeatureBackup,
		Name:      utils.MinioCertificates,
		Data:      data,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sMinioCertificates), false, false)
}

func (generator *Generator) generateElasticsearchConfigMap() error {
	ca, error := utils.ReadFile(generator.config.GetFullLocalAssetFilename(utils.PemCa))
	if error != nil {
		return error
	}

	elasticsearch, error := utils.ReadFile(generator.config.GetFullLocalAssetFilename(utils.PemElasticsearch))
	if error != nil {
		return error
	}

	elasticsearchKey, error := utils.ReadFile(generator.config.GetFullLocalAssetFilename(utils.PemElasticsearchKey))
	if error != nil {
		return error
	}

	data := map[string]string{"ca.pem": ca, "elasticsearch.pem": elasticsearch, "elasticsearch-key.pem": elasticsearchKey}

	return utils.ApplyTemplateAndSave(utils.ElasticsearchCertificates, utils.TemplateConfigMap, struct {
		Namespace string
		Name      string
		Data      map[string]string
	}{
		Namespace: utils.FeatureLogging,
		Name:      utils.ElasticsearchCertificates,
		Data:      data,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sElasticsearchCertificates), false, false)
}

func (generator *Generator) generateGrafanaCredentials() error {
	password, error := generator.generatePassword()
	if error != nil {
		return error
	}

	return utils.ApplyTemplateAndSave(utils.GrafanaCredentials, utils.TemplateCredentials, struct {
		Namespace  string
		SecretName string
		Data       map[string]string
	}{
		Namespace:  utils.FeatureMonitoring,
		SecretName: utils.GrafanaCredentials,
		Data:       map[string]string{utils.KeyUsername: utils.Username, utils.KeyPassword: password},
	}, generator.config.GetFullLocalAssetFilename(utils.K8sGrafanaCredentials), false, false)
}

func (generator *Generator) generateGrafanaSetup() error {
	return utils.ApplyTemplateAndSave("grafana", utils.TemplateGrafanaSetup, struct {
		GrafanaImage string
		GrafanaPort  uint16
		GrafanaSize  uint16
		BusyboxImage string
	}{
		GrafanaImage: generator.config.Config.Versions.Grafana,
		GrafanaPort:  utils.PortGrafana,
		GrafanaSize:  generator.config.Config.GrafanaSize,
		BusyboxImage: generator.config.Config.Versions.Busybox,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sGrafanaSetup), true, true)
}

func (generator *Generator) generateAlertManagerSetup() error {
	counts := []string{}

	for i := uint16(0); i < generator.config.Config.AlertManagerCount; i++ {
		counts = append(counts, fmt.Sprintf("%d", i))
	}

	return utils.ApplyTemplateAndSave("alert-manager", utils.TemplateAlertManagerSetup, struct {
		AlertManagerImage  string
		AlertManagerCount  uint16
		AlertManagerCounts []string
		AlertManagerSize   uint16
		BusyboxImage       string
	}{
		AlertManagerImage:  generator.config.Config.Versions.AlertManager,
		AlertManagerCount:  generator.config.Config.AlertManagerCount,
		AlertManagerCounts: counts,
		AlertManagerSize:   generator.config.Config.AlertManagerSize,
		BusyboxImage:       generator.config.Config.Versions.Busybox,
	}, generator.config.GetFullLocalAssetFilename(utils.K8sAlertManagerSetup), true, false)
}

func (generator *Generator) generateWordpressSetup() error {
	return utils.ApplyTemplateAndSave("wordpress", utils.TemplateWordpressSetup, struct {
		WordPressIngressDomain string
		MySQLImage             string
		WordPressImage         string
		WordPressPort          uint16
	}{
		WordPressIngressDomain: fmt.Sprintf("%s.%s", utils.IngressSubdomainWordpress, generator.config.Config.IngressDomain),
		MySQLImage:             generator.config.Config.Versions.MySQL,
		WordPressImage:         generator.config.Config.Versions.WordPress,
		WordPressPort:          utils.PortWordpress,
	}, generator.config.GetFullLocalAssetFilename(utils.WordpressSetup), true, false)
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

func (generator *Generator) generatePassword() (string, error) {
	result, error := password.Generate(12, 6, 0, false, true)
	if error != nil {
		return result, errors.Wrap(error, "Could not generate secret")
	}

	return result, nil
}
