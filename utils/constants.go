package utils

// Versions
const K8S_VERSION = "1.11.0"
const CONFIG_VERSION = "2.0.0"
const ETCD_VERSION = "3.3.7"
const FLANNELD_VERSION = "0.10.0"
const CNI_VERSION = "0.6.0"
const CONTAINERD_VERSION = "1.1.0"
const RUNC_VERSION = "1.0.0-rc5"
const CRICTL_VERSION = "1.0.0-beta.0"
const GOBETWEEN_VERSION = "0.5.0"
const HELM_VERSION = "2.9.1"

// Settings
const PROJECT_TITLE = "Kubernetes - The Easier Way"
const RSA_SIZE = 2048
const CA_VALIDITY_PERIOD = 5
const CLIENT_VALIDITY_PERIOD = 1
const BASE_DIRECTORY = "assets"
const CLUSTER_IP_RANGE = "10.32.0.0/24"
const CLUSTER_DNS_IP = "10.32.0.10"
const CLUSTER_CIDR = "10.200.0.0/16"
const RESOLV_CONF = "/etc/resolv.conf"
const API_SERVER_PORT = 6443
const PUBLIC_NETWORK = "192.168.0.0/24"
const LOAD_BALANCER_PORT = 16443
const HELM_SERVICE_ACCOUNT = "tiller"

// URLs
const K8S_DOWNLOAD_URL = "https://storage.googleapis.com/kubernetes-release/release/v{{.Versions.K8S}}/bin/linux/amd64/{{.Filename}}"
const ETCD_BASE_NAME = "etcd-v{{.Versions.Etcd}}-linux-amd64"
const ETCD_DOWNLOAD_URL = "https://github.com/coreos/etcd/releases/download/v{{.Versions.Etcd}}/{{.Filename}}.tar.gz"
const FLANNELD_DOWNLOAD_URL = "https://github.com/coreos/flannel/releases/download/v{{.Versions.Flanneld}}/flanneld-amd64"
const CNI_BASE_NAME = "cni-plugins-amd64-v{{.Versions.CNI}}"
const CNI_DOWNLOAD_URL = "https://github.com/containernetworking/plugins/releases/download/v{{.Versions.CNI}}/{{.Filename}}.tgz"
const CONTAINERD_BASE_NAME = "containerd-{{.Versions.Containerd}}.linux-amd64"
const CONTAINERD_DOWNLOAD_URL = "https://github.com/containerd/containerd/releases/download/v{{.Versions.Containerd}}/{{.Filename}}.tar.gz"
const RUNC_DOWNLOAD_URL = "https://github.com/opencontainers/runc/releases/download/v{{.Versions.Runc}}/runc.amd64"
const CRICTL_BASE_NAME = "crictl-v{{.Versions.CriCtl}}-linux-amd64"
const CRICTL_DOWNLOAD_URL = "https://github.com/kubernetes-incubator/cri-tools/releases/download/v{{.Versions.CriCtl}}/{{.Filename}}.tar.gz"
const GOBETWEEN_BASE_NAME = "gobetween_{{.Versions.Gobetween}}_linux_amd64"
const GOBETWEEN_DOWNLOAD_URL = "https://github.com/yyyar/gobetween/releases/download/{{.Versions.Gobetween}}/{{.Filename}}.tar.gz"
const HELM_BASE_NAME = "helm-v{{.Versions.Helm}}-linux-amd64"
const HELM_DOWNLOAD_URL = "https://storage.googleapis.com/kubernetes-helm/{{.Filename}}.tar.gz"

// Config
const CONFIG_FILENAME = "config.yaml"

// Node Labels
const NODE_BOOTSTRAPPER = "bootstrapper"
const NODE_CONTROLLER = "controller"
const NODE_WORKER = "worker"
const NODE_STORAGE_CONTROLLER = "storage-controller"
const NODE_STORAGE_NODE = "storage-node"

