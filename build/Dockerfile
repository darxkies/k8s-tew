FROM ubuntu:bionic

ARG GO_VERSION=1.11.1

RUN apt-get update && apt-get install -y git socat ipset conntrack wget build-essential

RUN wget https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz && tar xzfv go$GO_VERSION.linux-amd64.tar.gz -C /usr/local && rm go$GO_VERSION.linux-amd64.tar.gz

WORKDIR /go

ENV GOBIN=/go/bin
ENV GOCACHE=/go/bin/cache
ENV HOME=/go/bin
ENV GOROOT=/usr/local/go
ENV PATH=$GOBIN:$GOROOT/bin:$PATH

CMD ["build/build.sh"]

