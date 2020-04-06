# GOSSH

This repo is an experiement with creating a declarative IT automation and configuration management package for Golang. Think Ansible, but no Yaml, just plain Go. WOW - all teh power!

* Declarative - use `rules` to check and ensure state on any linux `host`
* Agentless - all work is done on remote hosts by issuing commands via SSH
* Efficient - leverage ready-made rules to kick-start your IT automation

**The project is in a super-early state. I am looking for API/usage/naming convention input and ideas in general on how to approach this problem.** Have a look at [examples](examples/random/main.go) to get a feel for what I currently think it would look like to use the package. If you have any ideas please reach out through a GH issue.

## Building blocks

The package has only a handful main concepts or building blocks.

### Rules

The base building block of the declarative mindset baked into this experiment is the notion of a [Rule](gossh.go).
A rule is an interface with a single _Ensure_ function. Ensure does what it sounds like: _Ensure_ ensures the rule is adhered to on the target.

Ensuring is a two-step process - _check_ and _enforce_. Enforcement is only done if the check was not successful.

An example of a rule is [apt.Package](rules/x/apt/apt.go). _Check_ verifies if the apt package is installed or not, and _enforcement_ is done by (un)installs the package.

Rules are made up of imperative code/logic, other declarative rules or a combination of both. Rules can be nested infinitely.

`Rules` are _applied_ to `Targets`.

Implemented/example rules (just to show some ideas):

* file.Exists - creates an empty file if it does not exists
* apt.Package - install/uninstall apt packages
* base.Cmd - run shell commands as Check and Ensure. Check depends on the ExitStatus code.
* base.Meta - for constructing meta-rules on the fly. This is where imperative mode kicks in.


### Target

A [Target](gossh.go) is a bare-metal server, virtual machine or container. It can be localhost, a remote host (SSH) or a docker container on localhost.

(Current) Requirements

* Running Linux
* Bash shell available
* Sudo installed
* SSH'able (remote)

### Inventories

An [Inventory](inventory.go) is a list of Hosts.

## Usage - give it a spin using docker

_Please remeber that this is very experimental_

Prerequisites:

* Go
* Docker

This is what you do:

0. Clone this repo
1. Build and run a SSH enabled Ubuntu container by running `make docker`
2. Cd over to [examples/random](examples/random) and try running it with `go run main.go`
3. Have a look at the output
4. Modify the example script however you like and run again.
5. Kill and remove the container using `make docker-down`

## Motivation

I've recently listened to [Pulumi: Infrastructure as Code with Joe Duffy on Software Engineering Daily](https://softwareengineeringdaily.com/2020/03/19/pulumi-infrastructure-as-code-with-joe-duffy/). The vision and ideas behind Pulumi really resonated with me. The promise of no YAML or DSL - and just using a progamming language and tooling I allready enjoy - was very appealing. Combining a full-fledged programming language (with package management) with a declarative structural representation of the state sounds powerful and like something I would like to have.

Ansible has been my favorite CM tool for a while. It's awesome! But if I'm honest, I'm not really fond of all the YAML. I also find that I ofted need to do quite a lot of imperative things in the playbooks (`register` & `when`, I'm looking at you), which is awkward. What I do love about Ansible though, it it's simplicy and low learing curve and that it is agentless and does all it's work over SSH.

In essence, the experiment aims to take all the things I love about Ansible and bring all the nice things that Pulumi promises, but for configuration, not provisioning.

I think the Go language, typechecking, compile-time checks, standard library, package manager and simplicity makes it perfect starting point for nice configuration management tool.

## Docs

#### SUDO

Gossh building blocks allows commands and rules to run as other users. It is done using Sudo.

## References

### Early feedback Reddit threads

* 2020-03-28 [Show me the gnarliest [Ansible] config you know?](https://www.reddit.com/r/ansible/comments/fq3v0b/show_me_the_gnarliest_config_you_know/)
* 2020-03-27 [Declarative configuration management in Go? Need input on what you think Ansible with no YAML, only Go, should look like](https://www.reddit.com/r/golang/comments/fpjavy/declarative_configuration_management_in_go_need/)

## Inspiration

This project is heavily inspired by

* Pulumi
* Puppet Bolt
* Ansible
* GOSS

## Licence

[GNU General Public License v3.0](LICENSE)
