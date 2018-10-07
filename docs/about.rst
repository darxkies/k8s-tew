About
=====

k8s-tew_ is a CLI tool to install a Kubernetes Cluster (local, single-node, multi-node or HA-cluster) on Bare Metal. It installs the most essential components needed by a cluster such as networking, storage, monitoring, logging, backuping/restoring and so on. Besides that, k8s-tew is also a supervisor that starts all cluster components on each node, once it setup the nodes.

.. _k8s-tew: https://github.com/darxkies/k8s-tew

Why
---

Kubernetes_ is a fairly complex project. For a newbie it is hard to understand and also to use. While `Kelsey Hightower's Kubernetes The Hard Way <https://github.com/kelseyhightower/kubernetes-the-hard-way>`_, on which this project is based, helps a lot to understand Kubernetes, it is optimized for the use with Google Cloud Platform.

Thus, this project's aim is to give newbies an easy to use tool that allows them to tinker with Kubernetes and later on to install HA production grade clusters.

.. _Kubernetes: https://kubernetes.io/

