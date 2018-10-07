#!/bin/sh

unset GOPATH

VERSION=$(git describe --tags)

echo "Version: $VERSION"

echo "Getting dependencies"

go mod vendor

echo "Build packr"

if [ ! -f $GOBIN/packr ]; then
  go get -u github.com/gobuffalo/packr/...
fi

echo "Build freezer"

CGO_ENABLED=0 go install -ldflags '-s -w' github.com/darxkies/k8s-tew/cmd/freezer

echo "Freeze binaries"

freezer freeze-binary /usr/bin/socat       embedded
freezer freeze-binary /sbin/ipset          embedded
freezer freeze-binary /usr/sbin/conntrack  embedded

echo "Prepare embedded files"

packr 

echo "Build k8s-tew"

CGO_ENABLED=0 go install -ldflags "-X github.com/darxkies/k8s-tew/version.Version=$VERSION -s -w" github.com/darxkies/k8s-tew/cmd/k8s-tew

echo "Done"
