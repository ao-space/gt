# GT

[English](README.md) | 简体中文

GT 是一个支持点对点直连（P2P）和互联网中转的反向代理开源项目。

具有以下设计特点：

- 注重隐私保护，在保证满足功能实现需要的情况下，最少化 server 端对数据包的分析，例如：基于 TCP 连接的实现方式，应用层 HTTP
  协议传输只分析第一个数据包的 HTTP 协议头的目标数据，不作任何多余分析，将后续数据直接转发。
- 注重性能，在代码实现上，倾向于采用性能更高的设计，例如：修改标准库来实现减少内存分配和复制的设计方案。
- 基于 WebRTC 实现的点对点连接功能，支持所有支持 WebRTC 的平台，例如：iOS，Android，浏览器等。

目前已经实现的主要功能有：

- 支持 HTTP(S)、WebSocket(S)、SSH、SMB 等基于 TCP 协议的通信协议转发
- 支持 WebRTC 点对点连接
- 多用户功能
  - 支持多种用户验证方式：API服务、本地配置
  - 每个用户独立配置
  - 限制用户速度
  - 限制客户端连接数
  - 验证失败达一定次数后，拒绝访问一段时间
- 服务端与客户端之间通信采用 TCP 连接池
- 保持命令行参数与 yaml 配置参数一致
- 支持日志上报到 Sentry 服务

## 目录

<!-- TOC -->

