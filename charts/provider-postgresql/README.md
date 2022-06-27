# provider-postgresql

![Version: 0.2.0](https://img.shields.io/badge/Version-0.2.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

VSHN-opinionated PostgreSQL operator for AppCat

## Installation

```bash
helm repo add appcat-service-postgresql https://vshn.github.io/appcat-service-postgresql
helm install provider-postgresql appcat-service-postgresql/provider-postgresql
```
```bash
kubectl apply -f https://github.com/vshn/appcat-service-postgresql/releases/download/provider-postgresql-0.2.0/crds.yaml
```

<!---
The README.md file is automatically generated with helm-docs!

Edit the README.gotmpl.md template instead.
-->

## Handling CRDs

* Always upgrade the CRDs before upgrading the Helm release.
* Watch out for breaking changes in the Provider-Postgresql release notes.

## Webhook support

This chart is capable of rendering the `MutatingWebhookConfiguration` and `ValidatingWebhookConfiguration` required for the operator.
While you can disable webhook support with `webhook.enabled=false`, you'll lose important business functionality.
But it may be easier to set up in testing environments.

In order for webhooks to work, Kubernetes requires a Secret with TLS.
The webhook configuration spec requires the CA bundle to be set as well so that Kubernetes trusts the certificate that is mounted in the operator.

You can set `webhook.externalSecretName` to a Secret that is managed outside of this chart.

### Automatic Certificate Provisioning on OpenShift 4

On Openshift 4, you can make use of the automatic service serving certificate provisioning as documented here:
https://docs.openshift.com/container-platform/4.10/security/certificates/service-serving-certificate.html

In short, setting the annotations as outlined below causes OpenShift to create the Secret and patch the Webhook configuration objects.

To make use of it, configure the following values:
```yaml
service:
  annotations:
    service.beta.openshift.io/serving-cert-secret-name: <secret_name>
webhook:
  externalSecretName: <secret_name>
  annotations:
    service.beta.openshift.io/inject-cabundle: "true"
```

<!---
The values below are generated with helm-docs!

Document your changes in values.yaml and let `make chart-docs` generate this section.
-->
## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Operator image pull policy If set to empty, then Kubernetes default behaviour applies. |
| image.registry | string | `"ghcr.io"` | Operator image registry |
| image.repository | string | `"vshn/appcat-service-postgresql"` | Operator image repository |
| image.tag | string | `"latest"` | Operator image tag |
| imagePullSecrets | list | `[]` | List of image pull secrets if custom image is behind authentication. |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| operator.args | list | `[]` | Overrides arguments passed to the entrypoint |
| podAnnotations | object | `{}` | Annotations to add to the Pod spec. |
| podSecurityContext | object | `{}` | Security context to add to the Pod spec. |
| replicaCount | int | `1` | How many operator pods should run. Follower pods reduce interruption time as they're on hot standby when leader is unresponsive. |
| resources | object | `{}` |  |
| securityContext | object | `{}` | Container security context |
| service.annotations | object | `{}` | Annotations to add to the service |
| service.port | int | `80` | Service port number |
| service.type | string | `"ClusterIP"` | Service type |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and `.create` is `true`, a name is generated using the fullname template |
| tolerations | list | `[]` |  |
| webhook.annotations | object | `{}` | Annotations to add to the webhook configuration resources. |
| webhook.caBundle | string | `""` | Certificate in PEM format for the ValidatingWebhookConfiguration. |
| webhook.certificate | string | `""` | Certificate in PEM format for the TLS secret. |
| webhook.enabled | bool | `true` | Enable admission webhooks |
| webhook.externalSecretName | string | `""` | Name of an existing or external Secret with TLS to mount in the operator. The secret is expected to have `tls.crt` and `tls.key` keys. Note: You will still need to set `.caBundle` if the certificate is not verifiable (self-signed) by Kubernetes. |
| webhook.privateKey | string | `""` | Private key in PEM format for the TLS secret. |

<!---
Common/Useful Link references from values.yaml
-->
[resource-units]: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes
[prometheus-operator]: https://github.com/coreos/prometheus-operator
