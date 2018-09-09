FROM ubuntu:bionic

RUN apt-get update && apt-get install -y golang git socat ipset conntrack ceph-common

RUN mkdir -p /go/src/github.com/darxkies/k8s-tew

ENV GOPATH=/go
ENV GOBIN=$GOPATH/bin
ENV PATH=$GOBIN:$PATH
ENV GOCACHE=$GOPATH/pkg

RUN go get -u github.com/gobuffalo/packr/...

WORKDIR /go/src/github.com/darxkies/k8s-tew

CMD ["build/build.sh"]