// Sub-Directories
const TEMPORARY_SUBDIRECTORY = "tmp"
const CONFIG_SUBDIRECTORY = "etc"
const SYSTEMD_SUBDIRECTORY = "systemd"
const SYSTEM_SUBDIRECTORY = "system"
const K8S_TEW_SUBDIRECTORY = "k8s-tew"
const CERTIFICATES_SUBDIRECTORY = "ssl"
const OPTIONAL_SUBDIRECTORY = "opt"
const VARIABLE_SUBDIRECTORY = "var"
const LOGGING_SUBDIRECTORY = "log"
const LIBRARY_SUBDIRECTORY = "lib"
const RUN_SUBDIRECTORY = "run"
const BINARY_SUBDIRECTORY = "bin"
const K8S_SUBDIRECTORY = "k8s"
const ETCD_SUBDIRECTORY = "etcd"
const CRI_SUBDIRECTORY = "cri"
const CNI_SUBDIRECTORY = "cni"
const KUBECONFIG_SUBDIRECTORY = "kubeconfig"
const SECURITY_SUBDIRECTORY = "security"
const SETUP_SUBDIRECTORY = "setup"
const CONTAINERD_SUBDIRECTORY = "containerd"
const PROFILE_D_SUBDIRECTORY = "profile.d"
const LOAD_BALANCER_SUBDIRECTORY = "lb"
const HELM_SUBDIRECTORY = "helm"
const KUBELET_SUBDIRECTORY = "kubelet"
const MANIFESTS_SUBDIRECTORY = "manifests"
const CEPH_SUBDIRECTORY = "ceph"
const CEPH_BOOTSTRAP_MDS_SUBDIRECTORY = "bootstrap-mds"
const CEPH_BOOTSTRAP_OSD_SUBDIRECTORY = "bootstrap-osd"
const CEPH_BOOTSTRAP_RBD_SUBDIRECTORY = "bootstrap-rbd"
const CEPH_BOOTSTRAP_RGW_SUBDIRECTORY = "bootstrap-rgw"

// Directories
const CONFIG_DIRECTORY = "config"
const CERTIFICATES_DIRECTORY = "certificates"
const CNI_CONFIG_DIRECTORY = "cni-config"
const CRI_CONFIG_DIRECTORY = "cri-config"
const K8S_SECURITY_CONFIG_DIRECTORY = "security-config"
const K8S_CONFIG_DIRECTORY = "k8s-config"
const K8S_KUBE_CONFIG_DIRECTORY = "kube-config"
const K8S_SETUP_CONFIG_DIRECTORY = "setup-config"
const BINARIES_DIRECTORY = "binaries"
const K8S_BINARIES_DIRECTORY = "k8s-binaries"
const ETCD_BINARIES_DIRECTORY = "etcd-binaries"
const CNI_BINARIES_DIRECTORY = "cni-binaries"
const CRI_BINARIES_DIRECTORY = "cri-binaries"
const DYNAMIC_DATA_DIRECTORY = "dynamic-data"
const ETCD_DATA_DIRECTORY = "etcd-data"
const CONTAINERD_DATA_DIRECTORY = "containerd-data"
const LOGGING_DIRECTORY = "logging"
const SERVICE_DIRECTORY = "service"
const CONTAINERD_STATE_DIRECTORY = "containerd-state"
const PROFILE_DIRECTORY = "profile"
const GOBETWEEN_BINARIES_DIRECTORY = "gobetween-binaries"
const GOBETWEEN_CONFIG_DIRECTORY = "gobetween-config"
const HELM_DATA_DIRECTORY = "helm-data"
const KUBELET_DATA_DIRECTORY = "kubelet-data"
const TEMPORARY_DIRECTORY = "temporary"
const K8S_MANIFESTS_DIRECTORY = "kubelet-manifests"
const CEPH_DIRECTORY = "ceph"
const CEPH_CONFIG_DIRECTORY = "ceph-config"
const CEPH_DATA_DIRECTORY = "ceph-data"
const CEPH_BOOTSTRAP_MDS_DIRECTORY = "bootstrap-mds"
const CEPH_BOOTSTRAP_OSD_DIRECTORY = "bootstrap-osd"
const CEPH_BOOTSTRAP_RBD_DIRECTORY = "bootstrap-rbd"
const CEPH_BOOTSTRAP_RGW_DIRECTORY = "bootstrap-rgw"

