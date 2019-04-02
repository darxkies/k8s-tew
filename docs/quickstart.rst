Quick Start
===========

The following snippet will create a cluster on the host computer or in a virtual machine:

  .. code:: shell

    # Switch to user root
    sudo su -

    # Download Binary
    wget https://github.com/darxkies/k8s-tew/releases/download/2.3.0/k8s-tew
    chmod a+x k8s-tew

    # Everything is installed relative to the root directory
    export K8S_TEW_BASE_DIRECTORY=/

    # Initialize cluster configuration
    ./k8s-tew initialize

    # Node the current machine to the cluster (the settings such as IP and hostname are inferred)
    ./k8s-tew node-add -s

    # Only on Ubuntu 18.04 to solve any DNS related issues
    ./k8s-tew configure --resolv-conf=/run/systemd/resolve/resolv.conf

    # Generate artefacts (e.g. certificates, configurations and so on)
    ./k8s-tew generate 

    # Activate and start service
    systemctl daemon-reload
    systemctl enable k8s-tew
    systemctl start k8s-tew

    # Activate environment variables and switch back to root
    exit
    sudo su -

    # Watch the pods being installed
    watch -n 1 kubectl get pods --all-namespaces

.. note:: You will need at least 20GB HDD, 8GB RAM and 4 CPU Cores.
.. note:: To use k8s-tew with Vagrant take a look at `https://github.com/darxkies/k8s-tew/tree/2.3.0/setup <https://github.com/darxkies/k8s-tew/tree/2.3.0/setup>`_.

