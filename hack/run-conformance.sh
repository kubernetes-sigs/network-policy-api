#!/usr/bin/env bash

# Copyright 2024 The Kubernetes Authors.
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

# This script is a helper script to run the conformance tests locally.
# It is a translation of the .github/workflows/conformance.yml file.

set -o errexit
set -o nounset
set -o pipefail

#
# All the following variables are inherited from the github workflow
#
# The github workflow sets the IP_FAMILY using the matrix strategy.
# For this script we default to ipv4, but can be overridden by setting the
# IP_FAMILY environment variable.
IP_FAMILY="${IP_FAMILY:-ipv4}"
GO_VERSION="${GO_VERSION:-1.24}"
K8S_VERSION="${K8S_VERSION:-v1.33.0}"
KIND_VERSION="${KIND_VERSION:-v0.30.0}"
IMAGE_NAME="${IMAGE_NAME:-registry.k8s.io/networking/kube-network-policies}"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-kind}"
NPAPI_VERSION="${NPAPI_VERSION:-v1alpha2}"

#
# The following functions are a translation of the steps in the github workflow
#
enable_ipv4_and_ipv6_forwarding() {
  echo "Enabling ipv4 and ipv6 forwarding"
  sudo sysctl -w net.ipv6.conf.all.forwarding=1
  sudo sysctl -w net.ipv4.ip_forward=1
}

setup_environment() {
  echo "Setting up environment (download dependencies)"
  TMP_DIR=$(mktemp -d)
  # kubectl
  curl -L https://dl.k8s.io/"${K8S_VERSION}"/bin/linux/amd64/kubectl -o "${TMP_DIR}"/kubectl
  # kind
  curl -Lo "${TMP_DIR}"/kind https://kind.sigs.k8s.io/dl/"${KIND_VERSION}"/kind-linux-amd64
  # Install
  sudo cp "${TMP_DIR}"/kubectl /usr/local/bin/kubectl
  sudo cp "${TMP_DIR}"/kind /usr/local/bin/kind
  sudo chmod +x /usr/local/bin/kubectl /usr/local/bin/kind
}

create_multinode_cluster() {
  echo "Creating multi node cluster"
  # output_dir
  mkdir -p _artifacts
  # create cluster
  cat <<EOF | /usr/local/bin/kind create cluster \
    --name "${KIND_CLUSTER_NAME}"           \
    --image kindest/node:"${K8S_VERSION}"  \
    -v7 --wait 1m --retain --config=-
  kind: Cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  networking:
    ipFamily: ${IP_FAMILY}
  nodes:
  - role: control-plane
  - role: worker
  - role: worker
EOF
  # newer kind version ship a kindnet version that implements network policies.
  # we need to downgrade it to a version that does not support network policies
  # to be able to avoid conflicts with kube-network-policies
  kubectl -n kube-system set image ds kindnet kindnet-cni=docker.io/kindest/kindnetd:v20230809-80a64d96
  # dump the kubeconfig for later
  /usr/local/bin/kind get kubeconfig --name "${KIND_CLUSTER_NAME}" > _artifacts/kubeconfig.conf
}

install_network_policy_apis() {
  echo "Installing network policy APIs"
  /usr/local/bin/kubectl apply -f ./config/crd/standard/policy.networking.k8s.io_clusternetworkpolicies.yaml
}

install_kube_network_policies_from_main() {
  echo "Installing kube-network-policies from main"
  TEMP_DIR=$(mktemp -d)
  cleanup() { rm -rf "${TEMP_DIR}"; }
  trap cleanup EXIT INT TERM
  (
    cd "$TEMP_DIR"
    # Clone the repo
    git clone --depth 1 https://github.com/kubernetes-sigs/kube-network-policies.git
    cd kube-network-policies/
    # Build the image with the network policy API support
    REGISTRY="registry.k8s.io/networking" IMAGE_NAME="kube-network-policies" TAG="test" make image-build-npa-"${NPAPI_VERSION}"
    # Preload the image in the kind cluster
    /usr/local/bin/kind load docker-image registry.k8s.io/networking/kube-network-policies:test-npa-"${NPAPI_VERSION}" --name "${KIND_CLUSTER_NAME}"
    # Install kube-network-policies with the image built from main
    sed -i s#registry.k8s.io/networking/kube-network-policies.*#registry.k8s.io/networking/kube-network-policies:test-npa-${NPAPI_VERSION}# install-cnp.yaml
    /usr/local/bin/kubectl apply -f ./install-cnp.yaml
  )
}

get_cluster_status() {
  echo "Getting cluster status"
  # wait network is ready
  sleep 5
  /usr/local/bin/kubectl get nodes -o wide
  /usr/local/bin/kubectl get pods -A
  /usr/local/bin/kubectl wait --timeout=1m --for=condition=ready pods --namespace=kube-system -l k8s-app=kube-dns
  /usr/local/bin/kubectl wait --timeout=1m --for=condition=ready pods --namespace=kube-system -l app=kube-network-policies
}

run_tests() {
  echo "Running tests"
  go mod download
  REPO_VERSION=$(git describe --always --dirty)
  go test  -v ./conformance -run TestConformanceProfiles -args --conformance-profiles=ClusterNetworkPolicy --organization=kubernetes --project=kube-network-policies --url=https://github.com/kubernetes-sigs/kube-network-policies --version="${REPO_VERSION}" --contact=https://github.com/kubernetes-sigs/kube-network-policies/issues/new --additional-info=https://github.com/kubernetes-sigs/kube-network-policies
}

export_logs() {
  echo "Exporting logs"
  /usr/local/bin/kind export logs --name "${KIND_CLUSTER_NAME}" -v7 ./_artifacts/logs
}

# TODO: make this script execute multiple steps if specified as args in the
# order the steps are specified.
#
# run-conformance.sh setup_environment run_tests
#
# Will execute setup_environment, then run_tests.
#

usage() {
  echo "Usage: $0 [commands]"
  echo "Commands:"
  echo "  enable_ipv4_and_ipv6_forwarding"
  echo "  setup_environment"
  echo "  create_multinode_cluster"
  echo "  install_network_policy_apis"
  echo "  install_kube_network_policies_from_main"
  echo "  get_cluster_status"
  echo "  run_tests"
  echo "  export_logs"
  echo "  all"
}

all() {
  enable_ipv4_and_ipv6_forwarding
  setup_environment
  create_multinode_cluster
  install_network_policy_apis
  install_kube_network_policies_from_main
  get_cluster_status
  run_tests
  export_logs
}

if [ "$#" -eq 0 ]; then
  usage
  exit 1
fi

case "$1" in
  enable_ipv4_and_ipv6_forwarding)
    enable_ipv4_and_ipv6_forwarding
    ;;
  setup_environment)
    setup_environment
    ;;
  create_multinode_cluster)
    create_multinode_cluster
    ;;
  install_network_policy_apis)
    install_network_policy_apis
    ;;
  install_kube_network_policies_from_main)
    install_kube_network_policies_from_main
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
  all)
    all
    ;;
  *)
    usage
    exit 1
    ;;
esac
