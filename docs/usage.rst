Usage
=====

All k8s-tew commands accept the argument :file:`--base-directory`, which defines where all the files (binaries, certificates, configurations and so on) will be stored. If no value is defined then it will create a subdirectory called "assets" in the working directory. Additionally, the environment variable K8S_TEW_BASE_DIRECTORY can be set to point to the assets directory instead of using :file:`--base-directory`.

To see all the commands and their arguments use the :file:`-h` argument.

To activate completion enter:

  .. code:: shell

    source <(k8s-tew completion)

Requirements
------------

k8s-tew was tested so far on Ubuntu 18.04 and CentOS 7.5. But it should be able to run on other Linux distributions.

Host related dependencies such as socat, conntrack, ipset and rbd are embedded in k8s-tew and put in place, once the cluster is running. Thus it is fairly portable to other Linux distributions.

Labels
------

k8s-tew uses labels to specify which files belong on a node, which commands should be executed on a node and also which components need to be started on a node. They are similar to Kubernetes' Labels and Selectors.

- bootstrapper - This label marks bootstrapping node/commands
- controller - Kubernetes Controler. At least three controller nodes are required for a HA cluster.
- worker - Kubernetes Node/Minion. At least one worker node is required.
- storage - The storage manager components are installed on the controller nodes. The worker nodes are used to store the actual data of the pods. If the storage label is omitted then all nodes are used. If you choose to use only some nodes for storage, then keep in mind that you need at least three storage managers and at least two data storage servers for a HA cluster.

Workflow
--------

To setup a local singl-node cluster, the workflow is: initialize -> configure -> generate -> run

For a remote cluster, the workflow is: initialize -> configure -> generate -> deploy

If something in one of the steps is changed, (e.g. configuration), then all the following steps have to be performed.

Initialization
^^^^^^^^^^^^^^

The first step in using k8s-tew is to create a config file. This is achieved by executing this command:

  .. code:: shell

    k8s-tew initialize

That command generates the config file called :file:`{base-directory}/etc/k8s-tew/config.yaml`. To overwrite the existing configuration use the argument -f.

Configuration
^^^^^^^^^^^^^^

