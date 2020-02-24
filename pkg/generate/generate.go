package generate

import (
	"fmt"
	"path"
	"strings"

	"github.com/darxkies/k8s-tew/pkg/config"

	"github.com/darxkies/k8s-tew/pkg/pki"
	"github.com/darxkies/k8s-tew/pkg/utils"
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
		// Generate metallb setup
		generator.generateMetalLBSetup,
		// Generate scheduler config
		generator.generateKubeSchedulerConfig,
		// Generate kubelet config
		generator.generateKubeletConfig,
		// Generate kubelet configuration
		generator.generateK8SKubeletConfigFile,
		// Generate dashboard admin user configuration
		generator.generateK8SAdminUserConfigFile,
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
		// Generate ElasticSearch/Fluent-Bit/Kibana setup file
		generator.generateEFKSetup,
		// Generate velero setup file
		generator.generateVeleroSetup,
		// Generate kubernetes dashboard setup file
		generator.generateKubernetesDashboardSetup,
		// Generate cert-manager setup file
		generator.generateCertManagerSetup,
		// Generate nginx ingress setup file
		generator.generateNginxIngressSetup,
		// Generate metrics server setup file
		generator.generateMetricsServerSetup,
		// Generate prometheus setup file
		generator.generatePrometheusSetup,
		// Generate kube state metrics setup file
		generator.generateKubeStateMetricsSetup,
		// Generate node exporter setup file
		generator.generateNodeExporterSetup,
		// Generate grafana setup file
		generator.generateGrafanaSetup,
		// Generate alert manager setup file
		generator.generateAlertManagerSetup,
		// Generate wordpress setup file
		generator.generateWordpressSetup,
		// Generate Bash Completion for K8S-TEW
		generator.generateBashCompletionK8STEW,
		// Generate Bash Completion for Kubectl
		generator.generateBashCompletionKubectl,
		// Generate Bash Completion for Helm
		generator.generateBashCompletionHelm,
		// Generate Bash Completion for Velero
		generator.generateBashCompletionVelero,
		// Generate Bash Completion for CriCtl
		generator.generateBashCompletionCriCtl,
		// Generate gobetween manifest
		generator.generateManifestGobetween,
		// Generate controller virtual-ip manifest
		generator.generateManifestControllerVirtualIP,
		// Generate worker virtual-ip manifest
		generator.generateManifestWorkerVirtualIP,
		// Generate etcd manifest
		generator.generateManifestEtcd,
		// Generate kube-apiserver manifest
		generator.generateManifestKubeApiserver,
		// Generate kube-controller-manager manifest
		generator.generateManifestKubeControllerManager,
		// Generate kube-scheduler manifest
		generator.generateManifestKubeScheduler,
		// Generate kube-proxy manifest
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
		CalicoTyphaImage           string
		CalicoNodeImage            string
		CalicoCNIImage             string
		CalicoKubeControllersImage string
	}{
		CalicoTyphaIP:              generator.config.Config.CalicoTyphaIP,
		ClusterCIDR:                generator.config.Config.ClusterCIDR,
		CNIConfigDirectory:         generator.config.GetFullTargetAssetDirectory(utils.DirectoryCniConfig),
		CNIBinariesDirectory:       generator.config.GetFullTargetAssetDirectory(utils.DirectoryCniBinaries),
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
		Name:      "admin-user",
		Namespace: "kube-system",
	}, generator.config.GetFullLocalAssetFilename(utils.K8sAdminUserSetup), true, false)
}

