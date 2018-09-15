<p align="center"><img src="logo.svg" width="360"></p>

<p align="center"><a href="https://github.com/cncf/k8s-conformance/tree/master/v1.10/k8s-tew"><img src="conformance/certified-kubernetes-1.10-color.svg" width="120"></a><a href="https://github.com/cncf/k8s-conformance/tree/master/v1.11/k8s-tew"><img src="conformance/certified-kubernetes-1.11-color.svg" width="120"></a></p>

# Kubernetes - The Easier Way (k8s-tew)

[![Build Status](https://travis-ci.org/darxkies/k8s-tew.svg?branch=master)](https://travis-ci.org/darxkies/k8s-tew)
[![Go Report Card](https://goreportcard.com/badge/github.com/darxkies/k8s-tew)](https://goreportcard.com/report/github.com/darxkies/k8s-tew)
[![GitHub release](https://img.shields.io/github/release/darxkies/k8s-tew.svg)](https://github.com/darxkies/k8s-tew/releases/latest)
![GitHub](https://img.shields.io/github/license/darxkies/k8s-tew.svg)


k8s-tew is a CLI tool to install a [Kubernetes](https://kubernetes.io/) Cluster (local, single-node, multi-node or HA-cluster) on Bare Metal. It installs the most essential components needed by a cluster such as networking, storage, monitoring, logging, backuping/restoring and so on. Besides that, k8s-tew is also a supervisor that starts all cluster components on each node, once it setup the nodes.

## TL;DR

[![k8s-tew](https://img.youtube.com/vi/53qQa5EkBTU/0.jpg)](https://www.youtube.com/watch?v=53qQa5EkBTU)

## Why

Kubernetes is a fairly complex project. For a newbie it is hard to understand and also to use. While [Kelsey Hightower's Kubernetes The Hard Way](https://github.com/kelseyhightower/kubernetes-the-hard-way), on which this project is based, helps a lot to understand Kubernetes, it is optimized for the use with Google Cloud Platform.

Thus, this project's aim is to give newbies an easy to use tool that allows them to tinker with Kubernetes and later on to install HA production grade clusters.

# Features

* Multi node setup passes all CNCF conformance tests ([Kubernetes 1.10](https://github.com/cncf/k8s-conformance/tree/master/v1.10/k8s-tew), [Kubernetes 1.11](https://github.com/cncf/k8s-conformance/tree/master/v1.11/k8s-tew))
* Container Management: [Containerd](https://containerd.io/)
* Networking: [Calico](https://www.projectcalico.org)
* Ingress: [NGINX Ingress](https://kubernetes.github.io/ingress-nginx/) and [cert-manager](http://docs.cert-manager.io/en/latest/) for [Let's Encrypt](https://letsencrypt.org/)
* Storage: [Ceph/RBD](https://ceph.com/)
* Metrics: [metering-metrics](https://github.com/kubernetes-incubator/metrics-server) and [Heapster](https://github.com/kubernetes/heapster)
* Monitoring: [Prometheus](https://prometheus.io/) and [Grafana](https://grafana.com/)
* Logging: [Fluent-Bit](https://fluentbit.io/), [Elasticsearch](https://www.elastic.co/), [Kibana](https://www.elastic.co/products/kibana) and [Cerebro](https://github.com/lmenezes/cerebro)
* Backups: [Ark](https://github.com/heptio/ark), [Restic](https://restic.net/) and [Minio](https://www.minio.io/)
* Controller Load Balancing: [gobetween](http://gobetween.io/)
* Package Manager: [Helm](https://helm.sh/)
* Dashboard: [Kubernetes Dashboard](https://github.com/kubernetes/dashboard)
* The communication between the components is encrypted
* RBAC is enabled
* The controllers and the workers have Floating/Virtual IPs
* Integrated Load Balancer for the API Servers
* Support for deployment to a HA cluster using ssh
* Only the changed files are deployed
* No [Docker](https://www.docker.com/) installation required
* No cloud provider required
* Single binary without any dependencies
* Runs locally
* Nodes management from the command line
* Downloads all the used binaries (kubernetes, etcd, flanneld...) from the Internet
* Probably lower storage and RAM footprint compared to other solutions (kubespray, kubeadm, minikube...)
* Uses systemd to install itself as a service on the remote machine
* Installs [WordPress](https://wordpress.com) and [MySQL](https://www.mysql.com) to test drive the installation

# Install

The commands in the upcoming sections will assume that k8s-tew is going to be installed in the directory /usr/local/bin. That means that the aforementioned directory exists and it is included in the PATH. If that is not the case use the following commands:

```shell
sudo mkdir -p /usr/local/bin
export PATH=/usr/local/bin:$PATH
```

## From binary

The 64-bit binary can be downloaded from the following address: https://github.com/darxkies/k8s-tew/releases

Additionally the these commands can be used to download it and install it in /usr/local/bin

```shell
curl -s https://api.github.com/repos/darxkies/k8s-tew/releases/latest | grep "browser_download_url" | cut -d : -f 2,3 | tr -d \" | sudo wget -O /usr/local/bin/k8s-tew -qi -
sudo chmod a+x /usr/local/bin/k8s-tew
```

## From source

To compile it from source you will need a Go (version 1.10+) environment, Git, Make and Docker installed. Once everything is installed, enter the following commands:

```shell
export GOPATH=~/go
export PATH=$GOPATH/bin:$PATH
mkdir -p $GOPATH/src/github.com/darxkies
cd $GOPATH/src/github.com/darxkies
git clone https://github.com/darxkies/k8s-tew.git
cd k8s-tew
make
sudo mv $GOPATH/bin/k8s-tew /usr/local/bin
```

# Requirements

k8s-tew was tested so far on Ubuntu 18.04 and CentOS 7.5. But it should be able to run on other Linux distributions.

Host related dependencies such as socat, conntrack, ipset and rbd are embedded in k8s-tew and put in place, once the cluster is running. Thus it is fairly portable to other Linux distributions.

# Usage

All k8s-tew commands accept the argument --base-directory, which defines where all the files (binaries, certificates, configurations and so on) will be stored. If no value is defined then it will create a subdirectory called "assets" in the working directory. Additionally, the environment variable K8S_TEW_BASE_DIRECTORY can be set to point to the assets directory instead of using --base-directory.

To see all the commands and their arguments use the -h argument.

To activate completion enter:

```shell
source <(k8s-tew completion)
```

## Workflow

To setup a local singl-node cluster, the workflow is: initialize -> configure -> generate -> run

For a remote cluster, the workflow is: initialize -> configure -> generate -> deploy

If something in one of the steps is changed, (e.g. configuration), then all the following steps have to be performed.

## Initialization

The first step in using k8s-tew is to create a config file. This is achieved by executing this command:

```shell
k8s-tew initialize
```

That command generates the config file called {base-directory}/etc/k8s-tew/config.yaml. To overwrite the existing configuration use the argument -f.

## Configuration

After the initialization step the parameters of the cluster should be be adapted. These are the configure parameters and their defaults:

* --apiserver-port -                          API Server Port (default 6443)
* --ca-certificate-validity-period -          CA Certificate Validity Period (default 20)
* --calico-typha-ip -                 Calico Typha IP (default "10.32.0.5")
* --client-certificate-validity-period -      Client Certificate Validity Period (default 15)
* --cluster-cidr -                            Cluster CIDR (default "10.200.0.0/16")
* --cluster-dns-ip -                          Cluster DNS IP (default "10.32.0.10")
* --cluster-domain  -                          Cluster domain (default "cluster.local")
* --cluster-ip-range  -                        Cluster IP range (default "10.32.0.0/24")
* --cluster-name  -                            Cluster Name used for Kubernetes Dashboard (default "k8s-tew")
* --controller-virtual-ip  -                   Controller Virtual/Floating IP for the cluster
* --controller-virtual-ip-interface  -         Controller Virtual/Floating IP interface for the cluster
* --dashboard-port  -                          Dashboard Port (default 32443)
* --deployment-directory  -                    Deployment directory (default "/")
* --email  -                                   Email address used for example for Let's Encrypt (default "k8s-tew@gmail.com")
* --ingress-domain  -                          Ingress domain name (default "k8s-tew.net")
* --load-balancer-port  -                      Load Balancer Port (default 16443)
* --public-network  -                          Public Network (default "192.168.0.0/24")
* --resolv-conf  -                             Custom resolv.conf (default "/etc/resolv.conf")
* --rsa-key-size  -                            RSA Key Size (default 2048)
* --version-addon-resizer  -                   Addon-Resizer version (default "k8s.gcr.io/addon-resizer:1.7")
* --version-ark  -                             Ark version (default "gcr.io/heptio-images/ark:v0.9.4")
* --version-calico-cni  -                      Calico CNI version (default "quay.io/calico/cni:v3.1.3")
* --version-calico-node  -                     Calico Node version (default "quay.io/calico/node:v3.1.3")
* --version-calico-typha  -                    Calico Typha version (default "quay.io/calico/typha:v0.7.4")
* --version-ceph  -                            Ceph version (default "docker.io/ceph/daemon:v3.0.7-stable-3.0-mimic-centos-7-x86_64")
* --version-cerebro  -                         Cerebro version (default "docker.io/upmcenterprises/cerebro:0.6.8")
* --version-cert-manager-controller  -         Cert Manager Controller version (default "quay.io/jetstack/cert-manager-controller:v0.4.1")
* --version-configmap-reload  -                ConfigMap Reload version (default "quay.io/coreos/configmap-reload:v0.0.1")
* --version-containerd  -                      Containerd version (default "1.1.3")
* --version-coredns  -                         CoreDNS version (default "docker.io/coredns/coredns:1.2.0")
* --version-crictl  -                          CriCtl version (default "1.11.1")
* --version-elasticsearch  -                   Elasticsearch version (default "docker.io/upmcenterprises/docker-elasticsearch-kubernetes:6.1.3_0")
* --version-elasticsearch-cron  -              Elasticsearch Cron version (default "docker.io/upmcenterprises/elasticsearch-cron:0.0.3")
* --version-elasticsearch-operator  -          Elasticsearch Operator version (default "docker.io/upmcenterprises/elasticsearch-operator:0.0.12")
* --version-etcd  -                            Etcd version (default "3.3.9")
* --version-fluent-bit  -                      Fluent-Bit version (default "docker.io/fluent/fluent-bit:0.13.0")
* --version-gobetween  -                       Gobetween version (default "0.6.0")
* --version-grafana  -                         Grafana version (default "docker.io/grafana/grafana:5.0.0")
* --version-grafana-watcher  -                 Grafana Watcher version (default "quay.io/coreos/grafana-watcher:v0.0.8")
* --version-heapster  -                        Heapster version (default "k8s.gcr.io/heapster:v1.3.0")
* --version-helm  -                            Helm version (default "2.9.1")
* --version-k8s  -                             Kubernetes version (default "1.11.2")
* --version-kibana  -                          Kibana version (default "docker.elastic.co/kibana/kibana-oss:6.1.3")
* --version-kube-state-metrics  -              Kube State Metrics version (default "gcr.io/google_containers/kube-state-metrics:v1.2.0")
* --version-kubernetes-dashboard  -            Kubernetes Dashboard version (default "k8s.gcr.io/kubernetes-dashboard-amd64:v1.8.3")
* --version-metrics-server  -                  Metrics Server version (default "gcr.io/google_containers/metrics-server-amd64:v0.2.1")
* --version-minio-client  -                    Minio client version (default "docker.io/minio/mc:RELEASE.2018-08-18T02-13-04Z")
* --version-minio-server  -                    Minio server version (default "docker.io/minio/minio:RELEASE.2018-08-18T03-49-57Z")
* --version-nginx-ingress-controller  -        Nginx Ingress Controller version (default "quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.18.0")
* --version-nginx-ingress-default-backend  -   Nginx Ingress Default Backend version (default "k8s.gcr.io/defaultbackend:1.4")
* --version-pause  -                           Pause version (default "k8s.gcr.io/pause:3.1")
* --version-prometheus  -                      Prometheus version (default "quay.io/prometheus/prometheus:v2.2.1")
* --version-prometheus-alert-manager  -        Prometheus Alert Manager version (default "quay.io/prometheus/alertmanager:v0.15.1")
* --version-prometheus-config-reloader  -      Prometheus Config Reloader version (default "quay.io/coreos/prometheus-config-reloader:v0.20.0")
* --version-prometheus-node-exporter  -        Prometheus Node Exporter version (default "quay.io/prometheus/node-exporter:v0.15.2")
* --version-prometheus-operator  -             Prometheus Operator version (default "quay.io/coreos/prometheus-operator:v0.20.0")
* --version-rbd-provisioner  -                 RBD-Provisioner version (default "quay.io/external_storage/rbd-provisioner:v2.1.1-k8s1.11")
* --version-runc  -                            Runc version (default "1.0.0-rc5")
* --vip-raft-controller-port  -                VIP Raft Controller Port (default 16277)
* --vip-raft-worker-port  -                    VIP Raft Worker Port (default 16728)
* --worker-virtual-ip  -                       Worker Virtual/Floating IP for the cluster
* --worker-virtual-ip-interface  -             Worker Virtual/Floating IP interface for the cluster

The email and the ingress-domain parameters need to be changed if you want a working Ingress and Lets' Encrypt configuration. It goes like this:

```shell
k8s-tew configure --email john.doe@gmail.com --ingress-domain example.com
```

Another important argument is --resolv-conf which is used to define which resolv.conf file should be used for DNS.

The Virtual/Floating IP parameters should be accordingly changed if you want true HA. This is especially for the controllers important. Then if there are for example three controllers then the IP of the first controller is used by the whole cluster and if that one fails then the whole cluster will stop working. k8s-tew uses internally [RAFT](https://raft.github.io/) and its leader election functionality to select one node on which the Virtual IP is set. If the leader fails, one of the remaining nodes gets the Virtual IP assigned.

## Labels

k8s-tew uses labels to specify which files belong on a node, which commands should be executed on a node and also which components need to be started on a node. They are similar to Kubernetes' Labels and Selectors.

* bootstrapper - This label marks bootstrapping node/commands
* controller - Kubernetes Controler. At least three controller nodes are required for a HA cluster.
* worker - Kubernetes Node/Minion. At least one worker node is required.
* storage - The storage manager components are installed on the controller nodes. The worker nodes are used to store the actual data of the pods. If the storage label is omitted then all nodes are used. If you choose to use only some nodes for storage, then keep in mind that you need at least three storage managers and at least two data storage servers for a HA cluster.

## Nodes

So far the configuration has no nodes defined yet.

### Add Remote Node

A remote node can be added with the following command:

```shell
k8s-tew node-add -n controller00 -i 192.168.100.100 -x 0 -l controller
```

The arguments:

* -n - The name of the node. This name has to match the hostname of that node.
* -i - The IP of the node
* -x - Each node needs a unique number. Do not reuse this number even though the node does not exist anymore.
* -l - The role of the node in the cluster: controller and/or worker and/or storage

__NOTE__: Make sure the IP address of the node matches the public network set using the configuration argument --public-network.

### Add Local Node
k8s-tew is also able to start a cluster on the local computer and for that the local computer has to be added as a node:

```shell
k8s-tew node-add -s
```

The arguments:

* -s - it overrides the previously described flags by overwriting the name and the ip after inferring them.

### Remove Node

A node can be removed like this:

```shell
k8s-tew node-remove -n controller00
```
### List Nodes

And all the nodes can be listed with the command:

```shell
k8s-tew node-list
```

## Generating Files

Once all the nodes were added, the required files (third party binares, certificates, kubeconfigs and so on) have to be put in place. And this goes like this:

```shell
k8s-tew generate
```

__NOTE__: Depending on your internet connection, you could speed up the download process by using the argument --parallel.

## Run

With this command the local cluster can be started:

```shell
k8s-tew run
```

## Deploy

In case remote nodes were added with the deploy command, the remotely missing files are copied to the nodes. k8s-tew is installed and started as a service.

The deployment is executed with the command:

```shell
k8s-tew deploy
```

The files are copied using scp and the ssh private key $HOME/.ssh/id_rsa. In case the file  $HOME/.ssh/id_rsa does not exist it should be generated using the command ssh-keygen. If another private key should be used, it can be specified using the command line argument -i.

__NOTE__: The argument --pull-images downloads the required Docker Images on the nodes, before the setup process is executed. That could speed up the whole setup process later on. Furthermore, by using --parallel the process of uploading files to the nodes and the download of Docker Images can be again considerable shortened. Use these parameters with caution, as they can starve your network.

## Environment

After starting the cluster, the user will need some environment variables set locally to make the interaction with the cluster easier. This is done with this command:

```shell
eval $(k8s-tew environment)
```

## Kubernetes Dashboard

k8s-tew also installs the Kubernetes Dashboard. Invoke the command to display the admin token:

```shell
k8s-tew dashboard
```

If you have a GUI web browser installed, then you can use the following command to display the admin token for three seconds, enough time to copy the token, and to also open the web browser:

```shell
k8s-tew dashboard -o
```

__NOTE__: It takes minutes to actually download the dashboard. Use the following command to check the status of the pods:

```shell
kubectl get pods -n kube-system
```

Once the pod is running the dashboard can be accessed through the TCP port 32443. Regarding the IP address, use the IP address of a worker node or the worker Virtual IP if one was specified.

When asked to login, enter the admin token.

## Ingress

For working Ingress make sure ports 80 and 443 are available. The Ingress Domain have to be also configured before 'generate' and 'deploy' are executed:

```shell
k8s-tew configure --ingress-domain [ingress-domain]
```

## WordPress

* Address: http://[worker-ip]:30100
* Address: https://wordpress.[ingress-domain]

__NOTE__: Wordpress is installed for testing purposes and [ingress-domain] can be set using the configure command.

## Minio

* Address: http://[worker-ip]:30800
* Username: minio
* Password: changeme

## Grafana

* Address: http://[worker-ip]:30900
* Username: admin
* Password: changeme

## Kibana

* Address: https://[worker-ip]:30980

## Cerebro

* Address: http://[worker-ip]:30990

## Ceph Dashboard

 * Address: https://[worker-ip]:7000
 * Username: admin
 * Password: changeme

__NOTE__: [worker-ip] is the IP of the worker where ceph-mgr is running. The port for this service is not exposed as a Kubernetes NodePort.

# Cluster Setups

Vagrant/VirtualBox can be used to test drive k8s-tew. The host is used to bootstrap the cluster which runs in VirtualBox. The Vagrantfile included in the repository can be used for single-node/multi-node & Ubuntu 18.04/CentOS 7 setups.

The Vagrantfile can be configured using the environment variables:

* OS - define the operating system. It accepts ubuntu, the default value, and centos.
* MULTI_NODE - if set then a HA cluster is generated. Otherwise a single-node setup is used.
* CONTROLLERS - defines the number of controller nodes. The default number is 3.
* WORKERS - specifies the number of worker nodes. The default number is 2.
* SSH_PUBLIC_KEY - if this environment variable is not set, then $HOME/.ssh/id_rsa is used by default.
* IP_PREFIX - this value is used to generate the IP addresses of the nodes. If not set 192.168.100 will be used.

__NOTE__: The multi-node setup with the default settings needs about 20G RAM for itself.

## Usage

The directory called setup contains sub-directories for various cluster setup configurations:

* local - it starts single-node cluster locally without using any kind of virtualization. This kind of setup needs root rights.
* ubuntu-single-node - Ubuntu 18.04 single-node cluster. It needs about 8G Ram.
* ubuntu-multi-node - Ubuntu 18.04 HA cluster. It needs around 20G Ram.
* centos-single-node - CentOS 7.5 single-node cluster. It needs about 8G Ram.
* centos-multi-node - CentOS 7.5 HA cluster. It needs around 20G Ram.

__NOTE__: Regardless of the setup, once the deployment is done it will take a while to download all required containers from the internet. So better use kubectl to check the status of the pods.

__NOTE__: For the local setup, to access the Kubernetes Dashboard use the internal IP address (e.g. 192.168.x.y or 10.x.y.z) and not 127.0.0.1/localhost. Depending on the hardware used, it might take a while until it starts and setups everything.

### Create

Change to one of the sub-directories and enter the following command to start the cluster:

```shell
make
```
__NOTE__: This will destroy any existing VMs, creates new VMs and performs all the steps (forced initialization, configuration, generation and deployment) to create the cluster.

### Stop

For the local setup, just press CTRL+C.

For the other setups enter:

```shell
make halt
```

### Start

To start an existing setup/VMs enter:

```shell
make up
```

__NOTE__: This and the following commands work only for Vagrant based setups.

### SSH

For single-node setups enter:

```shell
make ssh
```

And for multi-node setups:

```shell
make ssh-controller00
make ssh-controller01
make ssh-controller02
make ssh-worker00
make ssh-worker01
```

### Kubernetes Dashboard

This will display the token for three seconds, and then it will open the web browser pointing to the address of Kubernetes Dashboard:

```shell
make dashboard
```

### Ingress Port Forwarding

In order to start port forwarding from your host's ports 80 and 443 to Vagrant's VMs for Ingress enter:

```shell
make forward-80
make forward-443
```

__NOTE__: Both commands are blocking. So you need two different terminal sessions.

### ubuntu-single-node Setup Snippets

#### Configuration

This is what the configuration looks like when using ubuntu-single-node setup:

```yaml
version: 2.1.0
cluster-id: 33a54ca2-7d4a-47f9-95ad-f40a07f70465
cluster-name: k8s-tew
email: k8s-tew@gmail.com
ingress-domain: k8s-tew.net
load-balancer-port: 16443
vip-raft-controller-port: 16277
vip-raft-worker-port: 16728
dashboard-port: 32443
apiserver-port: 6443
public-network: 192.168.110.0/24
cluster-domain: cluster.local
cluster-ip-range: 10.32.0.0/24
cluster-dns-ip: 10.32.0.10
cluster-cidr: 10.200.0.0/16
calico-typha-ip: 10.32.0.5
resolv-conf: /run/systemd/resolve/resolv.conf
deployment-directory: /
rsa-size: 2048
ca-validity-period: 20
client-validity-period: 15
versions:
  etcd: 3.3.9
  kubernetes: 1.11.2
  helm: 2.9.1
  containerd: 1.1.3
  runc: 1.0.0-rc5
  crictl: 1.11.1
  gobetween: 0.6.0
  ark: gcr.io/heptio-images/ark:v0.9.4
  minio-server: docker.io/minio/minio:RELEASE.2018-08-18T03-49-57Z
  minio-client: docker.io/minio/mc:RELEASE.2018-08-18T02-13-04Z
  pause: k8s.gcr.io/pause:3.1
  core-dns: docker.io/coredns/coredns:1.2.0
  elasticsearch: docker.io/upmcenterprises/docker-elasticsearch-kubernetes:6.1.3_0
  elasticsearch-cron: docker.io/upmcenterprises/elasticsearch-cron:0.0.3
  elasticsearch-operator: docker.io/upmcenterprises/elasticsearch-operator:0.0.12
  kibana: docker.elastic.co/kibana/kibana-oss:6.1.3
  cerebro: docker.io/upmcenterprises/cerebro:0.6.8
  fluent-bit: docker.io/fluent/fluent-bit:0.13.0
  calico-typha: quay.io/calico/typha:v0.7.4
  calico-node: quay.io/calico/node:v3.1.3
  calico-cni: quay.io/calico/cni:v3.1.3
  rbd-provisioner: quay.io/external_storage/rbd-provisioner:v2.1.1-k8s1.11
  ceph: docker.io/ceph/daemon:v3.0.7-stable-3.0-mimic-centos-7-x86_64
  heapster: k8s.gcr.io/heapster:v1.3.0
  addon-resizer: k8s.gcr.io/addon-resizer:1.7
  kubernetes-dashboard: k8s.gcr.io/kubernetes-dashboard-amd64:v1.8.3
  cert-manager-controller: quay.io/jetstack/cert-manager-controller:v0.4.1
  nginx-ingress-controller: quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.18.0
  nginx-ingress-default-backend: k8s.gcr.io/defaultbackend:1.4
  metrics-server: gcr.io/google_containers/metrics-server-amd64:v0.2.1
  prometheus-operator: quay.io/coreos/prometheus-operator:v0.20.0
  prometheus-config-reloader: quay.io/coreos/prometheus-config-reloader:v0.20.0
  configmap-reload: quay.io/coreos/configmap-reload:v0.0.1
  kube-state-metrics: gcr.io/google_containers/kube-state-metrics:v1.2.0
  grafana: docker.io/grafana/grafana:5.0.0
  grafana-watcher: quay.io/coreos/grafana-watcher:v0.0.8
  prometheus: quay.io/prometheus/prometheus:v2.2.1
  prometheus-node-exporter: quay.io/prometheus/node-exporter:v0.15.2
  prometheus-alert-manager: quay.io/prometheus/alertmanager:v0.15.1
assets:
  directories:
    absolute-containerd-state:
      directory: run/containerd
      absolute: true
    ark:
      directory: opt/k8s-tew/bin/ark
    bash-completion:
      directory: etc/bash_completion.d
    binaries:
      directory: opt/k8s-tew/bin
    bootstrap-mds:
      labels:
      - worker
      directory: var/lib/k8s-tew/ceph/bootstrap-mds
    bootstrap-osd:
      labels:
      - worker
      directory: var/lib/k8s-tew/ceph/bootstrap-osd
    bootstrap-rbd:
      labels:
      - worker
      directory: var/lib/k8s-tew/ceph/bootstrap-rbd
    bootstrap-rgw:
      labels:
      - worker
      directory: var/lib/k8s-tew/ceph/bootstrap-rgw
    ceph-config:
      labels:
      - worker
      directory: etc/k8s-tew/ceph
    ceph-data:
      labels:
      - worker
      directory: var/lib/k8s-tew/ceph
    certificates:
      directory: etc/k8s-tew/ssl
    cni-binaries:
      labels:
      - controller
      - worker
      directory: opt/k8s-tew/bin/cni
    cni-config:
      labels:
      - controller
      - worker
      directory: etc/k8s-tew/cni
    config:
      directory: etc/k8s-tew
    containerd-data:
      directory: var/lib/k8s-tew/containerd
    containerd-state:
      directory: var/run/k8s-tew/containerd
    cri-binaries:
      directory: opt/k8s-tew/bin/cri
    cri-config:
      directory: etc/k8s-tew/cri
    dynamic-data:
      directory: var/lib/k8s-tew
    etcd-binaries:
      directory: opt/k8s-tew/bin/etcd
    etcd-data:
      directory: var/lib/k8s-tew/etcd
    gobetween-binaries:
      directory: opt/k8s-tew/bin/lb
    gobetween-config:
      directory: etc/k8s-tew/lb
    helm-data:
      directory: var/lib/k8s-tew/helm
    host-binaries:
      labels:
      - controller
      - worker
      directory: opt/k8s-tew/bin/host
    k8s-binaries:
      directory: opt/k8s-tew/bin/k8s
    k8s-config:
      directory: etc/k8s-tew/k8s
    kube-config:
      directory: etc/k8s-tew/k8s/kubeconfig
    kubelet-data:
      directory: var/lib/k8s-tew/kubelet
    kubelet-manifests:
      labels:
      - worker
      directory: etc/k8s-tew/k8s/manifests
    logging:
      directory: var/log/k8s-tew
    pods-data:
      directory: var/lib/k8s-tew/kubelet/pods
    profile:
      directory: etc/profile.d
    security-config:
      directory: etc/k8s-tew/k8s/security
    service:
      directory: etc/systemd/system
    setup-config:
      directory: etc/k8s-tew/k8s/setup
    temporary:
      directory: tmp
  files:
    admin-key.pem:
      directory: certificates
    admin-user-setup.yaml:
      directory: setup-config
    admin.kubeconfig:
      directory: kube-config
    admin.pem:
      directory: certificates
    aggregator-key.pem:
      labels:
      - controller
      directory: certificates
    aggregator.pem:
      labels:
      - controller
      directory: certificates
    ark:
      directory: ark
    ark-restic-restore-helper:
      directory: ark
    ark-setup.yaml:
      directory: setup-config
    ark.bash-completion:
      directory: bash-completion
    ca-key.pem:
      labels:
      - controller
      directory: certificates
    ca.pem:
      labels:
      - controller
      - worker
      directory: certificates
    calico-setup.yaml:
      directory: setup-config
    ceph-secrets.yaml:
      directory: setup-config
    ceph-setup.yaml:
      directory: setup-config
    ceph.bootstrap.mds.keyring:
      labels:
      - controller
      - worker
      filename: ceph.keyring
      directory: bootstrap-mds
    ceph.bootstrap.osd.keyring:
      labels:
      - controller
      - worker
      filename: ceph.keyring
      directory: bootstrap-osd
    ceph.bootstrap.rbd.keyring:
      labels:
      - controller
      - worker
      filename: ceph.keyring
      directory: bootstrap-rbd
    ceph.bootstrap.rgw.keyring:
      labels:
      - controller
      - worker
      filename: ceph.keyring
      directory: bootstrap-rgw
    ceph.client.admin.keyring:
      labels:
      - controller
      - worker
      directory: ceph-config
    ceph.conf:
      labels:
      - controller
      - worker
      directory: ceph-config
    ceph.mon.keyring:
      labels:
      - controller
      - worker
      directory: ceph-config
    cert-manager-setup.yaml:
      directory: setup-config
    config-{{.Name}}.toml:
      labels:
      - controller
      - worker
      directory: cri-config
    config.toml:
      labels:
      - controller
      directory: gobetween-config
    config.yaml:
      labels:
      - controller
      - worker
      directory: config
    containerd:
      labels:
      - controller
      - worker
      directory: cri-binaries
    containerd-shim:
      labels:
      - controller
      - worker
      directory: cri-binaries
    containerd.sock:
      directory: absolute-containerd-state
    controller-manager-key.pem:
      directory: certificates
    controller-manager.kubeconfig:
      labels:
      - controller
      directory: kube-config
    controller-manager.pem:
      directory: certificates
    coredns-setup.yaml:
      directory: setup-config
    crictl:
      labels:
      - controller
      - worker
      directory: cri-binaries
    crictl.bash-completion:
      labels:
      - controller
      directory: bash-completion
    ctr:
      labels:
      - controller
      - worker
      directory: cri-binaries
    efk-setup.yaml:
      directory: setup-config
    elasticsearch-operator-setup.yaml:
      directory: setup-config
    encryption-config.yaml:
      labels:
      - controller
      directory: security-config
    etcd:
      labels:
      - controller
      directory: etcd-binaries
    etcdctl:
      labels:
      - controller
      directory: etcd-binaries
    gobetween:
      labels:
      - controller
      directory: gobetween-binaries
    heapster-setup.yaml:
      directory: setup-config
    helm:
      directory: k8s-binaries
    helm-user-setup.yaml:
      directory: setup-config
    helm.bash-completion:
      directory: bash-completion
    k8s-tew:
      labels:
      - controller
      - worker
      directory: binaries
    k8s-tew.bash-completion:
      labels:
      - controller
      - worker
      directory: bash-completion
    k8s-tew.service:
      labels:
      - controller
      - worker
      directory: service
    k8s-tew.sh:
      labels:
      - controller
      - worker
      directory: profile
    kube-apiserver:
      labels:
      - controller
      directory: k8s-binaries
    kube-controller-manager:
      labels:
      - controller
      directory: k8s-binaries
    kube-prometheus-datasource-setup.yaml:
      directory: setup-config
    kube-prometheus-deployment-dashboard-setup.yaml:
      directory: setup-config
    kube-prometheus-kubernetes-capacity-planning-dashboard-setup.yaml:
      directory: setup-config
    kube-prometheus-kubernetes-cluster-health-dashboard-setup.yaml:
      directory: setup-config
    kube-prometheus-kubernetes-cluster-status-dashboard-setup.yaml:
      directory: setup-config
    kube-prometheus-kubernetes-control-plane-status-dashboard-setup.yaml:
      directory: setup-config
    kube-prometheus-kubernetes-resource-requests-dashboard-setup.yaml:
      directory: setup-config
    kube-prometheus-nodes-dashboard-setup.yaml:
      directory: setup-config
    kube-prometheus-pods-dashboard-setup.yaml:
      directory: setup-config
    kube-prometheus-setup.yaml:
      directory: setup-config
    kube-prometheus-statefulset-dashboard-setup.yaml:
      directory: setup-config
    kube-proxy:
      labels:
      - controller
      - worker
      directory: k8s-binaries
    kube-scheduler:
      labels:
      - controller
      directory: k8s-binaries
    kube-scheduler-config.yaml:
      labels:
      - controller
      directory: k8s-config
    kubectl:
      labels:
      - controller
      directory: k8s-binaries
    kubectl.bash-completion:
      labels:
      - controller
      directory: bash-completion
    kubelet:
      labels:
      - controller
      - worker
      directory: k8s-binaries
    kubelet-{{.Name}}-config.yaml:
      labels:
      - controller
      - worker
      directory: k8s-config
    kubelet-{{.Name}}-key.pem:
      labels:
      - controller
      - worker
      directory: certificates
    kubelet-{{.Name}}.kubeconfig:
      labels:
      - controller
      - worker
      directory: kube-config
    kubelet-{{.Name}}.pem:
      labels:
      - controller
      - worker
      directory: certificates
    kubelet-setup.yaml:
      directory: setup-config
    kubernetes-dashboard-setup.yaml:
      directory: setup-config
    kubernetes-key.pem:
      labels:
      - controller
      directory: certificates
    kubernetes.pem:
      labels:
      - controller
      directory: certificates
    letsencrypt-cluster-issuer.yaml:
      directory: setup-config
    metrics-server-setup.yaml:
      directory: setup-config
    nginx-ingress-setup.yaml:
      directory: setup-config
    prometheus-operator-setup.yaml:
      directory: setup-config
    proxy-key.pem:
      directory: certificates
    proxy.kubeconfig:
      labels:
      - controller
      - worker
      directory: kube-config
    proxy.pem:
      directory: certificates
    runc:
      labels:
      - controller
      - worker
      directory: cri-binaries
    scheduler-key.pem:
      directory: certificates
    scheduler.kubeconfig:
      labels:
      - controller
      directory: kube-config
    scheduler.pem:
      directory: certificates
    service-account-key.pem:
      labels:
      - controller
      directory: certificates
    service-account.pem:
      labels:
      - controller
      directory: certificates
    wordpress-setup.yaml:
      directory: setup-config
nodes:
  single-node:
    ip: 192.168.110.50
    index: 0
    labels:
    - controller
    - worker
commands:
- name: setup-ubuntu
  command: apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y apt-transport-https
    bash-completion
  labels:
  - controller
  - worker
  os:
  - ubuntu
- name: setup-centos
  command: systemctl disable firewalld && systemctl stop firewalld && setenforce 0
    && sed -i --follow-symlinks 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/sysconfig/selinux
  labels:
  - controller
  - worker
  os:
  - centos{base-directory}
- name: setup-centos-disable-selinux
  command: setenforce 0
  labels:
  - controller
  - worker
  os:
  - centos
- name: swapoff
  command: swapoff -a
  labels:
  - controller
  - worker
- name: load-overlay
  command: modprobe overlay
  labels:
  - controller
  - worker
- name: load-btrfs
  command: modprobe btrfs
  labels:
  - controller
  - worker
- name: load-br_netfilter
  command: modprobe br_netfilter
  labels:
  - controller
  - worker
- name: enable-br_netfilter
  command: echo '1' > /proc/sys/net/bridge/bridge-nf-call-iptables
  labels:
  - controller
  - worker
- name: enable-net-forwarding
  command: sysctl net.ipv4.conf.all.forwarding=1
  labels:
  - controller
  - worker
- name: kubelet-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kubelet-setup.yaml
  labels:
  - bootstrapper
- name: admin-user-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/admin-user-setup.yaml
  labels:
  - bootstrapper
- name: calico-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/calico-setup.yaml
  labels:
  - bootstrapper
- name: coredns-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/coredns-setup.yaml
  labels:
  - bootstrapper
- name: helm-user-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/helm-user-setup.yaml
  labels:
  - bootstrapper
- name: ceph-secrets
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/ceph-secrets.yaml
  labels:
  - bootstrapper
- name: ceph-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/ceph-setup.yaml
  labels:
  - bootstrapper
- name: helm-init
  command: KUBECONFIG=/workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    HELM_HOME=/workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/var/lib/k8s-tew/helm
    /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/helm
    init --service-account tiller --upgrade
  labels:
  - bootstrapper
- name: kubernetes-dashboard-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kubernetes-dashboard-setup.yaml
  labels:
  - bootstrapper
- name: cert-manager-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/cert-manager-setup.yaml
  labels:
  - bootstrapper
- name: nginx-ingress-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f{base-directory} /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/nginx-ingress-setup.yaml
  labels:
  - bootstrapper
- name: letsencrypt-cluster-issuer-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/letsencrypt-cluster-issuer.yaml
  labels:
  - bootstrapper
- name: heapster-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/heapster-setup.yaml
  labels:
  - bootstrapper
- name: metrics-server-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/metrics-server-setup.yaml
  labels:
  - bootstrapper
- name: prometheus-operator-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/prometheus-operator-setup.yaml
  labels:
  - bootstrapper
- name: kube-prometheus-datasource-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kube-prometheus-datasource-setup.yaml
  labels:
  - bootstrapper
- name: kube-prometheus-kuberntes-cluster-status-dashboard-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kube-prometheus-kubernetes-cluster-status-dashboard-setup.yaml
  labels:
  - bootstrapper
- name: kube-prometheus-kuberntes-cluster-health-dashboard-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kube-prometheus-kubernetes-cluster-health-dashboard-setup.yaml
  labels:
  - bootstrapper
- name: kube-prometheus-kuberntes-control-plane-status-dashboard-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kube-prometheus-kubernetes-control-plane-status-dashboard-setup.yaml
  labels:
  - bootstrapper
- name: kube-prometheus-kuberntes-capacity-planning-dashboard-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kube-prometheus-kubernetes-capacity-planning-dashboard-setup.yaml
  labels:
  - bootstrapper
- name: kube-prometheus-kuberntes-resource-requests-dashboard-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kube-prometheus-kubernetes-resource-requests-dashboard-setup.yaml
  labels:
  - bootstrapper
- name: kube-prometheus-nodes-dashboard-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kube-prometheus-nodes-dashboard-setup.yaml
  labels:
  - bootstrapper
- name: kube-prometheus-deployment-dashboard-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kube-prometheus-deployment-dashboard-setup.yaml
  labels:
  - bootstrapper
- name: kube-prometheus-statefulset-dashboard-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kube-prometheus-statefulset-dashboard-setup.yaml
  labels:
  - bootstrapper
- name: kube-prometheus-pods-dashboard-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kube-prometheus-pods-dashboard-setup.yaml
  labels:
  - bootstrapper
- name: kube-prometheus-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/kube-prometheus-setup.yaml
  labels:
  - bootstrapper
- name: elasticsearch-operator-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/elasticsearch-operator-setup.yaml
  labels:
  - bootstrapper
- name: efk-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/efk-setup.yaml
  labels:
  - bootstrapper
- name: patch-kibana-service
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    get svc kibana-elasticsearch-cluster -n logging --output=jsonpath={.spec..nodePort}
    | grep 30980 || /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    patch service kibana-elasticsearch-cluster -n logging -p '{"spec":{"type":"NodePort","ports":[{"port":80,"nodePort":30980}]}}'
  labels:
  - bootstrapper
- name: patch-cerebro-service
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    get svc cerebro-elasticsearch-cluster -n logging --output=jsonpath={.spec..nodePort}
    | grep 30990 || /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    patch service cerebro-elasticsearch-cluster -n logging -p '{"spec":{"type":"NodePort","ports":[{"port":80,"nodePort":30990}]}}'
  labels:
  - bootstrapper
- name: ark-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/ark-setup.yaml
  labels:
  - bootstrapper
- name: wordpress-setup
  command: /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/opt/k8s-tew/bin/k8s/kubectl
    --request-timeout 30s --kubeconfig /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig
    apply -f /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/etc/k8s-tew/k8s/setup/wordpress-setup.yaml
  labels:
  - bootstrapper
servers:
- name: etcd
  enabled: true
  labels:
  - controller
  logger:
    enabled: true
    filename: '{{asset_directory "logging"}}/etcd.log'
  command: '{{asset_file "etcd"}}'
  arguments:
    advertise-client-urls: https://{{.Node.IP}}:2379
    cert-file: '{{asset_file "kubernetes.pem"}}'
    client-cert-auth: ""
    data-dir: '{{asset_directory "etcd-data"}}'
    initial-advertise-peer-urls: https://{{.Node.IP}}:2380
    initial-cluster: '{{etcd_cluster}}'
    initial-cluster-state: new
    initial-cluster-token: etcd-cluster
    key-file: '{{asset_file "kubernetes-key.pem"}}'
    listen-client-urls: https://{{.Node.IP}}:2379
    listen-peer-urls: https://{{.Node.IP}}:2380
    name: '{{.Name}}'
    peer-cert-file: '{{asset_file "kubernetes.pem"}}'
    peer-client-cert-auth: ""
    peer-key-file: '{{asset_file "kubernetes-key.pem"}}'
    peer-trusted-ca-file: '{{asset_file "ca.pem"}}'
    trusted-ca-file: '{{asset_file "ca.pem"}}'
- name: containerd
  enabled: true
  labels:
  - controller
  - worker
  logger:
    enabled: true
    filename: '{{asset_directory "logging"}}/containerd.log'
  command: '{{asset_file "containerd"}}'
  arguments:
    config: '{{asset_file "config-{{.Name}}.toml"}}'
- name: gobetween
  enabled: true
  labels:
  - controller
  logger:
    enabled: true
    filename: '{{asset_directory "logging"}}/gobetween.log'
  command: '{{asset_file "gobetween"}}'
  arguments:
    config: '{{asset_file "config.toml"}}'
- name: kube-apiserver
  enabled: true
  labels:
  - controller
  logger:
    enabled: true
    filename: '{{asset_directory "logging"}}/kube-apiserver.log'
  command: '{{asset_file "kube-apiserver"}}'
  arguments:
    advertise-address: '{{.Node.IP}}'
    allow-privileged: "true"
    apiserver-count: '{{controllers_count}}'
    audit-log-maxage: "30"
    audit-log-maxbackup: "3"
    audit-log-maxsize: "100"
    audit-log-path: '{{asset_directory "logging"}}/audit.log'
    authorization-mode: Node,RBAC
    bind-address: 0.0.0.0
    client-ca-file: '{{asset_file "ca.pem"}}'
    enable-admission-plugins: Initializers,NamespaceLifecycle,NodeRestriction,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota
    enable-aggregator-routing: "true"
    enable-swagger-ui: "true"
    etcd-cafile: '{{asset_file "ca.pem"}}'
    etcd-certfile: '{{asset_file "kubernetes.pem"}}'
    etcd-keyfile: '{{asset_file "kubernetes-key.pem"}}'
    etcd-servers: '{{etcd_servers}}'
    event-ttl: 1h
    experimental-encryption-provider-config: '{{asset_file "encryption-config.yaml"}}'
    kubelet-certificate-authority: '{{asset_file "ca.pem"}}'
    kubelet-client-certificate: '{{asset_file "kubernetes.pem"}}'
    kubelet-client-key: '{{asset_file "kubernetes-key.pem"}}'
    kubelet-https: "true"
    proxy-client-cert-file: '{{asset_file "aggregator.pem"}}'
    proxy-client-key-file: '{{asset_file "aggregator-key.pem"}}'
    requestheader-allowed-names: aggregator,admin,system:kube-controller-manager,system:kube-controller-manager,system:kube-scheduler,system:node:single-node
    requestheader-client-ca-file: '{{asset_file "ca.pem"}}'
    requestheader-extra-headers-prefix: X-Remote-Extra-
    requestheader-group-headers: X-Remote-Group
    requestheader-username-headers: X-Remote-User
    runtime-config: api/all
    secure-port: '{{.Config.APIServerPort}}'
    service-account-key-file: '{{asset_file "service-account.pem"}}'
    service-cluster-ip-range: '{{.Config.ClusterIPRange}}'
    service-node-port-range: 30000-32767
    tls-cert-file: '{{asset_file "kubernetes.pem"}}'
    tls-private-key-file: '{{asset_file "kubernetes-key.pem"}}'
    v: "0"
- name: kube-controller-manager
  enabled: true
  labels:
  - controller
  logger:
    enabled: true
    filename: '{{asset_directory "logging"}}/kube-controller-manager.log'
  command: '{{asset_file "kube-controller-manager"}}'
  arguments:
    address: 0.0.0.0
    allocate-node-cidrs: "true"
    cluster-cidr: '{{.Config.ClusterCIDR}}'
    cluster-name: kubernetes
    cluster-signing-cert-file: '{{asset_file "ca.pem"}}'
    cluster-signing-key-file: '{{asset_file "ca-key.pem"}}'
    kubeconfig: '{{asset_file "controller-manager.kubeconfig"}}'
    leader-elect: "true"
    root-ca-file: '{{asset_file "ca.pem"}}'
    service-account-private-key-file: '{{asset_file "service-account-key.pem"}}'
    service-cluster-ip-range: '{{.Config.ClusterIPRange}}'
    use-service-account-credentials: "true"
    v: "0"
- name: kube-scheduler
  enabled: true
  labels:
  - controller
  logger:
    enabled: true
    filename: '{{asset_directory "logging"}}/kube-scheduler.log'
  command: '{{asset_file "kube-scheduler"}}'
  arguments:
    config: '{{asset_file "kube-scheduler-config.yaml"}}'
    v: "0"
- name: kube-proxy
  enabled: true
  labels:
  - controller
  - worker
  logger:
    enabled: true
    filename: '{{asset_directory "logging"}}/kube-proxy.log'
  command: '{{asset_file "kube-proxy"}}'
  arguments:
    cluster-cidr: '{{.Config.ClusterCIDR}}'
    kubeconfig: '{{asset_file "proxy.kubeconfig"}}'
    proxy-mode: iptables
    v: "0"
- name: kubelet
  enabled: true
  labels:
  - controller
  - worker
  logger:
    enabled: true
    filename: '{{asset_directory "logging"}}/kubelet.log'
  command: '{{asset_file "kubelet"}}'
  arguments:
    config: '{{asset_file "kubelet-{{.Name}}-config.yaml"}}'
    container-runtime: remote
    container-runtime-endpoint: unix://{{asset_file "containerd.sock"}}
    fail-swap-on: "false"
    image-pull-progress-deadline: 2m
    kubeconfig: '{{asset_file "kubelet-{{.Name}}.kubeconfig"}}'
    network-plugin: cni
    read-only-port: "10255"
    register-node: "true"
    resolv-conf: '{{.Config.ResolvConf}}'
    root-dir: '{{asset_directory "kubelet-data"}}'
    v: "0"
```

#### File structure

This is the content of {base-directory}:

```shell
assets/
├── etc
│   ├── bash_completion.d
│   │   ├── ark.bash-completion
│   │   ├── crictl.bash-completion
│   │   ├── helm.bash-completion
│   │   ├── k8s-tew.bash-completion
│   │   └── kubectl.bash-completion
│   ├── k8s-tew
│   │   ├── ceph
│   │   │   ├── ceph.client.admin.keyring
│   │   │   ├── ceph.conf
│   │   │   └── ceph.mon.keyring
│   │   ├── cni
│   │   ├── config.yaml
│   │   ├── cri
│   │   │   └── config-single-node.toml
│   │   ├── k8s
│   │   │   ├── kubeconfig
│   │   │   │   ├── admin.kubeconfig
│   │   │   │   ├── controller-manager.kubeconfig
│   │   │   │   ├── kubelet-single-node.kubeconfig
│   │   │   │   ├── proxy.kubeconfig
│   │   │   │   └── scheduler.kubeconfig
│   │   │   ├── kubelet-single-node-config.yaml
│   │   │   ├── kube-scheduler-config.yaml
│   │   │   ├── manifests
│   │   │   ├── security
│   │   │   │   └── encryption-config.yaml
│   │   │   └── setup
│   │   │       ├── admin-user-setup.yaml
│   │   │       ├── ark-setup.yaml
│   │   │       ├── calico-setup.yaml
│   │   │       ├── ceph-secrets.yaml
│   │   │       ├── ceph-setup.yaml
│   │   │       ├── cert-manager-setup.yaml
│   │   │       ├── coredns-setup.yaml
│   │   │       ├── efk-setup.yaml
│   │   │       ├── elasticsearch-operator-setup.yaml
│   │   │       ├── heapster-setup.yaml
│   │   │       ├── helm-user-setup.yaml
│   │   │       ├── kubelet-setup.yaml
│   │   │       ├── kube-prometheus-datasource-setup.yaml
│   │   │       ├── kube-prometheus-deployment-dashboard-setup.yaml
│   │   │       ├── kube-prometheus-kubernetes-capacity-planning-dashboard-setup.yaml
│   │   │       ├── kube-prometheus-kubernetes-cluster-health-dashboard-setup.yaml
│   │   │       ├── kube-prometheus-kubernetes-cluster-status-dashboard-setup.yaml
│   │   │       ├── kube-prometheus-kubernetes-control-plane-status-dashboard-setup.yaml
│   │   │       ├── kube-prometheus-kubernetes-resource-requests-dashboard-setup.yaml
│   │   │       ├── kube-prometheus-nodes-dashboard-setup.yaml
│   │   │       ├── kube-prometheus-pods-dashboard-setup.yaml
│   │   │       ├── kube-prometheus-setup.yaml
│   │   │       ├── kube-prometheus-statefulset-dashboard-setup.yaml
│   │   │       ├── kubernetes-dashboard-setup.yaml
│   │   │       ├── letsencrypt-cluster-issuer.yaml
│   │   │       ├── metrics-server-setup.yaml
│   │   │       ├── nginx-ingress-setup.yaml
│   │   │       ├── prometheus-operator-setup.yaml
│   │   │       └── wordpress-setup.yaml
│   │   ├── lb
│   │   │   └── config.toml
│   │   └── ssl
│   │       ├── admin-key.pem
│   │       ├── admin.pem
│   │       ├── aggregator-key.pem
│   │       ├── aggregator.pem
│   │       ├── ca-key.pem
│   │       ├── ca.pem
│   │       ├── controller-manager-key.pem
│   │       ├── controller-manager.pem
│   │       ├── kubelet-single-node-key.pem
│   │       ├── kubelet-single-node.pem
│   │       ├── kubernetes-key.pem
│   │       ├── kubernetes.pem
│   │       ├── proxy-key.pem
│   │       ├── proxy.pem
│   │       ├── scheduler-key.pem
│   │       ├── scheduler.pem
│   │       ├── service-account-key.pem
│   │       └── service-account.pem
│   ├── profile.d
│   │   └── k8s-tew.sh
│   └── systemd
│       └── system
│           └── k8s-tew.service
├── opt
│   └── k8s-tew
│       └── bin
│           ├── ark
│           │   ├── ark
│           │   └── ark-restic-restore-helper
│           ├── cni
│           ├── cri
│           │   ├── containerd
│           │   ├── containerd-shim
│           │   ├── crictl
│           │   ├── ctr
│           │   └── runc
│           ├── etcd
│           │   ├── etcd
│           │   └── etcdctl
│           ├── host
│           ├── k8s
│           │   ├── helm
│           │   ├── kube-apiserver
│           │   ├── kube-controller-manager
│           │   ├── kubectl
│           │   ├── kubelet
│           │   ├── kube-proxy
│           │   └── kube-scheduler
│           ├── k8s-tew
│           └── lb
│               └── gobetween
├── tmp
└── var
    ├── lib
    │   └── k8s-tew
    │       ├── ceph
    │       │   ├── bootstrap-mds
    │       │   │   └── ceph.keyring
    │       │   ├── bootstrap-osd
    │       │   │   └── ceph.keyring
    │       │   ├── bootstrap-rbd
    │       │   │   └── ceph.keyring
    │       │   └── bootstrap-rgw
    │       │       └── ceph.keyring
    │       ├── containerd
    │       ├── etcd
    │       ├── helm
    │       │   ├── cache
    │       │   │   └── archive
    │       │   ├── plugins
    │       │   ├── repository
    │       │   │   ├── cache
    │       │   │   │   ├── local-index.yaml -> /workspace/k8s-tew/src/github.com/darxkies/k8s-tew/setup/ubuntu-single-node/assets/var/lib/k8s-tew/helm/repository/local/index.yaml
    │       │   │   │   └── stable-index.yaml
    │       │   │   ├── local
    │       │   │   │   └── index.yaml
    │       │   │   └── repositories.yaml
    │       │   └── starters
    │       └── kubelet
    │           └── pods
    ├── log
    │   └── k8s-tew
    └── run
        └── k8s-tew
            └── containerd

52 directories, 94 files
```

#### Kubectl Output

And this is what `kubectl get all --all-namespaces` returns:

```shell
NAMESPACE     NAME                                                       READY     STATUS      RESTARTS   AGE
backup        pod/ark-68c56f6d75-q5crq                                   1/1       Running     0          5m
backup        pod/minio-7895b9d495-szkm6                                 1/1       Running     0          5m
backup        pod/minio-setup-8k9l4                                      0/1       Completed   0          5m
backup        pod/restic-9knmd                                           1/1       Running     0          5m
default       pod/elasticsearch-operator-sysctl-pmgkg                    1/1       Running     0          5m
kube-system   pod/coredns-646944c5c4-lm6xq                               1/1       Running     0          12m
kube-system   pod/coredns-646944c5c4-smmnv                               1/1       Running     0          12m
kube-system   pod/heapster-heapster-6b7985754b-mqhn4                     2/2       Running     0          7m
kube-system   pod/kubernetes-dashboard-845c9dbcdf-nmjcj                  1/1       Running     0          12m
kube-system   pod/tiller-deploy-759cb9df9-w2nhw                          1/1       Running     0          12m
logging       pod/cerebro-elasticsearch-cluster-567468c475-d6j8j         1/1       Running     0          5m
logging       pod/elasticsearch-operator-76769d959c-pfxd4                1/1       Running     0          8m
logging       pod/es-client-elasticsearch-cluster-7d6fb8dcdd-dlwtz       1/1       Running     0          5m
logging       pod/es-data-elasticsearch-cluster-default-0                1/1       Running     0          5m
logging       pod/es-master-elasticsearch-cluster-default-0              1/1       Running     0          5m
logging       pod/fluent-bit-nzvc7                                       1/1       Running     0          8m
logging       pod/kibana-elasticsearch-cluster-66778d655d-l6gjh          1/1       Running     0          5m
monitoring    pod/alertmanager-kube-prometheus-0                         2/2       Running     0          8m
monitoring    pod/kube-prometheus-exporter-kube-state-65c6c77579-hw7vq   2/2       Running     0          8m
monitoring    pod/kube-prometheus-exporter-node-m5k8d                    1/1       Running     0          8m
monitoring    pod/kube-prometheus-grafana-749496574c-dsr6r               2/2       Running     0          8m
monitoring    pod/metrics-server-6486f65987-wws9d                        1/1       Running     0          12m
monitoring    pod/prometheus-kube-prometheus-0                           3/3       Running     1          8m
monitoring    pod/prometheus-operator-6bc587f9fc-b62rw                   1/1       Running     0          12m
networking    pod/calico-node-js5lx                                      2/2       Running     0          12m
networking    pod/cert-manager-86b95f4dc8-6gpmq                          1/1       Running     0          12m
networking    pod/nginx-ingress-controller-wdxxw                         1/1       Running     0          12m
networking    pod/nginx-ingress-default-backend-789c7df7cb-x5ph5         1/1       Running     0          12m
storage       pod/ceph-mgr-54b46c94c4-q92km                              1/1       Running     0          12m
storage       pod/ceph-mon-single-node-8c98868fb-sjzqm                   1/1       Running     0          12m
storage       pod/ceph-osd-single-node-7b7f848b97-zbtp9                  1/1       Running     0          12m
storage       pod/ceph-setup-5cvhb                                       0/1       Completed   1          12m
storage       pod/rbd-provisioner-789795cf94-flxdw                       1/1       Running     0          12m
wordpress     pod/cm-acme-http-solver-hkrkp                              1/1       Running     0          5m
wordpress     pod/mysql-65d54b75b4-zb5s2                                 1/1       Running     0          5m
wordpress     pod/wordpress-6b74676664-p66kf                             1/1       Running     1          5m
NAMESPACE     NAME                                                       TYPE           CLUSTER-IP    EXTERNAL-IP   PORT(S)                      AGE
backup        service/minio                                              NodePort       10.32.0.170   <none>        9000:30800/TCP               5m
default       service/kubernetes                                         ClusterIP      10.32.0.1     <none>        443/TCP                      12m
kube-system   service/heapster                                           ClusterIP      10.32.0.168   <none>        8082/TCP                     12m
kube-system   service/kube-dns                                           ClusterIP      10.32.0.10    <none>        53/UDP,53/TCP                12m
kube-system   service/kube-prometheus-exporter-kube-controller-manager   ClusterIP      10.32.0.186   <none>        10252/TCP                    8m
kube-system   service/kube-prometheus-exporter-kube-dns                  ClusterIP      10.32.0.73    <none>        10054/TCP,10055/TCP          8m
kube-system   service/kube-prometheus-exporter-kube-etcd                 ClusterIP      10.32.0.102   <none>        4001/TCP                     8m
kube-system   service/kube-prometheus-exporter-kube-scheduler            ClusterIP      10.32.0.229   <none>        10251/TCP                    8m
kube-system   service/kubelet                                            ClusterIP      None          <none>        10250/TCP                    5m
kube-system   service/kubernetes-dashboard                               NodePort       10.32.0.38    <none>        443:32443/TCP                12m
kube-system   service/tiller-deploy                                      ClusterIP      10.32.0.66    <none>        44134/TCP                    12m
logging       service/cerebro-elasticsearch-cluster                      NodePort       10.32.0.196   <none>        80:30990/TCP                 5m
logging       service/elasticsearch-discovery-elasticsearch-cluster      ClusterIP      10.32.0.51    <none>        9300/TCP                     5m
logging       service/elasticsearch-elasticsearch-cluster                ClusterIP      10.32.0.224   <none>        9200/TCP                     5m
logging       service/es-data-svc-elasticsearch-cluster                  ClusterIP      10.32.0.20    <none>        9300/TCP                     5m
logging       service/kibana-elasticsearch-cluster                       NodePort       10.32.0.118   <none>        80:30980/TCP                 5m
monitoring    service/alertmanager-operated                              ClusterIP      None          <none>        9093/TCP,6783/TCP            8m
monitoring    service/kube-prometheus                                    ClusterIP      10.32.0.19    <none>        9090/TCP                     8m
monitoring    service/kube-prometheus-alertmanager                       ClusterIP      10.32.0.253   <none>        9093/TCP                     8m
monitoring    service/kube-prometheus-exporter-kube-state                ClusterIP      10.32.0.234   <none>        80/TCP                       8m
monitoring    service/kube-prometheus-exporter-node                      ClusterIP      10.32.0.35    <none>        9100/TCP                     8m
monitoring    service/kube-prometheus-grafana                            NodePort       10.32.0.238   <none>        80:30900/TCP                 8m
monitoring    service/metrics-server                                     ClusterIP      10.32.0.141   <none>        443/TCP                      12m
monitoring    service/prometheus-operated                                ClusterIP      None          <none>        9090/TCP                     8m
networking    service/calico-typha                                       ClusterIP      10.32.0.5     <none>        5473/TCP                     12m
networking    service/nginx-ingress-controller                           LoadBalancer   10.32.0.108   <pending>     80:30525/TCP,443:30104/TCP   12m
networking    service/nginx-ingress-default-backend                      ClusterIP      10.32.0.52    <none>        80/TCP                       12m
wordpress     service/cm-acme-http-solver-fr8f8                          NodePort       10.32.0.91    <none>        8089:32756/TCP               5m
wordpress     service/mysql                                              ClusterIP      None          <none>        3306/TCP                     5m
wordpress     service/wordpress                                          NodePort       10.32.0.82    <none>        80:30100/TCP                 5m
NAMESPACE    NAME                                           DESIRED   CURRENT   READY     UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
backup       daemonset.apps/restic                          1         1         1         1            1           <none>          5m
default      daemonset.apps/elasticsearch-operator-sysctl   1         1         1         1            1           <none>          5m
logging      daemonset.apps/fluent-bit                      1         1         1         1            1           <none>          8m
monitoring   daemonset.apps/kube-prometheus-exporter-node   1         1         1         1            1           <none>          8m
networking   daemonset.apps/calico-node                     1         1         1         1            1           <none>          12m
networking   daemonset.apps/nginx-ingress-controller        1         1         1         1            1           <none>          12m
NAMESPACE     NAME                                                  DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
backup        deployment.apps/ark                                   1         1         1            1           5m
backup        deployment.apps/minio                                 1         1         1            1           5m
kube-system   deployment.apps/coredns                               2         2         2            2           12m
kube-system   deployment.apps/heapster-heapster                     1         1         1            1           12m
kube-system   deployment.apps/kubernetes-dashboard                  1         1         1            1           12m
kube-system   deployment.apps/tiller-deploy                         1         1         1            1           12m
logging       deployment.apps/cerebro-elasticsearch-cluster         1         1         1            1           5m
logging       deployment.apps/elasticsearch-operator                1         1         1            1           8m
logging       deployment.apps/es-client-elasticsearch-cluster       1         1         1            1           5m
logging       deployment.apps/kibana-elasticsearch-cluster          1         1         1            1           5m
monitoring    deployment.apps/kube-prometheus-exporter-kube-state   1         1         1            1           8m
monitoring    deployment.apps/kube-prometheus-grafana               1         1         1            1           8m
monitoring    deployment.apps/metrics-server                        1         1         1            1           12m
monitoring    deployment.apps/prometheus-operator                   1         1         1            1           12m
networking    deployment.apps/calico-typha                          0         0         0            0           12m
networking    deployment.apps/cert-manager                          1         1         1            1           12m
networking    deployment.apps/nginx-ingress-default-backend         1         1         1            1           12m
storage       deployment.apps/ceph-mgr                              1         1         1            1           12m
storage       deployment.apps/ceph-mon-single-node                  1         1         1            1           12m
storage       deployment.apps/ceph-osd-single-node                  1         1         1            1           12m
storage       deployment.apps/rbd-provisioner                       1         1         1            1           12m
wordpress     deployment.apps/mysql                                 1         1         1            1           5m
wordpress     deployment.apps/wordpress                             1         1         1            1           5m
NAMESPACE     NAME                                                             DESIRED   CURRENT   READY     AGE
backup        replicaset.apps/ark-68c56f6d75                                   1         1         1         5m
backup        replicaset.apps/minio-7895b9d495                                 1         1         1         5m
kube-system   replicaset.apps/coredns-646944c5c4                               2         2         2         12m
kube-system   replicaset.apps/heapster-heapster-6b7985754b                     1         1         1         7m
kube-system   replicaset.apps/heapster-heapster-7bb6d67b9d                     0         0         0         12m
kube-system   replicaset.apps/kubernetes-dashboard-845c9dbcdf                  1         1         1         12m
kube-system   replicaset.apps/tiller-deploy-759cb9df9                          1         1         1         12m
logging       replicaset.apps/cerebro-elasticsearch-cluster-567468c475         1         1         1         5m
logging       replicaset.apps/elasticsearch-operator-76769d959c                1         1         1         8m
logging       replicaset.apps/es-client-elasticsearch-cluster-7d6fb8dcdd       1         1         1         5m
logging       replicaset.apps/kibana-elasticsearch-cluster-66778d655d          1         1         1         5m
monitoring    replicaset.apps/kube-prometheus-exporter-kube-state-65c6c77579   1         1         1         8m
monitoring    replicaset.apps/kube-prometheus-grafana-749496574c               1         1         1         8m
monitoring    replicaset.apps/metrics-server-6486f65987                        1         1         1         12m
monitoring    replicaset.apps/prometheus-operator-6bc587f9fc                   1         1         1         12m
networking    replicaset.apps/calico-typha-679bd5b97f                          0         0         0         12m
networking    replicaset.apps/cert-manager-86b95f4dc8                          1         1         1         12m
networking    replicaset.apps/nginx-ingress-default-backend-789c7df7cb         1         1         1         12m
storage       replicaset.apps/ceph-mgr-54b46c94c4                              1         1         1         12m
storage       replicaset.apps/ceph-mon-single-node-8c98868fb                   1         1         1         12m
storage       replicaset.apps/ceph-osd-single-node-7b7f848b97                  1         1         1         12m
storage       replicaset.apps/rbd-provisioner-789795cf94                       1         1         1         12m
wordpress     replicaset.apps/mysql-65d54b75b4                                 1         1         1         5m
wordpress     replicaset.apps/wordpress-6b74676664                             1         1         1         5m
NAMESPACE    NAME                                                       DESIRED   CURRENT   AGE
logging      statefulset.apps/es-data-elasticsearch-cluster-default     1         1         5m
logging      statefulset.apps/es-master-elasticsearch-cluster-default   1         1         5m
monitoring   statefulset.apps/alertmanager-kube-prometheus              1         1         8m
monitoring   statefulset.apps/prometheus-kube-prometheus                1         1         8m
NAMESPACE   NAME                    DESIRED   SUCCESSFUL   AGE
backup      job.batch/minio-setup   1         1            5m
storage     job.batch/ceph-setup    1         1            12m
```

# Troubleshooting

k8s-tew enables logging for all components by default. The log files are stored in {base-directory}/var/log/k8s-tew/.

# Caveats

* The local setup uses for ingress the ports 80, 443 so they need to be free on the host. It also turns swapping off which is a requirement for kubelet.
* On CentOS nodes the firewall and SELinux are disabled to not interfere with Kubernetes.

# Feedback

* E-Mail: darxkies@gmail.com