After the initialization step the parameters of the cluster should be be adapted. These are the configure parameters and their defaults:

      --apiserver-port uint16                          API Server Port (default 6443)
      --ca-certificate-validity-period uint16          CA Certificate Validity Period (default 20)
      --calico-typha-ip string                         Calico Typha IP (default "10.32.0.5")
      --client-certificate-validity-period uint16      Client Certificate Validity Period (default 15)
      --cluster-cidr string                            Cluster CIDR (default "10.200.0.0/16")
      --cluster-dns-ip string                          Cluster DNS IP (default "10.32.0.10")
      --cluster-domain string                          Cluster domain (default "cluster.local")
      --cluster-ip-range string                        Cluster IP range (default "10.32.0.0/24")
      --cluster-name string                            Cluster Name used for Kubernetes Dashboard (default "k8s-tew")
      --controller-virtual-ip string                   Controller Virtual/Floating IP for the cluster
      --controller-virtual-ip-interface string         Controller Virtual/Floating IP interface for the cluster
      --deployment-directory string                    Deployment directory (default "/")
      --email string                                   Email address used for example for Let's Encrypt (default "k8s-tew@gmail.com")
      --ingress-domain string                          Ingress domain name (default "k8s-tew.net")
      --kubernetes-dashboard-port uint16               Kubernetes Dashboard Port (default 32443)
      --load-balancer-port uint16                      Load Balancer Port (default 32443)
      --metallb-addresses string                       Comma separated MetalLB address ranges and CIDR (e.g 192.168.0.16/28,192.168.0.75-192.168.0.100) (default "192.168.0.16/28")
      --public-network string                          Public Network (default "192.168.100.0/24")
      --resolv-conf string                             Custom resolv.conf (default "/etc/resolv.conf")
      --rsa-key-size uint16                            RSA Key Size (default 2048)
      --version-addon-resizer string                   Addon-Resizer version (default "k8s.gcr.io/addon-resizer:1.7")
      --version-busybox string                         Busybox version (default "docker.io/library/busybox:1.30.1")
      --version-calico-cni string                      Calico CNI version (default "quay.io/calico/cni:v3.6.0")
      --version-calico-kube-controllers string         Calico Kube Controllers  version (default "quay.io/calico/kube-controllers:v3.6.0")
      --version-calico-node string                     Calico Node version (default "quay.io/calico/node:v3.6.0")
      --version-calico-typha string                    Calico Typha version (default "quay.io/calico/typha:v3.6.0")
      --version-ceph string                            Ceph version (default "docker.io/ceph/daemon:v3.2.1-stable-3.2-mimic-centos-7-x86_64")
      --version-cerebro string                         Cerebro version (default "docker.io/upmcenterprises/cerebro:0.7.2")
      --version-cert-manager-controller string         Cert Manager Controller version (default "quay.io/jetstack/cert-manager-controller:v0.4.1")
      --version-configmap-reload string                ConfigMap Reload version (default "quay.io/coreos/configmap-reload:v0.0.1")
      --version-containerd string                      Containerd version (default "1.2.5")
      --version-coredns string                         CoreDNS version (default "docker.io/coredns/coredns:1.4.0")
      --version-crictl string                          CriCtl version (default "1.14.0")
      --version-csi-attacher string                    CSI Attacher version (default "quay.io/k8scsi/csi-attacher:v1.0.1")
      --version-csi-ceph-fs-plugin string              CSI Ceph FS Plugin version (default "quay.io/cephcsi/cephfsplugin:v1.0.0")
      --version-csi-ceph-rbd-plugin string             CSI Ceph RBD Plugin version (default "quay.io/cephcsi/rbdplugin:v1.0.0")
      --version-csi-ceph-snapshotter string            CSI Ceph Snapshotter version (default "quay.io/k8scsi/csi-snapshotter:v1.0.1")
      --version-csi-driver-registrar string            CSI Driver Registrar version (default "quay.io/k8scsi/csi-node-driver-registrar:v1.0.2")
      --version-csi-provisioner string                 CSI Provisioner version (default "quay.io/k8scsi/csi-provisioner:v1.0.1")
      --version-elasticsearch string                   Elasticsearch version (default "docker.io/upmcenterprises/docker-elasticsearch-kubernetes:6.1.3_0")
      --version-elasticsearch-cron string              Elasticsearch Cron version (default "docker.io/upmcenterprises/elasticsearch-cron:0.1.0")
      --version-elasticsearch-operator string          Elasticsearch Operator version (default "docker.io/upmcenterprises/elasticsearch-operator:0.3.0")
      --version-etcd string                            Etcd version (default "quay.io/coreos/etcd:v3.3.12")
      --version-fluent-bit string                      Fluent-Bit version (default "docker.io/fluent/fluent-bit:0.13.0")
      --version-gobetween string                       Gobetween version (default "docker.io/yyyar/gobetween:0.6.1")
      --version-grafana string                         Grafana version (default "docker.io/grafana/grafana:5.0.0")
      --version-grafana-watcher string                 Grafana Watcher version (default "quay.io/coreos/grafana-watcher:v0.0.8")
      --version-heapster string                        Heapster version (default "k8s.gcr.io/heapster:v1.3.0")
      --version-helm string                            Helm version (default "2.13.1")
      --version-k8s string                             Kubernetes version (default "k8s.gcr.io/hyperkube:v1.14.0")
      --version-kibana string                          Kibana version (default "docker.elastic.co/kibana/kibana-oss:6.1.3")
      --version-kube-state-metrics string              Kube State Metrics version (default "gcr.io/google_containers/kube-state-metrics:v1.2.0")
      --version-kubernetes-dashboard string            Kubernetes Dashboard version (default "k8s.gcr.io/kubernetes-dashboard-amd64:v1.10.1")
      --version-metallb-controller string              MetalLB Controller version (default "docker.io/metallb/controller:v0.7.3")
      --version-metallb-speaker string                 MetalLB Speaker version (default "docker.io/metallb/speaker:v0.7.3")
      --version-metrics-server string                  Metrics Server version (default "k8s.gcr.io/metrics-server-amd64:v0.3.1")
      --version-minio-client string                    Minio client version (default "docker.io/minio/mc:RELEASE.2019-03-20T21-29-03Z")
      --version-minio-server string                    Minio server version (default "docker.io/minio/minio:RELEASE.2019-03-20T22-38-47Z")
      --version-mysql string                           MySQL version (default "docker.io/library/mysql:5.6")
      --version-nginx-ingress-controller string        Nginx Ingress Controller version (default "quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.23.0")
      --version-nginx-ingress-default-backend string   Nginx Ingress Default Backend version (default "k8s.gcr.io/defaultbackend:1.4")
      --version-pause string                           Pause version (default "k8s.gcr.io/pause:3.1")
      --version-prometheus string                      Prometheus version (default "quay.io/prometheus/prometheus:v2.2.1")
      --version-prometheus-alert-manager string        Prometheus Alert Manager version (default "quay.io/prometheus/alertmanager:v0.15.1")
      --version-prometheus-config-reloader string      Prometheus Config Reloader version (default "quay.io/coreos/prometheus-config-reloader:v0.20.0")
      --version-prometheus-node-exporter string        Prometheus Node Exporter version (default "quay.io/prometheus/node-exporter:v0.15.2")
      --version-prometheus-operator string             Prometheus Operator version (default "quay.io/coreos/prometheus-operator:v0.20.0")
      --version-runc string                            Runc version (default "1.0.0-rc6")
      --version-velero string                          Velero version (default "gcr.io/heptio-images/velero:v0.11.0")
      --version-virtual-ip string                      Virtual-IP version (default "docker.io/darxkies/virtual-ip:0.1.4")
      --version-wordpress string                       WordPress version (default "docker.io/library/wordpress:4.8-apache")
      --vip-raft-controller-port uint16                VIP Raft Controller Port (default 16277)
      --vip-raft-worker-port uint16                    VIP Raft Worker Port (default 16728)
      --worker-virtual-ip string                       Worker Virtual/Floating IP for the cluster
      --worker-virtual-ip-interface string             Worker Virtual/Floating IP interface for the cluster

