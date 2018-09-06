FROM ubuntu:18.04 AS source

RUN apt-get update && apt-get install -y socat ipset conntrack ceph-common golang

ADD freezer.go /

RUN mkdir /assets

RUN CGO_ENABLED=0 go build -o /assets/freezer /freezer.go

RUN /assets/freezer freeze /usr/bin/socat       /assets
RUN /assets/freezer freeze /sbin/ipset          /assets
RUN /assets/freezer freeze /usr/sbin/conntrack  /assets
RUN /assets/freezer freeze /usr/bin/rbd         /assets

RUN mkdir /assets/deployment

FROM scratch

COPY --from=source /assets/ /

ENV LD_LIBRARY_PATH=/lib

ENTRYPOINT ["/lib/ld-linux-x86-64.so.2"]
