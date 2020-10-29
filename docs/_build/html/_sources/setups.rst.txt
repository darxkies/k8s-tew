Setups
======

Vagrant
-------

Vagrant/VirtualBox can be used to test drive k8s-tew. The host is used to bootstrap the cluster which runs in VirtualBox. The Vagrantfile included in the repository can be used for single-node/multi-node & Ubuntu 20.04/CentOS 8.2 setups.

The Vagrantfile can be configured using the environment variables:

- OS - define the operating system. It accepts ubuntu, the default value, and centos.
- MULTI_NODE - if set then a HA cluster is generated. Otherwise a single-node setup is used.
- CONTROLLERS - defines the number of controller nodes. The default number is 3.
- WORKERS - specifies the number of worker nodes. The default number is 2.
- SSH_PUBLIC_KEY - if this environment variable is not set, then :file:`$HOME/.ssh/id_rsa` is used by default.
- IP_PREFIX - this value is used to generate the IP addresses of the nodes. If not set 192.168.100 will be used. The single node has the IP address 192.168.100.50. The controllers start with the IP address 192.168.100.200 and the workers with 192.168.100.100.
- CONTROLLERS_RAM - amount of RAM for one controller
- WORKERS_RAM - amount of RAM for one worker
- STORAGE_RAM - amount of RAM for one storage
- CONTROLLERS_CPUS - number of CPUs per controller
- WORKERS_CPUS - number of CPUs per worker
- STORAGE_CPUS - number of CPUs per storage

.. note:: The multi-node setup with the default settings needs about 20GB RAM for itself.


Usage
^^^^^

The directory called :file:`setup` (`https://github.com/darxkies/k8s-tew/tree/2.4.0/setup <https://github.com/darxkies/k8s-tew/tree/2.4.0/setup>`_) contains sub-directories for various cluster setup configurations:

- local - it starts a single-node cluster locally without using any kind of virtualization. This kind of setup needs root rights. It is meant for local development where it might be important to fire the cluster up and shut it down fast. If you want it to start automatically, take a look at the quickstart section.
- ubuntu-single-node - Ubuntu 20.04 single-node cluster. It needs about 8GB Ram.
- ubuntu-multi-node - Ubuntu 20.04 HA cluster. It needs around 20GB Ram.
- centos-single-node - CentOS 8.2 single-node cluster. It needs about 8GB Ram.
- centos-multi-node - CentOS 8.2 HA cluster. It needs around 20GB Ram.

.. note:: Regardless of the setup, once the deployment is done it will take a while to download all required containers from the internet. So better use kubectl to check the status of the pods.

.. note:: For the local setup, to access the Kubernetes Dashboard use the internal IP address (e.g. 192.168.x.y or 10.x.y.z) and not 127.0.0.1/localhost. Depending on the hardware used, it might take a while until it starts and setups everything.

Create
^^^^^^

Change to one of the sub-directories and enter the following command to start the cluster:

  .. code:: shell

    make

.. note:: This will destroy any existing VMs, creates new VMs and performs all the steps (forced initialization, configuration, generation and deployment) to create the cluster.

Stop
^^^^^^

For the local setup, just press CTRL+C.

For the other setups enter:

  .. code:: shell

    make halt

Start
^^^^^

To start an existing setup/VMs enter:

  .. code:: shell

    make up

.. note:: This and the following commands work only for Vagrant based setups.

SSH
^^^

For single-node setups enter:

  .. code:: shell

    make ssh

And for multi-node setups:

  .. code:: shell

    make ssh-controller00
    make ssh-controller01
    make ssh-controller02
    make ssh-worker00
    make ssh-worker01

Kubernetes Dashboard
^^^^^^^^^^^^^^^^^^^^

This will display the token, and then it will open the web browser pointing to the address of Kubernetes Dashboard:

  .. code:: shell

    make dashboard

Ingress Port Forwarding
^^^^^^^^^^^^^^^^^^^^^^^

In order to start port forwarding from your host's ports 80 and 443 to Vagrant's VMs for Ingress enter:

  .. code:: shell

    make forward-80
    make forward-443

.. note:: Both commands are blocking. So you need two different terminal sessions.

