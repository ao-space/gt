#!/bin/sh
# Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
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

cat > /opt/aonetwork-server.yml <<EOF
version: 1.0
options:
  addr: ${NETWORK_ADDR:-80}
  tlsAddr: ${NETWORK_TLSADDR:-443}
  certFile: /opt/crt/tls.crt
  keyFile: /opt/crt/tls.key
  logLevel: ${NETWORK_LOGLEVEL:-info}
  apiAddr: ${NETWORK_API_ADDR:-0.0.0.0:81}
  sentryDSN: ${NETWORK_SENTRYDSN}
  authAPI: ${NETWORK_AUTHAPI}
  timeout: ${NETWORK_TIMEOUT:-90s}
  httpMUXHeader: EID
  stunAddr: ${NETWORK_STUNADDR:-3478}
EOF

exec /usr/bin/server -config /opt/aonetwork-server.yml
