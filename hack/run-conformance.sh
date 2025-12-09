#!/usr/bin/env bash

# Copyright 2025 The Kubernetes Authors.
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

usage() {
  cat <<EOF
Usage: $0 [commands]

Commands:

  enable_ipv4_and_ipv6_forwarding
  setup_environment
  create_cluster
  install_apis
  install_np_impl
  get_cluster_status
  run_tests
  export_logs

  all (run all steps)

Example:

  # Run the conformance test locally. Skip some of the github
  # action steps.
  ARTIFACTS_DIR=/tmp TMP_DIR=/tmp \\
    $0 create_cluster \\
      install_kube_network_policices_from_main \\
      install_apis \\
      run_tests

EOF
}

if [[ "$#" -eq 0 ]]; then
  usage
  exit 1
fi

for arg in "$@"; do
  case "$arg" in
    -h|--help)
      usage
      exit 0
      ;;
  esac
done

# Use pre-installed kubectl and kind if not running setup_environment.
KUBECTL=$(which kubectl || echo "not-found")
KIND=$(which kind || echo "not-found")

echo "[RUN] found KUBECTL=${KUBECTL}"
echo "[RUN] found KIND=${KIND}"

# These values can be overriden by external vars.
IP_FAMILY="${IP_FAMILY:-ipv4}"
GO_VERSION="${GO_VERSION:-1.24}"
K8S_VERSION="${K8S_VERSION:-v1.33.0}"
KIND_VERSION="${KIND_VERSION:-v0.30.0}"
IMAGE_NAME="${IMAGE_NAME:-registry.k8s.io/networking/kube-network-policies}"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-kind}"
NPAPI_VERSION="${NPAPI_VERSION:-v1alpha2}"
ARTIFACTS_DIR="${ARTIFACTS_DIR:-./_artifacts}"

CLEANUP_ON_EXIT="${CLEANUP_ON_EXIT:-true}"
CLEANUP_CLUSTER="false"
if [[ -v TMP_DIR ]]; then
  CLEANUP_ON_EXIT="false"
else
  TMP_DIR=$(mktemp -d)
fi
echo "[RUN] Using TMP_DIR='${TMP_DIR}'"

cleanup() {
  if [[ ${CLEANUP_CLUSTER} = "true" ]]; then
    echo "[RUN] Stopping Kind cluster if it exists"
    ${KIND} delete cluster --name "${KIND_CLUSTER_NAME}" || true
  fi
  if [[ ${CLEANUP_ON_EXIT} != "true" ]]; then
    echo "[RUN] Warning: not cleaning up TMP_DIR (${TMP_DIR})"
  else
    echo "[RUN] Removing TMP_DIR ${TMP_DIR})"
    rm -rf "${TMP_DIR}"
  fi
}
trap cleanup EXIT INT TERM

enable_ipv4_and_ipv6_forwarding() {
  echo "[RUN] Enabling ipv4 and ipv6 forwarding"
  sudo sysctl -w net.ipv6.conf.all.forwarding=1
  sudo sysctl -w net.ipv4.ip_forward=1
}

setup_environment() {
  echo "[RUN] Setting up environment (download dependencies)"
  # kubectl
  curl -L https://dl.k8s.io/"${K8S_VERSION}"/bin/linux/amd64/kubectl -o "${TMP_DIR}"/kubectl
  # kind
  curl -Lo "${TMP_DIR}"/kind https://kind.sigs.k8s.io/dl/"${KIND_VERSION}"/kind-linux-amd64
  # Install
  sudo cp "${TMP_DIR}"/kubectl /usr/local/bin/kubectl
  sudo cp "${TMP_DIR}"/kind /usr/local/bin/kind
  sudo chmod +x /usr/local/bin/kubectl /usr/local/bin/kind

  KUBECTL="/usr/local/bin/kubectl"
  KIND="/usr/local/bin/kind"

  echo "[RUN] using KUBECTL=${KUBECTL}"
  echo "[RUN] using KIND=${KIND}"

  if [[ ! -d "${ARTIFACTS_DIR}" ]]; then
    mkdir -p "${ARTIFACTS_DIR}"
  fi
  echo "[RUN] using ARTIFACTS_DIR=${ARTIFACTS_DIR}"
}

