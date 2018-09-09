<p align="center"><img src="logo.svg" width="360"></p>

<p align="center"><a href="https://github.com/cncf/k8s-conformance/tree/master/v1.10/k8s-tew"><img src="conformance/certified-kubernetes-1.10-color.svg" width="120"></a><a href="https://github.com/cncf/k8s-conformance/tree/master/v1.11/k8s-tew"><img src="conformance/certified-kubernetes-1.11-color.svg" width="120"></a></p>

# Kubernetes - The Easier Way (k8s-tew)

[![Build Status](https://travis-ci.org/darxkies/k8s-tew.svg?branch=master)](https://travis-ci.org/darxkies/k8s-tew)
[![Go Report Card](https://goreportcard.com/badge/github.com/darxkies/k8s-tew)](https://goreportcard.com/report/github.com/darxkies/k8s-tew)
![GitHub](https://img.shields.io/github/license/darxkies/k8s-tew.svg)


k8s-tew is a CLI tool to install a [Kubernetes](https://kubernetes.io/) Cluster (local, single-node, multi-node or HA-cluster) on Bare Metal. It installs the most essential components needed by a cluster such as networking, storage, monitoring, logging, backuping/restoring and so on. Besides that, k8s-tew is also a supervisor that starts all cluster components on each node, once it setup the nodes.

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

# Troubleshooting

k8s-tew enables logging for all components by default. The log files are stored in {base-directory}/var/log/k8s-tew/.

# Caveats

* The local setup uses for ingress the ports 80, 443 so they need to be free on the host. It also turns swapping off which is a requirement for kubelet.
* On CentOS nodes the firewall and SELinux are disabled to not interfere with Kubernetes.

# Feedback

* E-Mail: darxkies@gmail.com
