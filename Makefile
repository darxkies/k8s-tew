BUILD_IMAGE=darxkies/k8s-tew-build

build:
	docker build -t $(BUILD_IMAGE) build
	docker run --rm -v $$(pwd):/go $(BUILD_IMAGE)

watch-and-compile:
	go get github.com/cespare/reflex
	reflex -r '\.go$$' -R '^vendor' -R '^utils/a_utils-packr\.go$$' build/build.sh

watch-and-update-documentation:
	(cd docs && reflex -r '\.rst' make html)

.PHONY: build
