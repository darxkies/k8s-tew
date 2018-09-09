BUILD_IMAGE=darxkies/k8s-tew-build

build:
	docker build -t $(BUILD_IMAGE) build
	docker run --rm -v $(GOPATH):/go $(BUILD_IMAGE)

.PHONY: build
