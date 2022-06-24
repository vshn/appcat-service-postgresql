```bash
kubectl apply -f https://github.com/vshn/appcat-service-postgresql/releases/download/provider-postgresql-{{ template "chart.version" . }}/crds.yaml
```

<!---
The README.md file is automatically generated with helm-docs!

Edit the README.gotmpl.md template instead.
-->

## Handling CRDs

* Always upgrade the CRDs before upgrading the Helm release.
* Watch out for breaking changes in the {{ title .Name }} release notes.

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

{{ template "chart.sourcesSection" . }}

{{ template "chart.requirementsSection" . }}
<!---
The values below are generated with helm-docs!

Document your changes in values.yaml and let `make chart-docs` generate this section.
-->
{{ template "chart.valuesSection" . }}
