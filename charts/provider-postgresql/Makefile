webhook_gen_src = templates/manifests.yaml
webhook_gen_tgt = templates/webhook.yaml
rbac_gen_src = templates/role.yaml
rbac_gen_tgt = templates/clusterrole.yaml

ifeq ($(shell uname -s),Darwin)
	sed := sed -i ""
else
	sed := sed -i
endif

$(webhook_gen_tgt):
	@yq -i e '.metadata.name="{{ include \"provider-postgresql.fullname\" . }}", del(.metadata.creationTimestamp)' $(webhook_gen_src)
	@yq -i e '.metadata.labels.replace="LABELS"' $(webhook_gen_src)
	@yq -i e '.webhooks[0].clientConfig.caBundle="{{ .Values.webhook.caBundle }}"' $(webhook_gen_src)
	@yq -i e '.webhooks[0].clientConfig.service.name="{{ include \"provider-postgresql.fullname\" . }}"' $(webhook_gen_src)
	@yq -i e '.webhooks[0].clientConfig.service.namespace="{{ .Release.Namespace }}"' $(webhook_gen_src)
	@$(sed) -e 's/replace: LABELS/{{- include "provider-postgresql.labels" . | nindent 4 }}/g' $(webhook_gen_src)
	@mv $(webhook_gen_src) $(webhook_gen_tgt)

$(rbac_gen_tgt):
	@yq -i e '.metadata.name="{{ include \"provider-postgresql.fullname\" . }}", del(.metadata.creationTimestamp)' $(rbac_gen_src)
	@yq -i e '.metadata.labels.replace="LABELS"' $(rbac_gen_src)
	@$(sed) -e 's/replace: LABELS/{{- include "provider-postgresql.labels" . | nindent 4 }}/g' $(rbac_gen_src)
	@mv $(rbac_gen_src) $(rbac_gen_tgt)

.PHONY: prepare
prepare: $(rbac_gen_tgt) $(webhook_gen_tgt) ## Helmify generated artifacts
