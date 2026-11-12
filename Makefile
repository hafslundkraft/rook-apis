LOCALBIN ?= $(shell pwd)/bin
CONTROLLER_TOOLS_VERSION ?= v0.22.0
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen-$(CONTROLLER_TOOLS_VERSION)
ROOK_VERSION := $(shell cat ROOK_VERSION)
UPSTREAM_DIR ?= $(LOCALBIN)/rook-$(ROOK_VERSION)

# Rook's types reference core/v1 types, which take client-go's apply configurations: without this
# mapping, the builder generated for Rook's embedded SecretReference does not compile.
CLIENT_GO_CORE_V1_AC := k8s.io/client-go/applyconfigurations/core/v1
EXTERNAL_ACS := k8s.io/api/core/v1.SecretKeySelector@$(CLIENT_GO_CORE_V1_AC);k8s.io/api/core/v1.SecretReference@$(CLIENT_GO_CORE_V1_AC)

# `rm -rf` first because controller-gen never deletes a file whose type is gone.
.PHONY: generate
generate: $(CONTROLLER_GEN) ## Generate deepcopy functions and apply configurations.
	rm -rf applyconfiguration/
	$(CONTROLLER_GEN) object paths="./..."
	$(CONTROLLER_GEN) "applyconfiguration:externalApplyConfigurations=$(EXTERNAL_ACS)" paths="./ceph.rook.io/..."

.PHONY: check-upstream
check-upstream: ## Check every type in ceph.rook.io/v1 against Rook at ROOK_VERSION.
	rm -rf $(UPSTREAM_DIR)
	git -c advice.detachedHead=false clone --quiet --depth 1 --branch $(ROOK_VERSION) --filter=blob:none --sparse \
		https://github.com/rook/rook $(UPSTREAM_DIR)
	git -C $(UPSTREAM_DIR) sparse-checkout set pkg/apis/ceph.rook.io/v1
	go run ./hack/check-upstream ./ceph.rook.io/v1 $(UPSTREAM_DIR)/pkg/apis/ceph.rook.io/v1

$(CONTROLLER_GEN):
	mkdir -p $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
	mv $(LOCALBIN)/controller-gen $(CONTROLLER_GEN)