// Binaries
const K8S_TEW_BINARY = "k8s-tew"

// Helm Binary
const HELM_BINARY = "helm"

// CNI Binaries
const BRIDGE_BINARY = "bridge"
const FLANNEL_BINARY = "flannel"
const LOOPBACK_BINARY = "loopback"
const HOST_LOCAL_BINARY = "host-local"

// ContainerD Binaries
const CONTAINERD_BINARY = "containerd"
const CONTAINERD_SHIM_BINARY = "containerd-shim"
const CTR_BINARY = "ctr"
const RUNC_BINARY = "runc"
const CRICTL_BINARY = "crictl"

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

// Gobeween Binary
const GOBETWEEN_BINARY = "gobetween"

// Certificates
const CA_PEM = "ca.pem"
const CA_KEY_PEM = "ca-key.pem"
const KUBERNETES_PEM = "kubernetes.pem"
const KUBERNETES_KEY_PEM = "kubernetes-key.pem"
const ADMIN_PEM = "admin.pem"
const ADMIN_KEY_PEM = "admin-key.pem"
const PROXY_PEM = "proxy.pem"
const PROXY_KEY_PEM = "proxy-key.pem"
const CONTROLLER_MANAGER_PEM = "controller-manager.pem"
const CONTROLLER_MANAGER_KEY_PEM = "controller-manager-key.pem"
const SCHEDULER_PEM = "scheduler.pem"
const SCHEDULER_KEY_PEM = "scheduler-key.pem"
const KUBELET_PEM = "kubelet-{{.Name}}.pem"
const KUBELET_KEY_PEM = "kubelet-{{.Name}}-key.pem"
const SERVICE_ACCOUNT_PEM = "service-account.pem"
const SERVICE_ACCOUNT_KEY_PEM = "service-account-key.pem"
const FLANNELD_PEM = "flanneld.pem"
const FLANNELD_KEY_PEM = "flanneld-key.pem"
const VIRTUAL_IP_PEM = "virtual-ip.pem"
const VIRTUAL_IP_KEY_PEM = "virtual-ip-key.pem"
const AGGREGATOR_PEM = "aggregator.pem"
const AGGREGATOR_KEY_PEM = "aggregator-key.pem"

// Kubeconfig
const ADMIN_KUBECONFIG = "admin.kubeconfig"
const CONTROLLER_MANAGER_KUBECONFIG = "controller-manager.kubeconfig"
const SCHEDULER_KUBECONFIG = "scheduler.kubeconfig"
const PROXY_KUBECONFIG = "proxy.kubeconfig"
const KUBELET_KUBECONFIG = "kubelet-{{.Name}}.kubeconfig"

// CNI
const CNI_CONFIG = "cni-config.json"
const NET_CONFIG = "net-config.json"

// Security
const ENCRYPTION_CONFIG = "encryption-config.yml"

// Containerd
const CONTAINERD_CONFIG = "config-{{.Name}}.toml"
const CONTAINERD_SOCK = "containerd.sock"

// K8S Config
const K8S_KUBELET_SETUP = "kubelet-setup.yml"
const K8S_ADMIN_USER_SETUP = "admin-user-setup.yml"
const K8S_HELM_USER_SETUP = "helm-user-setup.yml"
const K8S_KUBE_SCHEDULER_CONFIG = "kube-scheduler-config.yaml"
const K8S_KUBELET_CONFIG = "kubelet-{{.Name}}-config.yaml"

// Gobetween Config
const GOBETWEEN_CONFIG = "config.toml"

// Profile
const K8S_TEW_PROFILE = "k8s-tew.sh"

// Logging
const AUDIT_LOG = "audit.log"

// Deployment
const DEPLOYMENT_USER = "root"

// Service
const SERVICE_NAME = "k8s-tew"
const SERVICE_CONFIG = SERVICE_NAME + ".service"

