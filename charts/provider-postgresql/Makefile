webhook_gen_src = ../../package/webhook/manifests.yaml
webhook_gen_tgt = templates/webhook.yaml

rbac_gen_src = ../../package/rbac/role.yaml
rbac_gen_tgt = templates/clusterrole.yaml

ifeq ($(shell uname -s),Darwin)
	sed := gsed -i
else
	sed := sed -i
endif

$(webhook_gen_tgt):
	@cp $(webhook_gen_src) $@
	@yq -i e '.metadata.name="{{ include \"provider-postgresql.fullname\" . }}", del(.metadata.creationTimestamp)' $@
	@yq -i e '.metadata.labels.replace="LABELS"' $@
	@yq -i e '.metadata.annotations.replace="ANNOTATIONS"' $@
	@yq -i e '.webhooks[0].clientConfig.caBundle="{{ . }}"' $@
	@yq -i e '.webhooks[0].clientConfig.service.name="{{ include \"provider-postgresql.fullname\" . }}"' $@
	@yq -i e '.webhooks[0].clientConfig.service.namespace="{{ .Release.Namespace }}"' $@
	@$(sed) -e '1s/^/{{- if .Values.webhook.enabled -}}\n/' $@
	@$(sed) -e '/^      caBundle:.*/i {{- with .Values.webhook.caBundle }}' $@
	@$(sed) -e '/^      caBundle:.*/a {{- end }}' $@
	@$(sed) -e 's/replace: LABELS/{{- include "provider-postgresql.labels" . | nindent 4 }}/g' $@
	@$(sed) -e 's/replace: ANNOTATIONS/{{- toYaml .Values.webhook.annotations | nindent 4 }}/g' $@
	@echo "{{- end -}}" >> $@

$(rbac_gen_tgt):
	@cp $(rbac_gen_src) $@
	@yq -i e '.metadata.name="{{ include \"provider-postgresql.fullname\" . }}", del(.metadata.creationTimestamp)' $@
	@yq -i e '.metadata.labels.replace="LABELS"' $@
	@$(sed) -e 's/replace: LABELS/{{- include "provider-postgresql.labels" . | nindent 4 }}/g' $@

.PHONY: prepare
prepare: $(rbac_gen_tgt) $(webhook_gen_tgt) ## Helmify generated artifacts

.PHONY: clean
clean: ## Clean generated artifacts
	rm -rf $(rbac_gen_tgt) $(webhook_gen_tgt)