The email and the ingress-domain parameters need to be changed if you want a working Ingress and Lets' Encrypt configuration. It goes like this:

  .. code:: shell

    k8s-tew configure --email john.doe@gmail.com --ingress-domain example.com

Another important argument is :file:`--resolv-conf` which is used to define which resolv.conf file should be used for DNS.

The Virtual/Floating IP parameters should be accordingly changed if you want true HA. This is especially for the controllers important. Then if there are for example three controllers then the IP of the first controller is used by the whole cluster and if that one fails then the whole cluster will stop working. k8s-tew uses internally RAFT_ and its leader election functionality to select one node on which the Virtual IP is set. If the leader fails, one of the remaining nodes gets the Virtual IP assigned.

.. _RAFT: https://raft.github.io/ 


Add Remote Node
"""""""""""""""

A remote node can be added with the following command:

  .. code:: shell

    k8s-tew node-add -n controller00 -i 192.168.100.100 -x 0 -l controller

The arguments:

  -x, --index uint      The unique index of the node which should never be reused
  -i, --ip string       IP of the node (default "192.168.100.50")
  -l, --labels string   The labels of the node which define the attributes of the node (default "controller,worker")
  -n, --name string     The hostname of the node (default "single-node")


.. note:: Make sure the IP address of the node matches the public network set using the configuration argument :file:`--public-network`.

Add Local Node
""""""""""""""

k8s-tew is also able to start a cluster on the local computer and for that the local computer has to be added as a node:

  .. code:: shell

    k8s-tew node-add -s
    
The arguments:

  -s, --self            Add this machine by infering the host's name & ip and by setting the labels controller,worker,bootstrapper - The public-network and the deployment-directory are also updated

Remove Node
"""""""""""

A node can be removed like this:

  .. code:: shell

    k8s-tew node-remove -n controller00