// Ceph
const CEPH_POOL_NAME = "ceph"
const CEPH_CONFIG = "ceph.conf"
const CEPH_CLIENT_ADMIN_KEYRING = "ceph.client.admin.keyring"
const CEPH_MONITOR_KEYRING = "ceph.mon.keyring"
const CEPH_KEYRING = "ceph.keyring"
const CEPH_BOOTSTRAP_MDS_KEYRING = "ceph.bootstrap.mds.keyring"
const CEPH_BOOTSTRAP_OSD_KEYRING = "ceph.bootstrap.osd.keyring"
const CEPH_BOOTSTRAP_RBD_KEYRING = "ceph.bootstrap.rbd.keyring"
const CEPH_BOOTSTRAP_RGW_KEYRING = "ceph.bootstrap.rgw.keyring"
const CEPH_SECRETS = "ceph-secrets.yaml"
const CEPH_SETUP = "ceph-setup.yaml"

// Environment variables
const K8S_TEW_BASE_DIRECTORY = "K8S_TEW_BASE_DIRECTORY"

// Virtual IP Manager
const ELECTION_NAMESPACE = "/k8s-tew"
const ELECTION_CONTROLLER = "/controller-vip-manager"
const ELECTION_WORKER = "/worker-vip-manager"

// Common Names
const CN_ADMIN = "admin"
const CN_AGGREGATOR = "aggregator"
const CN_SYSTEM_KUBE_CONTROLLER_MANAGER = "system:kube-controller-manager"
const CN_SYSTEM_KUBE_SCHEDULER = "system:kube-scheduler"
const CN_SYSTEM_KUBE_PROXY = "system:kube-proxy"
const CN_SYSTEM_NODE_PREFIX = "system:node:%s"

// Templates
const CONTAINERD_CONFIG_TEMPLATE = `root = "{{.ContainerdRootDirectory}}"

state = "{{.ContainerdStateDirectory}}"
oom_score = 0

[grpc]
  address = "{{.ContainerdSock}}"
  uid = 0
  gid = 0
  max_recv_message_size = 16777216
  max_send_message_size = 16777216

[debug]
  address = ""
  uid = 0
  gid = 0
  level = ""

[metrics]
  address = ""
  grpc_histogram = false

[cgroup]
  path = ""

[plugins]
  [plugins.cgroups]
    no_prometheus = false
  [plugins.cri]
    stream_server_address = "{{.IP}}"
    stream_server_port = "10010"
    enable_selinux = false
    sandbox_image = "k8s.gcr.io/pause:3.1"
    stats_collect_period = 10
    systemd_cgroup = false
    enable_tls_streaming = false
    [plugins.cri.containerd]
      snapshotter = "overlayfs"
      [plugins.cri.containerd.default_runtime]
        runtime_type = "io.containerd.runtime.v1.linux"
        runtime_engine = ""
        runtime_root = ""
      [plugins.cri.containerd.untrusted_workload_runtime]
        runtime_type = "io.containerd.runtime.v1.linux"
        runtime_engine = ""
        runtime_root = ""
    [plugins.cri.cni]
      bin_dir = "{{.CNIBinariesDirectory}}"
      conf_dir = "{{.CNIConfigDirectory}}"
      conf_template = ""
    [plugins.cri.registry]
      [plugins.cri.registry.mirrors]
        [plugins.cri.registry.mirrors."docker.io"]
          endpoint = ["https://registry-1.docker.io"]
  [plugins.diff-service]
    default = ["walking"]
  [plugins.linux]
    shim = "{{.CRIBinariesDirectory}}/containerd-shim"
    runtime = "{{.CRIBinariesDirectory}}/runc"
    runtime_root = ""
    no_shim = false
    shim_debug = false
  [plugins.scheduler]
    pause_threshold = 0.02
    deletion_threshold = 0
    mutation_threshold = 100
    schedule_delay = "0s"
    startup_delay = "100ms"
`

