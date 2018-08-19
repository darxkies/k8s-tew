package config

import "github.com/darxkies/k8s-tew/utils"

type Versions struct {
	Etcd        string
	Flanneld    string
	K8S         string
	Helm        string
	CNI         string
	Containerd  string
	Runc        string
	CriCtl      string
	Gobetween   string
	Ark         string
	MinioServer string
	MinioClient string
}

func NewVersions() Versions {
	return Versions{
		Etcd:        utils.ETCD_VERSION,
		K8S:         utils.K8S_VERSION,
		Helm:        utils.HELM_VERSION,
		Containerd:  utils.CONTAINERD_VERSION,
		Runc:        utils.RUNC_VERSION,
		CriCtl:      utils.CRICTL_VERSION,
		Gobetween:   utils.GOBETWEEN_VERSION,
		Ark:         utils.ARK_VERSION,
		MinioServer: utils.MINIO_SERVER_VERSION,
		MinioClient: utils.MINIO_CLIENT_VERSION,
	}
}
