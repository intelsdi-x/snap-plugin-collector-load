#!/bin/bash
# File managed by pluginsync

# http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
# Copyright 2016 Intel Corporation
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

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

UNAME=$(uname)

plugin_name=${__proj_dir##*/}
build_dir="${__proj_dir}/build"
if [ ${UNAME} == "FreeBSD" ]; then
  go_build=(go build -ldflags "-w" --ldflags '-linkmode external -extldflags "-static"')
else
  go_build=(go build -ldflags "-w")
fi

_info "project path: ${__proj_dir}"
_info "plugin name: ${plugin_name}"

if [ ${UNAME} != "FreeBSD" ]; then
export CGO_ENABLED=0
fi

# rebuild binaries:
_debug "removing: ${build_dir:?}/*"
rm -rf "${build_dir:?}/"*

_info "building plugin: ${plugin_name}"
if [ ${UNAME} == "FreeBSD" ]; then
  export GOOS=freebsd
  export CC=clang
else
  export GOOS=linux
fi
export GOARCH=amd64
mkdir -p "${build_dir}/${GOOS}/x86_64"
"${go_build[@]}" -o "${build_dir}/${GOOS}/x86_64/${plugin_name}" . || exit 1
