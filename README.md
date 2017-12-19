# Kubernetes - The Easier Way (k8s-tew)

[Kubernetes](https://kubernetes.io/) is a fairly complex project. For a newbie it is hard to understand and also to use. While [Kelsey Hightower's Kubernetes The Hard Way](https://github.com/kelseyhightower/kubernetes-the-hard-way), on which this project is based, helps a lot to understand Kubernetes, it is optimized for the use with Google Cloud Platform.

This project's aim is to give newbies a tool that allows them to easily tinker with Kubernetes. k8s-tew is a CLI tool to generate the configuration for a Kubernetes cluster (local single node or remote multi node with support for HA). Besides that, k8s-tew is also a supervisor that starts all cluster components. And finally, k8s-tew is a proxy for the HA setup.

# Requirements

k8s-tew was tested so far only on Ubuntu 17.10 and Ubuntu Server 16.04.3. But it should be able to run on other Linux distributions.

# Features

* No docker installation required (uses cri-containerd)
* No cloud provider required
* Runs locally
* Support for deployment to a HA cluster using ssh
* Only the changed files are deployed
* Nodes management from the command line
* Downloads all the used binaries (kubernetes, etcd, flanneld...) from the Internet
* Lower storage and RAM footprint compared to other solutions (kubespray, kubeadm, minikube...)

# Install

## From binary

The 64-bit binary can be downloaded from the following address: https://github.com/darxkies/k8s-tew/releases

## From source

To compile it from source you will need a Go (version 1.8+) environment. Once Go is configured, enter the following command:

```shell
go install github.com/darxkies/k8s-tew/cmd/k8s-tew
```

# Usage

This section will assume that k8s-tew was copied to the folder /usr/local/bin and that the commands are executed using root privileges.

All k8s-tew commands accept the argument --base-directory which defines where all the files will be stored. If no value is defined then it will create a subdirectory called artifacts in the working directory.

To see all the commands and arguments use the -h argument.

## Initialization

The first step in using k8s-tew is to create a config file. This is achieved by executing this command:

```shell
k8s-tew initialize
```

That command generates the config file called artifacts/etc/k8s-tew/config.yaml.

## Nodes

The configuration has no nodes defined yet. A remote node can be added with the following command:

```shell
k8s-tew node-add -n controller00 -i 192.168.122.157 -x 0 -l controller
```

The arguments:

* -n - the name of the node. This name has to match the hostname of that node.
* -i - the ip of the node
* -x - each node needs a unique number
* -l - the role of the node in the cluster: controller and/or worker

k8s-tew is also able to start a cluster on the local computer and for that the local computer has to be added as a node:

```shell
k8s-tew node-add -s
```

The arguments:

* -s - it overrides the previously described flags by overwriting the name and the ip after inferring them.

A node can be removed like this:

```shell
k8s-tew node-add -n controller00
```

And all the nodes can be listed with the command:

```shell
k8s-tew node-list
```

## Generating Files

Once all the nodes were added, the required files (download binares, certificates, kubeconfigs and so on) have to be put in place. And this goes like this:

```shell
k8s-tew generate
```

The versions of the binaries used can be specified as command line arguments:

* --k8s-version - Kubernetes version (default "1.8.5")
* --cni-version - CNI version (default "0.6.0")
* --cri-version - CRI version (default "1.0.0-alpha.1")
* --etcd-version - Etcd version (default "3.2.11")
* --flanneld-version - Flanneld version (default "0.9.1")

## Run

With this command the local cluster can be started:

```shell
k8s-tew run
```

## Deploy

In case remote nodes were added with the deploy command, the missing files are copied to the nodes, k8s-tew is installed and started as a service.

```shell
k8s-tew deploy
```

The files are copied using scp and the ssh private key $HOME/.ssh/id_rsa. If another private key should be used, it can be specified using the command line argument -i.

## Environment

After starting the cluster locally, the user will need some environment variables set to make the work with the cluster easier. This is done with this command:

```shell
eval $(k8s-tew environment)
```

For a remote cluster, additionally to the command above the following command has to be executed outside the cluster:


```shell
export KUBECTL=<base-directory>/etc/k8s-tew/kubeconfig/admin-<controller-name>.kubeconfig
```

<base-directory> and <controller-name> are place holders, that need to be replaced.

## Dashboard

k8s-tew also installs the Kubernetes Dashboard. To access it, the token of the admin user has to be retrieved:

```shell
kubectl -n kube-system describe secret $(kubectl -n kube-system get secret | grep admin-user | awk '{print $1}') | grep token: | awk '{print $2}'
```

Next, the dashboard has to be made accessible:

```shell
kubectl proxy
```

And finally in the web browser, the following page has to be opened:

[http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/](http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/)

When asked to login, enter the token from the first step.

## Config File

By running the following commands for a remote cluster

```shell
k8s-tew initialize
k8s-tew node-add -n controller00 -i 192.168.122.157 -x 0 -l controller
k8s-tew node-add -n controller01 -i 192.168.122.158 -x 1 -l controller
k8s-tew node-add -n controller02 -i 192.168.122.159 -x 2 -l controller
k8s-tew node-add -n worker00 -i 192.168.122.160 -x 3 -l worker
k8s-tew node-add -n worker01 -i 192.168.122.161 -x 4 -l worker
k8s-tew generate
```

a config file will be created with this content

```shell
version: 1.0.0
apiserver-port: 6443
deployment-files:
  admin-{{.Name}}-key.pem:
    file: etc/k8s-tew/ssl/admin-{{.Name}}-key.pem
  admin-{{.Name}}.kubeconfig:
    file: etc/k8s-tew/kubeconfig/admin-{{.Name}}.kubeconfig
  admin-{{.Name}}.pem:
    file: etc/k8s-tew/ssl/admin-{{.Name}}.pem
  bridge:
    labels:
    - worker
    file: opt/k8s-tew/bin/cni/bridge
  ca-key.pem:
    labels:
    - controller
    file: etc/k8s-tew/ssl/ca-key.pem
  ca.pem:
    labels:
    - controller
    - worker
    file: etc/k8s-tew/ssl/ca.pem
  cni-config.json:
    labels:
    - worker
    file: etc/k8s-tew/cni/cni-config.json
  config.yaml:
    labels:
    - controller
    - worker
    file: etc/k8s-tew/config.yaml
  containerd:
    labels:
    - worker
    file: opt/k8s-tew/bin/cri/containerd
  containerd-shim:
    labels:
    - worker
    file: opt/k8s-tew/bin/cri/containerd-shim
  cri-containerd:
    labels:
    - worker
    file: opt/k8s-tew/bin/cri/cri-containerd
  crictl:
    labels:
    - worker
    file: opt/k8s-tew/bin/cri/crictl
  ctr:
    labels:
    - worker
    file: opt/k8s-tew/bin/cri/ctr
  encryption-config.yml:
    labels:
    - controller
    file: etc/k8s-tew/security/encryption-config.yml
  etcd:
    labels:
    - controller
    file: opt/k8s-tew/bin/etcd/etcd
  etcdctl:
    labels:
    - controller
    file: opt/k8s-tew/bin/etcd/etcdctl
  flannel:
    labels:
    - worker
    file: opt/k8s-tew/bin/cni/flannel
  flanneld:
    labels:
    - worker
    file: opt/k8s-tew/bin/etcd/flanneld
  host-local:
    labels:
    - worker
    file: opt/k8s-tew/bin/cni/host-local
  k8s-tew:
    labels:
    - controller
    - worker
    file: opt/k8s-tew/bin/k8s-tew
  k8s-tew.service:
    labels:
    - controller
    - worker
    file: etc/systemd/system/k8s-tew.service
  kube-apiserver:
    labels:
    - controller
    file: opt/k8s-tew/bin/k8s/kube-apiserver
  kube-controller-manager:
    labels:
    - controller
    file: opt/k8s-tew/bin/k8s/kube-controller-manager
  kube-proxy:
    labels:
    - worker
    file: opt/k8s-tew/bin/k8s/kube-proxy
  kube-scheduler:
    labels:
    - controller
    file: opt/k8s-tew/bin/k8s/kube-scheduler
  kubectl:
    labels:
    - controller
    file: opt/k8s-tew/bin/k8s/kubectl
  kubelet:
    labels:
    - worker
    file: opt/k8s-tew/bin/k8s/kubelet
  kubelet-{{.Name}}-key.pem:
    labels:
    - worker
    file: etc/k8s-tew/ssl/kubelet-{{.Name}}-key.pem
  kubelet-{{.Name}}.kubeconfig:
    labels:
    - worker
    file: etc/k8s-tew/kubeconfig/kubelet-{{.Name}}.kubeconfig
  kubelet-{{.Name}}.pem:
    labels:
    - worker
    file: etc/k8s-tew/ssl/kubelet-{{.Name}}.pem
  kubernetes-key.pem:
    labels:
    - controller
    file: etc/k8s-tew/ssl/kubernetes-key.pem
  kubernetes.pem:
    labels:
    - controller
    file: etc/k8s-tew/ssl/kubernetes.pem
  loopback:
    labels:
    - worker
    file: opt/k8s-tew/bin/cni/loopback
  net-config.json:
    labels:
    - worker
    file: etc/k8s-tew/cni/net-config.json
  proxy-key.pem:
    labels:
    - worker
    file: etc/k8s-tew/ssl/proxy-key.pem
  proxy.kubeconfig:
    labels:
    - worker
    file: etc/k8s-tew/kubeconfig/proxy.kubeconfig
  proxy.pem:
    labels:
    - worker
    file: etc/k8s-tew/ssl/proxy.pem
  runc:
    labels:
    - worker
    file: opt/k8s-tew/bin/cri/runc
nodes:
  controller00:
    ip: 192.168.122.157
    index: 0
    labels:
    - controller
  controller01:
    ip: 192.168.122.158
    index: 1
    labels:
    - controller
  controller02:
    ip: 192.168.122.159
    index: 2
    labels:
    - controller
  worker00:
    ip: 192.168.122.160
    index: 3
    labels:
    - worker
  worker01:
    ip: 192.168.122.161
    index: 4
    labels:
    - worker
commands:
  flanneld-configuration:
    command: '{{deployment_file "etcdctl"}} --ca-file={{deployment_file "ca.pem"}}
      --cert-file={{deployment_file "kubernetes.pem"}} --key-file={{deployment_file
      "kubernetes-key.pem"}} set /coreos.com/network/config ''{ "Network": "10.200.0.0/16"
      }'''
    labels:
    - controller
  load-overlay:
    command: modprobe overlay
    labels:
    - controller
    - worker
servers:
  containerd:
    labels:
    - worker
    command: '{{deployment_file "containerd"}}'
    arguments: {}
  cri-containerd:
    labels:
    - worker
    command: '{{deployment_file "cri-containerd"}}'
    arguments:
      network-bin-dir: '{{.BaseDirectory}}/opt/k8s-tew/bin/cni'
      network-conf-dir: '{{.BaseDirectory}}/etc/k8s-tew/cni'
  etcd:
    labels:
    - controller
    command: '{{deployment_file "etcd"}}'
    arguments:
      advertise-client-urls: https://{{.Node.IP}}:2379
      cert-file: '{{deployment_file "kubernetes.pem"}}'
      client-cert-auth: ""
      data-dir: '{{.BaseDirectory}}/var/lib/k8s-tew/etcd'
      initial-advertise-peer-urls: https://{{.Node.IP}}:2380
      initial-cluster: '{{etcd_cluster}}'
      initial-cluster-state: new
      initial-cluster-token: etcd-cluster
      key-file: '{{deployment_file "kubernetes-key.pem"}}'
      listen-client-urls: https://{{.Node.IP}}:2379,http://127.0.0.1:2379
      listen-peer-urls: https://{{.Node.IP}}:2380
      name: '{{.Name}}'
      peer-cert-file: '{{deployment_file "kubernetes.pem"}}'
      peer-client-cert-auth: ""
      peer-key-file: '{{deployment_file "kubernetes-key.pem"}}'
      peer-trusted-ca-file: '{{deployment_file "ca.pem"}}'
      trusted-ca-file: '{{deployment_file "ca.pem"}}'
  flanneld:
    labels:
    - worker
    command: '{{deployment_file "flanneld"}}'
    arguments:
      etcd-cafile: '{{deployment_file "ca.pem"}}'
      etcd-certfile: '{{deployment_file "kubelet-{{.Name}}.pem"}}'
      etcd-endpoints: '{{etcd_servers}}'
      etcd-keyfile: '{{deployment_file "kubelet-{{.Name}}-key.pem"}}'
      v: "0"
  kube-apiserver:
    labels:
    - controller
    command: '{{deployment_file "kube-apiserver"}}'
    arguments:
      admission-control: Initializers,NamespaceLifecycle,NodeRestriction,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota
      advertise-address: '{{.Node.IP}}'
      allow-privileged: "true"
      apiserver-count: '{{controllers_count}}'
      audit-log-maxage: "30"
      audit-log-maxbackup: "3"
      audit-log-maxsize: "100"
      audit-log-path: '{{.BaseDirectory}}/var/log/k8s-tew/audit.log'
      authorization-mode: Node,RBAC
      bind-address: '{{.Node.IP}}'
      client-ca-file: '{{deployment_file "ca.pem"}}'
      enable-swagger-ui: "true"
      etcd-cafile: '{{deployment_file "ca.pem"}}'
      etcd-certfile: '{{deployment_file "kubernetes.pem"}}'
      etcd-keyfile: '{{deployment_file "kubernetes-key.pem"}}'
      etcd-servers: '{{etcd_servers}}'
      event-ttl: 1h
      experimental-encryption-provider-config: '{{deployment_file "encryption-config.yml"}}'
      insecure-bind-address: 127.0.0.1
      kubelet-certificate-authority: '{{deployment_file "ca.pem"}}'
      kubelet-client-certificate: '{{deployment_file "kubernetes.pem"}}'
      kubelet-client-key: '{{deployment_file "kubernetes-key.pem"}}'
      kubelet-https: "true"
      runtime-config: api/all
      secure-port: '{{.Config.APIServerPort}}'
      service-account-key-file: '{{deployment_file "ca-key.pem"}}'
      service-cluster-ip-range: 10.32.0.0/24
      service-node-port-range: 30000-32767
      tls-ca-file: '{{deployment_file "ca.pem"}}'
      tls-cert-file: '{{deployment_file "kubernetes.pem"}}'
      tls-private-key-file: '{{deployment_file "kubernetes-key.pem"}}'
      v: "0"
  kube-controller-manager:
    labels:
    - controller
    command: '{{deployment_file "kube-controller-manager"}}'
    arguments:
      address: 0.0.0.0
      cluster-cidr: 10.200.0.0/16
      cluster-name: kubernetes
      cluster-signing-cert-file: '{{deployment_file "ca.pem"}}'
      cluster-signing-key-file: '{{deployment_file "ca-key.pem"}}'
      leader-elect: "true"
      master: http://127.0.0.1:8080
      root-ca-file: '{{deployment_file "ca.pem"}}'
      service-account-private-key-file: '{{deployment_file "ca-key.pem"}}'
      service-cluster-ip-range: 10.32.0.0/24
      v: "0"
  kube-proxy:
    labels:
    - worker
    command: '{{deployment_file "kube-proxy"}}'
    arguments:
      cluster-cidr: 10.200.0.0/16
      kubeconfig: '{{deployment_file "proxy.kubeconfig"}}'
      proxy-mode: iptables
      v: "0"
  kube-scheduler:
    labels:
    - controller
    command: '{{deployment_file "kube-scheduler"}}'
    arguments:
      leader-elect: "true"
      master: http://127.0.0.1:8080
      v: "0"
  kubelet:
    labels:
    - worker
    command: '{{deployment_file "kubelet"}}'
    arguments:
      allow-privileged: "true"
      anonymous-auth: "false"
      authorization-mode: Webhook
      client-ca-file: '{{deployment_file "ca.pem"}}'
      cluster-dns: 10.32.0.10
      cluster-domain: cluster.local
      container-runtime: remote
      container-runtime-endpoint: unix:///var/run/cri-containerd.sock
      image-pull-progress-deadline: 2m
      kubeconfig: '{{deployment_file "kubelet-{{.Name}}.kubeconfig"}}'
      network-plugin: cni
      pod-cidr: 10.200.{{.Node.Index}}.0/24
      register-node: "true"
      require-kubeconfig: ""
      runtime-request-timeout: 15m
      tls-cert-file: '{{deployment_file "kubelet-{{.Name}}.pem"}}'
      tls-private-key-file: '{{deployment_file "kubelet-{{.Name}}-key.pem"}}'
      v: "0"
```
and the file structure in artifacts (provided no other base directory was used) will look like this

```
.
├── etc
│   ├── k8s-tew
│   │   ├── cni
│   │   │   ├── cni-config.json
│   │   │   └── net-config.json
│   │   ├── config.yaml
│   │   ├── kubeconfig
│   │   │   ├── admin-controller00.kubeconfig
│   │   │   ├── admin-controller01.kubeconfig
│   │   │   ├── admin-controller02.kubeconfig
│   │   │   ├── kubelet-controller00.kubeconfig
│   │   │   ├── kubelet-controller01.kubeconfig
│   │   │   ├── kubelet-controller02.kubeconfig
│   │   │   ├── kubelet-worker00.kubeconfig
│   │   │   ├── kubelet-worker01.kubeconfig
│   │   │   └── proxy.kubeconfig
│   │   ├── security
│   │   │   └── encryption-config.yml
│   │   └── ssl
│   │       ├── admin-controller00-key.pem
│   │       ├── admin-controller00.pem
│   │       ├── admin-controller01-key.pem
│   │       ├── admin-controller01.pem
│   │       ├── admin-controller02-key.pem
│   │       ├── admin-controller02.pem
│   │       ├── ca-key.pem
│   │       ├── ca.pem
│   │       ├── kubelet-controller00-key.pem
│   │       ├── kubelet-controller00.pem
│   │       ├── kubelet-controller01-key.pem
│   │       ├── kubelet-controller01.pem
│   │       ├── kubelet-controller02-key.pem
│   │       ├── kubelet-controller02.pem
│   │       ├── kubelet-worker00-key.pem
│   │       ├── kubelet-worker00.pem
│   │       ├── kubelet-worker01-key.pem
│   │       ├── kubelet-worker01.pem
│   │       ├── kubernetes-key.pem
│   │       ├── kubernetes.pem
│   │       ├── proxy-key.pem
│   │       └── proxy.pem
│   └── systemd
│       └── system
│           └── k8s-tew.service
└── opt
    └── k8s-tew
        └── bin
            ├── cni
            │   ├── bridge
            │   ├── dhcp
            │   ├── flannel
            │   ├── host-local
            │   ├── ipvlan
            │   ├── loopback
            │   ├── macvlan
            │   ├── portmap
            │   ├── ptp
            │   ├── sample
            │   ├── tuning
            │   └── vlan
            ├── cri
            │   ├── containerd
            │   ├── containerd-shim
            │   ├── cri-containerd
            │   ├── crictl
            │   ├── ctr
            │   └── runc
            ├── etcd
            │   ├── etcd
            │   ├── etcdctl
            │   └── flanneld
            ├── k8s
            │   ├── kube-apiserver
            │   ├── kube-controller-manager
            │   ├── kubectl
            │   ├── kubelet
            │   ├── kube-proxy
            │   └── kube-scheduler
            └── k8s-tew
```

The k8s-tew labels are very similiar to the Kubernetes' Labels and Selectors. The difference is that k8s-tew uses no keys. The purpose of the labels is to specify which files belong on a node, which commands should be executed on a node and also which componets need to be started on a node.

# Caveat

* k8s-tew needs root privileges to be executed. Thus, it should be executed on a virtual machine.

# Feedback

* Gmail: darxkies@gmail.com

