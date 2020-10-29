Usage
=====

All k8s-tew commands accept the argument :file:`--base-directory`, which defines where all the files (binaries, certificates, configurations and so on) will be stored. If no value is defined then it will create a subdirectory called "assets" in the working directory. Additionally, the environment variable K8S_TEW_BASE_DIRECTORY can be set to point to the assets directory instead of using :file:`--base-directory`.

To see all the commands and their arguments use the :file:`-h` argument.

To activate completion for bash enter:

  .. code:: shell

    source <(k8s-tew completion)

For zsh run:

  .. code:: shell

    source <(k8s-tew completion zsh)

Requirements
------------

k8s-tew was tested so far on Ubuntu 20.04 and CentOS 8.2. But it should be able to run on other Linux distributions.

Host related dependencies such as socat, conntrack, ipset and rbd are embedded in k8s-tew and put in place, once the cluster is running. Thus it is fairly portable to other Linux distributions.

Labels
------

k8s-tew uses labels to specify which files belong on a node, which commands should be executed on a node and also which components need to be started on a node. They are similar to Kubernetes' Labels and Selectors.

- bootstrapper - This label marks bootstrapping node/commands
- controller - Kubernetes Controler. At least three controller nodes are required for a HA cluster.
- worker - Kubernetes Node/Minion. At least one worker node is required.
- storage - The Ceph Monitors are installed on the controller nodes. The storage nodes are used to store the actual data of the pods, and also to run the containers related to Ceph and CSI. 

Workflow
--------

