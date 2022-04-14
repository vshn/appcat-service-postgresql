# AppCat Service PostgreSQL

[![Build](https://img.shields.io/github/workflow/status/vshn/appcat-service-postgresql/Test)][build]
![Go version](https://img.shields.io/github/go-mod/go-version/vshn/appcat-service-postgresql)
[![Version](https://img.shields.io/github/v/release/vshn/appcat-service-postgresql)][releases]
[![Maintainability](https://img.shields.io/codeclimate/maintainability/vshn/appcat-service-postgresql)][codeclimate]
[![Coverage](https://img.shields.io/codeclimate/coverage/vshn/appcat-service-postgresql)][codeclimate]
[![GitHub downloads](https://img.shields.io/github/downloads/vshn/appcat-service-postgresql/total)][releases]

[build]: https://github.com/vshn/appcat-service-postgresql/actions?query=workflow%3ATest
[releases]: https://github.com/vshn/appcat-service-postgresql/releases
[codeclimate]: https://codeclimate.com/github/vshn/appcat-service-postgresql

This service provider installs PostgreSQL instances of various architecture types using the AppCat and Crossplane frameworks.

## Local Development

### Requirements

* `docker`
* `go`
* `helm`
* `kubectl`
* `yq`

### Common make targets

* `make build` to build the binary and docker image
* `make generate` to (re)generate additional code artifacts
* `make test` run test suite
* `make package-install` to package the provider and install via Crossplane
* `make install-samples` to run the provider in local cluster and apply a sample instance
* `make run-operator` to run the code in operator mode against local cluster

See all targets with `make help`
