#!/usr/bin/env bash
# shellcheck disable=SC2086
#Double quote to prevent globbing and word splitting. dont apply for this specific scenario

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

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
readonly SCRIPT_ROOT

# Keep outer module cache so we don't need to redownload them each time.
# The build cache already is persisted.
GOMODCACHE="$(go env GOMODCACHE)"
readonly GOMODCACHE
readonly GO111MODULE="on"
readonly GOFLAGS="-mod=readonly"
GOPATH="$(mktemp -d)"
readonly GOPATH
readonly MIN_REQUIRED_GO_VER="1.19"

function go_version_matches {
  go version | perl -ne "exit 1 unless m{go version go([0-9]+.[0-9]+)}; exit 1 if (\$1 < ${MIN_REQUIRED_GO_VER})"
  return $?
}

if ! go_version_matches; then
  echo "Go v${MIN_REQUIRED_GO_VER} or later is required to run code generation"
  exit 1
fi

export GOMODCACHE GO111MODULE GOFLAGS GOPATH

readonly APIS_PKG=sigs.k8s.io/network-policy-api
readonly OUTPUT_PKG=sigs.k8s.io/network-policy-api/pkg/client
readonly OUTPUT_DIR=pkg/client
readonly APIS_PATH=apis
readonly CLIENTSET_NAME=versioned
readonly CLIENTSET_PKG_NAME=clientset
readonly APPLYCONFIG_PKG_NAME=applyconfiguration

readonly COMMON_FLAGS="${VERIFY_FLAG:-} --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.generatego.txt"

echo "Generating CRDs"
go run ./pkg/generator

INPUT_DIRS_SPACE=""
INPUT_DIRS_CLIENTSET=""

mapfile -t VERSIONS < <(find "${APIS_PATH}" -maxdepth 1 -type d -name "v*" -printf '%f\n' | LC_ALL=C sort -u)
for VERSION in "${VERSIONS[@]}"; do
  INPUT_DIRS_SPACE+="${APIS_PKG}/${APIS_PATH}/${VERSION} "
  INPUT_DIRS_CLIENTSET+="${APIS_PATH}/${VERSION},"
done

INPUT_DIRS_SPACE="${INPUT_DIRS_SPACE%,}" # drop trailing space
INPUT_DIRS_CLIENTSET="${INPUT_DIRS_CLIENTSET%,}" # drop trailing comma

echo "Generating applyconfig at ${APIS_PKG}/${APPLYCONFIG_PKG_NAME}"
go run k8s.io/code-generator/cmd/applyconfiguration-gen \
  --output-pkg "${OUTPUT_PKG}/${APPLYCONFIG_PKG_NAME}" \
  --output-dir "${OUTPUT_DIR}/${APPLYCONFIG_PKG_NAME}" \
  ${COMMON_FLAGS} \
  ${INPUT_DIRS_SPACE}

echo "Generating clientset at ${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}"
go run k8s.io/code-generator/cmd/client-gen \
  --clientset-name "${CLIENTSET_NAME}" \
  --input-base "${APIS_PKG}" \
  --input "${INPUT_DIRS_CLIENTSET}" \
  --output-dir "${OUTPUT_DIR}/${CLIENTSET_PKG_NAME}" \
  --output-pkg "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}" \
  --apply-configuration-package "${OUTPUT_PKG}/${APPLYCONFIG_PKG_NAME}" \
  ${COMMON_FLAGS}

echo "Generating listers at ${OUTPUT_PKG}/listers"
go run k8s.io/code-generator/cmd/lister-gen \
  --output-dir "${OUTPUT_DIR}/listers" \
  --output-pkg "${OUTPUT_PKG}/listers" \
  ${COMMON_FLAGS} \
  ${INPUT_DIRS_SPACE}

echo "Generating informers at ${OUTPUT_PKG}/informers"
go run k8s.io/code-generator/cmd/informer-gen \
  --versioned-clientset-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}/${CLIENTSET_NAME}" \
  --listers-package "${OUTPUT_PKG}/listers" \
  --output-dir "${OUTPUT_DIR}/informers" \
  --output-pkg "${OUTPUT_PKG}/informers" \
  ${COMMON_FLAGS} \
  ${INPUT_DIRS_SPACE}

echo "Generating register"
go run k8s.io/code-generator/cmd/register-gen \
  --output-file "zz_generated.register.go" \
  ${COMMON_FLAGS} \
  ${INPUT_DIRS_SPACE}

echo "Generating deepcopy at ${APIS_PATH}"
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
  object:headerFile="${SCRIPT_ROOT}/hack/boilerplate.generatego.txt" \
  paths="./${APIS_PATH}"