const CNI_CONFIG_TEMPLATE = `{
	"name": "cbr0",
	"type": "flannel",
	"delegate": {
		"hairpinMode": true,
		"isDefaultGateway": true
	}
}
`

const NET_CONFIG_TEMPLATE = `{
	"Network": "{{.ClusterCIDR}}",
	"Backend": {
		"Type": "vxlan"
	}
}
`

const SERVICE_CONFIG_TEMPLATE = `[Unit]
Description={{.ProjectTitle}}

[Service]
ExecStart={{.Command}} run --base-directory={{.BaseDirectory}}
ExecStop=/usr/bin/killall -INT {{.Binary}}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`

const K8S_TEW_PROFILE_TEMPLATE = `#!/bin/sh

export K8S_TEW_BASE_DIRECTORY={{.BaseDirectory}}
eval $({{.Binary}} environment)
`

const ENVIRONMENT_TEMPLATE = `
export PATH={{.K8STEWPath}}:{{.K8SPath}}:{{.EtcdPath}}:{{.CRIPath}}:{{.CNIPath}}:{{.CurrentPath}}
export KUBECONFIG={{.KubeConfig}}
export CONTAINER_RUNTIME_ENDPOINT=unix://{{.ContainerdSock}}
`

const GOBETWEEN_CONFIG_TEMPLATE = `[servers.kube-apiserver]
bind = "0.0.0.0:{{ .LoadBalancerPort }}"
protocol = "tcp" 
balance = "roundrobin"

max_connections = 10000
client_idle_timeout = "10m"
backend_idle_timeout = "10m"
backend_connection_timeout = "2s"

[servers.kube-apiserver.discovery]
kind = "static"
static_list = [ {{ .KubeAPIServers | quoted_string_list }} ]
`

const KUBE_SCHEDULER_CONFIGURATION_TEMPLATE = `apiVersion: componentconfig/v1alpha1
kind: KubeSchedulerConfiguration
clientConnection:
  kubeconfig: "{{.KubeConfig}}"
leaderElection:
  leaderElect: true
`

const KUBELET_CONFIGURATION_TEMPLATE = `apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
allowPrivileged: true
authentication:
  anonymous:
    enabled: false
  webhook:
    enabled: true
  x509:
    clientCAFile: "{{.CA}}"
authorization:
  mode: Webhook
clusterDomain: "cluster.local"
clusterDNS:
  - "{{.ClusterDNSIP}}"
podCIDR: "{{.PODCIDR}}"
runtimeRequestTimeout: "15m"
tlsCertFile: "{{.CertificateFilename}}"
tlsPrivateKeyFile: "{{.KeyFilename}}"
staticPodPath: "{{.StaticPodPath}}"
`

const ENCRYPTION_CONFIG_TEMPLATE = `apiVersion: v1
kind: EncryptionConfig
resources:
  - resources:
      - secrets
    providers:
      - aescbc:
          keys:
            - name: key1
              secret: {{.EncryptionKey | unescape}}
      - identity: {}
`

const KUBE_CONFIG_TEMPLATE = `apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: {{.CAData}}
    server: https://{{.APIServer}}
  name: kubernetes-the-easier-way
users:
- name: {{.Name}}
  user:
    client-certificate-data: {{.CertificateData}}
    client-key-data: {{.KeyData}}
contexts:
- context:
    cluster: kubernetes-the-easier-way
    user: {{.User}}
  name: default
current-context: default
`

