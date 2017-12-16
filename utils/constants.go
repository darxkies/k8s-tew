package utils

import "path"

const CONFIG_VERSION = "1.0.0"

// TODO use go templating
const ENCRYPTION_CONFIG_TEMPLATE = `apiVersion: v1
kind: EncryptionConfig
resources:
  - resources:
      - secrets
    providers:
      - aescbc:
          keys:
            - name: key1
              secret: %s
      - identity: {}
`

// TODO use go templating
const KUBE_CONFIG_TEMPLATE = `apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    certificate-authority-data: %s
    server: https://%s
  name: kubernetes-the-easier-way
users:
- name: %s
  user:
    as-user-extra: {}
    client-certificate-data: %s
    client-key-data: %s
contexts:
- context:
    cluster: kubernetes-the-easier-way
    user: %s
  name: default
current-context: default`

const CNI_CONFIG_TEMPLATE = `{
	"name": "cbr0",
	"type": "flannel",
	"delegate": {
		"hairpinMode": true,
		"isDefaultGateway": true
	}
}
`

// TODO use go template
const NET_CONFIG_TEMPLATE = `{
	"Network": "` + CIDR_PREFIX + `.0.0/16",
	"Backend": {
		"Type": "vxlan"
	}
}
`

// TODO use go template
const SERVICE_CONFIG_TEMPLATE = `[Unit]
Description=%s

[Service]
ExecStart=%s
ExecStart=pkill -INT %s
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`

const PROJECT_TITLE = "Kubernetes - The Easier Way"
const RSA_SIZE = 2048
const CA_VALIDITY_PERIOD = 5
const CLIENT_VALIDITY_PERIOD = 1
const ETCD_VERSION = "3.2.11"
const FLANNELD_VERSION = "0.9.1"
const K8S_VERSION = "1.9.0"
const CNI_VERSION = "0.6.0"
const CRI_VERSION = "1.0.0-alpha.1"

const K8S_DOWNLOAD_URL = "https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/%s"
const ETCD_BASE_NAME = "etcd-v%s-linux-amd64"
const ETCD_DOWNLOAD_URL = "https://github.com/coreos/etcd/releases/download/v%s/%s.tar.gz"
const FLANNELD_DOWNLOAD_URL = "https://github.com/coreos/flannel/releases/download/v%s/flanneld-amd64"
const CNI_BASE_NAME = "cni-plugins-amd64-v%s"
const CNI_DOWNLOAD_URL = "https://github.com/containernetworking/plugins/releases/download/v%s/%s.tgz"
const CRI_DOWNLOAD_URL = "https://github.com/kubernetes-incubator/cri-containerd/releases/download/v%s/cri-containerd-%s.tar.gz"

const BASE_DIRECTORY = "artifacts"

const CIDR_PREFIX = "10.200"

// Config
const CONFIG_FILENAME = "config.yaml"

// Node Labels
const NODE_CONTROLLER = "controller"
const NODE_WORKER = "worker"

// Directories
const TEMPORARY_DIRECTORY = "tmp"
const CONFIG_DIRECTORY = "etc"
const SYSTEMD_DIRECTORY = "systemd"
const SYSTEM_DIRECTORY = "system"
const K8S_TEW_DIRECTORY = "k8s-tew"
const CERTIFICATES_DIRECTORY = "ssl"
const OPTIONAL_DIRECTORY = "opt"
const VARIABLE_DIRECTORY = "var"
const LOGGING_DIRECTORY = "log"
const LIBRARY_DIRECTORY = "lib"
const BINARY_DIRECTORY = "bin"
const K8S_DIRECTORY = "k8s"
const ETCD_DIRECTORY = "etcd"
const CRI_DIRECTORY = "cri"
const CNI_DIRECTORY = "cni"
const KUBECONFIG_DIRECTORY = "kubeconfig"
const SECURITY_DIRECTORY = "security"

// Binaries
const K8S_TEW_BINARY = "k8s-tew"

// CNI Binaries
const BRIDGE_BINARY = "bridge"
const FLANNEL_BINARY = "flannel"
const LOOPBACK_BINARY = "loopback"
const HOST_LOCAL_BINARY = "host-local"

