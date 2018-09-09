package config

import "github.com/darxkies/k8s-tew/utils"

type Versions struct {
	Etcd                       string `yaml:"etcd"`
	K8S                        string `yaml:"kubernetes"`
	Helm                       string `yaml:"helm"`
	Containerd                 string `yaml:"containerd"`
	Runc                       string `yaml:"runc"`
	CriCtl                     string `yaml:"crictl"`
	Gobetween                  string `yaml:"gobetween"`
	Ark                        string `yaml:"ark"`
	MinioServer                string `yaml:"minio-server"`
	MinioClient                string `yaml:"minio-client"`
	Pause                      string `yaml:"pause"`
	CoreDNS                    string `yaml:"core-dns"`
	Elasticsearch              string `yaml:"elasticsearch"`
	ElasticsearchCron          string `yaml:"elasticsearch-cron"`
	ElasticsearchOperator      string `yaml:"elasticsearch-operator"`
	Kibana                     string `yaml:"kibana"`
	Cerebro                    string `yaml:"cerebro"`
	FluentBit                  string `yaml:"fluent-bit"`
	CalicoTypha                string `yaml:"calico-typha"`
	CalicoNode                 string `yaml:"calico-node"`
	CalicoCNI                  string `yaml:"calico-cni"`
	RBDProvisioner             string `yaml:"rbd-provisioner"`
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
}

func NewVersions() Versions {
	return Versions{
		Etcd:                       utils.VERSION_ETCD,
		K8S:                        utils.VERSION_K8S,
		Helm:                       utils.VERSION_HELM,
		Containerd:                 utils.VERSION_CONTAINERD,
		Runc:                       utils.VERSION_RUNC,
		CriCtl:                     utils.VERSION_CRICTL,
		Gobetween:                  utils.VERSION_GOBETWEEN,
		Ark:                        utils.VERSION_ARK,
		MinioServer:                utils.VERSION_MINIO_SERVER,
		MinioClient:                utils.VERSION_MINIO_CLIENT,
		Pause:                      utils.VERSION_PAUSE,
		CoreDNS:                    utils.VERSION_COREDNS,
		Elasticsearch:              utils.VERSION_ELASTICSEARCH,
		ElasticsearchCron:          utils.VERSION_ELASTICSEARCH_CRON,
		ElasticsearchOperator:      utils.VERSION_ELASTICSEARCH_OPERATOR,
		Kibana:                     utils.VERSION_KIBANA,
		Cerebro:                    utils.VERSION_CEREBRO,
		FluentBit:                  utils.VERSION_FLUENT_BIT,
		CalicoTypha:                utils.VERSION_CALICO_TYPHA,
		CalicoNode:                 utils.VERSION_CALICO_NODE,
		CalicoCNI:                  utils.VERSION_CALICO_CNI,
		RBDProvisioner:             utils.VERSION_RBD_PROVISIONER,
		Ceph:                       utils.VERSION_CEPH,
		Heapster:                   utils.VERSION_HEAPSTER,
		AddonResizer:               utils.VERSION_ADDON_RESIZER,
		KubernetesDashboard:        utils.VERSION_KUBERNETES_DASHBOARD,
		CertManagerController:      utils.VERSION_CERT_MANAGER_CONTROLLER,
		NginxIngressController:     utils.VERSION_NGINX_INGRESS_CONTROLLER,
		NginxIngressDefaultBackend: utils.VERSION_NGINX_INGRESS_DEFAULT_BACKEND,
		MetricsServer:              utils.VERSION_METRICS_SERVER,
		PrometheusOperator:         utils.VERSION_PROMETHEUS_OPERATOR,
		PrometheusConfigReloader:   utils.VERSION_PROMETHEUS_CONFIG_RELOADER,
		ConfigMapReload:            utils.VERSION_CONFIGMAP_RELOAD,
		KubeStateMetrics:           utils.VERSION_KUBE_STATE_METRICS,
		Grafana:                    utils.VERSION_GRAFANA,
		GrafanaWatcher:             utils.VERSION_GRAFANA_WATCHER,
		Prometheus:                 utils.VERSION_PROMETHEUS,
		PrometheusNodeExporter:     utils.VERSION_PROMETHEUS_NODE_EXPORTER,
		PrometheusAlertManager:     utils.VERSION_PROMETHEUS_ALERT_MANAGER,
	}
}
