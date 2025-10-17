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

# Command-line flags passed to "go test" for the conformance
# test. These are passed after the "-args" flag.
CONFORMANCE_FLAGS ?=

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: ## Run go test
	go test ./apis/...
	go test ./pkg/...

.PHONY: generate
generate: ## Generate the code from the API definitions.
	./hack/update-codegen.sh

all: generate fmt vet ## Runs all the development targets

.PHONY: verify
verify: ## Verify the code.
	hack/verify-all.sh -v

crd-e2e: ## Run the CRD e2e tests.
	hack/crd-e2e.sh -v

.PHONY: conformance
conformance: ## Run the conformance tests.
	go test ${GO_TEST_FLAGS} -v ./conformance \
		-run '^TestConformance$$' -args ${CONFORMANCE_FLAGS}

.PHONY: conformance-profiles
conformance-profiles: ## Run the conformance profiles.
	go test ${GO_TEST_FLAGS} -v ./conformance \
		-run '^TestConformanceProfiles$$' -args ${CONFORMANCE_FLAGS}

.PHONY: conformance-profiles-default
conformance-profiles-default: ## Run the default conformance profile.
	go test ${GO_TEST_FLAGS} -v ./conformance \
		-run '^TestConformanceProfiles$$' -args \
		--conformance-profiles=ClusterNetworkPolicy

##@ Deployment
install: generate ## Install standard CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl kustomize config/crd/standard | kubectl apply -f -

uninstall: generate kustomize ## Uninstall standard CRDs from the K8s cluster specified in ~/.kube/config.
	kubectl kustomize config/crd/standard | kubectl delete -f -

##@ Deployment
install-experimental: generate ## Install experimental CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl kustomize config/crd/experimental | kubectl apply -f -

uninstall-experimental: generate kustomize ## Uninstall experimental CRDs from the K8s cluster specified in ~/.kube/config.
	kubectl kustomize config/crd/experimental | kubectl delete -f -

.PHONY: docs ## Build the documentation website
docs:
	./hack/make-docs.sh
 
.PHONY: local-docs ## Deploy the docs locally 
local-docs:
	mkdocs serve

.PHONY: build-install-yaml ## Build the install YAML.
build-install-yaml:
	./hack/build-install-yaml.sh

.PHONY: crdtest # use go to check build freshness
crdtest: ## Build the crdtest executable
	go test -c ./pkg/crdtest -o crdtest
