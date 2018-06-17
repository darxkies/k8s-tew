<p align="center"><img src="logo.svg" width="360"></p>

# Kubernetes - The Easier Way (k8s-tew)

[Kubernetes](https://kubernetes.io/) is a fairly complex project. For a newbie it is hard to understand and also to use. While [Kelsey Hightower's Kubernetes The Hard Way](https://github.com/kelseyhightower/kubernetes-the-hard-way), on which this project is based, helps a lot to understand Kubernetes, it is optimized for the use with Google Cloud Platform.

This project's aim is to give newbies a tool that allows them to easily tinker with Kubernetes. k8s-tew is a CLI tool to generate the configuration for a Kubernetes cluster (local single node or remote multi node with support for HA). Besides that, k8s-tew is also a deployment tools and also a supervisor that starts all cluster components.

# Requirements

k8s-tew was tested so far on Ubuntu 18.04 and CentOS 7. But it should be able to run on other Linux distributions.

# Features

* Multi node setup (Ubuntu 18.04 / Kubernetes 1.10.4 / 3 Controllers / 2 Workers) passes all CNCF conformance tests
* Kubernetes Dashboard and Helm installed
* The communication between the components is encrypted
* RBAC is enabled
* The controllers and the workers have each a Virtual IP
* Integrated Load Balancer for the API Servers
* Support for deployment to a HA cluster using ssh
* Only the changed files are deployed
* No docker installation required (uses containerd)
* No cloud provider required
* Single binary without any dependencies
* Runs locally
* Nodes management from the command line
* Downloads all the used binaries (kubernetes, etcd, flanneld...) from the Internet
* Lower storage and RAM footprint compared to other solutions (kubespray, kubeadm, minikube...)
* Uses systemd to install itself as a service on the remote machine
* All components are written in Go

# Install

## From binary

The 64-bit binary can be downloaded from the following address: https://github.com/darxkies/k8s-tew/releases

Additionally the these commands can be used to download it and install it in /usr/local/bin

```shell
curl -s https://api.github.com/repos/darxkies/k8s-tew/releases/latest | grep "browser_download_url" | cut -d : -f 2,3 | tr -d \" | sudo wget -O /usr/local/bin/k8s-tew -qi - && sudo chmod a+x /usr/local/bin/k8s-tew
```

## From source

To compile it from source you will need a Go (version 1.10+) environment and Git installed. Once Go is configured, enter the following commands:

```shell
export GOPATH=~/go
mkdir $GOPATH
go get github.com/darxkies/k8s-tew
cd ~/go/src/github.com/darxkies/k8s-tew
make
sudo mv ~/go/bin/k8s-tew /usr/local/bin
```

The k8s-tew binary is moved to directory /usr/local/bin.

# Usage

This section will assume that k8s-tew is already in the folder /usr/local/bin, which should be defined in $PATH and that the commands are executed using root privileges.

All k8s-tew commands accept the argument --base-directory, which defines where all the files will be stored. If no value is defined then it will create a subdirectory called "assets" in the working directory. Additionally, the environment variable K8S_TEW_BASE_DIRECTORY can be set to point to the assets directory instead of using --base-directory.

To see all the commands and their arguments use the -h argument.

## Initialization

The first step in using k8s-tew is to create a config file. This is achieved by executing this command:

```shell
k8s-tew initialize
```

That command generates the config file called artifacts/etc/k8s-tew/config.yaml. To overwrite the existing configuration use the argument -f.

## Configuration

The Virtual IP functionality has to be first specified otherwise it is disabled:

* --controller-virtual-ip - Controller Virtual IP for the cluster
* --controller-virtual-ip-interface - Controller Virtual IP interface for the cluster
* --worker-virtual-ip - Worker Virtual IP for the cluster
* --worker-virtual-ip-interface - Worker Virtual IP interface for the cluster

Another important argument is --resolv-conf which is used to define which resolv.conf file should be used for DNS.

## Nodes

The configuration has no nodes defined yet. A remote node can be added with the following command:

```shell
k8s-tew node-add -n controller00 -i 192.168.100.100 -x 0 -l controller
```

The arguments:

* -n - the name of the node. This name has to match the hostname of that node.
* -i - the ip of the node
* -x - each node needs a unique number
* -l - the role of the node in the cluster: controller and/or worker and/or bootstrapper

The pods are are created on the worker nodes. The bootstrapper label/role is used for local setups.

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

* --k8s-version - Kubernetes version (default "1.10.4")
* --runc-version - runc version (default "1.0.0-rc5")
* --cni-version - CNI version (default "0.6.0")
* --containerd-version - containerd version (default "1.1.0")
* --crictl-version - crictl version (default "1.0.0-beta.0")
* --etcd-version - Etcd version (default "3.3.7")
* --flanneld-version - Flanneld version (default "0.10.0")
* --gobetween-version - gobetween version (default "0.5.0")
* --helm-version - helm version (default "2.9.1")

The command generate also provides the argument --deployment-directory to specify the target directory to which the assets will be deployed remotely.

## Run

With this command the local cluster can be started:

```shell
k8s-tew run
```

## Deploy

In case remote nodes were added with the deploy command, the missing files are copied to the nodes, k8s-tew is installed and started as a service.

The deployment is executed with the command:

```shell
k8s-tew deploy
```

The files are copied using scp and the ssh private key $HOME/.ssh/id_rsa. In case the file  $HOME/.ssh/id_rsa does not exist it should be generated using the command ssh-keygen. If another private key should be used, it can be specified using the command line argument -i.

## Environment

After starting the cluster, the user will need some environment variables set locally to make the interaction with the cluster easier. This is done with this command:

```shell
eval $(k8s-tew environment)
```

## Kubernetes Dashboard

k8s-tew also installs the Kubernetes Dashboard. To access it, the token of the admin user has to be retrieved:

```shell
kubectl -n kube-system describe secret $(kubectl -n kube-system get secret | grep admin-user | awk '{print $1}') | grep token: | awk '{print $2}'
```

__NOTE__: It takes minutes to actually download dashboard. Use the following command to check the status of the pods:

```shell
kubectl get pods -n kube-system
```

Once the pod is running the dashboard can be accessed through the TCP port 32443. Regarding the IP address, use the IP address of a worker node or the worker Virtual IP if one was specified.

When asked to login, enter the token from the first step.

## Labels

The k8s-tew labels are very similar to the Kubernetes' Labels and Selectors. The difference is that k8s-tew uses no keys. The purpose of the labels is to specify which files belong on a node, which commands should be executed on a node and also which components need to be started on a node.

# Setup

Vagrant/VirtualBox can be used to test drive k8s-tew. The host is used to bootstrap the cluster which runs in VirtualBox. The Vagrantfile included in the repository can be used for single-node/multi-node & Ubuntu 18.04/CentOS 7 setups.

The Vagrantfile can be configured using the environment variables:

* OS - define the operating system. It accepts ubuntu and centos.
* MULTI_NODE - if set then a HA cluster is generated. Otherwise a single-node setup is used.
* CONTROLLERS - defines the number of controller nodes. The default number is 3.
* WORKERS - specifies the number of worker nodes. The default number is 2.

__NOTE__: The multi-node setup with the default settings needs about 16GB RAM for itself.

## Ubuntu Single-Node

Steps to generate a single-node / Ubuntu cluster:

```shell
# Create the single-node VM
vagrant destroy single-node -f
OS=ubuntu vagrant up

# Create the assets and deploy them
k8s-tew initialize -f
k8s-tew configure --controller-virtual-ip=192.168.100.10 --controller-virtual-ip-interface=enp0s8 --worker-virtual-ip=192.168.100.20 --worker-virtual-ip-interface=enp0s8 --resolv-conf=/run/systemd/resolve/resolv.conf
k8s-tew node-add -n single-node -i 192.168.100.50 -x 0 -l controller,worker
k8s-tew generate --deployment-directory=/
k8s-tew deploy

# Setup local environment to execute kubectl, helm, etcdcl and so on
eval $(k8s-tew environment)

# Get token for Kubernetes Dashboard
kubectl -n kube-system describe secret $(kubectl -n kube-system get secret | grep admin-user | awk '{print $1}') | grep token: | awk '{print $2}'

# Access Kubernetes dashboard
xdg-open https://192.168.100.20:32443
```

## Ubuntu Multi-Node

Steps to generate a multi-node / Ubuntu cluster:

```shell
# Create the multi-node VM (3 controllers & 2 workers)
OS=ubuntu MULTI_NODE=true vagrant destroy -f
OS=ubuntu MULTI_NODE=true vagrant up

# Create the assets and deploy them
k8s-tew initialize -f
k8s-tew configure --controller-virtual-ip=192.168.100.10 --controller-virtual-ip-interface=enp0s8 --worker-virtual-ip=192.168.100.20 --worker-virtual-ip-interface=enp0s8 --resolv-conf=/run/systemd/resolve/resolv.conf
k8s-tew node-add -n controller00 -i 192.168.100.100 -x 0 -l controller
k8s-tew node-add -n controller01 -i 192.168.100.101 -x 1 -l controller
k8s-tew node-add -n controller02 -i 192.168.100.102 -x 2 -l controller
k8s-tew node-add -n worker00 -i 192.168.100.200 -x 3 -l worker
k8s-tew node-add -n worker01 -i 192.168.100.201 -x 4 -l worker
k8s-tew generate --deployment-directory=/
k8s-tew deploy

# Setup local environment to execute kubectl, helm, etcdcl and so on
eval $(k8s-tew environment)

# Get token for Kubernetes Dashboard
kubectl -n kube-system describe secret $(kubectl -n kube-system get secret | grep admin-user | awk '{print $1}') | grep token: | awk '{print $2}'

# Access Kubernetes dashboard
xdg-open https://192.168.100.20:32443
```

## Local

This steps can be used to get k8s-tew to run locally without any virtual machine.

```shell
k8s-tew initialize -f
k8s-tew node-add -s
k8s-tew generate
sudo bin/k8s-tew run
```

__NOTE__: To access Kuberntes Dashboard use the internal IP address and 127.0.0.1/localhost. Depending on the hardware used, it might take a while until it starts and setups everything.

# Troubleshooting

k8s-tew enables logging for all components by default. The log files are stored in {base-directory}/var/log/k8s-tew/.

# Caveats

* k8s-tew needs root privileges to be executed. Thus, it should be executed on a virtual machine or in a Docker container to generate the assets and to deploy the cluster.

# Feedback

* Gmail: darxkies@gmail.com
