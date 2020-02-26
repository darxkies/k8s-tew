package config

import (
	"github.com/darxkies/k8s-tew/pkg/utils"
)

type Versions struct {
	Etcd                       string `yaml:"etcd"`
	K8S                        string `yaml:"kubernetes"`
	Helm                       string `yaml:"helm"`
	Containerd                 string `yaml:"containerd"`
	Runc                       string `yaml:"runc"`
	CriCtl                     string `yaml:"crictl"`
	Gobetween                  string `yaml:"gobetween"`
	VirtualIP                  string `yaml:"virtual-ip"`
	Busybox                    string `yaml:"busybox"`
	Velero                     string `yaml:"velero"`
	VeleroPluginAWS            string `yaml:"velero-plugin-aws"`
	MinioServer                string `yaml:"minio-server"`
	MinioClient                string `yaml:"minio-client"`
	Pause                      string `yaml:"pause"`
	CoreDNS                    string `yaml:"core-dns"`
	Elasticsearch              string `yaml:"elasticsearch"`
	Kibana                     string `yaml:"kibana"`
	Cerebro                    string `yaml:"cerebro"`
	FluentBit                  string `yaml:"fluent-bit"`
	CalicoTypha                string `yaml:"calico-typha"`
	CalicoNode                 string `yaml:"calico-node"`
	CalicoCNI                  string `yaml:"calico-cni"`
	CalicoKubeControllers      string `yaml:"calico-kube-controllers"`
	MetalLBController          string `yaml:"metallb-controller"`
	MetalLBSpeaker             string `yaml:"metallb-speaker"`
	Ceph                       string `yaml:"ceph"`
	KubernetesDashboard        string `yaml:"kubernetes-dashboard"`
	CertManagerController      string `yaml:"cert-manager-controller"`
	CertManagerCAInjector      string `yaml:"cert-manager-cainjector"`
	CertManagerWebHook         string `yaml:"cert-manager-webhook"`
	CertManagerACMEResolver    string `yaml:"cert-manager-acmeresolver"`
	NginxIngressController     string `yaml:"nginx-ingress-controller"`
	NginxIngressDefaultBackend string `yaml:"nginx-ingress-default-backend"`
	MetricsScraper             string `yaml:"metrics-scraper"`
	MetricsServer              string `yaml:"metrics-server"`
	ConfigMapReload            string `yaml:"configmap-reload"`
	KubeStateMetrics           string `yaml:"kube-state-metrics"`
	Grafana                    string `yaml:"grafana"`
	Prometheus                 string `yaml:"prometheus"`
	NodeExporter               string `yaml:"node-exporter"`
	AlertManager               string `yaml:"alert-manager"`
	CSIAttacher                string `yaml:"csi-attacher"`
	CSIProvisioner             string `yaml:"csi-provisioner"`
	CSIDriverRegistrar         string `yaml:"csi-driver-registrar"`
	CSISnapshotter             string `yaml:"csi-snapshotter"`
	CSIResizer                 string `yaml:"csi-resizer"`
	CSICephPlugin              string `yaml:"csi-ceph-plugin"`
	WordPress                  string `yaml:"wordpress"`
	MySQL                      string `yaml:"mysql"`
}

