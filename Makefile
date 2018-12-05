BUILD_IMAGE = darxkies/k8s-tew-build
VERSION = $(shell git describe --tags)
PACKAGE = github.com/darxkies/k8s-tew

compile:
	docker build -t $(BUILD_IMAGE) .
	docker run --rm -v $$(pwd):/go/src/$(PACKAGE) $(BUILD_IMAGE)

build-binaries:
	mkdir -p embedded
	CGO_ENABLED=0 go install -ldflags '-s -w' ${PACKAGE}/cmd/freezer
	packr
	CGO_ENABLED=0 go build -ldflags "-X ${PACKAGE}/version.Version=${VERSION} -s -w" -o k8s-tew ${PACKAGE}/cmd/k8s-tew 

watch-and-compile:
	go get github.com/cespare/reflex
	reflex -r '\.go$$' -R '^vendor' -R '^utils/a_utils-packr\.go$$' make build-binaries

watch-and-update-documentation:
	(cd docs && reflex -r '\.rst' -R "^_build" make clean html)

clean:
	sudo rm -Rf bin vendor

.PHONY: build clean
