.EXPORT_ALL_VARIABLES:

NAME=single-node
IP=$(shell multipass info $(NAME) | grep IPv4 | sed -e "s/IPv4://" | xargs)
SSH_KEY=$(shell cat ~/.ssh/id_rsa.pub)
CLOUD_INIT="disable_root: 0\nssh_authorized_keys:\\n  - $(SSH_KEY)" 

run: destroy create deploy 

deploy:
	k8s-tew initialize -f
	k8s-tew configure --ingress-domain=darxkies.duckdns.org
	k8s-tew configure --email=darxkies@gmail.com
	k8s-tew configure --public-network=$(IP)/24
	k8s-tew configure --resolv-conf=/run/systemd/resolve/resolv.conf 
	k8s-tew configure --metallb-addresses=$(IP)/32
	k8s-tew node-add -n single-node -i $(IP) -x 0 -l controller,worker
	k8s-tew generate 
	k8s-tew deploy --parallel 

refresh:
	k8s-tew generate --parallel 
	k8s-tew deploy --parallel 

destroy:
	multipass delete $(NAME) -p

create:
	echo $(CLOUD_INIT) | multipass launch -n $(NAME) -c 4 -m 8G -d 40G --cloud-init - 18.04 

start:
	multipass start $(NAME)

stop:
	multipass stop $(NAME)

info:
	multipass info $(NAME)

reboot: halt up

dashboard:
	k8s-tew dashboard -o

open-web-browser-all:
	k8s-tew open-web-browser --all

shell:
	multipass shell $(NAME)

forward-80:
	sudo socat -d -v TCP-LISTEN:80,fork TCP:$(IP):80

forward-443:
	sudo socat -d -v TCP-LISTEN:443,fork TCP:$(IP):443

k9s:
	KUBECONFIG=assets/etc/k8s-tew/k8s/kubeconfig/admin.kubeconfig k9s