func (generator *Generator) generateEncryptionFile() error {
	fullEncryptionConfigFilename := generator.config.GetFullLocalAssetFilename(utils.EncryptionConfig)

	if utils.FileExists(fullEncryptionConfigFilename) {
		utils.LogFilename("skipped", fullEncryptionConfigFilename)

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

func (generator *Generator) generateKubeSchedulerConfig() error {
	return utils.ApplyTemplateAndSave("kube-scheduler-config", utils.TemplateKubeSchedulerConfiguration, struct {
		KubeConfig string
	}{
		KubeConfig: generator.config.GetFullTargetAssetFilename(utils.KubeconfigScheduler),
	}, generator.config.GetFullLocalAssetFilename(utils.K8sKubeSchedulerConfig), true, false)
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

		if !node.IsController() {
			continue
		}

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
			KubernetesImage:   generator.config.Config.Versions.K8S,
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
			KubernetesImage:      generator.config.Config.Versions.K8S,
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
			KubernetesImage:         generator.config.Config.Versions.K8S,
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
		}{
			KubernetesImage:     generator.config.Config.Versions.K8S,
			ClusterCIDR:         generator.config.Config.ClusterCIDR,
			KubeProxyKubeconfig: generator.config.GetFullTargetAssetFilename(utils.KubeconfigProxy),
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

	// Generate admin certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, utils.CnAdmin, "system:masters", []string{}, []string{}, generator.config.GetFullLocalAssetFilename(utils.PemAdmin), generator.config.GetFullLocalAssetFilename(utils.PemAdminKey), false); error != nil {
		return error
	}

	// Generate kuberentes certificate
	if error := pki.GenerateClient(generator.ca, generator.config.Config.RSASize, generator.config.Config.ClientValidityPeriod, "kubernetes", "Kubernetes", kubernetesDNSNames, kubernetesIPAddresses, generator.config.GetFullLocalAssetFilename(utils.PemKubernetes), generator.config.GetFullLocalAssetFilename(utils.PemKubernetesKey), true); error != nil {
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

func (generator *Generator) generateKubeConfigs() error {
	apiServer, error := generator.config.GetAPIServerIP()
	if error != nil {
		return error
	}

	apiServer = fmt.Sprintf("%s:%d", apiServer, generator.config.Config.LoadBalancerPort)

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KubeconfigAdmin), generator.ca.CertificateFilename, "admin", apiServer, generator.config.GetFullLocalAssetFilename(utils.PemAdmin), generator.config.GetFullLocalAssetFilename(utils.PemAdminKey), true); error != nil {
		return error
	}

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KubeconfigControllerManager), generator.ca.CertificateFilename, "system:kube-controller-manager", apiServer, generator.config.GetFullLocalAssetFilename(utils.PemControllerManager), generator.config.GetFullLocalAssetFilename(utils.PemControllerManagerKey), true); error != nil {
		return error
	}

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KubeconfigScheduler), generator.ca.CertificateFilename, "system:kube-scheduler", apiServer, generator.config.GetFullLocalAssetFilename(utils.PemScheduler), generator.config.GetFullLocalAssetFilename(utils.PemSchedulerKey), true); error != nil {
		return error
	}

	if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KubeconfigProxy), generator.ca.CertificateFilename, "system:kube-proxy", apiServer, generator.config.GetFullLocalAssetFilename(utils.PemProxy), generator.config.GetFullLocalAssetFilename(utils.PemProxyKey), true); error != nil {
		return error
	}

	for nodeName, node := range generator.config.Config.Nodes {
		generator.config.SetNode(nodeName, node)

		if error := generator.generateConfigKubeConfig(generator.config.GetFullLocalAssetFilename(utils.KubeconfigKubelet), generator.ca.CertificateFilename, fmt.Sprintf("system:node:%s", nodeName), apiServer, generator.config.GetFullLocalAssetFilename(utils.PemKubelet), generator.config.GetFullLocalAssetFilename(utils.PemKubeletKey), true); error != nil {
			return error
		}
	}

	return nil
}

func (generator *Generator) generateCephConfig() error {
	return utils.ApplyTemplateAndSave("ceph-config", utils.TemplateCephConfig, struct {
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
	}, generator.config.GetFullLocalAssetFilename(utils.CephConfig), true, false)
}

func (generator *Generator) generateCephSetup() error {
	return utils.ApplyTemplateAndSave("ceph-setup", utils.TemplateCephSetup, struct {
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
		CephRBDPoolName:      utils.CephRbdPoolName,
		CephFSPoolName:       utils.CephFsPoolName,
		PublicNetwork:        generator.config.Config.PublicNetwork,
		StorageControllers:   generator.config.GetStorageControllers(),
		StorageNodes:         generator.config.GetStorageNodes(),
		CephConfigDirectory:  generator.config.GetFullTargetAssetDirectory(utils.DirectoryCephConfig),
		CephDataDirectory:    generator.config.GetFullTargetAssetDirectory(utils.DirectoryCephData),
		CephImage:            generator.config.Config.Versions.Ceph,
		CephManagerPort:      utils.PortCephManager,
		CephRadosGatewayPort: utils.PortCephRadosGateway,
	}, generator.config.GetFullLocalAssetFilename(utils.CephSetup), true, false)
}

