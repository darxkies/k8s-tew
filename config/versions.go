package config

import (
	"github.com/darxkies/k8s-tew/utils"
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
	Heapster                   string `yaml:"heapster"`
	AddonResizer               string `yaml:"addon-resizer"`
	KubernetesDashboard        string `yaml:"kubernetes-dashboard"`
	CertManagerController      string `yaml:"cert-manager-controller"`
	NginxIngressController     string `yaml:"nginx-ingress-controller"`
	NginxIngressDefaultBackend string `yaml:"nginx-ingress-default-backend"`
	MetricsServer              string `yaml:"metrics-server"`
	PrometheusOperator         string `yaml:"prometheus-operator"`
	PrometheusConfigReloader   string `yaml:"prometheus-config-reloader"`
	ConfigMapReload            string `yaml:"configmap-reload"`
	KubeStateMetrics           string `yaml:"kube-state-metrics"`
	Grafana                    string `yaml:"grafana"`
	GrafanaWatcher             string `yaml:"grafana-watcher"`
	Prometheus                 string `yaml:"prometheus"`
	PrometheusNodeExporter     string `yaml:"prometheus-node-exporter"`
	PrometheusAlertManager     string `yaml:"prometheus-alert-manager"`
	CSIAttacher                string `yaml:"csi-attacher"`
	CSIProvisioner             string `yaml:"csi-provisioner"`
	CSIDriverRegistrar         string `yaml:"csi-driver-registrar"`
	CSICephRBDPlugin           string `yaml:"csi-ceph-rbd-plugin"`
	CSICephFSPlugin            string `yaml:"csi-ceph-fs-plugin"`
	CSICephSnapshotter         string `yaml:"csi-ceph-snapshotter"`
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
		MinioServer:                utils.VersionMinioServer,
		MinioClient:                utils.VersionMinioClient,
		Pause:                      utils.VersionPause,
		CoreDNS:                    utils.VersionCoredns,
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
		Heapster:                   utils.VersionHeapster,
		AddonResizer:               utils.VersionAddonResizer,
		KubernetesDashboard:        utils.VersionKubernetesDashboard,
		CertManagerController:      utils.VersionCertManagerController,
		NginxIngressController:     utils.VersionNginxIngressController,
		NginxIngressDefaultBackend: utils.VersionNginxIngressDefaultBackend,
		MetricsServer:              utils.VersionMetricsServer,
		PrometheusConfigReloader:   utils.VersionPrometheusConfigReloader,
		ConfigMapReload:            utils.VersionConfigmapReload,
		KubeStateMetrics:           utils.VersionKubeStateMetrics,
		Grafana:                    utils.VersionGrafana,
		GrafanaWatcher:             utils.VersionGrafanaWatcher,
		Prometheus:                 utils.VersionPrometheus,
		PrometheusNodeExporter:     utils.VersionPrometheusNodeExporter,
		PrometheusAlertManager:     utils.VersionPrometheusAlertManager,
		CSIAttacher:                utils.VersionCsiAttacher,
		CSIProvisioner:             utils.VersionCsiProvisioner,
		CSIDriverRegistrar:         utils.VersionCsiDriverRegistrar,
		CSICephRBDPlugin:           utils.VersionCsiCephRbdPlugin,
		CSICephFSPlugin:            utils.VersionCsiCephFsPlugin,
		CSICephSnapshotter:         utils.VersionCsiCephSnapshotter,
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
		{Name: versions.Ceph, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSIAttacher, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSIProvisioner, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSIDriverRegistrar, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSICephRBDPlugin, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSICephFSPlugin, Features: Features{utils.FeatureStorage}},
		{Name: versions.CSICephSnapshotter, Features: Features{utils.FeatureStorage}},
		{Name: versions.FluentBit, Features: Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: versions.Elasticsearch, Features: Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: versions.Kibana, Features: Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: versions.Cerebro, Features: Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: versions.Heapster, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.AddonResizer, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.MetricsServer, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.KubernetesDashboard, Features: Features{utils.FeaturePackaging}},
		{Name: versions.Helm, Features: Features{utils.FeaturePackaging}},
		{Name: versions.PrometheusConfigReloader, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.ConfigMapReload, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.KubeStateMetrics, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.Grafana, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.GrafanaWatcher, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.Prometheus, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.PrometheusNodeExporter, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.PrometheusAlertManager, Features: Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: versions.CertManagerController, Features: Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: versions.NginxIngressDefaultBackend, Features: Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: versions.NginxIngressController, Features: Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: versions.MySQL, Features: Features{utils.FeatureShowcase, utils.FeatureStorage}},
		{Name: versions.WordPress, Features: Features{utils.FeatureShowcase, utils.FeatureStorage}},
	}
}
