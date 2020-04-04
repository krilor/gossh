package docker

import "fmt"

// this file contains

// Image is a reference to a docker image
type Image interface {
	Dockerfile() string
	Name() string
}

// DebianImage is a image for ubuntu
type DebianImage struct {
	distro  string
	version string
}

// NewDebianImage returns a debian image
func NewDebianImage(distro string, version string) DebianImage {
	return DebianImage{distro, version}
}

// Dockerfile returns the dockerfile for a debian image
func (d DebianImage) Dockerfile() string {
	return fmt.Sprintf(`FROM %s:%s

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

    RUN echo "#!/usr/bin/env bash\nset -e\n/usr/sbin/sshd -D" > /run.sh
    RUN chmod +x /run.sh

    EXPOSE 22
    CMD ["/run.sh"]`, d.distro, d.version)
}

// Name returns an image name
func (d DebianImage) Name() string {
	return fmt.Sprintf("gossh_throwaway_%s_%s", d.distro, d.version)
}

// RHELImage is a image for ubuntu
type RHELImage struct {
	distro  string
	version string
}

// NewRHELImage returns a debian image
func NewRHELImage(distro string, version string) RHELImage {
	return RHELImage{distro, version}
}

// Dockerfile returns the dockerfile for a debian image
func (r RHELImage) Dockerfile() string {
	return fmt.Sprintf(`FROM %s:%s

    RUN yum -y erase vim-minimal iputils libss && \
        yum -y install openssh openssh-server openssh-clients sudo && \
        yum -y clean all

    RUN ssh-keygen -A

    RUN echo "Defaults lecture = never" >> /etc/sudoers.d/privacy

    RUN groupadd -r gossh && \
        useradd -m -s /bin/bash -g gossh gossh && \
        usermod -g wheel gossh

    RUN groupadd -r hobgob && \
        useradd -m -s /bin/bash -g hobgob hobgob && \
        usermod -g wheel hobgob

    RUN echo 'root:rootpwd' | chpasswd
    RUN echo 'gossh:gosshpwd' | chpasswd
    RUN echo 'hobgob:hobgobpwd' | chpasswd

    RUN echo -e "#!/usr/bin/env bash\nset -e\n/usr/sbin/sshd -D" > /run.sh
    RUN chmod +x /run.sh

    EXPOSE 22
    CMD ["/run.sh"]`, r.distro, r.version)
}

// Name returns an image name
func (r RHELImage) Name() string {
	return fmt.Sprintf("gossh_throwaway_%s_%s", r.distro, r.version)
}

// CustomImage allows the user to specify a custom image name and dockerfile content
type CustomImage struct {
	dockerfile string
	name       string
}

// NewCustomImage returns a Image based on the input
func NewCustomImage(dockerfile, name string) CustomImage {
	return CustomImage{dockerfile, name}
}

// Dockerfile returns dockerfile contents
func (c CustomImage) Dockerfile() string {
	return c.dockerfile
}

// Name returns name
func (c CustomImage) Name() string {
	return c.name
}
