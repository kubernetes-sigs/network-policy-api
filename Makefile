# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: ## Generate ClusterRole and CustomResourceDefinition objects.
	go run sigs.k8s.io/controller-tools/cmd/controller-gen rbac:roleName=manager-role crd paths=./apis/... output:crd:dir=./config/crd/bases output:stdout

generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	go run sigs.k8s.io/controller-tools/cmd/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

all: generate manifests fmt vet ## Runs all the development targets

verify:
	hack/verify-all.sh -v

crd-e2e:
	hack/crd-e2e.sh -v

##@ Deployment
install: manifests ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl kustomize config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	kubectl kustomize config/crd | kubectl delete -f -

.PHONY: docs ## Build the documentation website
docs:
	./hack/make-docs.sh
 
.PHONY: local-docs ## Deploy the docs locally 
local-docs:
	mkdocs serve	