List Nodes
""""""""""
  And all the nodes can be listed with the command:

  .. code:: shell

    k8s-tew node-list

Generating Files
^^^^^^^^^^^^^^^^

Once all the nodes were added, the required files (third party binares, certificates, kubeconfigs and so on) have to be put in place. And this goes like this:

  .. code:: shell

    k8s-tew generate

The arguments:

  -r, --command-retries uint   The count of command retries during the setup (default 300)
  --force-download         Force downloading all binary dependencies from the internet
  --parallel               Download binary dependencies in parallel
  --pull-images            Pull and convert images to OCI to be deployed later on

Run
^^^

With this command the local cluster can be started:

  .. code:: shell

    k8s-tew run

.. note:: This command will run in the foreground and it will supervise all the programs it started in the background. 

Deploy
^^^^^^

In case remote nodes were added with the deploy command, the remotely missing files are copied to the nodes. k8s-tew is installed and started as a service.

The deployment is executed with the command:

  .. code:: shell

    k8s-tew deploy

The arguments:

  -r, --command-retries uint    The count of command retries during the setup (default 300)
      --force-upload            Files are uploaded without checking if they are already installed
  -i, --identity-file string    SSH identity file (default "/home/darxkies/.ssh/id_rsa")
      --import-images           Install images
      --parallel                Run steps in parallel
      --skip-backup-setup       Skip backup setup
      --skip-ingress-setup      Skip ingress setup
      --skip-logging-setup      Skip logging setup
      --skip-monitoring-setup   Skip monitoring setup
      --skip-packaging-setup    Skip packaging setup
      --skip-setup              Skip setup steps
      --skip-showcase-setup     Skip showcase setup
      --skip-storage-setup      Skip storage setup and all other feature setup steps

The files are copied using scp and the ssh private key :file:`$HOME/.ssh/id_rsa`. In case the file :file:`$HOME/.ssh/id_rsa` does not exist it should be generated using the command :file:`ssh-keygen`. If another private key should be used, it can be specified using the command line argument :file:`-i`.

.. note:: The argument :file:`--pull-images` downloads the required Docker Images on the nodes, before the setup process is executed. That could speed up the whole setup process later on. Furthermore, by using :file:`--parallel` the process of uploading files to the nodes and the download of Docker Images can be again considerable shortened. Use these parameters with caution, as they can starve your network.


Environment
-----------

After starting the cluster, the user will need some environment variables set locally to make the interaction with the cluster easier. This is done with this command:

  .. code:: shell

      eval $(k8s-tew environment)

This command sets KUBECONFIG needed by kubectl to communicate with the cluster and it also updates PATH to point to the downloaded third-party binaries.

Services
--------

Depending on the configuration of the cluster, the installation of all containers can take a while. Once everything is installed, the following command can be used to open the web browser pointing to the web sites hosted by the cluster:

  .. code:: shell

    k8s-tew open-web-browser

The arguments:

      --all                    Open all websites
      --ceph-manager           Open Ceph Manager website
      --ceph-rados-gateway     Open Ceph Rados Gateway website
      --cerebro                Open Cerebro website
      --grafana                Open Grafana website
      --kibana                 Open Kibana website
      --kubernetes-dashboard   Open Kubernetes Dashboard website
      --minio                  Open Minio website
      --wordpress-ingress      Open WordPress Ingress website
      --wordpress-nodeport     Open WordPress NodePort website

.. note:: One of the parameters has to be used, otherwise no web site will be opened.

Alternatively, the web sites can be accessed manually.

Kubernetes Dashboard
^^^^^^^^^^^^^^^^^^^^

k8s-tew installs the Kubernetes Dashboard. Invoke the following command to display the admin token:

  .. code:: shell

    k8s-tew dashboard

If you have a GUI web browser installed, then you can use the following command to display the admin token for three seconds, enough time to copy the token, and to also open the web browser:

  .. code:: shell

    k8s-tew dashboard -o

.. note:: It takes minutes to actually download the dashboard. Use the following command to check the status of the pods:

  .. code:: shell

    kubectl get pods -n kube-system

Once the pod is running the dashboard can be accessed through the TCP port 32443. Regarding the IP address, use the IP address of a worker node or the worker Virtual IP if one was specified.

When asked to login, enter the admin token.


Ingress
^^^^^^^

For working Ingress make sure ports 80 and 443 are available. The Ingress Domain have to be also configured before 'generate' and 'deploy' are executed:

  .. code:: shell

    k8s-tew configure --ingress-domain [ingress-domain]

WordPress
^^^^^^^^^

Wordpress/MySQL are installed for testing purposes and [ingress-domain] can be set using the configure command.

- Address: http://[worker-ip]:30100
- Address: https://wordpress.[ingress-domain]


Minio
^^^^^

Minio is used by Ark to store the backups.

- Address: http://[worker-ip]:30800
- Username: minio
- Password: changeme

Grafana
^^^^^^^

Grafana provides an overview of the cluster's status.

- Address: http://[worker-ip]:30900
- Username: admin
- Password: changeme

Kibana
^^^^^^^

Kibana can be used to inspect the log messages of all pods in the cluster.

- Address: https://[worker-ip]:30980

Cerebro
^^^^^^^

Cerebro allows the user to manage the collected log message from the cluster.

- Address: http://[worker-ip]:30990

Ceph Dashboard
^^^^^^^^^^^^^^

Ceph Dashboard gives an overview of the storage status.

- Address: https://[worker-ip]:30700
- Username: admin
- Password: changeme

