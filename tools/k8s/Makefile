IMAGE=docker.io/darxkies/kubernetes-tools:18.04 
docker:
	docker build -t $(IMAGE) .

push: docker
	docker push $(IMAGE)

deploy: docker
	docker run -ti --rm -v $$(pwd)/deployment:/deployment --entrypoint /freezer $(IMAGE) deploy / /deployment $$(pwd)/deployment

rbd: docker
	docker run -ti --rm -v $$(pwd)/deployment:/deployment $(IMAGE) /bin/rbd

ctr:
	ctr run --rm --mount type=bind,src=/tmp/deployment,dst=/deployment,options=rbind:rw $(IMAGE) kubernetes-tools /freezer deploy / /deployment /tmp/deployment 

.PHONY: docker