* [工作原理](#工作原理)
* [用法](#用法)
  * [配置文件](#配置文件)
  * [服务端配置 users](#服务端配置-users)
    * [通过命令行配置 users](#通过命令行配置-users)
    * [通过 users 配置文件配置 users](#通过-users-配置文件配置-users)
    * [通过 config 配置文件配置 users](#通过-config-配置文件配置-users)
    * [允许所有的客户端](#允许所有的客户端)
  * [服务端配置 TCP](#服务端配置-tcp)
    * [通过 users 配置文件配置 TCP](#通过-users-配置文件配置-tcp)
    * [通过 config 配置文件配置 TCP](#通过-config-配置文件配置-tcp)
  * [命令行参数](#命令行参数)
    * [HTTP 内网穿透](#http-内网穿透)
    * [HTTPS 内网穿透](#https-内网穿透)
    * [HTTPS SNI 内网穿透](#https-sni-内网穿透)
    * [TLS 加密客户端服务端之间的通信](#tls-加密客户端服务端之间的通信)
    * [TCP 内网穿透](#tcp-内网穿透)
    * [客户端同时开启多个服务](#客户端同时开启多个服务)
    * [服务端 API](#服务端-api)
* [性能测试](#性能测试)
  * [GT benchmark](#gt-benchmark)
  * [frp dev branch 42745a3](#frp-dev-branch-42745a3)
* [运行](#运行)
  * [Docker 容器运行](#docker-容器运行)
* [编译](#编译)
  * [在 Ubuntu/Debian 上编译](#在-ubuntudebian-上编译)
    * [安装依赖](#安装依赖)
    * [获取代码并编译](#获取代码并编译)
      * [从 ISCAS 镜像获取 WebRTC 并编译 GT](#从-iscas-镜像获取-webrtc-并编译-gt)
      * [从官方获取 WebRTC 并编译 GT](#从官方获取-webrtc-并编译-gt)
  * [在 Ubuntu/Debian 上通过 Docker 编译](#在-ubuntudebian-上通过-docker-编译)
    * [安装依赖](#安装依赖-1)
    * [获取代码并编译](#获取代码并编译-1)
      * [从 ISCAS 镜像获取 WebRTC 并编译 GT](#从-iscas-镜像获取-webrtc-并编译-gt-1)
      * [从官方获取 WebRTC 并编译 GT](#从官方获取-webrtc-并编译-gt-1)
* [演进计划](#演进计划)
* [贡献指南](#贡献指南)
  * [贡献代码](#贡献代码)
  * [代码质量](#代码质量)
  * [提交信息](#提交信息)
  * [问题报告](#问题报告)
  * [功能请求](#功能请求)
  * [感谢您的贡献](#感谢您的贡献)
  * [贡献者](#贡献者)

<!-- TOC -->

## 工作原理

```text
      ┌──────────────────────────────────────┐
      │  Web    Android     iOS    PC    ... │
      └──────────────────┬───────────────────┘
                  ┌──────┴──────┐
                  │  GT Server  │
                  └──────┬──────┘
       ┌─────────────────┼─────────────────┐
┌──────┴──────┐   ┌──────┴──────┐   ┌──────┴──────┐
│  GT Client  │   │  GT Client  │   │  GT Client  │ ...
└──────┬──────┘   └──────┬──────┘   └──────┬──────┘
┌──────┴──────┐   ┌──────┴──────┐   ┌──────┴──────┐
│     SSH     │   │   HTTP(S)   │   │     SMB     │ ...
└─────────────┘   └─────────────┘   └─────────────┘
```

## 用法

### 配置文件

配置文件使用 yaml 格式，客户端与服务端均可以使用配置文件。

```shell
./release/linux-amd64-server -config server.yaml
./release/linux-amd64-client -config client.yaml
```

基础服务端配置可以参考 [server.yaml](example/config/server.yaml) 文件。
基础客户端配置可以参考 [client.yaml](example/config/client.yaml) 文件。

### 服务端配置 users

以下四种方式可同时使用，如果冲突则按照从上到下优先级依次降低的方式解决。

#### 通过命令行配置 users

第 i 个 id 与第 i 个 secret 相匹配。下面两种启动方式是等价的。

```shell
./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -id id2 -secret secret2
```

```shell
./release/linux-amd64-server -addr 8080 -id id1 -id id2 -secret secret1 -secret secret2
```

#### 通过 users 配置文件配置 users

```yaml
id3:
  secret: secret3
id1:
  secret: secret1-overwrite
```

#### 通过 config 配置文件配置 users

```yaml
version: 1.0
users:
  id1:
    secret: secret1
  id2:
    secret: secret2
options:
  apiAddr: 1.2.3.4:1234
  certFile: /path
  keyFile: /path
  logFile: /path
  logFileMaxCount: 1234
  logFileMaxSize: 1234
  logLevel: debug
  addr: 1234
  timeout: 1234m1234ms
  tlsAddr: 1234
  tlsVersion: tls1.3
  users: testdata/users.yaml
```

#### 允许所有的客户端

在服务端的启动参数上添加 `-allowAnyClient`，所有的客户端无需在服务端配置即可连接服务端，但 `id`
相同的客户端只将第一个连接服务端的客户端的 `secret` 作为正确的 `secret`，不能被后续连接服务端的客户端的 `secret`
覆盖，保证安全性。

### 服务端配置 TCP

以下三种方式可同时使用。优先级：用户 > 全局。用户优先级：users 配置文件 > config 配置文件。全局优先级：命令行 > config
配置文件。如果没有配置 TCP 则表示不启用 TCP 功能。

#### 通过 users 配置文件配置 TCP

通过 users 配置文件可以配置单个用户的 TCP。下面的配置文件表示用户 id1 可以开启任意数量的任意 TCP 端口，用户 id2 没有开启
TCP 端口的权限。

```yaml
id1:
  secret: secret1
  tcp:
    - range: 1-65535
id2:
  secret: secret2
```

#### 通过 config 配置文件配置 TCP

通过 config 配置文件可以配置全局 TCP 和单个用户的 TCP。下面的配置文件表示用户 id1 可以开启任意数量在 10000 到 20000 的
TCP 端口，用户 id2 可以在 50000 到 65535 的 TCP 端口之间开启 1 个 TCP 端口。

```yaml
version: 1.0
users:
  id1:
    secret: secret1
    tcp:
      - range: 10000-20000
    tcpNumber: 0
  id2:
    secret: secret2
tcp:
  - range: 50000-65535
options:
  apiAddr: 1.2.3.4:1234
  certFile: /path
  keyFile: /path
  logFile: /path
  logFileMaxCount: 1234
  logFileMaxSize: 1234
  logLevel: debug
  addr: 1234
  timeout: 1234m1234ms
  tlsAddr: 1234
  tlsVersion: tls1.3
  users: testdata/users.yaml
  tcpNumber: 1
```

### 命令行参数

```shell
./release/linux-amd64-server -h
./release/linux-amd64-client -h
```

#### HTTP 内网穿透

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:8080
  来访问内网服务器上 80 端口服务的网页。

- 服务端（公网服务器）

```shell
./release/linux-amd64-server -addr 8080 -id id1 -secret secret1
```

- 客户端（内网服务器）

```shell
./release/linux-amd64-client -local http://127.0.0.1:80 -remote tcp://id1.example.com:8080 -id id1 -secret secret1
```

#### HTTPS 内网穿透

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 <https://id1.example.com>
  来访问内网服务器上 80 端口提供的 HTTP 网页。

- 服务端（公网服务器）

```shell
./release/linux-amd64-server -addr "" -tlsAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
```

- 客户端（内网服务器），因为使用了自签名证书，所以使用了 `-remoteCertInsecure` 选项，其它情况禁止使用此选项（中间人攻击导致加密内容被解密）

```shell
./release/linux-amd64-client -local http://127.0.0.1 -remote tls://id1.example.com -remoteCertInsecure -id id1 -secret secret1
```

#### HTTPS SNI 内网穿透

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 <https://id1.example.com>
  来访问内网服务器上 443 端口提供的 HTTPS 网页。

- 服务端（公网服务器）

```shell
./release/linux-amd64-server -addr 8080 -sniAddr 443 -id id1 -secret secret1
```

- 客户端（内网服务器）

```shell
./release/linux-amd64-client -local https://127.0.0.1 -remote tcp://id1.example.com:8080 -id id1 -secret secret1
```

#### TLS 加密客户端服务端之间的通信

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:8080
  来访问内网服务器上 80 端口服务的网页。同时用 TLS
  加密客户端与服务端之间的通信。

- 服务端（公网服务器）

```shell
./release/linux-amd64-server -addr 8080 -tlsAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
```

- 客户端（内网服务器），因为使用了自签名证书，所以使用了 `-remoteCertInsecure` 选项，其它情况禁止使用此选项（中间人攻击导致加密内容被解密）

```shell
./release/linux-amd64-client -local http://127.0.0.1:80 -remote tls://id1.example.com -remoteCertInsecure -id id1 -secret secret1
```

#### TCP 内网穿透

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:2222
  来访问内网服务器上 22 端口上的 SSH 服务，如果服务端 2222 端口不可以，则由服务端选择一个随机端口。

- 服务端（公网服务器）

```shell
./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 1 -tcpRange 1024-65535
```

- 客户端（内网服务器）

```shell
./release/linux-amd64-client -local tcp://127.0.0.1:22 -remote tcp://id1.example.com:8080 -id id1 -secret secret1 -remoteTCPPort 2222 -remoteTCPRandom
```

#### QUIC 内网穿透

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:8080
  来访问内网服务器上 80 端口服务的网页。使用 QUIC 为客户端与服务端之间构建传输连接，QUIC 使用 TLS 1.3 进行传输加密。当用户同时给出certFile
  和keyFile时，使用他们进行加密通信。否则，会使用 ECDSA 加密算法自动生成密钥和证书。

- 服务端（公网服务器）

```shell
./release/linux-amd64-server -addr 8080 -quicAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
```

- 客户端（内网服务器），因为使用了自签名证书，所以使用了 `-remoteCertInsecure` 选项，其它情况禁止使用此选项（中间人攻击导致加密内容被解密）。

```shell
./release/linux-amd64-client -local http://127.0.0.1:80 -remote quic://id1.example.com:443 -remoteCertInsecure -id id1 -secret secret1
```

#### 智能内网穿透（自适应选择 TCP/QUIC ）

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:8080
  来访问内网服务器上 80 端口服务的网页。GT根据内网服务器和公网服务器之间网络的时延和丢包率，自适应选择使用 TCP+TLS 还是 QUIC 进行内网穿透。

- 服务端（公网服务器）

```shell
./release/linux-amd64-server -addr 8080 -autoAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
```

- 客户端（内网服务器）。如果选择了QUIC则需要使用自签名证书，所以使用了 `-remoteCertInsecure` 选项。

```shell
./release/linux-amd64-client -local http://127.0.0.1:80 -remote auto://id1.example.com:443 -remoteCertInsecure -id id1 -secret secret1
```

#### 客户端同时开启多个服务

- 需求：有一台内网服务器和一台公网服务器，id1-1.example.com 和 id1-2.example.com 解析到公网服务器的地址。希望通过访问
  id1-1.example.com:8080 来访问内网服务器上 80 端口上的服务，希望通过访问 id1-2.example.com:8080 来访问内网服务器上 8080
  端口上的服务，希望通过访问 id1-1.example.com:2222 来访问内网服务器上 2222 端口上的服务，希望通过访问 id1-1.example.com:
  2223 来访问内网服务器上 2223 端口上的服务。同时服务端限制客户端的 hostPrefix 只能由纯数字或纯字母组成。

- 注意：在这种模式下客户端 local 对应的参数（remoteTCPPort，hostPrefix 等）位置要在此 local 和下一个 local 之间。

- 服务端（公网服务器）

```shell
./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 2 -tcpRange 1024-65535 -hostNumber 2 -hostWithID -hostRegex ^[0-9]+$ -hostRegex ^[a-zA-Z]+$
```

- 客户端（内网服务器）

```shell
./release/linux-amd64-client -remote tcp://id1.example.com:8080 -id id1 -secret secret1 \
>     -local http://127.0.0.1:80 -useLocalAsHTTPHost -hostPrefix 1 \
>     -local http://127.0.0.1:8080 -useLocalAsHTTPHost -hostPrefix 2 \
>     -local tcp://127.0.0.1:2222 -remoteTCPPort 2222 \
>     -local tcp://127.0.0.1:2223 -remoteTCPPort 2223
```

上面的命令行也可以使用配置文件来启动

```shell
./release/linux-amd64-client -config client.yaml
```

client.yaml 文件内容：

```yaml
services:
  - local: http://127.0.0.1:80
    useLocalAsHTTPHost: true
    hostPrefix: 1
  - local: http://127.0.0.1:8080
    useLocalAsHTTPHost: true
    hostPrefix: 2
  - local: tcp://127.0.0.1:2222
    remoteTCPPort: 2222
  - local: tcp://127.0.0.1:2223
    remoteTCPPort: 2223
options:
  remote: tcp://id1.example.com:8080
  id: id1
  secret: secret1
```

#### 服务端 API

服务端 API 通过模拟客户端检测服务是否正常。下面的例子可以帮助你更好地理解这一点，其中，id1.example.com 解析到公网服务器的地址。当
apiCertFile 和 apiKeyFile 选项不为空时使用 HTTPS，其他情况使用 HTTP。

- 服务端（公网服务器）

```shell
./release/linux-amd64-server -addr 8080 -apiAddr 8081
```

- 用户

```shell
# curl http://id1.example.com:8081/status
{"status": "ok", "version":"linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f"}
```

## 性能测试

通过 wrk 进行压力测试本项目与 frp 进行对比，内网服务指向在本地运行 nginx 的测试页面，测试结果如下：

```text
Model Name: MacBook Pro
Model Identifier: MacBookPro17,1
Chip: Apple M1
Total Number of Cores: 8 (4 performance and 4 efficiency)
Memory: 16 GB
```

### GT benchmark

```shell
$ wrk -c 100 -d 30s -t 10 http://pi.example.com:7001
Running 30s test @ http://pi.example.com:7001
  10 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     2.22ms  710.73us  37.99ms   98.30%
    Req/Sec     4.60k   231.54     4.86k    91.47%
  1374783 requests in 30.01s, 1.09GB read
Requests/sec:  45811.08
Transfer/sec:     37.14MB

$ ps aux
  PID  %CPU %MEM      VSZ    RSS   TT  STAT STARTED      TIME COMMAND
 2768   0.0  0.1 408697792  17856 s008  S+    4:55PM   0:52.34 ./client -local http://localhost:8080 -remote tcp://localhost:7001 -id pi -threads 3
 2767   0.0  0.1 408703664  17584 s007  S+    4:55PM   0:52.16 ./server -port 7001
```

### frp dev branch 42745a3

```shell
$ wrk -c 100 -d 30s -t 10 http://pi.example.com:7000
Running 30s test @ http://pi.example.com:7000
  10 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    76.92ms   73.46ms 748.61ms   74.21%
    Req/Sec   154.63    308.28     2.02k    93.75%
  45487 requests in 30.10s, 31.65MB read
  Non-2xx or 3xx responses: 20610
Requests/sec:   1511.10
Transfer/sec:      1.05MB

$ ps aux
  PID  %CPU %MEM      VSZ    RSS   TT  STAT STARTED      TIME COMMAND
 2975   0.3  0.5 408767328  88768 s004  S+    5:01PM   0:21.88 ./frps -c ./frps.ini
 2976   0.0  0.4 408712832  66112 s005  S+    5:01PM   1:06.51 ./frpc -c ./frpc.ini
```

## 运行

### Docker 容器运行

更多容器镜像信息可以从<https://github.com/ao-space/gt/pkgs/container/gt>获取。

```shell
docker pull ghcr.io/ao-space/gt:server-dev

docker pull ghcr.io/ao-space/gt:client-dev
```

## 编译

### 在 Ubuntu/Debian 上编译

#### 安装依赖

```shell
apt-get update
apt-get install make git gn ninja-build python3 python3-pip libgtk-3-dev gcc-aarch64-linux-gnu g++-aarch64-linux-gnu gcc-x86-64-linux-gnu g++-x86-64-linux-gnu -y
```

#### 获取代码并编译

你可以选择从镜像或者官方获取 WebRTC 并编译 GT：

##### 从 ISCAS 镜像获取 WebRTC 并编译 GT

1. 获取代码

      ```shell
      git clone <url>
      cd <folder>
      ```

2. 编译

      ```shell
      make release
      ```

   编译后的可执行文件在 release 目录下。

##### 从官方获取 WebRTC 并编译 GT

1. 获取代码

      ```shell
      git clone <url>
      cd <folder>
      ```

2. 从官方获取 WebRTC

      ```shell
      mkdir -p dep/_google-webrtc
      cd dep/_google-webrtc
      git clone https://webrtc.googlesource.com/src
      ```

   然后按照[这个链接中的步骤](https://webrtc.googlesource.com/src/+/main/docs/native-code/development/)检出构建工具链和许多依赖项。

3. 编译

      ```shell
      WITH_OFFICIAL_WEBRTC=1 make release
      ```

   编译后的可执行文件在 release 目录下。

### 在 Ubuntu/Debian 上通过 Docker 编译

#### 安装依赖

[安装 Docker](https://docs.docker.com/engine/install/)

#### 获取代码并编译

你可以选择从镜像或者官方获取 WebRTC 并编译 GT：

##### 从 ISCAS 镜像获取 WebRTC 并编译 GT

1. 获取代码

      ```shell
      git clone <url>
      cd <folder>
      ```

2. 编译

      ```shell
      make docker_release_linux_amd64 # docker_release_linux_arm64
      ```

   编译后的可执行文件在 release 目录下。

##### 从官方获取 WebRTC 并编译 GT

1. 获取代码

      ```shell
      git clone <url>
      cd <folder>
      ```

2. 从官方获取 WebRTC

      ```shell
      mkdir -p dep/_google-webrtc
      cd dep/_google-webrtc
      git clone https://webrtc.googlesource.com/src
      ```

   然后按照[这个链接中的步骤](https://webrtc.googlesource.com/src/+/main/docs/native-code/development/)检出构建工具链和许多依赖项。

3. 编译

      ```shell
      WITH_OFFICIAL_WEBRTC=1 make docker_release_linux_amd64 # docker_release_linux_arm64
      ```

   编译后的可执行文件在 release 目录下。

## 演进计划

- 添加网页管理功能
- 支持使用QUIC协议，BBR拥塞算法
- 支持配置P2P连接转发数据到多个服务
- 认证功能支持公钥和私钥

## 贡献指南

我们非常欢迎对本项目进行贡献。以下是一些指导原则和建议，希望能够帮助您参与到项目中来。

### 贡献代码

如果您想要为项目做出贡献，最好的方式就是提交代码。在提交代码之前，请确保您已经下载并熟悉了项目代码库，并且您的代码遵循了以下指导原则：

- 代码应当尽量简洁明了，并且易于维护和扩展。
- 代码应遵循项目约定的命名规范，以确保代码的一致性。
- 代码应当遵循项目的代码风格指南，可以参考项目代码库中已有的代码。

如果您想向项目提交代码，可以按照以下步骤进行：

- 在 GitHub 上 fork 该项目。
- 克隆您 fork 的项目到本地。
- 在本地进行您的修改和改进。
- 执行测试确保任何更改都无影响。
- 提交您的更改并新建一个 pull request。

### 代码质量

我们非常注重代码的质量，因此您提交的代码应当符合以下要求：

- 代码应当经过充分的测试，确保其正确性和稳定性。
- 代码应当遵循良好的设计原则和最佳实践。
- 代码应当尽可能地符合您所提交代码贡献的相关要求。

### 提交信息

在提交代码之前，请确保您提供了有意义而且详细的提交信息。这有助于我们更好地理解您的代码贡献并且更快速地合并它。

提交信息应当包含以下内容：

- 描述本次代码贡献的目的或者原因。
- 描述本次代码贡献的内容或者变化。
- （可选）描述本次代码贡献的测试方法或者结果。

提交信息应当清晰明了，并且符合项目代码库的提交信息约定。

### 问题报告

如果您在项目中遇到了问题，或者发现了错误，欢迎向我们提交问题报告。在提交问题报告之前，请确保您已经对问题进行了充分的调查和试验，并且尽量提供以下信息：

- 描述问题的现象和表现。
- 描述问题出现的场景和条件。
- 描述上下文信息或任何相关的背景信息。
- 描述您期望的行为信息。
- （可选）提供相关的截图或者报错信息。

问题报告应当清晰明了，并且符合项目代码库的问题报告约定。

### 功能请求

如果您想要向项目中添加新的功能或者特性，欢迎向我们提交功能请求。在提交功能请求之前，请确保您已经了解了项目的历史和现状，并且尽量提供以下信息：

- 描述您想要添加的功能或者特性。
- 描述这个功能或者特性的用途和目的。
- （可选）提供相关的实现思路或者建议。

功能请求应当清晰明了，并且符合项目代码库的功能请求约定。

### 感谢您的贡献

最后，感谢您对本项目的贡献。我们欢迎各种形式的贡献，包括但不限于代码贡献、问题报告、功能请求、文档编写等。我们相信在您的帮助下，本项目会变得更加完善和强大。

### 贡献者

感谢以下人员为项目做出的贡献：

- [zhiyi](https://github.com/vyloy)
- [jianti](https://github.com/FH0)
