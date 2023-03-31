# GT

专注于高性能，低延迟的内网穿透解决方案。

- 支持 HTTP(S)、WebSocket(S)、SSH、SMB 等基于 TCP 协议的通信协议转发
- 支持 WebRTC 点对点连接
- 支持日志上报到 Sentry 服务
- 多用户功能
  - 支持多种用户验证方式：API服务、本地配置
  - 每个用户独立配置
  - 限制用户速度
  - 限制客户端连接数
  - 验证失败达一定次数后，拒绝访问一段时间
- 服务端与客户端之间通信采用 TCP 连接池
- 保持命令行参数与 yaml 配置参数一致

## 目录

- [工作原理](#%E5%B7%A5%E4%BD%9C%E5%8E%9F%E7%90%86)
- [示例](#%E7%A4%BA%E4%BE%8B)
  - [HTTP 内网穿透](#http-%E5%86%85%E7%BD%91%E7%A9%BF%E9%80%8F)
  - [HTTPS 解密成 HTTP 后内网穿透](#https-%E8%A7%A3%E5%AF%86%E6%88%90-http-%E5%90%8E%E5%86%85%E7%BD%91%E7%A9%BF%E9%80%8F)
  - [HTTPS 直接内网穿透](#https-%E7%9B%B4%E6%8E%A5%E5%86%85%E7%BD%91%E7%A9%BF%E9%80%8F)
  - [TLS 加密客户端服务端之间的 HTTP 通信](#tls-%E5%8A%A0%E5%AF%86%E5%AE%A2%E6%88%B7%E7%AB%AF%E6%9C%8D%E5%8A%A1%E7%AB%AF%E4%B9%8B%E9%97%B4%E7%9A%84-http-%E9%80%9A%E4%BF%A1)
  - [TCP 内网穿透](#tcp-%E5%86%85%E7%BD%91%E7%A9%BF%E9%80%8F)
  - [客户端同时开启多个服务](#%E5%AE%A2%E6%88%B7%E7%AB%AF%E5%90%8C%E6%97%B6%E5%BC%80%E5%90%AF%E5%A4%9A%E4%B8%AA%E6%9C%8D%E5%8A%A1)
- [参数](#%E5%8F%82%E6%95%B0)
  - [客户端参数](#%E5%AE%A2%E6%88%B7%E7%AB%AF%E5%8F%82%E6%95%B0)
  - [服务端参数](#%E6%9C%8D%E5%8A%A1%E7%AB%AF%E5%8F%82%E6%95%B0)
  - [配置文件](#%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6)
  - [服务端配置 users](#%E6%9C%8D%E5%8A%A1%E7%AB%AF%E9%85%8D%E7%BD%AE-users)
    - [通过命令行配置 users](#%E9%80%9A%E8%BF%87%E5%91%BD%E4%BB%A4%E8%A1%8C%E9%85%8D%E7%BD%AE-users)
    - [通过 users 配置文件配置 users](#%E9%80%9A%E8%BF%87-users-%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E9%85%8D%E7%BD%AE-users)
    - [通过 config 配置文件配置 users](#%E9%80%9A%E8%BF%87-config-%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E9%85%8D%E7%BD%AE-users)
    - [允许所有的客户端](#%E5%85%81%E8%AE%B8%E6%89%80%E6%9C%89%E7%9A%84%E5%AE%A2%E6%88%B7%E7%AB%AF)
  - [服务端配置 TCP](#%E6%9C%8D%E5%8A%A1%E7%AB%AF%E9%85%8D%E7%BD%AE-tcp)
    - [通过 users 配置文件配置 TCP](#%E9%80%9A%E8%BF%87-users-%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E9%85%8D%E7%BD%AE-tcp)
    - [通过 config 配置文件配置 TCP](#%E9%80%9A%E8%BF%87-config-%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E9%85%8D%E7%BD%AE-tcp)
    - [通过命令行配置 TCP](#%E9%80%9A%E8%BF%87%E5%91%BD%E4%BB%A4%E8%A1%8C%E9%85%8D%E7%BD%AE-tcp)
  - [服务端 API](#%E6%9C%8D%E5%8A%A1%E7%AB%AF-api)
- [性能测试](#%E6%80%A7%E8%83%BD%E6%B5%8B%E8%AF%95)
  - [AONetwork benchmark](#aonetwork-benchmark)
  - [frp dev branch 42745a3](#frp-dev-branch-42745a3)
- [编译](#%E7%BC%96%E8%AF%91)

## 工作原理

```
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

## 示例

### HTTP 内网穿透

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:8080
  来访问内网服务器上 80 端口服务的网页。

- 服务端（公网服务器）

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -secret secret1
Sat Nov 19 20:16:33 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:16:33 CST 2022 INF Listening addr=:8080
Sat Nov 19 20:16:33 CST 2022 INF acceptLoop started addr=[::]:8080
```

- 客户端（内网服务器）

```shell
# ./release/linux-amd64-client -local http://127.0.0.1:80 -remote tcp://id1.example.com:8080 -id id1 -secret secret1
Sat Nov 19 20:18:59 CST 2022 INF linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e config={"Config":"","ID":"id1","Local":"http://127.0.0.1:80","LocalTimeout":120000000000,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tcp://id1.example.com:8080","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":false,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":0,"RemoteTCPRandom":false,"RemoteTimeout":5000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","UseLocalAsHTTPHost":false,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Sat Nov 19 20:18:59 CST 2022 INF remote url remote=tcp://id1.example.com:8080 stun=
Sat Nov 19 20:18:59 CST 2022 INF trying to connect to remote connID=1
Sat Nov 19 20:18:59 CST 2022 INF tunnel started connID=1
```

### HTTPS 解密成 HTTP 后内网穿透

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 https://id1.example.com
  来访问内网服务器上 80 端口提供的 HTTP 网页。

- 服务端（公网服务器）

```shell
# ./release/linux-amd64-server -addr "" -tlsAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
Sat Nov 19 20:19:53 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"","AllowAnyClient":false,"AuthAPI":"","CertFile":"/root/openssl_crt/tls.crt","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"/root/openssl_crt/tls.key","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"443","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:19:53 CST 2022 INF Listening TLS addr=:443
Sat Nov 19 20:19:53 CST 2022 INF acceptLoop started addr=[::]:443
```

- 客户端（内网服务器），因为使用了自签名证书，所以使用了 `-remoteCertInsecure` 选项，其它情况禁止使用此选项（中间人攻击导致加密内容被解密）

```shell
# ./release/linux-amd64-client -local http://127.0.0.1 -remote tls://id1.example.com -remoteCertInsecure -id id1 -secret secret1
Sat Nov 19 20:20:05 CST 2022 INF linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e config={"Config":"","ID":"id1","Local":"http://127.0.0.1","LocalTimeout":120000000000,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tls://id1.example.com","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":true,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":0,"RemoteTCPRandom":false,"RemoteTimeout":5000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","UseLocalAsHTTPHost":false,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Sat Nov 19 20:20:05 CST 2022 INF remote url remote=tls://id1.example.com stun=
Sat Nov 19 20:20:05 CST 2022 INF trying to connect to remote connID=1
Sat Nov 19 20:20:06 CST 2022 INF tunnel started connID=1
```

### HTTPS 直接内网穿透

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 https://id1.example.com
  来访问内网服务器上 443 端口提供的 HTTPS 网页。

- 服务端（公网服务器）

```shell
# ./release/linux-amd64-server -addr "" -sniAddr 443 -id id1 -secret secret1
Sat Nov 19 20:25:15 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"443","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:25:15 CST 2022 INF Listening sniAddr=:443
Sat Nov 19 20:25:15 CST 2022 INF acceptLoop started addr=[::]:443
```

- 客户端（内网服务器）

```shell
# ./release/linux-amd64-client -local https://127.0.0.1 -remote tcp://id1.example.com:443 -id id1 -secret secret1
Sat Nov 19 20:25:49 CST 2022 INF linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e config={"Config":"","ID":"id1","Local":"https://127.0.0.1","LocalTimeout":120000000000,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tcp://id1.example.com:443","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":false,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":0,"RemoteTCPRandom":false,"RemoteTimeout":5000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","UseLocalAsHTTPHost":false,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Sat Nov 19 20:25:49 CST 2022 INF remote url remote=tcp://id1.example.com:443 stun=
Sat Nov 19 20:25:49 CST 2022 INF trying to connect to remote connID=1
Sat Nov 19 20:25:49 CST 2022 INF tunnel started connID=1
```

### TLS 加密客户端服务端之间的 HTTP 通信

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:8080
  来访问内网服务器上 80 端口服务的网页。同时用 TLS
  加密客户端与服务端之间的通信。

- 服务端（公网服务器）

```shell
# ./release/linux-amd64-server -addr 8080 -tlsAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
Sat Nov 19 20:20:59 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"/root/openssl_crt/tls.crt","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"/root/openssl_crt/tls.key","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"443","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:20:59 CST 2022 INF Listening TLS addr=:443
Sat Nov 19 20:20:59 CST 2022 INF Listening addr=:8080
Sat Nov 19 20:20:59 CST 2022 INF acceptLoop started addr=[::]:8080
Sat Nov 19 20:20:59 CST 2022 INF acceptLoop started addr=[::]:443
```

- 客户端（内网服务器），因为使用了自签名证书，所以使用了 `-remoteCertInsecure` 选项，其它情况禁止使用此选项（中间人攻击导致加密内容被解密）

```shell
# ./release/linux-amd64-client -local http://127.0.0.1:80 -remote tls://id1.example.com -remoteCertInsecure -id id1 -secret secret1
Sat Nov 19 20:26:33 CST 2022 INF linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e config={"Config":"","ID":"id1","Local":"http://127.0.0.1:80","LocalTimeout":120000000000,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tls://id1.example.com","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":true,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":0,"RemoteTCPRandom":false,"RemoteTimeout":5000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","UseLocalAsHTTPHost":false,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Sat Nov 19 20:26:33 CST 2022 INF remote url remote=tls://id1.example.com stun=
Sat Nov 19 20:26:33 CST 2022 INF trying to connect to remote connID=1
Sat Nov 19 20:26:33 CST 2022 INF tunnel started connID=1
```

### TCP 内网穿透

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:2222
  来访问内网服务器上 22 端口上的 SSH 服务，如果服务端 2222 端口不可以，则由服务端选择一个随机端口。

- 服务端（公网服务器）

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 1 -tcpRange 1024-65535
Sat Nov 19 20:21:34 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":["1"],"TCPRanges":["1024-65535"],"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:21:34 CST 2022 INF Listening addr=:8080
Sat Nov 19 20:21:34 CST 2022 INF acceptLoop started addr=[::]:8080
```

- 客户端（内网服务器）

```shell
# ./release/linux-amd64-client -local tcp://127.0.0.1:22 -remote tcp://id1.example.com:8080 -id id1 -secret secret1 -remoteTCPPort 2222 -remoteTCPRandom
Sat Nov 19 20:21:53 CST 2022 INF linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e config={"Config":"","ID":"id1","Local":"tcp://127.0.0.1:22","LocalTimeout":120000000000,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tcp://id1.example.com:8080","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":false,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":2222,"RemoteTCPRandom":true,"RemoteTimeout":5000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","UseLocalAsHTTPHost":false,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Sat Nov 19 20:21:53 CST 2022 INF remote url remote=tcp://id1.example.com:8080 stun=
Sat Nov 19 20:21:53 CST 2022 INF trying to connect to remote connID=1
Sat Nov 19 20:21:53 CST 2022 INF receive server information: tcp port 2222 opened successfully connID=1
Sat Nov 19 20:21:53 CST 2022 INF tunnel started connID=1
```

### TCP 内网穿透

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:2222
  来访问内网服务器上 22 端口上的 SSH 服务，如果服务端 2222 端口不可以，则由服务端选择一个随机端口。

- 服务端（公网服务器）

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 1 -tcpRange 1024-65535
Fri Dec  9 18:38:21 CST 2022 INF linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":10,"HTTPMUXHeader":"Host","Host":{"Number":null,"Regex":null,"RegexStr":null,"WithID":null},"HostNumber":1,"HostRegex":null,"HostWithID":false,"IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDuration":300000000000,"ReconnectTimes":3,"SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":["1"],"TCPRanges":["1024-65535"],"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Fri Dec  9 18:38:21 CST 2022 INF Listening addr=:8080
Fri Dec  9 18:38:21 CST 2022 INF acceptLoop started addr=[::]:8080
```

- 客户端（内网服务器）

```shell
# ./release/linux-amd64-client -local tcp://127.0.0.1:22 -remote tcp://id1.example.com:8080 -id id1 -secret secret1 -remoteTCPPort 2222 -remoteTCPRandom
Fri Dec  9 18:39:05 CST 2022 INF linux-amd64-client - 2022-12-09 05:20:39 - dev 88d322f config={"Config":"","HostPrefix":null,"ID":"id1","Local":[{"Position":0,"Value":"tcp://127.0.0.1:22"}],"LocalTimeout":null,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tcp://id1.example.com:8080","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":false,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":[{"Position":1,"Value":2222}],"RemoteTCPRandom":[{"Position":2,"Value":true}],"RemoteTimeout":45000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-12-09 05:20:39 - dev 88d322f","SentrySampleRate":1,"SentryServerName":"","Services":null,"UseLocalAsHTTPHost":null,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Fri Dec  9 18:39:05 CST 2022 INF remote url remote=tcp://id1.example.com:8080 stun=
Fri Dec  9 18:39:05 CST 2022 INF trying to connect to remote connID=1
Fri Dec  9 18:39:05 CST 2022 INF receive server information: tcp port 2222 opened successfully connID=1
Fri Dec  9 18:39:05 CST 2022 INF tunnel started connID=1
```

### 客户端同时开启多个服务

- 需求：有一台内网服务器和一台公网服务器，id1-1.example.com 和 id1-2.example.com 解析到公网服务器的地址。希望通过访问
  id1-1.example.com:8080 来访问内网服务器上 80 端口上的服务，希望通过访问 id1-2.example.com:8080 来访问内网服务器上 8080
  端口上的服务，希望通过访问 id1-1.example.com:2222 来访问内网服务器上 2222 端口上的服务，希望通过访问 id1-1.example.com:
  2223 来访问内网服务器上 2223 端口上的服务。同时服务端限制客户端的 hostPrefix 只能由纯数字或纯字母组成。

- 注意：在这种模式下客户端 local 对应的参数（remoteTCPPort，hostPrefix 等）位置要在此 local 和下一个 local 之间。

- 服务端（公网服务器）

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 2 -tcpRange 1024-65535 -hostNumber 2 -hostWithID -hostRegex ^[0-9]+$ -hostRegex ^[a-zA-Z]+$
Fri Dec  9 18:39:22 CST 2022 INF linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":10,"HTTPMUXHeader":"Host","Host":{"Number":null,"Regex":null,"RegexStr":null,"WithID":null},"HostNumber":2,"HostRegex":["^[0-9]+$","^[a-zA-Z]+$"],"HostWithID":true,"IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDuration":300000000000,"ReconnectTimes":3,"SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":["2"],"TCPRanges":["1024-65535"],"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Fri Dec  9 18:39:22 CST 2022 INF Listening addr=:8080
Fri Dec  9 18:39:22 CST 2022 INF acceptLoop started addr=[::]:8080
```

- 客户端（内网服务器）

```shell
# ./release/linux-amd64-client -remote tcp://id1.example.com:8080 -id id1 -secret secret1 \
>     -local http://127.0.0.1:80 -useLocalAsHTTPHost -hostPrefix 1 \
>     -local http://127.0.0.1:8080 -useLocalAsHTTPHost -hostPrefix 2 \
>     -local tcp://127.0.0.1:2222 -remoteTCPPort 2222 \
>     -local tcp://127.0.0.1:2223 -remoteTCPPort 2223
Fri Dec  9 18:40:10 CST 2022 INF linux-amd64-client - 2022-12-09 05:20:39 - dev 88d322f config={"Config":"","HostPrefix":[{"Position":2,"Value":"1"},{"Position":5,"Value":"2"}],"ID":"id1","Local":[{"Position":0,"Value":"http://127.0.0.1:80"},{"Position":3,"Value":"http://127.0.0.1:8080"},{"Position":6,"Value":"tcp://127.0.0.1:2222"},{"Position":8,"Value":"tcp://127.0.0.1:2223"}],"LocalTimeout":null,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tcp://id1.example.com:8080","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":false,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":[{"Position":7,"Value":2222},{"Position":9,"Value":2223}],"RemoteTCPRandom":null,"RemoteTimeout":45000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-12-09 05:20:39 - dev 88d322f","SentrySampleRate":1,"SentryServerName":"","Services":null,"UseLocalAsHTTPHost":[{"Position":1,"Value":true},{"Position":4,"Value":true}],"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Fri Dec  9 18:40:10 CST 2022 INF remote url remote=tcp://id1.example.com:8080 stun=
Fri Dec  9 18:40:10 CST 2022 INF trying to connect to remote connID=1
Fri Dec  9 18:40:10 CST 2022 INF receive server information: tcp port 2222 opened successfully connID=1
Fri Dec  9 18:40:10 CST 2022 INF receive server information: tcp port 2223 opened successfully connID=1
Fri Dec  9 18:40:10 CST 2022 INF tunnel started connID=1
```

上面的命令行也可以使用配置文件来启动

```shell
# ./release/linux-amd64-client -config client.yaml
Fri Dec  9 18:41:03 CST 2022 INF linux-amd64-client - 2022-12-09 05:20:39 - dev 88d322f config={"Config":"client.yaml","HostPrefix":null,"ID":"id1","Local":null,"LocalTimeout":null,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tcp://id1.example.com:8080","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":false,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":null,"RemoteTCPRandom":null,"RemoteTimeout":45000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-12-09 05:20:39 - dev 88d322f","SentrySampleRate":1,"SentryServerName":"","Services":[{"HostPrefix":"1","LocalTimeout":0,"LocalURL":{"ForceQuery":false,"Fragment":"","Host":"127.0.0.1:80","OmitHost":false,"Opaque":"","Path":"","RawFragment":"","RawPath":"","RawQuery":"","Scheme":"http","User":null},"RemoteTCPPort":0,"RemoteTCPRandom":null,"UseLocalAsHTTPHost":true},{"HostPrefix":"2","LocalTimeout":0,"LocalURL":{"ForceQuery":false,"Fragment":"","Host":"127.0.0.1:8080","OmitHost":false,"Opaque":"","Path":"","RawFragment":"","RawPath":"","RawQuery":"","Scheme":"http","User":null},"RemoteTCPPort":0,"RemoteTCPRandom":null,"UseLocalAsHTTPHost":true},{"HostPrefix":"","LocalTimeout":0,"LocalURL":{"ForceQuery":false,"Fragment":"","Host":"127.0.0.1:2222","OmitHost":false,"Opaque":"","Path":"","RawFragment":"","RawPath":"","RawQuery":"","Scheme":"tcp","User":null},"RemoteTCPPort":2222,"RemoteTCPRandom":null,"UseLocalAsHTTPHost":false},{"HostPrefix":"","LocalTimeout":0,"LocalURL":{"ForceQuery":false,"Fragment":"","Host":"127.0.0.1:2223","OmitHost":false,"Opaque":"","Path":"","RawFragment":"","RawPath":"","RawQuery":"","Scheme":"tcp","User":null},"RemoteTCPPort":2223,"RemoteTCPRandom":null,"UseLocalAsHTTPHost":false}],"UseLocalAsHTTPHost":null,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Fri Dec  9 18:41:03 CST 2022 INF remote url remote=tcp://id1.example.com:8080 stun=
Fri Dec  9 18:41:03 CST 2022 INF trying to connect to remote connID=1
Fri Dec  9 18:41:03 CST 2022 INF receive server information: tcp port 2222 opened successfully connID=1
Fri Dec  9 18:41:03 CST 2022 INF receive server information: tcp port 2223 opened successfully connID=1
Fri Dec  9 18:41:03 CST 2022 INF tunnel started connID=1
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

## 参数

### 客户端参数

```shell
# ./release/linux-amd64-client -h
Usage of ./release/linux-amd64-client:
  -config string
        配置文件路径。
  -hostPrefix value
        服务端将识别此 host 前缀并转发到 local
  -id string
        唯一的用户标识符。目前为域名的前缀。
  -local value
        需要转发的本地服务地址
  -localTimeout value
        本地服务超时时间。支持像‘30s’，‘5m’这样的值
  -logFile string
        保存日志文件的路径
  -logFileMaxCount uint
        日志文件数量限制（默认 7）
  -logFileMaxSize int
        日志文件大小（默认 536870912）
  -logLevel string
        日志级别: trace, debug, info, warn, error, fatal, panic, disable (默认 "info")。
  -reconnectDelay duration
        重连等待时间 (默认 5s)
  -remote string
        服务端地址。支持 tcp:// 和 tls://, 默认 tcp://。
  -remoteAPI string
        获取服务端地址的 API
  -remoteCert string
        服务器证书路径
  -remoteCertInsecure
        允许自签名的服务器证书
  -remoteConnections uint
        服务器的连接数。有效值为 1 到 10（默认 3）
  -remoteIdleConnections uint
        在池中保留的空闲服务器连接数（默认为1）
  -remoteSTUN string
        STUN 服务端地址
  -remoteTCPPort value
        服务端将开启的 TCP 端口
  -remoteTCPRandom
        是否由服务端选择一个随机的 TCP 端口
  -remoteTimeout duration
        服务器连接超时。支持像‘30s’，‘5m’这样的值（默认 5s）
  -secret value
        用于校验 ID 的机密
  -sentryDSN string
        开启上报日志到 Sentry  DSN 的功能。
  -sentryDebug
        开启 Sentry debug 模式
  -sentryEnvironment string
        发送到 Sentry 的 environment
  -sentryLevel value
        发送到 Sentry 的日志级别: trace, debug, info, warn, error, fatal, panic (默认 ["error", "fatal", "panic"])
  -sentryRelease string
        发送到 Sentry 的 release
  -sentrySampleRate float
        发送到 Sentry 的 sample rate : [0.0 - 1.0] (默认 1)
  -sentryServerName string
        发送到 Sentry 的 server name
  -tcpForwardAddr string
        TCP 转发的监听地址
  -tcpForwardConnections uint
        TCP 转发所建立的 peer connection 数量。有效值为 1 到 10（默认 3）
  -tcpForwardHostPrefix string
        TCP 转发的对方客户端的 HostPrefix
  -useLocalAsHTTPHost
        转发请求到 local 参数指定的地址时将 local 参数作为 HTTP Host
  -version
        打印此程序的版本
  -webrtcConnectionIdleTimeout duration
        WebRTC 连接的超时时间。支持像‘30s’，‘5m’这样的值（默认 5m0s）
  -webrtcLogLevel string
        WebRTC 日志级别：verbose, info, warning, error (默认 "warning")
  -webrtcMaxPort uint
        WebRTC peer connection 的最大端口
  -webrtcMinPort uint
        WebRTC peer connection 的最小端口
```

### 服务端参数

```shell
# ./release/linux-amd64-server -h
Usage of ./release/linux-amd64-server:
  -addr string
        监听地址（默认 80）。支持像‘80’，‘:80’或‘0.0.0.0:80’这样的值
  -allowAnyClient
        允许任意的客户端连接服务端
  -apiAddr string
        api 监听地址。支持像‘80’，‘:80’或‘0.0.0.0:80’这样的值
  -apiCertFile string
        cert 文件路径
  -apiKeyFile string
        key 文件路径
  -apiTLSVersion string
        最低 tls 版本，支持的值： tls1.1, tls1.2, tls1.3 (默认 "tls1.2")
  -authAPI string
        验证用户的 ID 和 secret 的 API
  -certFile string
        cert 路径
  -config string
        配置文件路径
  -connections uint
        客户端隧道的最大连接数 (默认 10)
  -hostNumber value
        客户端可开启的基于 host 的服务数量
  -hostRegex value
        客户端开启的 host 前缀必须满足其中的一条规则
  -hostWithID
        host 前缀的形式变为 id-host
  -httpMUXHeader string
        HTTP 多路复用的头部（默认“Host”）
  -id value
        用户标识符
  -keyFile string
        key 路径
  -logFile string
        保存日志文件的路径
  -logFileMaxCount uint
        日志文件数量限制（默认 7）
  -logFileMaxSize int
        日志文件大小（默认 536870912）
  -logLevel string
        日志级别: trace, debug, info, warn, error, fatal, panic, disable (默认 "info")。
  -reconnectDuration duration
        客户端达到失败重连的最大数后不能连接服务端的时间 (默认 5m0s)
  -reconnectTimes uint
        客户端失败重连的最大数 (默认 3)
  -secret value
        用于校验 ID 的机密
  -sentryDSN string
        开启上报日志到 Sentry  DSN 的功能。
  -sentryDebug
        开启 Sentry debug 模式
  -sentryEnvironment string
        发送到 Sentry 的 environment
  -sentryLevel value
        发送到 Sentry 的日志级别: trace, debug, info, warn, error, fatal, panic (默认 ["error", "fatal", "panic"])
  -sentryRelease string
        发送到 Sentry 的 release
  -sentrySampleRate float
        发送到 Sentry 的 sample rate : [0.0 - 1.0] (默认 1)
  -sentryServerName string
        发送到 Sentry 的 server name
  -sniAddr string
        原生的 TLS 代理的监听地址。Host 来源于 Server Name Indication。支持像‘80’，‘:80’或‘0.0.0.0:80’这样的值
  -speed uint
        用户每秒能传输的最大字节数
  -stunAddr string
        STUN 服务的监听地址。支持像‘3478’，‘:3478’或‘0.0.0.0:3478’这样的值
  -tcpNumber value
        允许每一个用户开启的 TCP 端口数量
  -tcpRange value
        TCP 端口范围, 比如 1024-65535
  -timeout duration
        全局超时。支持像‘30s’，‘5m’这样的值（默认 90s）
  -timeoutOnUnidirectionalTraffic
        当流量是单向的时会发生超时
  -tlsAddr string
        tls 监听地址。支持像‘80’，‘:80’或‘0.0.0.0:80’这样的值
  -tlsVersion string
        最低 tls 版本，支持的值： tls1.1, tls1.2, tls1.3 (默认 "tls1.2")
  -users string
        yaml 格式的用户配置文件
  -version
        打印此程序的版本
```

### 配置文件

配置文件使用 yaml 格式，客户端与服务端均可以使用配置文件。[HTTP 内网穿透](#HTTP-内网穿透)
示例中的客户端也可以使用下面的文件（client.yaml）启动。启动命令为：`./release/linux-amd64-client -config client.yaml`

```yaml
version: 1.0 # 保留关键字，目前暂未使用
options:
  local: http://127.0.0.1:80
  remote: tcp://id1.example.com:8080
  id: id1
  secret: secret1
```

### 服务端配置 users

以下四种方式可同时使用，如果冲突则按照从上到下优先级依次降低的方式解决。

#### 通过命令行配置 users

第 i 个 id 与第 i 个 secret 相匹配。下面两种启动方式是等价的。

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -id id2 -secret secret2
Sat Nov 19 20:22:27 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1","id2"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1","secret2"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:22:27 CST 2022 INF Listening addr=:8080
Sat Nov 19 20:22:27 CST 2022 INF acceptLoop started addr=[::]:8080
```

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -id id2 -secret secret1 -secret secret2
Sat Nov 19 20:22:47 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1","id2"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1","secret2"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:22:47 CST 2022 INF Listening addr=:8080
Sat Nov 19 20:22:47 CST 2022 INF acceptLoop started addr=[::]:8080
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
  host: 1.2.3.4
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
    - number: 65535
      range: 1-65535
id2:
  secret: secret2
```

#### 通过 config 配置文件配置 TCP

通过 config 配置文件可以配置全局 TCP 和单个用户的 TCP。下面的配置文件表示用户 id1 可以开启任意数量的任意 TCP 端口，用户
id2 可以在 1024 到 65535 的 TCP 端口之间开启 1 个 TCP 端口。

```yaml
version: 1.0
users:
  id1:
    secret: secret1
    tcp:
      - number: 65535
        range: 1-65535
  id2:
    secret: secret2
tcp:
  - number: 1
    range: 1024-65535
options:
  apiAddr: 1.2.3.4:1234
  certFile: /path
  host: 1.2.3.4
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

#### 通过命令行配置 TCP

通过命令行可以配置全局 TCP。下面的命令表示同一时间内每个用户都可以在 1024 到 65535 的 TCP 端口之间开启 1 个 TCP 端口。

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 1 -tcpRange 1024-65535
Sat Nov 19 20:27:41 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":["1"],"TCPRanges":["1024-65535"],"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:27:41 CST 2022 INF Listening addr=:8080
Sat Nov 19 20:27:41 CST 2022 INF acceptLoop started addr=[::]:8080
```

### 服务端 API

服务端 API 通过模拟客户端检测服务是否正常。下面的例子可以帮助你更好地理解这一点，其中，id1.example.com 解析到公网服务器的地址。当
apiCertFile 和 apiKeyFile 选项不为空时使用 HTTPS，其他情况使用 HTTP。

- 服务端（公网服务器）

```shell
# ./release/linux-amd64-server -addr 8080 -apiAddr 8081
Sat Nov 19 20:23:16 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"8081","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":null,"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":null,"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:23:16 CST 2022 WRN working on -allowAnyClient mode, because no user is configured
Sat Nov 19 20:23:16 CST 2022 INF Listening addr=:8080
Sat Nov 19 20:23:16 CST 2022 INF acceptLoop started addr=[::]:8080
```

- 用户

```shell
# curl http://id1.example.com:8081/status
{"status": "ok", "version":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e"}
```

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
    - number: 65535
      range: 1-65535
id2:
  secret: secret2
```

#### 通过 config 配置文件配置 TCP

通过 config 配置文件可以配置全局 TCP 和单个用户的 TCP。下面的配置文件表示用户 id1 可以开启任意数量的任意 TCP 端口，用户
id2 可以在 1024 到 65535 的 TCP 端口之间开启 1 个 TCP 端口。

```yaml
version: 1.0
users:
  id1:
    secret: secret1
    tcp:
      - number: 65535
        range: 1-65535
  id2:
    secret: secret2
tcp:
  - number: 1
    range: 1024-65535
options:
  apiAddr: 1.2.3.4:1234
  certFile: /path
  host: 1.2.3.4
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

#### 通过命令行配置 TCP

通过命令行可以配置全局 TCP。下面的命令表示同一时间内每个用户都可以在 1024 到 65535 的 TCP 端口之间开启 1 个 TCP 端口。

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 1 -tcpRange 1024-65535
Sat Nov 19 20:27:41 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":["1"],"TCPRanges":["1024-65535"],"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:27:41 CST 2022 INF Listening addr=:8080
Sat Nov 19 20:27:41 CST 2022 INF acceptLoop started addr=[::]:8080
```

### 服务端 API

服务端 API 通过模拟客户端检测服务是否正常。下面的例子可以帮助你更好地理解这一点，其中，id1.example.com 解析到公网服务器的地址。当
apiCertFile 和 apiKeyFile 选项不为空时使用 HTTPS，其他情况使用 HTTP。

- 服务端（公网服务器）

```shell
# ./release/linux-amd64-server -addr 8080 -apiAddr 8081
Fri Dec  9 18:41:46 CST 2022 INF linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f config={"APIAddr":"8081","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":10,"HTTPMUXHeader":"Host","Host":{"Number":null,"Regex":null,"RegexStr":null,"WithID":null},"HostNumber":1,"HostRegex":null,"HostWithID":false,"IDs":null,"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDuration":300000000000,"ReconnectTimes":3,"SNIAddr":"","STUNAddr":"","Secrets":null,"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Fri Dec  9 18:41:46 CST 2022 WRN working on -allowAnyClient mode, because no user is configured
Fri Dec  9 18:41:46 CST 2022 INF Listening addr=:8080
Fri Dec  9 18:41:46 CST 2022 INF acceptLoop started addr=[::]:8080
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

### AONetwork benchmark

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

## 编译

```shell
git clone <url>
cd <folder>
make release
```

编译好的执行文件在 release 目录下。
