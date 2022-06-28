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
* `sed` (or `gsed` for Mac)

Some other requirements (e.g. `kind`) will be compiled on-the-fly and put in the local cache dir `.kind` as needed.

### Common make targets

* `make build` to build the binary and docker image
* `make generate` to (re)generate additional code artifacts
* `make test` run test suite
* `make local-install` to install the operator in local cluster
* `make install-samples` to run the provider in local cluster and apply a sample instance
* `make run-operator` to run the code in operator mode against local cluster

See all targets with `make help`

### QuickStart Demonstration

TL;DR: `make local-install install-samples`

### Kubernetes Webhook Troubleshooting

The provider comes with mutating and validation admission webhook server.
However, in this setup this currently only works in the kind cluster when installed as package using `make package-install`.

To test and troubleshoot the webhooks, do a port-forward and send an admission request sample of the spec:
```bash
# port-forward webhook server
kubectl -n crossplane-system port-forward $(kubectl -n crossplane-system get pods -o name -l pkg.crossplane.io/provider=appcat-service-postgresql) 9443:9443

# send an admission request
curl -k -v -H "Content-Type: application/json" --data @package/samples/admission.k8s.io_admissionreview.json https://localhost:9443/validate-postgresql-appcat-vshn-io-v1alpha1-postgresqlstandalone
```