To setup a local single-node cluster, the workflow is: initialize -> configure -> generate -> run

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


      --alert-manager-count uint16                       Number of Alert Manager Servers (default 1)
      --alert-manager-size uint16                        Size of Alert Manager Persistent Volume (default 2)
      --apiserver-port uint16                            API Server Port (default 6443)
      --ca-certificate-validity-period uint16            CA Certificate Validity Period (default 20)
      --calico-typha-ip string                           Calico Typha IP (default "10.32.0.5")
      --ceph-cluster-name string                         Ceph Cluster Name (default "ceph")
      --ceph-expected-number-of-objects uint             Ceph Expected Number of Objects (default 1000000)
      --ceph-placement-groups uint                       Ceph Placement Groups (default 256)
      --client-certificate-validity-period uint16        Client Certificate Validity Period (default 15)
      --cluster-cidr string                              Cluster CIDR (default "10.200.0.0/16")
      --cluster-dns-ip string                            Cluster DNS IP (default "10.32.0.10")
      --cluster-domain string                            Cluster domain (default "cluster.local")
      --cluster-ip-range string                          Cluster IP range (default "10.32.0.0/24")
      --cluster-name string                              Cluster Name used for Kubernetes Dashboard (default "k8s-tew")
      --controller-virtual-ip string                     Controller Virtual/Floating IP for the cluster
      --controller-virtual-ip-interface string           Controller Virtual/Floating IP interface for the cluster
      --deployment-directory string                      Deployment directory (default "/")
      --elasticsearch-count uint16                       Number of Elasticsearch Servers (default 1)
      --elasticsearch-size uint16                        Size of Elasticsearch Persistent Volume (default 10)
      --email string                                     Email address used for example for Let's Encrypt (default "k8s-tew@gmail.com")
      --grafana-size uint16                              Size of Grafana Persistent Volume (default 2)
      --help                                             help for configure
      --ingress-domain string                            Ingress domain name (default "k8s-tew.net")
      --kube-state-metrics-count uint16                  Number of Kube State Metrics Servers (default 1)
      --kubernetes-dashboard-port uint16                 Kubernetes Dashboard Port (default 32443)
      --load-balancer-port uint16                        Load Balancer Port (default 32443)
      --max-pods uint16                                  MaxPods (default 110)
      --metallb-addresses string                         Comma separated MetalLB address ranges and CIDR (e.g 192.168.0.16/28,192.168.0.75-192.168.0.100) (default "192.168.0.16/28")
      --minio-size uint16                                Size of Minio Persistent Volume (default 2)
      --prometheus-size uint16                           Size of Prometheus Persistent Volume (default 2)
      --public-network string                            Public Network (default "192.168.100.0/24")
      --resolv-conf string                               Custom resolv.conf (default "/etc/resolv.conf")
      --rsa-key-size uint16                              RSA Key Size (default 2048)
      --san-dns-names string                             SAN DNS Names (comma separated)
      --san-ip-addresses string                          SAN IP Addresses (comma separated)
      --version-alert-manager string                     Alert Manager version (default "quay.io/prometheus/alertmanager:v0.21.0")
      --version-busybox string                           Busybox version (default "docker.io/library/busybox:1.32.0")
      --version-calico-cni string                        Calico CNI version (default "quay.io/calico/cni:v3.16.4")
      --version-calico-kube-controllers string           Calico Kube Controllers  version (default "quay.io/calico/kube-controllers:v3.16.4")
      --version-calico-node string                       Calico Node version (default "quay.io/calico/node:v3.16.4")
      --version-calico-pod2daemon string                 Calico Pod2Daemon version (default "quay.io/calico/pod2daemon-flexvol:v3.16.4")
      --version-calico-typha string                      Calico Typha version (default "quay.io/calico/typha:v3.16.4")
      --version-ceph string                              Ceph version (default "docker.io/ceph/ceph:v15.2.5")
      --version-cerebro string                           Cerebro version (default "docker.io/lmenezes/cerebro:0.9.2")
      --version-cert-manager-cainjector string           Cert Manager CA Injector version (default "quay.io/jetstack/cert-manager-cainjector:v1.0.3")
      --version-cert-manager-controller string           Cert Manager Controller version (default "quay.io/jetstack/cert-manager-controller:v1.0.3")
      --version-cert-manager-webhook string              Cert Manager Web Hook version (default "quay.io/jetstack/cert-manager-webhook:v1.0.3")
      --version-containerd string                        Containerd version (default "1.4.1")
      --version-coredns string                           CoreDNS version (default "docker.io/coredns/coredns:1.8.0")
      --version-crictl string                            CriCtl version (default "1.18.0")
      --version-csi-attacher string                      CSI Attacher version (default "quay.io/k8scsi/csi-attacher:v2.1.1")
      --version-csi-ceph-plugin string                   CSI Ceph Plugin version (default "quay.io/cephcsi/cephcsi:v3.1.1")
      --version-csi-driver-registrar string              CSI Driver Registrar version (default "quay.io/k8scsi/csi-node-driver-registrar:v1.3.0")
      --version-csi-provisioner string                   CSI Provisioner version (default "quay.io/k8scsi/csi-provisioner:v1.6.0")
      --version-csi-resizer string                       CSI Resizer version (default "quay.io/k8scsi/csi-resizer:v0.5.0")
      --version-csi-snapshot-controller string           CSI Snapshot Controller  version (default "quay.io/k8scsi/snapshot-controller:v2.0.1")
      --version-csi-snapshotter string                   CSI Snapshotter version (default "quay.io/k8scsi/csi-snapshotter:v2.1.1")
      --version-elasticsearch string                     Elasticsearch version (default "docker.elastic.co/elasticsearch/elasticsearch:7.9.2")
      --version-etcd string                              Etcd version (default "quay.io/coreos/etcd:v3.4.13")
      --version-fluent-bit string                        Fluent-Bit version (default "docker.io/fluent/fluent-bit:1.6.1")
      --version-gobetween string                         Gobetween version (default "docker.io/yyyar/gobetween:0.8.0")
      --version-grafana string                           Grafana version (default "docker.io/grafana/grafana:7.2.2")
      --version-helm string                              Helm version (default "3.4.0")
      --version-k8s string                               Kubernetes version (default "v1.18.10")
      --version-kibana string                            Kibana version (default "docker.elastic.co/kibana/kibana:7.9.2")
      --version-kube-apiserver string                    Kubernetes API Server version (default "k8s.gcr.io/kube-apiserver:v1.18.10")
      --version-kube-controller-manager string           Kubernetes Controller Manager (default "k8s.gcr.io/kube-controller-manager:v1.18.10")
      --version-kube-proxy string                        Kubernetes Proxy version (default "k8s.gcr.io/kube-proxy:v1.18.10")
      --version-kube-scheduler string                    Kubernetes Scheduler (default "k8s.gcr.io/kube-scheduler:v1.18.10")
      --version-kube-state-metrics string                Kube State Metrics version (default "quay.io/coreos/kube-state-metrics:v1.9.7")
      --version-kubernetes-dashboard string              Kubernetes Dashboard version (default "docker.io/kubernetesui/dashboard:v2.0.4")
      --version-metallb-controller string                MetalLB Controller version (default "docker.io/metallb/controller:v0.9.4")
      --version-metallb-speaker string                   MetalLB Speaker version (default "docker.io/metallb/speaker:v0.9.4")
      --version-metrics-server string                    Metrics Server version (default "k8s.gcr.io/metrics-server/metrics-server:v0.3.7")
      --version-minio-client string                      Minio client version (default "docker.io/minio/mc:RELEASE.2020-10-03T02-54-56Z")
      --version-minio-server string                      Minio server version (default "docker.io/minio/minio:RELEASE.2020-10-18T21-54-12Z")
      --version-mysql string                             MySQL version (default "docker.io/library/mysql:8.0.19")
      --version-nginx-ingress-admission-webhook string   Nginx Ingress Admission Webhook version (default "docker.io/jettech/kube-webhook-certgen:v1.3.0")
      --version-nginx-ingress-controller string          Nginx Ingress Controller version (default "k8s.gcr.io/ingress-nginx/controller:v0.40.2")
      --version-node-exporter string                     Node Exporter version (default "quay.io/prometheus/node-exporter:v1.0.1")
      --version-pause string                             Pause version (default "k8s.gcr.io/pause:3.3")
      --version-prometheus string                        Prometheus version (default "quay.io/prometheus/prometheus:v2.22.0")
      --version-runc string                              Runc version (default "1.0.0-rc92")
      --version-velero string                            Velero version (default "docker.io/velero/velero:v1.5.2")
      --version-velero-plugin-aws string                 Velero Plugin AWS version (default "docker.io/velero/velero-plugin-for-aws:v1.1.0")
      --version-velero-plugin-csi string                 Velero Plugin CSI version (default "docker.io/velero/velero-plugin-for-csi:v0.1.2")
      --version-velero-restic-restore-helper string      Velero Restic Restore Helper (default "docker.io/velero/velero-restic-restore-helper:v1.5.2")
      --version-virtual-ip string                        Virtual-IP version (default "docker.io/darxkies/virtual-ip:0.1.4")
      --version-wordpress string                         WordPress version (default "docker.io/library/wordpress:5.4-apache")
      --vip-raft-controller-port uint16                  VIP Raft Controller Port (default 16277)
      --vip-raft-worker-port uint16                      VIP Raft Worker Port (default 16728)
      --worker-virtual-ip string                         Worker Virtual/Floating IP for the cluster
      --worker-virtual-ip-interface string               Worker Virtual/Floating IP interface for the cluster

