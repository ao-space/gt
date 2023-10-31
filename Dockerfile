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

FROM golang:1.20-bullseye

# apt 切换国内源，安装 google webrtc 依赖安装脚本的依赖
# mirrors.163.com
# repo.huaweicloud.com
RUN sed -i 's|deb.debian.org|mirrors.163.com|g' /etc/apt/sources.list && \
    apt update && \
    apt install xz-utils bzip2 sudo lsb-release ninja-build generate-ninja file patch -y

# 安装nodejs 20
RUN apt-get install -y ca-certificates curl gnupg && \
    mkdir -p /etc/apt/keyrings && \
    curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg && \
    NODE_MAJOR=20 && \
    echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_$NODE_MAJOR.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list && \
    apt-get update && \
    apt-get install nodejs -y

# 安装msquic依赖
RUN apt-get update && \
    apt-get install -y cmake build-essential liblttng-ust-dev lttng-tools libssl-dev && \
    wget https://cmake.org/files/v3.23/cmake-3.23.0.tar.gz && \
    tar -zxvf cmake-3.23.0.tar.gz && cd cmake-3.23.0 && ./configure && make -j8 && make install && \
    cmake --version
# RUN apt-get install -y cmake build-essential liblttng-ust-dev lttng-tools

# golang 切换国内源并且提前安装好依赖
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
ADD ./go.mod /go/src/github.com/isrc-cas/gt/
RUN cd /go/src/github.com/isrc-cas/gt && \
    go mod download && \
    rm -rf /go/src/github.com/isrc-cas/gt

# 安装 gnu 编译链
RUN apt install gcc-aarch64-linux-gnu g++-aarch64-linux-gnu gcc-x86-64-linux-gnu g++-x86-64-linux-gnu -y

# 安装 google webrtc 依赖
ADD ./dep/_google-webrtc/src/build /root/build
RUN	sed -i 's/egrep -q "i686|x86_64"/true/g' /root/build/install-build-deps.sh && \
    DEBIAN_FRONTEND=noninteractive TZ=Asia/Shanghai bash /root/build/install-build-deps.sh --no-chromeos-fonts && \
    rm -rf /root/build

RUN if uname -m | grep aarch64; then \
    file /usr/x86_64-linux-gnu/lib/libm.a | grep "ASCII text" && \
    rm -f /usr/x86_64-linux-gnu/lib/libm.a && \
    ln -s /usr/x86_64-linux-gnu/lib/libm-2.31.a /usr/x86_64-linux-gnu/lib/libm.a && echo "success"; \
    fi