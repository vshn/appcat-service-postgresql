
webhook_gen_result = chart/templates/manifests.yaml
rbac_gen_result = chart/templates/role.yaml

.PHONY: $(webhook_gen_result)
$(webhook_gen_result):
	@yq -i e '.metadata.name="{{ include \"provider-postgresql.fullname\" . }}", del(.metadata.creationTimestamp)' $@
	@yq -i e '.metadata.labels.replace="LABELS"' $@
	@yq -i e '.webhooks[0].clientConfig.caBundle="{{ .Values.webhook.caBundle }}"' $@
	@yq -i e '.webhooks[0].clientConfig.service.name="{{ include \"provider-postgresql.fullname\" . }}"' $@
	@yq -i e '.webhooks[0].clientConfig.service.namespace="{{ .Release.Namespace }}"' $@
	@sed -i -e 's/replace: LABELS/{{- include "provider-postgresql.labels" . | nindent 4 }}/g' $@
	@mv $(webhook_gen_result) chart/templates/webhook.yaml

.PHONY: $(rbac_gen_result)
$(rbac_gen_result):
	@yq -i e '.metadata.name="{{ include \"provider-postgresql.fullname\" . }}", del(.metadata.creationTimestamp)' $@
	@yq -i e '.metadata.labels.replace="LABELS"' $@
	@sed -i -e 's/replace: LABELS/{{- include "provider-postgresql.labels" . | nindent 4 }}/g' $@
