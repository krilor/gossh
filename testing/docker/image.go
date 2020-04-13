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
`

// TODO - Set disable_coredump false is a workaround for
// https://ask.fedoraproject.org/t/sudo-setrlimit-rlimit-core-operation-not-permitted/4223
// https://bugs.launchpad.net/ubuntu/+source/sudo/+bug/1857036
// https://bugzilla.redhat.com/show_bug.cgi?id=1773148
// RUN echo "Set disable_coredump false" >> /etc/sudo.conf

var commonInstructions string = `
RUN echo "Defaults lecture = never" >> /etc/sudoers.d/0_privacy

RUN echo 'root:rootpwd' | chpasswd

RUN groupadd -r gossh && \
	useradd -m -s /bin/bash -g gossh gossh && \
	echo 'gossh:gosshpwd' | chpasswd && \
	printf "%s\n" 'gossh ALL=(ALL) ALL' 'Defaults:gossh lecture = never' > /etc/sudoers.d/gossh

RUN groupadd -r hobgob && \
	useradd -m -s /bin/bash -g hobgob hobgob && \
	echo 'hobgob:hobgobpwd' | chpasswd && \
	printf "%s\n" 'hobgob ALL=(ALL) NOPASSWD:ALL' 'Defaults:hobgob lecture = never' > /etc/sudoers.d/hobgob

RUN groupadd -r joxter && \
	useradd -m -s /bin/bash -g joxter joxter && \
	echo 'joxter:joxterpwd' | chpasswd && \
	printf "%s\n" 'joxter ALL=(ALL) ALL' 'Defaults:joxter lecture = always' > /etc/sudoers.d/joxter

RUN groupadd -r groke && \
	useradd -m -s /bin/bash -g groke groke && \
	echo 'groke:grokepwd' | chpasswd && \
	printf "%s\n" 'groke ALL=(ALL) NOPASSWD:ALL' 'Defaults:groke lecture = always' > /etc/sudoers.d/groke

RUN groupadd -r stinky && \
	useradd -m -s /bin/bash -g stinky stinky && \
	echo 'stinky:stinkypwd' | chpasswd

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
var Bench []Image = []Image{
	Ubuntu("bionic"),
	CentOS(7),
}

// FullBench is a map of images that can be used as a test bench.
var FullBench []Image = []Image{

	// Debian - https://hub.docker.com/_/debian
	Debian("bullseye"),
	Debian("buster"),
	Debian("stretch"),

	// Ubuntu - https://hub.docker.com/_/ubuntu
	Ubuntu("bionic"), // 18
	Ubuntu("eoan"),   // 19
	Ubuntu("focal"),  // 20
	Ubuntu("trusty"), // 16
	Ubuntu("xenial"), // 14

	// CentOS - https://hub.docker.com/_/centos
	CentOS(7),
	CentOS(6),

	// RedHat - https://catalog.redhat.com/software/containers/search?q=rhel&p=1&vendor_name=Red%20Hat%2C%20Inc.&build_categories_list=Base%20Image&product=Red%20Hat%20Enterprise%20Linux&release_categories=Generally%20Available&rows=60
	RedHat(6),
	RedHat(7),

	// Oracle Linux - https://hub.docker.com/_/oraclelinux/
	Oracle(6),
	Oracle(7),
	Oracle(8),

	// Fedora - https://hub.docker.com/_/fedora
	Fedora(26),
	Fedora(27),
	Fedora(28),
	Fedora(29),
	Fedora(30),
	Fedora(31),
	Fedora(32),
	Fedora(33),
}
