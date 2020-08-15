FROM ubuntu:20.04

ENV container docker
ENV LC_ALL C
ENV DEBIAN_FRONTEND noninteractive

EXPOSE 22

VOLUME ["/sys/fs/cgroup"]

STOPSIGNAL SIGRTMIN+3

RUN sed -i 's/# deb/deb/g' /etc/apt/sources.list

RUN apt-get update \
    && apt-get install -y systemd systemd-sysv openssh-server iproute2 iputils-ping vim less iptables kmod ca-certificates curl libseccomp2 conntrack ethtool socat util-linux mount ebtables udev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* 

RUN cd /lib/systemd/system/sysinit.target.wants/ \
    && ls | grep -v systemd-tmpfiles-setup | xargs rm -fv $1 \
	&& mv /etc/systemd/system/multi-user.target.wants/ssh.service /tmp/ \
	&& rm -fv /lib/systemd/system/multi-user.target.wants/* \
    /etc/systemd/system/*.wants/* \
    /lib/systemd/system/local-fs.target.wants/* \
    /lib/systemd/system/sockets.target.wants/*udev* \
    /lib/systemd/system/sockets.target.wants/*initctl* \
    /lib/systemd/system/basic.target.wants/* \
    /lib/systemd/system/anaconda.target.wants/* \
    /lib/systemd/system/plymouth* \
    /lib/systemd/system/systemd-update-utmp* \
	&& echo "ReadKMsg=no" >> /etc/systemd/journald.conf \
	&& mv /tmp/ssh.service  /etc/systemd/system/multi-user.target.wants/ 

RUN mkdir /var/run/sshd

RUN echo 'root:root' | chpasswd

RUN sed -ri 's/^#?PermitRootLogin\s+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -ri 's/UsePAM yes/#UsePAM yes/g' /etc/ssh/sshd_config

RUN mkdir /root/.ssh && \
  touch /root/.ssh/authorized_keys && \
  chmod 700 /root/.ssh && \
  chmod 600 /root/.ssh/authorized_keys

COPY files/ssh_public_key /root/.ssh/authorized_keys

CMD ["/lib/systemd/systemd"]
