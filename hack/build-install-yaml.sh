#!/usr/bin/env bash

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
shopt -s nullglob
thisyear=$(date +"%Y")

mkdir -p release/

# Make clean files with boilerplate
cat hack/boilerplate.sh.txt > release/experimental-install.yaml
sed -i "s/YEAR/$thisyear/g" release/experimental-install.yaml
cat << EOF >> release/experimental-install.yaml
#
# NetworkPolicy API Experimental channel install
#
EOF

cat hack/boilerplate.sh.txt > release/standard-install.yaml
sed -i "s/YEAR/$thisyear/g" release/standard-install.yaml
cat << EOF >> release/standard-install.yaml
#
# NetworkPolicy API Standard channel install
#
EOF

for file in config/crd/experimental/policy*.yaml
do
    [[ -e "$file" ]] || break;
    {
      echo "---"
      echo "#"
      echo "# $file"
      echo "#"
      cat "$file"
    } >> release/experimental-install.yaml
done

for file in config/crd/standard/policy*.yaml
do
    [[ -e "$file" ]] || break;
    {
      echo "---"
      echo "#"
      echo "# $file"
      echo "#"
      cat "$file"
    } >> release/standard-install.yaml
done


echo "Generated:" install yamls
