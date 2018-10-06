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

# Documentation

The project documentation can be found here: [https://darxkies.github.io/k8s-tew](https://darxkies.github.io/k8s-tew)

# Caveats

* The local setup uses for ingress the ports 80, 443 so they need to be free on the host. It also turns swapping off which is a requirement for kubelet.
* On CentOS nodes the firewall and SELinux are disabled to not interfere with Kubernetes.

# Feedback

* E-Mail: darxkies@gmail.com
