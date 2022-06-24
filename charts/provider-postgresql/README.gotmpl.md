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

{{ template "chart.sourcesSection" . }}

{{ template "chart.requirementsSection" . }}
<!---
The values below are generated with helm-docs!

Document your changes in values.yaml and let `make chart-docs` generate this section.
-->
{{ template "chart.valuesSection" . }}
