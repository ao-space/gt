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

cat >/opt/aonetwork-client.yml <<EOF
version: 1.0
services:
  - local: http://${API_GATEWAY_HTTP_IP:-aospace-gateway}:${API_GATEWAY_HTTP_PORT:-8080}
options:
  id: ${SPACE_NAME_DOMAIN}
  secret: ${NETWORK_SECRET}
  remoteConnections: ${NETWORK_THREADS:-5}
  logLevel: ${NETWORK_LOGLEVEL:-info}
  reconnectDelay: ${NETWORK_RECONNECTDELAY:-15s}
  sentryDSN: ${NETWORK_SENTRYDSN}
  remoteAPI: ${NETWORK_REMOTEAPI}
  localTimeout: ${NETWORK_LOCALTIMEOUT:-120s}
  remoteTimeout: ${NETWORK_REMOTETIMEOUT:-70s}
  webrtcMaxPort: ${NETWORK_WEBRTCMAXPORT:-62000}
  webrtcMinPort: ${NETWORK_WEBRTCMINPORT:-61001}
EOF

exec /usr/bin/client -config /opt/aonetwork-client.yml
