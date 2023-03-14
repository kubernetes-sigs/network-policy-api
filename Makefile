# Get the current working directory for this code
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

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

CLIENTSET_NAME ?= versioned
CLIENTSET_PKG_NAME ?= clientset
API_PKG ?= sigs.k8s.io/network-policy-api
API_GROUP_NAME ?= policy.networking.k8s.io
API_DIR ?= ${API_PKG}/apis/v1alpha1
OUTPUT_PKG ?= sigs.k8s.io/network-policy-api/pkg/client
COMMON_FLAGS ?= --go-header-file $(shell pwd)/hack/boilerplate.go.txt

.PHONY: manifests
manifests: ## Generate ClusterRole and CustomResourceDefinition objects.
	go run sigs.k8s.io/controller-tools/cmd/controller-gen rbac:roleName=manager-role crd paths=./apis/... output:crd:dir=./config/crd/bases output:stdout

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: generate
generate: generate-setup generate-deepcopy generate-typed-clients generate-typed-listers generate-typed-informers generate-cleanup

.PHONY: generate-setup
generate-setup:
	# Even when modules are enabled, the code-generator tools always write to
	# a traditional GOPATH directory, so fake on up to point to the current
	# workspace.
	mkdir -p "${GOPATH}/src/sigs.k8s.io"
	ln -sf "${ROOT_DIR}" "${GOPATH}/src/sigs.k8s.io/network-policy-api"

.PHONY: generate-cleanup
generate-cleanup:
	rm -rf "${GOPATH}/src/sigs.k8s.io"

.PHONY: generate-deepcopy
generate-deepcopy: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	go run sigs.k8s.io/controller-tools/cmd/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: generate-typed-clients
generate-typed-clients: ## Generate typed client code
	go run k8s.io/code-generator/cmd/client-gen \
	--clientset-name "${CLIENTSET_NAME}" \
	--input-base "" \
	--input "${API_DIR}" \
	--output-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}" \
	${COMMON_FLAGS}

.PHONY: generate-typed-listers
generate-typed-listers: ## Generate typed listers code
	go run k8s.io/code-generator/cmd/lister-gen \
	--input-dirs "${API_DIR}" \
	--output-package "${OUTPUT_PKG}/listers" \
	${COMMON_FLAGS}

.PHONY: generate-typed-informers
generate-typed-informers: ## Generate typed informers code
	go run k8s.io/code-generator/cmd/informer-gen \
	--input-dirs "${API_DIR}" \
	--versioned-clientset-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}/${CLIENTSET_NAME}" \
	--listers-package "${OUTPUT_PKG}/listers" \
	--output-package "${OUTPUT_PKG}/informers" \
	${COMMON_FLAGS}

all: generate manifests fmt vet ## Runs all the development targets

.PHONY: verify
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
