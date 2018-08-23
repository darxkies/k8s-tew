package config

import "github.com/darxkies/k8s-tew/utils"

type Versions struct {
	Etcd                  string
	K8S                   string
	Helm                  string
	Containerd            string
	Runc                  string
	CriCtl                string
	Gobetween             string
	Ark                   string
	MinioServer           string
	MinioClient           string
	Pause                 string
	CoreDNS               string
	Elasticsearch         string
	ElasticsearchCron     string
	ElasticsearchOperator string
	Kibana                string
	Cerebro               string
	FluentBit             string
	CalicoTypha           string
	CalicoNode            string
	CalicoCNI             string
	RBDProvisioner        string
	Ceph                  string
	KubernetesDashboard   string
	CertManager           string
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