The email and the ingress-domain parameters need to be changed if you want a working Ingress and Lets' Encrypt configuration. It goes like this:

  .. code:: shell

    k8s-tew configure --email john.doe@gmail.com --ingress-domain example.com

Another important argument is :file:`--resolv-conf` which is used to define which resolv.conf file should be used for DNS.

The Virtual/Floating IP parameters should be accordingly changed if you want true HA. k8s-tew uses internally RAFT_ and its leader election functionality to select one node on which the Virtual IP is set. If the leader fails, one of the remaining nodes gets the Virtual IP assigned.

.. _RAFT: https://raft.github.io/ 


Add Remote Node
"""""""""""""""

A remote node can be added with the following command:

  .. code:: shell

    k8s-tew node-add -n controller00 -i 192.168.100.100 -l controller,node,storage

The arguments:

  -x, --index uint           The unique index of the node which should never be reused; if it is already in use a new one is assigned
  -i, --ip string            IP of the node (default "192.168.100.50")
  -l, --labels string        The labels of the node which define the attributes of the node (default "controller,worker")
  -n, --name string          The hostname of the node (default "single-node")
  -s, --self                 Add this machine by inferring the host's name & IP and by setting the labels controller,worker,bootstrapper - The public-network and the deployment-directory are also updated
  -r, --storage-index uint   The unique index of the storage node which should never be reused; if it is already in use a new one is assigned


.. note:: Make sure the IP address of the node matches the public network set using the configuration argument :file:`--public-network`.

Add Local Node
""""""""""""""

k8s-tew is also able to start a cluster on the local computer and for that the local computer has to be added as a node:

  .. code:: shell

    k8s-tew node-add -s
    
The arguments:

  -s, --self            Add this machine by inferring the host's name & IP and by setting the labels controller,worker,bootstrapper - The public-network and the deployment-directory are also updated

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

  -r, --command-retries uint    The number of command retries during the setup (default 1200)
  --force-upload            Files are uploaded without checking if they are already installed
  -h, --help                    help for deploy
  -i, --identity-file string    SSH identity file (default "/home/darxkies/.ssh/id_rsa")
  --import-images           Install images
  --parallel                Run steps in parallel
  --skip-backup-setup       Skip backup setup
  --skip-ingress-setup      Skip ingress setup
  --skip-logging-setup      Skip logging setup
  --skip-monitoring-setup   Skip monitoring setup
  --skip-restart            Skip restart steps
  --skip-setup              Skip setup steps
  --skip-showcase-setup     Skip showcase setup
  --skip-storage-setup      Skip storage setup and all other feature setup steps
  --skip-upload             Skip upload steps
  --wait uint               Wait for all cluster relevant pods to be ready and jobs to be completed. The parameter reflects the number of seconds in which the pods have to run stable.

The files are copied using scp and the ssh private key :file:`$HOME/.ssh/id_rsa`. In case the file :file:`$HOME/.ssh/id_rsa` does not exist it should be generated using the command :file:`ssh-keygen`. If another private key should be used, it can be specified using the command line argument :file:`-i`.

.. note:: The argument :file:`--pull-images` downloads the required Docker Images on the nodes, before the setup process is executed. That could speed up the whole setup process later on. Furthermore, by using :file:`--parallel` the process of uploading files to the nodes and the download of Docker Images can be again considerable shortened. Use these parameters with caution, as they can starve your network.


Environment
-----------

After starting the cluster, the user will need some environment variables set locally to make the interaction with the cluster easier. This is done with this command:

  .. code:: shell

      source <(k8s-tew environment)

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

- Address: http://[worker-ip]:30980

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

