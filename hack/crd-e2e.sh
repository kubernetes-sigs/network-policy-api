#!/bin/bash

# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

GOPATH="$(mktemp -d)"
readonly GOPATH
CLUSTER_NAME="verify-network-policy-api"
readonly CLUSTER_NAME

export KUBECONFIG="${GOPATH}/.kubeconfig"
export GOFLAGS GO111MODULE GOPATH
export PATH="${GOPATH}/bin:${PATH}"

# Cleanup logic for cleanup on exit
CLEANED_UP=false
cleanup() {
  if [ "$CLEANED_UP" = "true" ]; then
    return
  fi

  if [ "${KIND_CREATE_ATTEMPTED:-}" = true ]; then
    kind delete cluster --name "${CLUSTER_NAME}" || true
  fi
  CLEANED_UP=true
}

trap cleanup INT TERM

# For exit code
res=0

# Install kind
(go install sigs.k8s.io/kind@v0.25.0) || res=$?

# Create cluster
KIND_CREATE_ATTEMPTED=true
kind create cluster --name "${CLUSTER_NAME}" || res=$?
#echo $bases
#echo $patches
for _ in bases patches; do
  go run sigs.k8s.io/controller-tools/cmd/controller-gen rbac:roleName=manager-role crd paths=./apis/... output:crd:dir=./config/crd/bases output:stdout || res=$?
  kubectl kustomize config/crd/standard | kubectl apply -f - || res=$?
  kubectl kustomize config/crd/experimental | kubectl apply -f - || res=$?

  # Temporary workaround for https://github.com/kubernetes/kubernetes/issues/104090
  sleep 8

done

# Clean up and exit
cleanup || res=$?
exit $res