create_cluster() {
  echo "[RUN] Creating cluster"

  CLEANUP_CLUSTER=true

  local kind_cfg="# Kind cluster configuration.
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  ipFamily: ${IP_FAMILY}
nodes:
- role: control-plane
- role: worker
- role: worker
"
  echo "${kind_cfg}"
  echo "${kind_cfg}" | ${KIND} create cluster \
    --name "${KIND_CLUSTER_NAME}" \
    --image kindest/node:"${K8S_VERSION}"  \
    -v7 --wait 1m --retain --config=-


  # newer kind version ship a kindnet version that implements network policies.
  # we need to downgrade it to a version that does not support network policies
  # to be able to avoid conflicts with kube-network-policies
  local kindnetd_image="docker.io/kindest/kindnetd:v20230809-80a64d96"
  ${KUBECTL} -n kube-system set image ds kindnet kindnet-cni="${kindnetd_image}"

  # dump the kubeconfig
  ${KIND} get kubeconfig --name "${KIND_CLUSTER_NAME}" > "${ARTIFACTS_DIR}/kubeconfig.conf"
}

install_apis() {
  echo "[RUN] Installing network policy APIs"
  ${KUBECTL} apply -f ./config/crd/standard/policy.networking.k8s.io_clusternetworkpolicies.yaml
}

install_np_impl() {
  echo "[RUN] Installing kube-network-policies from main"
  ( # Subshell to preserve cwd.
    cd "$TMP_DIR"
    # Clone the repo
    git clone --depth 1 https://github.com/kubernetes-sigs/kube-network-policies.git
    cd kube-network-policies/
    # Build the image with the network policy API support
    REGISTRY="registry.k8s.io/networking" IMAGE_NAME="kube-network-policies" TAG="test" \
      make image-build-npa-"${NPAPI_VERSION}"
    # Preload the image in the kind cluster
    ${KIND} \
      load docker-image registry.k8s.io/networking/kube-network-policies:test-npa-"${NPAPI_VERSION}" \
        --name "${KIND_CLUSTER_NAME}"
    # Install kube-network-policies with the image built from main
    sed -i "s#registry.k8s.io/networking/kube-network-policies.*#registry.k8s.io/networking/kube-network-policies:test-npa-${NPAPI_VERSION}# install-cnp.yaml"
    ${KUBECTL} apply -f ./install-cnp.yaml
  )
}

get_cluster_status() {
  echo "[RUN] Getting cluster status"
  # wait for the network to be ready
  sleep 5
  ${KUBECTL} get nodes -o wide
  ${KUBECTL} get pods -A
  ${KUBECTL} wait \
    --timeout=1m \
    --for=condition=ready pods \
    --namespace=kube-system \
    -l k8s-app=kube-dns
  ${KUBECTL} wait \
    --timeout=1m \
    --for=condition=ready pods \
    --namespace=kube-system \
    -l app=kube-network-policies
}

run_tests() {
  echo "[RUN] Running tests"
  go mod download
  REPO_VERSION=$(git describe --always --dirty)
  go test  -v ./conformance \
    -run TestConformanceProfiles \
    -args \
    --conformance-profiles=ClusterNetworkPolicy \
    --organization=kubernetes \
    --project=kube-network-policies \
    --url=https://github.com/kubernetes-sigs/kube-network-policies \
    --version="${REPO_VERSION}" \
    --contact=https://github.com/kubernetes-sigs/kube-network-policies/issues/new \
    --additional-info=https://github.com/kubernetes-sigs/kube-network-policies
}

export_logs() {
  echo "[RUN] Exporting logs"
  ${KIND} export logs --name "${KIND_CLUSTER_NAME}" -v7 "${ARTIFACTS_DIR}/logs"
}

all() {
  enable_ipv4_and_ipv6_forwarding
  setup_environment
  create_cluster
  install_apis
  install_np_impl
  get_cluster_status
  run_tests
  export_logs
}

for arg in "$@"; do
  case "$arg" in
    all)
      all
      exit 0
      ;;
  esac
done

# Run through the commands.
for arg in "$@"; do
  case "$arg" in
    enable_ipv4_and_ipv6_forwarding)
      enable_ipv4_and_ipv6_forwarding
      ;;
    setup_environment)
      setup_environment
      ;;
    create_cluster)
      create_cluster
      ;;
    install_apis)
      install_apis
      ;;
    install_np_impl)
      install_np_impl
      ;;
    get_cluster_status)
      get_cluster_status
      ;;
    run_tests)
      run_tests
      ;;
    export_logs)
      export_logs
      ;;
    *)
      echo >&2 "Invalid command: $arg"
      usage
      exit 1
      ;;
  esac
done
