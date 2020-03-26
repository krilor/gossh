# GOSSH

This repo is an experiement with creating a declarative configuration management tool using Golang. Think Ansible, but no yaml, just plain Go. WOW - all teh power!

The project is in a super-early state. **I am looking for API/usage/naming convention input and ideas in general on how to approach this problem.** Have a look at [main.go](main.go) to get a feel for what I currently think it would look like to use the tool/SDK. If you have any ideas please reach out through a GH issue.

The tool will probably be limited to configuring linux machines over SSH.

## Rules

The basic building block of the declarative mindset baked into this experiment is the notion of a [Rule](rule/rule.go).
A rule is an interface with two functions, _Check_ and _Ensure_, that does what it sounds like. _Check_ checks the state, _Ensure_ ensures the state.

An example of such a rule is [apt.Package](apt/apt.go). _Check_ verifies if the apt package is installed or not, and _Ensure_ (un)installs the package. Both _Check_ and _Ensure_ are dependent on the declared PackageStatus (installed or not installed) and package name.

`Rules` run on `Machines`.

Implemented rules (just to show some ideas):

* file.Exists - creates an empty file if it does not exists
* apt.Package - install/uninstall apt packages
* rule.Cmd - run shell commands as Check and Ensure. Check depends on the ExitStatus code.
* rule.Meta - for constructing meta-rules on the fly. This is where imperative mode kicks in.

## Usage - give it a spin

_Please remeber that this is very experimental_

If you want to try it out, fire up a container or VM with SSH enabled. Then edit the `machine.New()` line in [main.go](main.go) and run using `go run main.go`.

Requirements

* Go
* a target vm running debian linux
* ssh key in ssh-agent

## Motivation

I've recently listened to [Pulumi: Infrastructure as Code with Joe Duffy on Software Engineering Daily](https://softwareengineeringdaily.com/2020/03/19/pulumi-infrastructure-as-code-with-joe-duffy/). The vision and ideas behind Pulumi really resonated with me. The promise of no YAML or DSL - and just using a progamming language and tooling I allready enjoy - was very appealing. Combining a full-fledged programming language (with package management) with a declarative structural representation of the state sounds powerful and like something I would like to have.

Ansible has been my favorite CM tool for a while. It's awesome! But if I'm honest, I'm not really fond of all the YAML. I also find that I ofted need to do quite a lot of imperative things in the playbooks (`register` & `when`, I'm looking at you), which is awkward. What I do love about Ansible though, it it's simplicy and low learing curve and that it is agentless and does all it's work over SSH.

In essence, the experiment aims to take all the things I love about Ansible and bring all the nice things that Pulumi promises, but for configuration, not provisioning.

I think the Go language, typechecking, compile-time checks, standard library, package manager and simplicity makes it perfect starting point for nice configuration management tool.

## Inspiration

This project is heavily inspired by

* Pulumi
* Ansible
* GOSS

## Licence

[GNU General Public License v3.0](LICENSE)
