.EXPORT_ALL_VARIABLES:

VAGRANT_VAGRANTFILE=../../Vagrantfile
OS=ubuntu
IP_PREFIX=192.168.110

run: destroy up deploy 

deploy:
	k8s-tew initialize -f
	k8s-tew configure --public-network=$(IP_PREFIX).0/24
	k8s-tew configure --resolv-conf=/run/systemd/resolve/resolv.conf 
	k8s-tew configure --metallb-addresses=$(IP_PREFIX).50/32
	k8s-tew node-add -n single-node -i $(IP_PREFIX).50 -x 0 -l controller,worker,storage
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

ssh:
	vagrant ssh

forward-80:
	sudo socat -d -v TCP-LISTEN:80,fork TCP:$(IP_PREFIX).50:80

forward-443:
	sudo socat -d -v TCP-LISTEN:443,fork TCP:$(IP_PREFIX).50:443

forward-grafana-30900:
	sudo socat -d -v TCP-LISTEN:30900,fork TCP:$(IP_PREFIX).50:30900

forward-dashboard-32443:
	sudo socat -d -v TCP-LISTEN:32443,fork TCP:$(IP_PREFIX).50:32443
