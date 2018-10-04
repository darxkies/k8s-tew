BUILD_IMAGE=darxkies/k8s-tew-build

build:
	docker build -t $(BUILD_IMAGE) build
	docker run --rm -v $$(pwd):/go $(BUILD_IMAGE)

.PHONY: build
