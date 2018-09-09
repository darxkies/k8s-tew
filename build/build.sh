#!/bin/sh

VERSION=$(git describe --tags)

echo "Version: $VERSION"

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
freezer freeze-binary /usr/bin/rbd         embedded

echo "Freeze libraries"

freezer freeze-library /usr/lib/x86_64-linux-gnu/ceph/libceph-common.so.0             embedded
freezer freeze-library /usr/lib/x86_64-linux-gnu/ceph/compressor/libceph_snappy.so.2  embedded
freezer freeze-library /usr/lib/x86_64-linux-gnu/ceph/compressor/libceph_zstd.so.2    embedded
freezer freeze-library /usr/lib/x86_64-linux-gnu/ceph/compressor/libceph_zlib.so.2    embedded
freezer freeze-library /usr/lib/x86_64-linux-gnu/ceph/crypto/libceph_crypto_isal.so.1 embedded
freezer freeze-library /usr/lib/x86_64-linux-gnu/libcephfs.so.2                       embedded
freezer freeze-library /usr/lib/x86_64-linux-gnu/libsqlite3.so.0                      embedded
freezer freeze-library /usr/lib/x86_64-linux-gnu/nss/libsoftokn3.so                   embedded
freezer freeze-library /usr/lib/x86_64-linux-gnu/nss/libfreeblpriv3.so                embedded

echo "Prepare embedded files"

packr 

echo "Build k8s-tew"

CGO_ENABLED=0 go install -ldflags "-X github.com/darxkies/k8s-tew/version.Version=$VERSION -s -w" github.com/darxkies/k8s-tew/cmd/k8s-tew
