.EXPORT_ALL_VARIABLES:

VAGRANT_VAGRANTFILE=../../Vagrantfile
OS=centos
IP_PREFIX=192.168.140
MULTI_NODE=true

run: destroy up deploy 

deploy:
	k8s-tew initialize -f
	k8s-tew configure --controller-virtual-ip=$(IP_PREFIX).20 
	k8s-tew configure --controller-virtual-ip-interface=eth1 
	k8s-tew configure --worker-virtual-ip=$(IP_PREFIX).10 
	k8s-tew configure --worker-virtual-ip-interface=eth1 
	k8s-tew configure --public-network=$(IP_PREFIX).0/24
	k8s-tew node-add -n controller00 -i $(IP_PREFIX).200 -l controller
	k8s-tew node-add -n controller01 -i $(IP_PREFIX).201 -l controller
	k8s-tew node-add -n controller02 -i $(IP_PREFIX).202 -l controller
	k8s-tew node-add -n worker00 -i $(IP_PREFIX).100 -l worker,storage
	k8s-tew node-add -n worker01 -i $(IP_PREFIX).101 -l worker,storage
	k8s-tew generate 
	k8s-tew deploy

destroy:
	vagrant destroy -f

up:
	vagrant up

halt:
	vagrant halt

dashboard:
	k8s-tew dashboard -o

ssh-controller00:
	vagrant ssh controller00

ssh-controller01:
	vagrant ssh controller01

ssh-controller02:
	vagrant ssh controller02

ssh-worker00:
	vagrant ssh worker00
	
ssh-worker01:
	vagrant ssh worker01

forward-80:
	sudo socat -d -v TCP-LISTEN:80,fork TCP:$(IP_PREFIX).20:80

forward-443:
	sudo socat -d -v TCP-LISTEN:443,fork TCP:$(IP_PREFIX).20:443