func NewVersions() Versions {
	return Versions{
		Etcd:                       utils.VersionEtcd,
		K8S:                        utils.VersionK8s,
		Helm:                       utils.VersionHelm,
		Containerd:                 utils.VersionContainerd,
		Runc:                       utils.VersionRunc,
		CriCtl:                     utils.VersionCrictl,
		Gobetween:                  utils.VersionGobetween,
		VirtualIP:                  utils.VersionVirtualIP,
		Busybox:                    utils.VersionBusybox,
		Velero:                     utils.VersionVelero,
		VeleroPluginAWS:            utils.VersionVeleroPluginAWS,
		MinioServer:                utils.VersionMinioServer,
		MinioClient:                utils.VersionMinioClient,
		Pause:                      utils.VersionPause,
		CoreDNS:                    utils.VersionCoreDNS,
		Elasticsearch:              utils.VersionElasticsearch,
		Kibana:                     utils.VersionKibana,
		Cerebro:                    utils.VersionCerebro,
		FluentBit:                  utils.VersionFluentBit,
		CalicoTypha:                utils.VersionCalicoTypha,
		CalicoNode:                 utils.VersionCalicoNode,
		CalicoCNI:                  utils.VersionCalicoCni,
		CalicoKubeControllers:      utils.VersionCalicoKubeControllers,
		MetalLBController:          utils.VersionMetalLBController,
		MetalLBSpeaker:             utils.VersionMetalLBSpeaker,
		Ceph:                       utils.VersionCeph,
		KubernetesDashboard:        utils.VersionKubernetesDashboard,
		CertManagerController:      utils.VersionCertManagerController,
		CertManagerCAInjector:      utils.VersionCertManagerCAInjector,
		CertManagerWebHook:         utils.VersionCertManagerWebHook,
		CertManagerACMEResolver:    utils.VersionCertManagerACMEResolver,
		NginxIngressController:     utils.VersionNginxIngressController,
		NginxIngressDefaultBackend: utils.VersionNginxIngressDefaultBackend,
		MetricsScraper:             utils.VersionMetricsScraper,
		MetricsServer:              utils.VersionMetricsServer,
		KubeStateMetrics:           utils.VersionKubeStateMetrics,
		Grafana:                    utils.VersionGrafana,
		Prometheus:                 utils.VersionPrometheus,
		NodeExporter:               utils.VersionNodeExporter,
		AlertManager:               utils.VersionAlertManager,
		CSIAttacher:                utils.VersionCsiAttacher,
		CSIProvisioner:             utils.VersionCsiProvisioner,
		CSIDriverRegistrar:         utils.VersionCsiDriverRegistrar,
		CSICephPlugin:              utils.VersionCsiCephPlugin,
		CSISnapshotter:             utils.VersionCsiSnapshotter,
		CSIResizer:                 utils.VersionCsiResizer,
		WordPress:                  utils.VersionWordpress,
		MySQL:                      utils.VersionMysql,
	}
}

func (versions Versions) GetImages() []Image {
	return []Image{
		{Name: versions.Pause, Features: Features{}},
		{Name: versions.Gobetween, Features: Features{}},
		{Name: versions.VirtualIP, Features: Features{}},
		{Name: versions.Etcd, Features: Features{}},
		{Name: versions.K8S, Features: Features{}},
		{Name: versions.CalicoCNI, Features: Features{}},
		{Name: versions.CalicoNode, Features: Features{}},
		{Name: versions.CalicoTypha, Features: Features{}},
		{Name: versions.CalicoKubeControllers, Features: Features{}},
		{Name: versions.MetalLBController, Features: Features{}},
		{Name: versions.MetalLBSpeaker, Features: Features{}},
		{Name: versions.CoreDNS, Features: Features{}},
		{Name: versions.Busybox, Features: Features{}},
		{Name: versions.MinioServer, Features: Features{utils.FeatureBackup, utils.FeatureStorage}},
		{Name: versions.MinioClient, Features: Features{utils.FeatureBackup, utils.FeatureStorage}},
		{Name: versions.Velero, Features: Features{utils.FeatureBackup, utils.FeatureStorage}},
		{Name: versions.VeleroPluginAWS, Features: Features{utils.FeatureBackup, utils.FeatureStorage}},
		{Name: versions.Ceph, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSIAttacher, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSIProvisioner, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSIDriverRegistrar, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSISnapshotter, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSIResizer, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSICephPlugin, Features: Features{utils.FeatureStorage}},
		{Name: versions.FluentBit, Features: Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: versions.Elasticsearch, Features: Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: versions.Kibana, Features: Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: versions.Cerebro, Features: Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: versions.MetricsScraper, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.MetricsServer, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.KubernetesDashboard, Features: Features{utils.FeaturePackaging}},
		{Name: versions.KubeStateMetrics, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.Grafana, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.Prometheus, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.NodeExporter, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.AlertManager, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.CertManagerController, Features: Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: versions.CertManagerCAInjector, Features: Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: versions.CertManagerWebHook, Features: Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: versions.CertManagerACMEResolver, Features: Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: versions.NginxIngressDefaultBackend, Features: Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: versions.NginxIngressController, Features: Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: versions.MySQL, Features: Features{utils.FeatureShowcase, utils.FeatureStorage}},
		{Name: versions.WordPress, Features: Features{utils.FeatureShowcase, utils.FeatureStorage}},
	}
}