func (generator *Generator) generateCephCSI() error {
	return utils.ApplyTemplateAndSave("ceph-csi", utils.TemplateCephCsi, struct {
		ClusterID                string
		KubeletDirectory         string
		PluginsDirectory         string
		PluginsRegistryDirectory string
		PodsDirectory            string
		CephRBDPoolName          string
		CephFSPoolName           string
		StorageControllers       []config.NodeData
		CSIAttacherImage         string
		CSIProvisionerImage      string
		CSIDriverRegistrarImage  string
		CSISnapshotterImage      string
		CSIResizerImage          string
		CSICephPluginImage       string
	}{
		ClusterID:                generator.config.Config.ClusterID,
		KubeletDirectory:         generator.config.GetFullTargetAssetDirectory(utils.DirectoryKubeletData),
		PluginsDirectory:         generator.config.GetFullTargetAssetDirectory(utils.DirectoryKubeletPlugins),
		PluginsRegistryDirectory: generator.config.GetFullTargetAssetDirectory(utils.DirectoryKubeletPluginsRegistry),
		PodsDirectory:            generator.config.GetFullTargetAssetDirectory(utils.DirectoryPodsData),
		CephRBDPoolName:          utils.CephRbdPoolName,
		CephFSPoolName:           utils.CephFsPoolName,
		StorageControllers:       generator.config.GetStorageControllers(),
		CSIAttacherImage:         generator.config.Config.Versions.CSIAttacher,
		CSIProvisionerImage:      generator.config.Config.Versions.CSIProvisioner,
		CSIDriverRegistrarImage:  generator.config.Config.Versions.CSIDriverRegistrar,
		CSISnapshotterImage:      generator.config.Config.Versions.CSISnapshotter,
		CSIResizerImage:          generator.config.Config.Versions.CSIResizer,
		CSICephPluginImage:       generator.config.Config.Versions.CSICephPlugin,
	}, generator.config.GetFullLocalAssetFilename(utils.CephCsi), true, false)
}

func (generator *Generator) generateCephFiles() error {
	if utils.FileExists(generator.config.GetFullLocalAssetFilename(utils.CephMonitorKeyring)) {
		return nil
	}

	monitorKey := utils.GenerateCephKey()
	clientAdminKey := utils.GenerateCephKey()
	clientBootstrapMetadataServerKey := utils.GenerateCephKey()
	clientBootstrapObjectStorageKey := utils.GenerateCephKey()
	clientBootstrapRadosBlockDeviceKey := utils.GenerateCephKey()
	clientBootstrapRadosGatewayKey := utils.GenerateCephKey()
	clientK8STEWKey := utils.GenerateCephKey()

	if error := utils.ApplyTemplateAndSave("ceph-monitor-keyring", utils.TemplateCephMonitorKeyring, struct {
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
		CephPoolName:                       utils.CephRbdPoolName,
	}, generator.config.GetFullLocalAssetFilename(utils.CephMonitorKeyring), false, false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("ceph-client-admin", utils.TemplateCephClientAdminKeyring, struct {
		Key string
	}{
		Key: clientAdminKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CephClientAdminKeyring), false, false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-mds-client-keyring", utils.TemplateCephClientKeyring, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-mds",
		Key:  clientBootstrapMetadataServerKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CephBootstrapMdsKeyring), false, false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-osd-client-keyring", utils.TemplateCephClientKeyring, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-osd",
		Key:  clientBootstrapObjectStorageKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CephBootstrapOsdKeyring), false, false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-rbd-client-keyring", utils.TemplateCephClientKeyring, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-rbd",
		Key:  clientBootstrapRadosBlockDeviceKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CephBootstrapRbdKeyring), false, false); error != nil {
		return error
	}

	if error := utils.ApplyTemplateAndSave("ceph-bootstrap-rgw-client-keyring", utils.TemplateCephClientKeyring, struct {
		Name string
		Key  string
	}{
		Name: "bootstrap-rgw",
		Key:  clientBootstrapRadosGatewayKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CephBootstrapRgwKeyring), false, false); error != nil {
		return error
	}

	return utils.ApplyTemplateAndSave("ceph-secrets", utils.TemplateCephSecrets, struct {
		ClientAdminKey  string
		ClientK8STEWKey string
	}{
		ClientAdminKey:  clientAdminKey,
		ClientK8STEWKey: clientK8STEWKey,
	}, generator.config.GetFullLocalAssetFilename(utils.CephSecrets), false, false)
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

func (generator *Generator) generateBashCompletion(binaryName, bashCompletionFilename string) error {
	binary := generator.config.GetFullLocalAssetFilename(binaryName)
	bashCompletionFullFilename := generator.config.GetFullLocalAssetFilename(bashCompletionFilename)

	command := fmt.Sprintf("%s completion bash > %s", binary, bashCompletionFullFilename)

	log.WithFields(log.Fields{"name": bashCompletionFilename}).Info("Generated")

	return utils.RunCommand(command)
}

func (generator *Generator) generateBashCompletionK8STEW() error {
	return generator.generateBashCompletion(utils.BinaryK8sTew, utils.BashCompletionK8sTew)
}

func (generator *Generator) generateBashCompletionKubectl() error {
	return generator.generateBashCompletion(utils.BinaryKubectl, utils.BashCompletionKubectl)
}

func (generator *Generator) generateBashCompletionHelm() error {
	return generator.generateBashCompletion(utils.BinaryHelm, utils.BashCompletionHelm)
}

func (generator *Generator) generateBashCompletionVelero() error {
	return generator.generateBashCompletion(utils.BinaryVelero, utils.BashCompletionVelero)
}

func (generator *Generator) generateBashCompletionCriCtl() error {
	return generator.generateBashCompletion(utils.BinaryCrictl, utils.BashCompletionCrictl)
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
