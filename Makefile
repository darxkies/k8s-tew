VERSION=$(shell git describe --tags)

compile:
	CGO_ENABLED=0 go build -ldflags '-X github.com/darxkies/k8s-tew/version.Version=${VERSION}' -ldflags '-s -w' -o ${GOPATH}/bin/k8s-tew github.com/darxkies/k8s-tew/cmd/k8s-tew
