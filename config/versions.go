package config

import "github.com/darxkies/k8s-tew/utils"

type Versions struct {
	Etcd                  string `yaml:"etcd"`
	K8S                   string `yaml:"kubernetes"`
	Helm                  string `yaml:"helm"`
	Containerd            string `yaml:"containerd"`
	Runc                  string `yaml:"runc"`
	CriCtl                string `yaml:"crictl"`
	Gobetween             string `yaml:"gobetween"`
	Ark                   string `yaml:"ark"`
	MinioServer           string `yaml:"minio-server"`
	MinioClient           string `yaml:"minio-client"`
	Pause                 string `yaml:"pause"`
	CoreDNS               string `yaml:"core-dns"`
	Elasticsearch         string `yaml:"elasticsearch"`
	ElasticsearchCron     string `yaml:"elasticsearch-cron"`
	ElasticsearchOperator string `yaml:"elasticsearch-operator"`
	Kibana                string `yaml:"kibana"`
	Cerebro               string `yaml:"cerebro"`
	FluentBit             string `yaml:"fluent-bit"`
	CalicoTypha           string `yaml:"calico-typha"`
	CalicoNode            string `yaml:"calico-node"`
	CalicoCNI             string `yaml:"calico-cni"`
	RBDProvisioner        string `yaml:"rbd-provisioner"`
	Ceph                  string `yaml:"ceph"`
	KubernetesDashboard   string `yaml:"kubernetes-dashboard"`
	CertManager           string `yaml:"cert-manager"`
}

func NewVersions() Versions {
	return Versions{
		Etcd:                  utils.VERSION_ETCD,
		K8S:                   utils.VERSION_K8S,
		Helm:                  utils.VERSION_HELM,
		Containerd:            utils.VERSION_CONTAINERD,
		Runc:                  utils.VERSION_RUNC,
		CriCtl:                utils.VERSION_CRICTL,
		Gobetween:             utils.VERSION_GOBETWEEN,
		Ark:                   utils.VERSION_ARK,
		MinioServer:           utils.VERSION_MINIO_SERVER,
		MinioClient:           utils.VERSION_MINIO_CLIENT,
		Pause:                 utils.VERSION_PAUSE,
		CoreDNS:               utils.VERSION_COREDNS,
		Elasticsearch:         utils.VERSION_ELASTICSEARCH,
		ElasticsearchCron:     utils.VERSION_ELASTICSEARCH_CRON,
		ElasticsearchOperator: utils.VERSION_ELASTICSEARCH_OPERATOR,
		Kibana:                utils.VERSION_KIBANA,
		Cerebro:               utils.VERSION_CEREBRO,
		FluentBit:             utils.VERSION_FLUENT_BIT,
		CalicoTypha:           utils.VERSION_CALICO_TYPHA,
		CalicoNode:            utils.VERSION_CALICO_NODE,
		CalicoCNI:             utils.VERSION_CALICO_CNI,
		RBDProvisioner:        utils.VERSION_RBD_PROVISIONER,
		Ceph:                  utils.VERSION_CEPH,
		KubernetesDashboard:   utils.VERSION_KUBERNETES_DASHBOARD,
		CertManager:           utils.VERSION_CERT_MANAGER,
	}
}
