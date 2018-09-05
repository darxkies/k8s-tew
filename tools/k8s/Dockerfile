FROM ubuntu:18.04 AS source

RUN apt-get update && apt-get install -y socat ipset conntrack ceph-common

ADD freezer.sh /

RUN ./freezer.sh /usr/bin/socat       /assets
RUN ./freezer.sh /sbin/ipset          /assets
RUN ./freezer.sh /usr/sbin/conntrack  /assets
RUN ./freezer.sh /usr/bin/rbd         /assets

FROM scratch

COPY --from=source /assets/ /

ENV LD_LIBRARY_PATH=/lib

ENTRYPOINT ["/lib/ld-linux-x86-64.so.2"]
