.EXPORT_ALL_VARIABLES:

VAGRANT_VAGRANTFILE=../../Vagrantfile
OS=ubuntu
IP_PREFIX=192.168.120
MULTI_NODE=true

run: destroy up deploy

deploy:
	k8s-tew initialize -f
	k8s-tew configure --controller-virtual-ip=$(IP_PREFIX).20 
	k8s-tew configure --controller-virtual-ip-interface=eth1 
	k8s-tew configure --worker-virtual-ip=$(IP_PREFIX).10 
	k8s-tew configure --worker-virtual-ip-interface=eth1 
	k8s-tew configure --public-network=$(IP_PREFIX).0/24 
	k8s-tew configure --resolv-conf=/run/systemd/resolve/resolv.conf 
	k8s-tew node-add -n controller00 -i $(IP_PREFIX).200 -x 0 -l controller
	k8s-tew node-add -n controller01 -i $(IP_PREFIX).201 -x 1 -l controller
	k8s-tew node-add -n controller02 -i $(IP_PREFIX).202 -x 2 -l controller
	k8s-tew node-add -n worker00 -i $(IP_PREFIX).100 -x 3 -l worker
	k8s-tew node-add -n worker01 -i $(IP_PREFIX).101 -x 4 -l worker
	k8s-tew generate 
	k8s-tew deploy 

destroy:
	vagrant destroy -f

up:
	vagrant up

recreate: destroy up save-clean

save-clean: 
	vagrant halt
	vagrant snapshot save controller00 clean
	vagrant snapshot save controller01 clean
	vagrant snapshot save controller02 clean
	vagrant snapshot save worker00 clean
	vagrant snapshot save worker01 clean
	vagrant up --parallel

restore-clean: 
	vagrant halt -f
	vagrant snapshot restore controller00 clean
	vagrant snapshot restore controller01 clean
	vagrant snapshot restore controller02 clean
	vagrant snapshot restore worker00 clean
	vagrant snapshot restore worker01 clean
	vagrant up --parallel

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
	sudo socat -d -v TCP-LISTEN:80,fork TCP:$(IP_PREFIX).20:30080

forward-443:
	sudo socat -d -v TCP-LISTEN:443,fork TCP:$(IP_PREFIX).20:30443
