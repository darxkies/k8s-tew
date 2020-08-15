CONTAINER ?= k8s-tew-in-docker 
NAME ?= $(CONTAINER)
SSH_PUBLIC_KEY ?= ~/.ssh/id_rsa.pub
IP ?= $(shell docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $(NAME))

build:
	mkdir -p files
	cp $(SSH_PUBLIC_KEY) files/ssh_public_key
	docker build -t $(CONTAINER) .

start:
	docker run \
		-d \
		--name $(NAME) \
		--hostname $(NAME) \
		-v /lib/modules:/lib/modules:ro \
		-v /sys/fs/cgroup:/sys/fs/cgroup \
		-v /var/lib/k8s-tew \
		-v /var/log/k8s-tew \
		--privileged \
		--security-opt seccomp=unconfined \
		--security-opt apparmor=unconfined \
		--rm $(CONTAINER) 
	

attach:
	docker exec -ti $(NAME) bash

stop:
	docker kill $(NAME) 

shell:
	ssh -o StrictHostKeyChecking=no root@$(IP)

deploy:
	k8s-tew initialize -f
	k8s-tew configure --public-network=$(IP)/24
	k8s-tew configure --metallb-addresses=$(IP)/32
	k8s-tew node-add -n $(NAME) -i $(IP) -x 0 -l controller,worker,storage
	k8s-tew generate --pull-images --parallel
	k8s-tew deploy --import-images --parallel
