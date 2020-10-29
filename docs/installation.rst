Installation
============

The commands in the upcoming sections will assume that k8s-tew is going to be installed in the directory :file:`/usr/local/bin`. That means that the aforementioned directory exists and it is included in the PATH. If that is not the case use the following commands:

  .. code:: shell

    sudo mkdir -p /usr/local/bin
    export PATH=/usr/local/bin:$PATH

Binary
------

The x86 64-bit binary can be downloaded from the following address: https://github.com/darxkies/k8s-tew/releases

Additionally the these commands can be used to download it and install it in :file:`/usr/local/bin`

  .. code:: shell

   curl -s https://api.github.com/repos/darxkies/k8s-tew/releases/latest | grep "browser_download_url" | cut -d : -f 2,3 | tr -d \" | sudo wget -O /usr/local/bin/k8s-tew -qi -
   sudo chmod a+x /usr/local/bin/k8s-tew

Source
------

To compile it from source you will need a Go (version 1.15+) environment, Git, Make and Docker installed. Once everything is installed, enter the following commands:

  .. code:: shell

    export GOPATH=~/go
    export PATH=$GOPATH/bin:$PATH
    mkdir -p $GOPATH/src/github.com/darxkies
    cd $GOPATH/src/github.com/darxkies
    git clone https://github.com/darxkies/k8s-tew.git
    cd k8s-tew
    make
    sudo mv k8s-tew /usr/local/bin

