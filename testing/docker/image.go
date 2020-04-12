package docker

import (
	"fmt"
	"strconv"
	"strings"
)

// this file contains

// Image is a reference to a docker image
type Image struct {
	// From and tag is the docker image (prefixed repo url, if not hub.docker.com)
	//
	// Used in the from statement in the top of the Dockerfile
	//
	// FROM <From>[:<tag>]
	From string
	Tag  string

	// Last part of the dockerfile, after FROM
	Instructions string

	// Slug should be a short sluglike name
	Slug string
}

// Dockerfile returns the dockerfile for a debian image
func (i Image) Dockerfile() string {

	b := strings.Builder{}

	b.WriteString("FROM ")
	b.WriteString(i.From)
	if i.Tag != "" {
		b.WriteString(":" + i.Tag)
	}
	b.WriteString("\n")
	b.WriteString(i.Instructions)
	return b.String()

}

// Name returns a gossh-prefixed name for the image
func (i Image) Name() string {
	return "gossh_throwaway_" + strings.ReplaceAll(i.Slug, ":", "_")
}

var debianInstructions string = `
RUN apt update
RUN apt -y install openssh-server sudo
RUN mkdir -p /var/run/sshd
`

var yumRepoInstructions = `
RUN yum -y install https://dl.fedoraproject.org/pub/epel/epel-release-latest-VERSION.noarch.rpm && \
	yum -y install https://download1.rpmfusion.org/free/el/rpmfusion-free-release-VERSION.noarch.rpm && \
	echo -e "[centos]\nname=CentOS-VERSION\nbaseurl=http://ftp.heanet.ie/pub/centos/VERSION/os/x86_64/\nenabled=1\ngpgcheck=1\ngpgkey=http://ftp.heanet.ie/pub/centos/VERSION/os/x86_64/RPM-GPG-KEY-CentOS-VERSION" > /etc/yum.repos.d/centosVERSION.repo

`

var rhelInstructions string = `
RUN yum -y install sudo && \
	sed -i.old '0,/# %wheel/{s/# %wheel.*/%wheel ALL=(ALL) ALL/}' /etc/sudoers

RUN yum -y install openssh openssh-server openssh-clients && \
	yum -y clean all

# ssh-keygen -A is not available on rhel6 based images
RUN ssh-keygen -q -N "" -t dsa -f /etc/ssh/ssh_host_dsa_key && \
	ssh-keygen -q -N "" -t rsa -b 4096 -f /etc/ssh/ssh_host_rsa_key && \
	ssh-keygen -q -N "" -t ecdsa -f /etc/ssh/ssh_host_ecdsa_key

RUN echo "Defaults lecture = never" >> /etc/sudoers.d/privacy`

var commonInstructions string = `
RUN groupadd -r gossh && \
	useradd -m -s /bin/bash -g gossh gossh && \
	echo "gossh  ALL=(ALL) ALL" > /etc/sudoers.d/gossh

RUN groupadd -r hobgob && \
	useradd -m -s /bin/bash -g hobgob hobgob && \
	echo "hobgob ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/hobgob

RUN echo 'root:rootpwd' | chpasswd
RUN echo 'gossh:gosshpwd' | chpasswd
RUN echo 'hobgob:hobgobpwd' | chpasswd

RUN printf "%s\n" '#!/usr/bin/env bash' 'set -e' '/usr/sbin/sshd -D' > /run.sh

RUN chmod +x /run.sh

EXPOSE 22
CMD ["/run.sh"]
`

// Ubuntu returns a ubuntu image
func Ubuntu(tag string) Image {
	return Image{"ubuntu", tag, debianInstructions + commonInstructions, fmt.Sprintf("ubuntu:%s", tag)}
}

// Debian returns a Debian image
//
// https://hub.docker.com/_/debian
func Debian(tag string) Image {
	return Image{"debian", tag, debianInstructions + commonInstructions, fmt.Sprintf("debian:%s", tag)}
}

// CentOS returns a CentOS image
func CentOS(version int) Image {
	return Image{"centos", strconv.Itoa(version), rhelInstructions + commonInstructions, fmt.Sprintf("centos:%d", version)}
}

// RedHat returns a RHEL image
func RedHat(version int) Image {
	return Image{fmt.Sprintf("registry.access.redhat.com/rhel%d/rhel", version), "", strings.ReplaceAll(yumRepoInstructions, "VERSION", strconv.Itoa(version)) + rhelInstructions + commonInstructions, fmt.Sprintf("rhel:%d", version)}
}

// Oracle returns a ol Image
func Oracle(version int) Image {
	return Image{"oraclelinux", strconv.Itoa(version), rhelInstructions + commonInstructions, fmt.Sprintf("ol:%d", version)}
}

// Fedora returns a fedora image
func Fedora(version int) Image {
	return Image{"fedora", strconv.Itoa(version), rhelInstructions + commonInstructions, fmt.Sprintf("fedora:%d", version)}
}

// Bench is
var Bench map[string]Image = map[string]Image{
	"fedora:33": Fedora(33),
}

// FullBench is a map of images that can be used as a test bench.
var FullBench map[string]Image = map[string]Image{

	// Debian - https://hub.docker.com/_/debian
	"debian:bullseye": Debian("bullseye"),
	"debian:buster":   Debian("buster"),
	"debian:stretch":  Debian("stretch"),

	// Ubuntu - https://hub.docker.com/_/ubuntu
	"ubuntu:bionic": Ubuntu("bionic"), // 18
	"ubuntu:eoan":   Ubuntu("eoan"),   // 19
	"ubuntu:focal":  Ubuntu("focal"),  // 20
	"ubuntu:trusty": Ubuntu("trusty"), // 16
	"ubuntu:xenial": Ubuntu("xenial"), // 14

	// CentOS - https://hub.docker.com/_/centos
	"centos:7": CentOS(7),
	"centos:6": CentOS(6),

	// RedHat - https://catalog.redhat.com/software/containers/search?q=rhel&p=1&vendor_name=Red%20Hat%2C%20Inc.&build_categories_list=Base%20Image&product=Red%20Hat%20Enterprise%20Linux&release_categories=Generally%20Available&rows=60
	"rhel:6": RedHat(6),
	"rhel:7": RedHat(7),

	// Oracle Linux - https://hub.docker.com/_/oraclelinux/
	"ol:6": Oracle(6),
	"ol:7": Oracle(7),
	"ol:8": Oracle(8),

	// Fedora - https://hub.docker.com/_/fedora
	"fedora:26": Fedora(26),
	"fedora:27": Fedora(27),
	"fedora:28": Fedora(28),
	"fedora:29": Fedora(29),
	"fedora:30": Fedora(30),
	"fedora:31": Fedora(31),
	"fedora:32": Fedora(32),
	"fedora:33": Fedora(33),
}
