#!/bin/bash

# Copyright 2022 The Kubernetes Authors.
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

readonly VERSION="v0.9.0"
OWNER=koalaman
REPO="shellcheck"
BINARY=shellcheck
FORMAT=tar.xz
#supported OS's linux and darwin
#supported cpu architectures: 
#linux: x86_64, aarch64, armv6hf
#darwin: x86_64 (only this architecture because the official shellcheck repo only delivers this binary)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
GITHUB_DOWNLOAD=https://github.com/${OWNER}/${REPO}/releases/download
NAME=${BINARY}-${VERSION}.${OS}.${ARCH}
TARBALL=${NAME}.${FORMAT}
TARBALL_URL=${GITHUB_DOWNLOAD}/${VERSION}/${TARBALL}
BINDIR=${BINDIR:-./bin}

echo "Installing shellcheck..."

http_download_curl() {
  local_file=$1
  source_url=$2
  code=$(curl -w '%{http_code}' -sL -o "$local_file" "$source_url")
  if [ "$code" != "200" ]; then
    log_debug "http_download_curl received HTTP status $code"
    return 1
  fi
  return 0
}

tmpdir=$(mktemp -d)
http_download_curl "${tmpdir}/${TARBALL}" "${TARBALL_URL}"
srcdir="${tmpdir}/${NAME}"
rm -rf "${srcdir}"
(cd "${tmpdir}" && tar --no-same-owner -xf "${TARBALL}")
binexe="./shellcheck-${VERSION}/shellcheck"
install "${tmpdir}/${binexe}" "${BINDIR}/"
rm -rf "${tmpdir}"

echo "Running shellcheck scan..."

#scanning all .sh scripts inside the hack directory
#ignoring SC2001 in shellcheck because complex sed substitution is required.
./bin/shellcheck -e SC2001 hack/*.sh