// ContainerD Binaries
const CONTAINERD_BINARY = "containerd"
const CONTAINERD_SHIM_BINARY = "containerd-shim"
const CRI_CONTAINERD_BINARY = "cri-containerd"
const CRICTL_BINARY = "crictl"
const CTR_BINARY = "ctr"
const RUNC_BINARY = "runc"

// Etcd Binaries
const ETCD_BINARY = "etcd"
const ETCDCTL_BINARY = "etcdctl"
const FLANNELD_BINARY = "flanneld"

// K8S Binaries
const KUBECTL_BINARY = "kubectl"
const KUBE_APISERVER_BINARY = "kube-apiserver"
const KUBE_CONTROLLER_MANAGER_BINARY = "kube-controller-manager"
const KUBELET_BINARY = "kubelet"
const KUBE_PROXY_BINARY = "kube-proxy"
const KUBE_SCHEDULER_BINARY = "kube-scheduler"

// Certificates
const CA_PEM = "ca.pem"
const CA_KEY_PEM = "ca-key.pem"
const KUBERNETES_PEM = "kubernetes.pem"
const KUBERNETES_KEY_PEM = "kubernetes-key.pem"
const ADMIN_PEM = "admin-{{.Name}}.pem"
const ADMIN_KEY_PEM = "admin-{{.Name}}-key.pem"
const PROXY_PEM = "proxy.pem"
const PROXY_KEY_PEM = "proxy-key.pem"
const KUBELET_PEM = "kubelet-{{.Name}}.pem"
const KUBELET_KEY_PEM = "kubelet-{{.Name}}-key.pem"

// Kubeconfig
const ADMIN_KUBECONFIG = "admin-{{.Name}}.kubeconfig"
const PROXY_KUBECONFIG = "proxy.kubeconfig"
const KUBELET_KUBECONFIG = "kubelet-{{.Name}}.kubeconfig"

// CNI
const CNI_CONFIG = "cni-config.json"
const NET_CONFIG = "net-config.json"

// Security
const ENCRYPTION_CONFIG = "encryption-config.yml"

// Logging
const AUDIT_LOG = "audit.log"

// Deployment
const DEPLOYMENT_USER = "root"

// Service
const SERVICE_NAME = "k8s-tew"
const SERVICE_CONFIG = SERVICE_NAME + ".service"

func GetFullConfigDirectory() string {
	return path.Join(CONFIG_DIRECTORY, K8S_TEW_DIRECTORY)
}

func GetFullCertificatesConfigDirectory() string {
	return path.Join(GetFullConfigDirectory(), CERTIFICATES_DIRECTORY)
}

func GetFullKubeConfigDirectory() string {
	return path.Join(GetFullConfigDirectory(), KUBECONFIG_DIRECTORY)
}

func GetFullCNIConfigDirectory() string {
	return path.Join(GetFullConfigDirectory(), CNI_DIRECTORY)
}

func GetFullSecurityConfigDirectory() string {
	return path.Join(GetFullConfigDirectory(), SECURITY_DIRECTORY)
}

func GetFullBinariesDirectory() string {
	return path.Join(OPTIONAL_DIRECTORY, K8S_TEW_DIRECTORY, BINARY_DIRECTORY)
}

func GetFullK8SBinariesDirectory() string {
	return path.Join(GetFullBinariesDirectory(), K8S_DIRECTORY)
}

func GetFullETCDBinariesDirectory() string {
	return path.Join(GetFullBinariesDirectory(), ETCD_DIRECTORY)
}

func GetFullCRIBinariesDirectory() string {
	return path.Join(GetFullBinariesDirectory(), CRI_DIRECTORY)
}

func GetFullCNIBinariesDirectory() string {
	return path.Join(GetFullBinariesDirectory(), CNI_DIRECTORY)
}

func GetFullETCDDataDirectory() string {
	return path.Join(VARIABLE_DIRECTORY, LIBRARY_DIRECTORY, K8S_TEW_DIRECTORY, ETCD_DIRECTORY)
}

func GetFullLoggingDirectory() string {
	return path.Join(VARIABLE_DIRECTORY, LOGGING_DIRECTORY, K8S_TEW_DIRECTORY)
}

func GetFullTemporaryDirectory() string {
	return path.Join("/", TEMPORARY_DIRECTORY)
}

func GetFullServiceDirectory() string {
	return path.Join(CONFIG_DIRECTORY, SYSTEMD_DIRECTORY, SYSTEM_DIRECTORY)
}
