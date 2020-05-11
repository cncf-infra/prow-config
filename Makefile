# AWS cluster variables
# PROJECT ?= cncf-apisnoop
# ZONE ?= ap-southeast-2
CLUSTER ?= prow-dev

# prow docker image variables
REPO ?= gcr.io/k8s-prow
TAG ?= v20200507-36c6a27f0
# v20200423-af610499d

CWD = $(shell pwd)

# ensure kubectl is pointing to the right cluster based on PROJECT/ZONE/CLUSTER
# gcloud container clusters get-credentials $(CLUSTER) --project=$(PROJECT) --zone=$(ZONE)
#	aws eks --region $(ZONE) update-kubeconfig --name $(CLUSTER)
get-cluster-credentials:
	aws eks update-kubeconfig --name $(CLUSTER)

.PHONY: get-cluster-credentials

update-config: get-cluster-credentials check-config
	kubectl create configmap config --from-file=config.yaml=$(CWD)/config.yaml --dry-run -o yaml | kubectl replace configmap config -f -

update-plugins: get-cluster-credentials check-config
	kubectl create configmap plugins --from-file=plugins.yaml=$(CWD)/plugins.yaml --dry-run -o yaml | kubectl replace configmap plugins -f -

.PHONY: update-config update-plugins

check-config:
	docker run -v $(CWD):/prow $(REPO)/checkconfig:$(TAG) \
		--config-path /prow/config.yaml \
		--plugin-config /prow/plugins.yaml \
		--strict \
		--warnings=mismatched-tide-lenient \
		--warnings=tide-strict-branch \
		--warnings=needs-ok-to-test \
		--warnings=validate-owners \
		--warnings=missing-trigger \
		--warnings=validate-urls \
		--warnings=unknown-fields

.PHONY: check-config
