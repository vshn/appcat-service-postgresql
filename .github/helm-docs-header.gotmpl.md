{{ template "chart.header" . }}
{{ template "chart.deprecationWarning" . }}

{{ template "chart.badgesSection" . }}

{{ template "chart.description" . }}

{{ template "chart.homepageLine" . }}

## Installation

```bash
helm repo add appcat-service-postgresql https://vshn.github.io/appcat-service-postgresql
helm install {{ template "chart.name" . }} appcat-service-postgresql/{{ template "chart.name" . }}
```
