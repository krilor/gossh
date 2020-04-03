package docker

// this file contains

// Dockerfile are strings representing dockerfiles
type Dockerfile string

var dockerFileUbuntu Dockerfile = `FROM ubuntu:bionic

RUN apt update
RUN apt -y install openssh-server sudo
RUN mkdir -p /var/run/sshd

RUN groupadd -r gossh && useradd -m -s /bin/bash -g gossh gossh
RUN adduser gossh sudo

RUN groupadd -r hobgob && useradd -m -s /bin/bash -g hobgob hobgob
RUN adduser hobgob sudo

RUN echo 'root:rootpwd' | chpasswd
RUN echo 'gossh:gosshpwd' | chpasswd
RUN echo 'hobgob:hobgobpwd' | chpasswd

RUN echo -e "#!/usr/bin/env bash\nset -e\n/usr/sbin/sshd -D"> /run.sh
RUN chmod +x /run.sh

EXPOSE 22
CMD ["/run.sh"]`
