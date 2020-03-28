FROM ubuntu:bionic

RUN apt update
RUN apt -y install openssh-server sudo
RUN mkdir -p /var/run/sshd

RUN groupadd -r gossh && useradd -m -s /bin/bash -g gossh gossh
RUN adduser gossh sudo

RUN echo 'root:rootpwd' | chpasswd
RUN echo 'gossh:gosshpwd' | chpasswd

COPY run.sh /run.sh
RUN chmod +x /run.sh

EXPOSE 22
CMD ["/run.sh"]
