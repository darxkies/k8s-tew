Features
========

* HA cluster setup passes all CNCF conformance tests (Kubernetes `1.10 <https://github.com/cncf/k8s-conformance/tree/master/v1.10/k8s-tew>`_, `1.11 <https://github.com/cncf/k8s-conformance/tree/master/v1.11/k8s-tew>`_, `1.12 <https://github.com/cncf/k8s-conformance/tree/master/v1.12/k8s-tew>`_, `1.13 <https://github.com/cncf/k8s-conformance/tree/master/v1.13/k8s-tew>`_, `1.14 <https://github.com/cncf/k8s-conformance/tree/master/v1.14/k8s-tew>`_, `1.15 <https://github.com/cncf/k8s-conformance/tree/master/v1.15/k8s-tew>`_, `1.16 <https://github.com/cncf/k8s-conformance/tree/master/v1.16/k8s-tew>`_, `1.17 <https://github.com/cncf/k8s-conformance/tree/master/v1.17/k8s-tew>`_, `1.18 <https://github.com/cncf/k8s-conformance/tree/master/v1.18/k8s-tew>`_, `1.19 <https://github.com/cncf/k8s-conformance/tree/master/v1.19/k8s-tew>`_, `1.20 <https://github.com/cncf/k8s-conformance/tree/master/v1.20/k8s-tew>`_ & `1.21 <https://github.com/cncf/k8s-conformance/tree/master/v1.21/k8s-tew>`_)
* Container Management: `Containerd <https://containerd.io/>`_
* Networking: `Calico <https://www.projectcalico.org>`_
* Ingress: `NGINX Ingress <https://kubernetes.github.io/ingress-nginx/>`_ and `cert-manager <http://docs.cert-manager.io/en/latest/>`_ for `Let's Encrypt <https://letsencrypt.org/>`_
* Storage: `Ceph/RBD <https://ceph.com/>`_
* Metrics: `metering-metrics <https://github.com/kubernetes-incubator/metrics-server>`_ and `Heapster <https://github.com/kubernetes/heapster>`_
* Monitoring: `Prometheus <https://prometheus.io/>`_ and `Grafana <https://grafana.com/>`_
* Logging: `Fluent-Bit <https://fluentbit.io/>`_, `Elasticsearch <https://www.elastic.co/>`_, `Kibana <https://www.elastic.co/products/kibana>`_ and `Cerebro <https://github.com/lmenezes/cerebro>`_
* Backups: `Velero <https://github.com/heptio/velero>`_, `Restic <https://restic.net/>`_ and `Minio <https://www.minio.io/>`_
* Cluster Load Balancing: `MetalLB <https://metallb.universe.tf>`_
* Controller Load Balancing: `gobetween <http://gobetween.io/>`_
* Package Manager: `Helm <https://helm.sh/>`_
* Dashboard: `Kubernetes Dashboard <https://github.com/kubernetes/dashboard>`_
* The communication between the components is encrypted
* RBAC is enabled
* The controllers and the workers have Floating/Virtual IPs
* Integrated Load Balancer for the API Servers
* Support for deployment to a HA cluster using ssh
* Only the changed files are deployed
* No `Docker <https://www.docker.com/>`_ installation required
* No cloud provider required
* Single binary without any dependencies
* Runs locally
* Nodes management from the command line
* Downloads all the used binaries (kubernetes, calico, ceph...) from the Internet
* Pull Images, Convert them to OCI and import them on the cluster for offline installations
* Uses systemd to install itself as a service on the remote machine
* Installs `WordPress <https://wordpress.com>`_ and `MySQL <https://www.mysql.com>`_ to test drive the installation

