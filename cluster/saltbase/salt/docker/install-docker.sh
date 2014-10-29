#!/bin/bash

# Copyright 2014 Google Inc. All rights reserved.
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

# This is gleaned from:
#   https://get.docker.com/ubuntu/dists/docker/main/binary-amd64/Packages
# ... which sent me to:
#   https://get.docker.com/ubuntu/pool/main/l/lxc-docker-1.3.0/${DEB}
SERVER=http://storage.googleapis.com
BUCKET=test-docker
DEB=lxc-docker-1.3.0_1.3.0-20141016165047-c78088f_amd64.deb
DOCS="/usr/share/doc/docker"

cd /tmp

# Get the license file.
curl -O "${SERVER}/${BUCKET}/apache2.txt"
mkdir -p "${DOCS}"
cp -a apache2.txt "${DOCS}"

# Get docker.
curl -O "${SERVER}/${BUCKET}/${DEB}"
dpkg -i "${DEB}"
apt-get -y -f install
