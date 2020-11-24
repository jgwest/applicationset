#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
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

GOPATH=$( if test -x `which go`; then go env GOPATH; else echo "$(HOME)/go"; fi)

echo $GOPATH

set -x
set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
. ${SCRIPT_ROOT}/hack/versions.sh
CODEGEN_PKG=$GOPATH/pkg/mod/k8s.io/code-generator@${kube_version}
TARGET_SCRIPT=/tmp/generate-groups.sh

(
  cd $CODEGEN_PKG
  go install ./cmd/{defaulter-gen,client-gen,lister-gen,informer-gen,deepcopy-gen}
)

#export GO111MODULE=off

# sed -e '/go install/d' ${CODEGEN_PKG}/generate-groups.sh > ${TARGET_SCRIPT}

bash -x ${TARGET_SCRIPT} "deepcopy,client,informer,lister" \
  jgw-generated github.com/argoproj-labs/applicationset/pkg/apis \
  "applicationset:v1alpha1" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt

# github.com/argoproj-labs/applicationset/pkg/client github.com/argoproj-labs/applicationset/pkg/apis \
#  github.com/argoproj-labs/applicationset/api/v1alpha1 \