const K8S_SERVICE_ACCOUNT_CONFIG_TEMPLATE = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: {{.Name}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: {{.Name}}
  namespace: {{.Namespace}}
`

const K8S_KUBELET_CONFIG_TEMPLATE = `apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:kube-apiserver-to-kubelet
rules:
  - apiGroups:
      - ""
    resources:
      - nodes/proxy
      - nodes/stats
      - nodes/log
      - nodes/spec
      - nodes/metrics
    verbs:
      - "*"
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: system:kube-apiserver
  namespace: ""
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:kube-apiserver-to-kubelet
subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: User
    name: kubernetes
`

const K8S_KUBERNETES_DASHBOARD_PORT_UPDATE_TEMPLATE = `kind: Service
apiVersion: v1
metadata:
  labels:
    k8s-app: kubernetes-dashboard
  name: kubernetes-dashboard
  namespace: kube-system
spec:
  type: NodePort
  ports:
    - port: 443
      nodePort: 32443
      targetPort: 8443
  selector:
    k8s-app: kubernetes-dashboard
`

const CEPH_KEYRING_TEMPLATE = `[client.{{.Name}}]
        key = {{.Key | unescape}}
        caps mon = "allow profile {{.Name}}"
`

const CEPH_CLIENT_ADMIN_KEYRING_TEMPLATE = `[client.admin]
        key = {{.Key | unescape}}
        auid = 0
        caps mds = "allow"
        caps mgr = "allow *"
        caps mon = "allow *"
        caps osd = "allow *"
`

const CEPH_MONITOR_KEYRING_TEMPLATE = `[mon.]
        key = {{.MonitorKey | unescape}}
        caps mon = "allow *"
[client.admin]
        key = {{.ClientAdminKey | unescape}}
        auid = 0
        caps mds = "allow"
        caps mgr = "allow *"
        caps mon = "allow *"
        caps osd = "allow *"
[client.bootstrap-mds]
        key = {{.ClientBootstrapMetadataServerKey | unescape}}
        caps mon = "allow profile bootstrap-mds"
[client.bootstrap-osd]
        key = {{.ClientBootstrapObjectStorageKey | unescape}}
        caps mon = "allow profile bootstrap-osd"
[client.bootstrap-rbd]
        key = {{.ClientBootstrapRadosBlockDeviceKey | unescape}}
        caps mon = "allow profile bootstrap-rbd"
[client.bootstrap-rgw]
        key = {{.ClientBootstrapRadosGatewayKey | unescape}}
        caps mon = "allow profile bootstrap-rgw"
[client.k8s-tew]
        key = {{.ClientK8STEWKey | unescape}}
		caps mon = "allow r"
		caps osd = "allow rwx pool={{.CephPoolName}}"
`

const CEPH_CONFIG_TEMPLATE = `[global]
fsid = {{.ClusterID}}

auth cluster required = cephx
auth service required = cephx
auth client required = cephx

mon initial members = {{range $index,$node := .StorageControllers}}{{if $index}},{{end}}{{$node.Name}}{{end}}
mon host = {{range $index,$node := .StorageControllers}}{{if $index}},{{end}}{{$node.IP}}{{end}}
public network = {{.PublicNetwork}}
cluster network = {{.ClusterNetwork}}
osd journal size = 100
log file = /dev/null
osd max object name len = 256
osd max object namespace len = 64
mon_max_pg_per_osd = 1000
osd pg bits = 11
osd pgp bits = 11
osd pool default size = {{len .StorageNodes}}
osd pool default min size = 1
osd pool default pg num = 100
osd pool default pgp num = 100
osd objectstore = filestore
rbd_default_features = 3
fatal signal handlers = false
mon_allow_pool_delete = true
`

const CEPH_SECRETS_TEMPLATE = `apiVersion: v1
kind: Secret
metadata:
    name: ceph-admin
    namespace: kube-system
type: "kubernetes.io/rbd"
data:
    key: {{.ClientAdminKey | base64}}
---
apiVersion: v1
kind: Secret
metadata:
    name: ceph-k8s-tew
    namespace: kube-system
type: "kubernetes.io/rbd"
data:
    key: {{.ClientK8STEWKey | base64}}
`

const CEPH_SETUP_TEMPLATE = `apiVersion: v1
kind: Namespace
metadata:
  name: ceph
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rbd-provisioner
  namespace: kube-system
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-provisioner
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
  - apiGroups: [""]
    resources: ["services"]
    resourceNames: ["kube-dns"]
    verbs: ["list", "get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: rbd-provisioner
  namespace: kube-system
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-provisioner
subjects:
  - kind: ServiceAccount
    name: rbd-provisioner
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: rbd-provisioner
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rbd-provisioner
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rbd-provisioner
subjects:
- kind: ServiceAccount
  name: rbd-provisioner
  namespace: kube-system
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ceph
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: ceph.com/rbd
parameters:
  monitors: {{range $index, $node := .StorageControllers}}{{if $index}},{{end}}{{$node.IP}}:6789{{end}}
  pool: {{.CephPoolName}}
  adminId: admin
  adminSecretName: ceph-admin
  adminSecretNamespace: kube-system
  userId: k8s-tew
  userSecretName: ceph-k8s-tew
  userSecretNamespace: kube-system
  imageFormat: "2"
  imageFeatures: layering
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: rbd-provisioner
  namespace: kube-system
spec:
  replicas: 1
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: rbd-provisioner
    spec:
      containers:
      - name: rbd-provisioner
        image: "quay.io/external_storage/rbd-provisioner:v1.0.0-k8s1.10"
        env:
        - name: PROVISIONER_NAME
          value: ceph.com/rbd
      serviceAccount: rbd-provisioner
{{range $index, $node := .StorageControllers}}---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: ceph-mon-{{$node.Name}}
  namespace: ceph
spec:
  replicas: 1
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: ceph-mon-{{$node.Name}}
    spec:
      hostNetwork: true
      volumes:
      - name: ceph-config
        hostPath:
          path: {{$.CephConfigDirectory}}
          type: DirectoryOrCreate
      - name: ceph-data
        hostPath:
          path: {{$.CephDataDirectory}}
          type: DirectoryOrCreate
      nodeSelector:
        kubernetes.io/hostname: {{$node.Name}}
      containers:
      - name: ceph-mon
        image: ceph/daemon:v3.0.5-stable-3.0-luminous-ubuntu-16.04-x86_64
        args: ["mon"]
        env:
        - name: MON_IP
          value: {{$node.IP}}
        - name: CEPH_PUBLIC_NETWORK
          value: {{$.PublicNetwork}}
        volumeMounts:
        - name: ceph-config
          mountPath: /etc/ceph
        - name: ceph-data
          mountPath: /var/lib/ceph
{{end}}{{range $index, $node := .StorageNodes}}---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: ceph-osd-{{$node.Name}}
  namespace: ceph
spec:
  replicas: 1
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: ceph-osd-{{$node.Name}}
    spec:
      hostNetwork: true
      volumes:
      - name: ceph-config
        hostPath:
          path: {{$.CephConfigDirectory}}
          type: DirectoryOrCreate
      - name: ceph-data
        hostPath:
          path: {{$.CephDataDirectory}}
          type: DirectoryOrCreate
      - name: ceph-dev
        hostPath:
          path: /dev
          type: DirectoryOrCreate
      nodeSelector:
        kubernetes.io/hostname: {{$node.Name}}
      containers:
      - name: ceph-osd
        image: ceph/daemon:v3.0.5-stable-3.0-luminous-ubuntu-16.04-x86_64
        args: ["osd"]
        securityContext:
          privileged: true
        env:
        - name: OSD_TYPE
          value: directory
        volumeMounts:
        - name: ceph-config
          mountPath: /etc/ceph
        - name: ceph-data
          mountPath: /var/lib/ceph
        - name: ceph-dev
          mountPath: /dev
{{end}}---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: ceph-mgr
  namespace: ceph
spec:
  replicas: 1
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: ceph-mgr
    spec:
      hostNetwork: true
      volumes:
      - name: ceph-config
        hostPath:
          path: {{$.CephConfigDirectory}}
          type: DirectoryOrCreate
      - name: ceph-data
        hostPath:
          path: {{$.CephDataDirectory}}
          type: DirectoryOrCreate
      containers:
      - name: ceph-mgr
        image: ceph/daemon:v3.0.5-stable-3.0-luminous-ubuntu-16.04-x86_64
        securityContext:
          privileged: true
        args: ["mgr"]
        volumeMounts:
        - name: ceph-config
          mountPath: /etc/ceph
        - name: ceph-data
          mountPath: /var/lib/ceph
`
