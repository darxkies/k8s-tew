<p align="center"><img src="logo.svg" width="360"></p>

<p align="center"><a href="https://github.com/cncf/k8s-conformance/tree/master/v1.20/k8s-tew"><img src="conformance/certified-kubernetes-color.svg"  alt="Kubernetes v1.20" width="120"></a></p>

# Kubernetes - The Easier Way (k8s-tew)

[![Build Status](https://travis-ci.com/darxkies/k8s-tew.svg?branch=master)](https://travis-ci.com/darxkies/k8s-tew)
[![Go Report Card](https://goreportcard.com/badge/github.com/darxkies/k8s-tew)](https://goreportcard.com/report/github.com/darxkies/k8s-tew)
[![GitHub release](https://img.shields.io/github/tag/darxkies/k8s-tew.svg)](https://github.com/darxkies/k8s-tew/releases/latest)
![GitHub](https://img.shields.io/github/license/darxkies/k8s-tew.svg)


k8s-tew is a CLI tool to install a [Kubernetes](https://kubernetes.io/) Cluster (local, single-node, multi-node or HA-cluster) on Bare Metal. It installs the most essential components needed by a cluster such as networking, storage, monitoring, logging, backuping/restoring and so on. Besides that, k8s-tew is also a supervisor that starts all cluster components on each node, once it setup the nodes.

## TL;DR

[![k8s-tew](https://img.youtube.com/vi/53qQa5EkBTU/0.jpg)](https://www.youtube.com/watch?v=53qQa5EkBTU)

# Documentation

The project documentation can be found here: [https://darxkies.github.io/k8s-tew](https://darxkies.github.io/k8s-tew)

# Caveats

* The local setup uses for ingress the ports 80, 443 so they need to be free on the host. It also turns swapping off which is a requirement for kubelet.
* On CentOS nodes the firewall and SELinux are disabled to not interfere with Kubernetes.

# Feedback

* E-Mail: darxkies@gmail.com
