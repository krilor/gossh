# Target

## Create, Open and Append vs Put and Get

Originally, the idea was to support Create, Open and Append methods on target. That is, methods that would return io.Reader/Wroter. This idea was scrapped because

* Append mode is not supported on all SFTP servers, like RHEL6-based Linux distros
* Since gossh is intended to do work over SSH, doing small reads/writes will hurt, performance wise
* It greatly simplifies the sudo implementations, both on local and remote targets
* Put and Get are methods used in scp, so one could envision supporting scp-only targets
