# GT

Focus on high-performance, low-latency intranet penetration solutions.

- Supports forwarding of TCP-based communication protocols such as HTTP(S), WebSocket(S), SSH, SMB, etc.
- Supports WebRTC peer-to-peer connection
- Supports log reporting to Sentry service
- Multi-user configurations
  - Support multiple user authentication methods: API service, local configuration
  - Independent configuration for each user
  - Limit user speed
  - Limit the number of client connections
  - Deny access for a period of time after a certain number of failed authentication attempts
- Uses TCP connection pooling for communication between server and client
- Keep command line parameters consistent with yaml configuration parameters

## Index

- [Working Principle](#working-principle)
- [Examples](#examples)
  - [HTTP](#http)
  - [HTTPS Decrypted Into HTTP](#https-decrypted-into-http)
  - [HTTPS Directly](#https-directly)
  - [Client HTTP Convert to HTTPS](#client-http-convert-to-https)
  - [TCP](#tcp)
  - [Client Starts Multiple Services at The Same Time](#client-starts-multiple-services-at-the-same-time)
- [Parameters](#parameters)
  - [Client Parameters](#client-parameters)
  - [Server Parameters](#server-parameters)
  - [Configuration](#configuration)
  - [Server User Configurations](#server-user-configurations)
    - [Configure Users Through Command Line](#configure-users-through-command-line)
    - [Configure Users Through Users Configuration File](#configure-users-through-users-configuration-file)
    - [Configure Users Through Config Configuration File](#configure-users-through-config-configuration-file)
    - [Allow Any Client](#allow-any-client)
  - [Server TCP Configurations](#server-tcp-configurations)
    - [Configure TCP Through Users Configuration File](#configure-tcp-through-users-configuration-file)
    - [Configure TCP Through Config Configuration File](#configure-tcp-through-config-configuration-file)
    - [Configure TCP Through Command Line](#configure-tcp-through-command-line)
  - [Server API](#server-api)
- [Benchmark](#benchmark)
  - [GT benchmark](#gt-benchmark)
  - [frp dev branch 42745a3](#frp-dev-branch-42745a3)
- [Compile](#compile)
- [TODO](#todo)
- [Contributors](#contributors)
- [中文文档](./README_CN.md)

## Working Principle

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

## Examples

### HTTP

- Requirements: There is an internal gt server and a public gt server, id1.example.com resolves to the address of the
  public gt server. Hope to visit the webpage served by port 80 on the intranet server by visiting id1.example.com:8080.

- Server (public)

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -secret secret1
Sat Nov 19 20:16:33 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:16:33 CST 2022 INF Listening addr=:8080
Sat Nov 19 20:16:33 CST 2022 INF acceptLoop started addr=[::]:8080
```

- Client (intranal)

```shell
# ./release/linux-amd64-client -local http://127.0.0.1:80 -remote tcp://id1.example.com:8080 -id id1 -secret secret1
Sat Nov 19 20:18:59 CST 2022 INF linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e config={"Config":"","ID":"id1","Local":"http://127.0.0.1:80","LocalTimeout":120000000000,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tcp://id1.example.com:8080","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":false,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":0,"RemoteTCPRandom":false,"RemoteTimeout":5000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","UseLocalAsHTTPHost":false,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Sat Nov 19 20:18:59 CST 2022 INF remote url remote=tcp://id1.example.com:8080 stun=
Sat Nov 19 20:18:59 CST 2022 INF trying to connect to remote connID=1
Sat Nov 19 20:18:59 CST 2022 INF tunnel started connID=1
```

### HTTPS Decrypted Into HTTP

- Requirements: There is an intranet server and a public network server, and id1.example.com resolves to the address of
  the public network server. Want to access the HTTP web page served on port 80 on the intranet server by
  visiting <https://id1.example.com>.
- Server (public)

```shell
# ./release/linux-amd64-server -addr "" -tlsAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
Sat Nov 19 20:19:53 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"","AllowAnyClient":false,"AuthAPI":"","CertFile":"/root/openssl_crt/tls.crt","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"/root/openssl_crt/tls.key","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"443","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:19:53 CST 2022 INF Listening TLS addr=:443
Sat Nov 19 20:19:53 CST 2022 INF acceptLoop started addr=[::]:443
```

- Client (internal), because it uses a self-signed certificate, uses the -remoteCertInsecure option, otherwise it is
  forbidden to use this option (man-in-the-middle attacks cause encrypted content to be decrypted)

```shell
# ./release/linux-amd64-client -local http://127.0.0.1 -remote tls://id1.example.com -remoteCertInsecure -id id1 -secret secret1
Sat Nov 19 20:20:05 CST 2022 INF linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e config={"Config":"","ID":"id1","Local":"http://127.0.0.1","LocalTimeout":120000000000,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tls://id1.example.com","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":true,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":0,"RemoteTCPRandom":false,"RemoteTimeout":5000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","UseLocalAsHTTPHost":false,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Sat Nov 19 20:20:05 CST 2022 INF remote url remote=tls://id1.example.com stun=
Sat Nov 19 20:20:05 CST 2022 INF trying to connect to remote connID=1
Sat Nov 19 20:20:06 CST 2022 INF tunnel started connID=1
```

### HTTPS Directly

- Requirements: There is an intranet server and a public network server, and id1.example.com resolves to the address of
  the public network server. Want to access the HTTPS webpage served on port 443 on the intranet server by
  visiting <https://id1.example.com>.
- Server (public)

```shell
# ./release/linux-amd64-server -addr "" -sniAddr 443 -id id1 -secret secret1
Sat Nov 19 20:25:15 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"443","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:25:15 CST 2022 INF Listening sniAddr=:443
Sat Nov 19 20:25:15 CST 2022 INF acceptLoop started addr=[::]:443
```

- Client (internal)

```shell
# ./release/linux-amd64-client -local https://127.0.0.1 -remote tcp://id1.example.com:443 -id id1 -secret secret1
Sat Nov 19 20:25:49 CST 2022 INF linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e config={"Config":"","ID":"id1","Local":"https://127.0.0.1","LocalTimeout":120000000000,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tcp://id1.example.com:443","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":false,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":0,"RemoteTCPRandom":false,"RemoteTimeout":5000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","UseLocalAsHTTPHost":false,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Sat Nov 19 20:25:49 CST 2022 INF remote url remote=tcp://id1.example.com:443 stun=
Sat Nov 19 20:25:49 CST 2022 INF trying to connect to remote connID=1
Sat Nov 19 20:25:49 CST 2022 INF tunnel started connID=1
```

### Client HTTP Convert to HTTPS

- Requirements: There is an internal gt server and a public gt server, id1.example.com resolves to the address of the
  public gt server. Hope to visit the webpage served by port 80 on the intranet server by visiting id1.example.com:8080.
  At the same time, TLS is used to encrypt the communication between the client and the server.

- Server (public)

```shell
# ./release/linux-amd64-server -addr 8080 -tlsAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
Sat Nov 19 20:20:59 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"/root/openssl_crt/tls.crt","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"/root/openssl_crt/tls.key","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"443","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:20:59 CST 2022 INF Listening TLS addr=:443
Sat Nov 19 20:20:59 CST 2022 INF Listening addr=:8080
Sat Nov 19 20:20:59 CST 2022 INF acceptLoop started addr=[::]:8080
Sat Nov 19 20:20:59 CST 2022 INF acceptLoop started addr=[::]:443
```

- Client (internal), because it uses a self-signed certificate, so the use of the -remoteCertInsecureoption,
  otherwise prohibit the use of this option (middle attack led to the encrypted content is decrypted)

```shell
# ./release/linux-amd64-client -local http://127.0.0.1:80 -remote tls://id1.example.com -remoteCertInsecure -id id1 -secret secret1
Sat Nov 19 20:26:33 CST 2022 INF linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e config={"Config":"","ID":"id1","Local":"http://127.0.0.1:80","LocalTimeout":120000000000,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tls://id1.example.com","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":true,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":0,"RemoteTCPRandom":false,"RemoteTimeout":5000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-11-19 11:07:33 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","UseLocalAsHTTPHost":false,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Sat Nov 19 20:26:33 CST 2022 INF remote url remote=tls://id1.example.com stun=
Sat Nov 19 20:26:33 CST 2022 INF trying to connect to remote connID=1
Sat Nov 19 20:26:33 CST 2022 INF tunnel started connID=1
```

### TCP

- Requirements: There is an intranet server and a public network server, and id1.example.com resolves to the address of
  the public network server. Hope to access the SSH service on port 22 on the intranet server by accessing
  id1.example.com:2222. If the server port 2222 cannot be used, the server will choose a random port.

- Server (public)

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 1 -tcpRange 1024-65535
Fri Dec  9 18:38:21 CST 2022 INF linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":10,"HTTPMUXHeader":"Host","Host":{"Number":null,"Regex":null,"RegexStr":null,"WithID":null},"HostNumber":1,"HostRegex":null,"HostWithID":false,"IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDuration":300000000000,"ReconnectTimes":3,"SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":["1"],"TCPRanges":["1024-65535"],"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Fri Dec  9 18:38:21 CST 2022 INF Listening addr=:8080
Fri Dec  9 18:38:21 CST 2022 INF acceptLoop started addr=[::]:8080
```

- Client (internal)

```shell
# ./release/linux-amd64-client -local tcp://127.0.0.1:22 -remote tcp://id1.example.com:8080 -id id1 -secret secret1 -remoteTCPPort 2222 -remoteTCPRandom
Fri Dec  9 18:39:05 CST 2022 INF linux-amd64-client - 2022-12-09 05:20:39 - dev 88d322f config={"Config":"","HostPrefix":null,"ID":"id1","Local":[{"Position":0,"Value":"tcp://127.0.0.1:22"}],"LocalTimeout":null,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tcp://id1.example.com:8080","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":false,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":[{"Position":1,"Value":2222}],"RemoteTCPRandom":[{"Position":2,"Value":true}],"RemoteTimeout":45000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-12-09 05:20:39 - dev 88d322f","SentrySampleRate":1,"SentryServerName":"","Services":null,"UseLocalAsHTTPHost":null,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Fri Dec  9 18:39:05 CST 2022 INF remote url remote=tcp://id1.example.com:8080 stun=
Fri Dec  9 18:39:05 CST 2022 INF trying to connect to remote connID=1
Fri Dec  9 18:39:05 CST 2022 INF receive server information: tcp port 2222 opened successfully connID=1
Fri Dec  9 18:39:05 CST 2022 INF tunnel started connID=1
```

### Client Starts Multiple Services at The Same Time

- Requirements: There is an intranet server and a public network server, and id1-1.example.com and id1-2.example.com
  resolve to the address of the public network server. Hope to access the service on port 80 on the intranet server by
  accessing id1-1.example.com:8080, and hope to access the service on port 8080 on the intranet server by accessing
  id1-2.example.com:8080. Visit id1-1.example.com:2222 to access the service on port 2222 on the intranet server, and
  hope to access the service on port 2223 on the intranet server by visiting id1-1.example.com:2223. At the same time,
  the server restricts the client's hostPrefix to only consist of pure numbers or pure letters.

- Note: In this mode, the parameters (remoteTCPPort, hostPrefix, etc.) corresponding to the client local should be
  located between this local and the next local.

- Server (public)

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 2 -tcpRange 1024-65535 -hostNumber 2 -hostWithID -hostRegex ^[0-9]+$ -hostRegex ^[a-zA-Z]+$
Fri Dec  9 18:39:22 CST 2022 INF linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":10,"HTTPMUXHeader":"Host","Host":{"Number":null,"Regex":null,"RegexStr":null,"WithID":null},"HostNumber":2,"HostRegex":["^[0-9]+$","^[a-zA-Z]+$"],"HostWithID":true,"IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDuration":300000000000,"ReconnectTimes":3,"SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":["2"],"TCPRanges":["1024-65535"],"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Fri Dec  9 18:39:22 CST 2022 INF Listening addr=:8080
Fri Dec  9 18:39:22 CST 2022 INF acceptLoop started addr=[::]:8080
```

- Client (internal)

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

The above command line can also be started using a configuration file

```shell
# ./release/linux-amd64-client -config client.yaml
Fri Dec  9 18:41:03 CST 2022 INF linux-amd64-client - 2022-12-09 05:20:39 - dev 88d322f config={"Config":"client.yaml","HostPrefix":null,"ID":"id1","Local":null,"LocalTimeout":null,"LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDelay":5000000000,"Remote":"tcp://id1.example.com:8080","RemoteAPI":"","RemoteCert":"","RemoteCertInsecure":false,"RemoteConnections":1,"RemoteSTUN":"","RemoteTCPPort":null,"RemoteTCPRandom":null,"RemoteTimeout":45000000000,"Secret":"secret1","SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-client - 2022-12-09 05:20:39 - dev 88d322f","SentrySampleRate":1,"SentryServerName":"","Services":[{"HostPrefix":"1","LocalTimeout":0,"LocalURL":{"ForceQuery":false,"Fragment":"","Host":"127.0.0.1:80","OmitHost":false,"Opaque":"","Path":"","RawFragment":"","RawPath":"","RawQuery":"","Scheme":"http","User":null},"RemoteTCPPort":0,"RemoteTCPRandom":null,"UseLocalAsHTTPHost":true},{"HostPrefix":"2","LocalTimeout":0,"LocalURL":{"ForceQuery":false,"Fragment":"","Host":"127.0.0.1:8080","OmitHost":false,"Opaque":"","Path":"","RawFragment":"","RawPath":"","RawQuery":"","Scheme":"http","User":null},"RemoteTCPPort":0,"RemoteTCPRandom":null,"UseLocalAsHTTPHost":true},{"HostPrefix":"","LocalTimeout":0,"LocalURL":{"ForceQuery":false,"Fragment":"","Host":"127.0.0.1:2222","OmitHost":false,"Opaque":"","Path":"","RawFragment":"","RawPath":"","RawQuery":"","Scheme":"tcp","User":null},"RemoteTCPPort":2222,"RemoteTCPRandom":null,"UseLocalAsHTTPHost":false},{"HostPrefix":"","LocalTimeout":0,"LocalURL":{"ForceQuery":false,"Fragment":"","Host":"127.0.0.1:2223","OmitHost":false,"Opaque":"","Path":"","RawFragment":"","RawPath":"","RawQuery":"","Scheme":"tcp","User":null},"RemoteTCPPort":2223,"RemoteTCPRandom":null,"UseLocalAsHTTPHost":false}],"UseLocalAsHTTPHost":null,"Version":"","WebRTCConnectionIdleTimeout":300000000000,"WebRTCLogLevel":"warning","WebRTCMaxPort":0,"WebRTCMinPort":0}
Fri Dec  9 18:41:03 CST 2022 INF remote url remote=tcp://id1.example.com:8080 stun=
Fri Dec  9 18:41:03 CST 2022 INF trying to connect to remote connID=1
Fri Dec  9 18:41:03 CST 2022 INF receive server information: tcp port 2222 opened successfully connID=1
Fri Dec  9 18:41:03 CST 2022 INF receive server information: tcp port 2223 opened successfully connID=1
Fri Dec  9 18:41:03 CST 2022 INF tunnel started connID=1
````

client.yaml file content：

```yaml
services:
  - id: id1
    secret: secret1
    local: http://127.0.0.1:80
    useLocalAsHTTPHost: true
  - id: id2
    secret: secret2
    local: http://127.0.0.1:8080
    useLocalAsHTTPHost: true
  - id: id1
    secret: secret1
    local: tcp://127.0.0.1:2222
    useLocalAsHTTPHost: true
    remoteTCPPort: 2222
  - id: id1
    secret: secret1
    local: tcp://127.0.0.1:2223
    useLocalAsHTTPHost: true
    remoteTCPPort: 2223
options:
  remote: tcp://id1.example.com:8080
```

## Parameters

### Client Parameters

```shell
# ./release/linux-amd64-client -h
Usage of ./release/linux-amd64-client:
  -config string
        The config file path to load
  -hostPrefix
        The server will recognize this host prefix and forward data to local
  -id string
        The unique id used to connect to server. Now it's the prefix of the domain.
  -local
        The local service url
  -localTimeout
        The timeout of local connections. Supports values like '30s', '5m'
  -logFile string
        Path to save the log file
  -logFileMaxCount uint
        Max count of the log files (default 7)
  -logFileMaxSize int
        Max size of the log files (default 536870912)
  -logLevel string
        Log level: trace, debug, info, warn, error, fatal, panic, disable (default "info")
  -reconnectDelay duration
        The delay before reconnect. Supports values like '30s', '5m' (default 5s)
  -remote string
        The remote server url. Supports tcp:// and tls://, default tcp://
  -remoteAPI string
        The API to get remote server url
  -remoteCert string
        The path to remote cert
  -remoteCertInsecure
        Accept self-signed SSL certs from remote
  -remoteConnections uint
        The max number of server connections in the pool. Valid value is 1 to 10 (default 3)
  -remoteIdleConnections uint
        The number of idle server connections kept in the pool (default 1)
  -remoteSTUN string
        The remote STUN server address
  -remoteTCPPort
        The TCP port that the remote server will open
  -remoteTCPRandom
        Whether to choose a random tcp port by the remote server
  -remoteTimeout duration
        The timeout of remote connections. Supports values like '30s', '5m' (default 45s)
  -secret string
        The secret used to verify the id
  -sentryDSN string
        Sentry DSN to use
  -sentryDebug
        Sentry debug mode, the debug information is printed to help you understand what sentry is doing
  -sentryEnvironment string
        Sentry environment to be sent with events
  -sentryLevel
        Sentry levels: trace, debug, info, warn, error, fatal, panic (default ["error", "fatal", "panic"])
  -sentryRelease string
        Sentry release to be sent with events
  -sentrySampleRate float
        Sentry sample rate for event submission: [0.0 - 1.0] (default 1)
  -sentryServerName string
        Sentry server name to be reported
  -tcpForwardAddr string
        The address of TCP forward
  -tcpForwardConnections uint
        The max number of TCP forward peer connections in the pool. Valid value is 1 to 10 (default 3)
  -tcpForwardHostPrefix string
        The host prefix of TCP forward
  -useLocalAsHTTPHost
        Use the local address as host
  -version
        Show the version of this program
  -webrtcConnectionIdleTimeout duration
        The timeout of WebRTC connection. Supports values like '30s', '5m' (default 5m0s)
  -webrtcLogLevel string
        WebRTC log level: verbose, info, warning, error (default "warning")
  -webrtcMaxPort uint
        The max port of WebRTC peer connection
  -webrtcMinPort uint
        The min port of WebRTC peer connection
```

### Server Parameters

```shell
# ./release/linux-amd64-server -h
Usage of ./release/linux-amd64-server:
  -addr string
        The address to listen on. Supports values like: '80', ':80' or '0.0.0.0:80' (default "80")
  -allowAnyClient
        Allow any client to connect to the server
  -apiAddr string
        The address to listen on for internal api service. Supports values like: '8080', ':8080' or '0.0.0.0:8080'
  -apiCertFile string
        The path to cert file
  -apiKeyFile string
        The path to key file
  -apiTLSVersion string
        The tls min version. Supports values: tls1.1, tls1.2, tls1.3 (default "tls1.2")
  -authAPI string
        The API to authenticate user with id and secret
  -certFile string
        The path to cert file
  -config string
        The config file path to load
  -connections uint
        The max number of tunnel connections for a client (default 10)
  -hostNumber value
        The number of host-based services that the user can start
  -hostRegex value
        The host prefix started by user must conform to one of these rules
  -hostWithID
        The prefix of host will become the form of id-host
  -httpMUXHeader string
        The http multiplexing header to be used (default "Host")
  -id value
        The user id
  -keyFile string
        The path to key file
  -logFile string
        Path to save the log file
  -logFileMaxCount uint
        Max count of the log files (default 7)
  -logFileMaxSize int
        Max size of the log files (default 536870912)
  -logLevel string
        Log level: trace, debug, info, warn, error, fatal, panic, disable (default "info")
  -reconnectDuration duration
        The time that the client cannot connect after the number of failed reconnections reaches the max number (default 5m0s)
  -reconnectTimes uint
        The max number of times the client fails to reconnect (default 3)
  -secret value
        The secret for user id
  -sentryDSN string
        Sentry DSN to use
  -sentryDebug
        Sentry debug mode, the debug information is printed to help you understand what sentry is doing
  -sentryEnvironment string
        Sentry environment to be sent with events
  -sentryLevel value
        Sentry levels: trace, debug, info, warn, error, fatal, panic (default ["error", "fatal", "panic"])
  -sentryRelease string
        Sentry release to be sent with events (default "server - 2022-05-28 17:37:00 - client-multi-service 7ad41e7")
  -sentrySampleRate float
        Sentry sample rate for event submission: [0.0 - 1.0] (default 1)
  -sentryServerName string
        Sentry server name to be reported
  -sniAddr string
        The address to listen on for raw tls proxy. Host comes from Server Name Indication. Supports values like: '443', ':443' or '0.0.0.0:443'
  -speed uint
        The max number of bytes the client can transfer per second
  -stunAddr string
        The address to listen on for STUN service. Supports values like: '3478', ':3478' or '0.0.0.0:3478'
  -tcpNumber value
        The number of tcp ports allowed to be opened for each id
  -tcpRange value
        The tcp port range, like 1024-65535
  -timeout duration
        The timeout of connections. Supports values like '30s', '5m' (default 1m30s)
  -timeoutOnUnidirectionalTraffic
        Timeout will happens when traffic is unidirectional
  -tlsAddr string
        The address for tls to listen on. Supports values like: '443', ':443' or '0.0.0.0:443'
  -tlsVersion string
        The tls min version. Supports values: tls1.1, tls1.2, tls1.3 (default "tls1.2")
  -users string
        The users yaml file to load
  -version
        Show the version of this program
```

### Configuration

The configuration file uses the yaml format, and both the client and the server can use configuration file. The client
in the [HTTP example](#http) can also be started with the following file (client.yaml). The startup command
is: `./release/linux-amd64-client -config client.yaml`

```yaml
version: 1.0 # Reserved keywords, currently not used
options:
  local: http://127.0.0.1:80
  remote: tcp://id1.example.com:8080
  id: id1
  secret: secret1
```

### Server User Configurations

The following four methods can be used at the same time. If conflicts are resolved, the priority will be lowered from
top to bottom.

#### Configure Users Through Command Line

The i-th id matches the i-th secret. The following two startup methods are equivalent.

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

#### Configure Users Through Users Configuration File

```yaml
id3:
  secret: secret3
id1:
  secret: secret1-overwrite
```

#### Configure Users Through Config Configuration File

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

#### Allow Any Client

Add `-allowAnyClient` to the startup parameters of the server, all clients can connect to the server without configuring
the server, but the clients with the same `id` only use the `secret` of the first client connected to the server as the
correct `secret`, which cannot be overwritten by subsequent clients to ensure security.

### Server TCP Configurations

The following three methods can be used at the same time. Priority: User > Global. User priority: users profile > config
profile. Global priority: command line > config configuration file. If TCP is not configured, it means that the TCP
function is not enabled.

#### Configure TCP Through Users Configuration File

TCP for a single user can be configured through the users configuration file. The following configuration file indicates
that user id1 can open any number of arbitrary TCP ports, and user id2 has no permission to open TCP ports.

```yaml
id1:
  secret: secret1
  tcp:
    - number: 65535
      range: 1-65535
id2:
  secret: secret2
```

#### Configure TCP Through Config Configuration File

Through the config configuration file, you can configure the global TCP and the TCP of a single user. The following
configuration file indicates that user id1 can open any number of arbitrary TCP ports, and user id2 can open 1 TCP port
between TCP ports 1024 to 65535.

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

#### Configure TCP Through Command Line

Global TCP can be configured through the command line. The following command indicates that each user can open 1 TCP
port between TCP ports 1024 to 65535 at the same time.

```shell
# ./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 1 -tcpRange 1024-65535
Sat Nov 19 20:27:41 CST 2022 INF linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e config={"APIAddr":"","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":0,"HTTPMUXHeader":"Host","IDs":["id1"],"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","SNIAddr":"","STUNAddr":"","Secrets":["secret1"],"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-11-19 11:07:19 - google-webrtc 9240c2e","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":["1"],"TCPRanges":["1024-65535"],"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Sat Nov 19 20:27:41 CST 2022 INF Listening addr=:8080
Sat Nov 19 20:27:41 CST 2022 INF acceptLoop started addr=[::]:8080
```

### Server API

The server API detects whether the service is working by simulating the client. The following example can help you
understand this better, where id1.example.com resolves to the address of the public server. Use HTTPS when the
apiCertFile and apiKeyFile options are not empty, otherwise use HTTP.

- Server (public)

```shell
# ./release/linux-amd64-server -addr 8080 -apiAddr 8081
Fri Dec  9 18:41:46 CST 2022 INF linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f config={"APIAddr":"8081","APICertFile":"","APIKeyFile":"","APITLSMinVersion":"tls1.2","Addr":"8080","AllowAnyClient":false,"AuthAPI":"","CertFile":"","Config":"","Connections":10,"HTTPMUXHeader":"Host","Host":{"Number":null,"Regex":null,"RegexStr":null,"WithID":null},"HostNumber":1,"HostRegex":null,"HostWithID":false,"IDs":null,"KeyFile":"","LogFile":"","LogFileMaxCount":7,"LogFileMaxSize":536870912,"LogLevel":"info","ReconnectDuration":300000000000,"ReconnectTimes":3,"SNIAddr":"","STUNAddr":"","Secrets":null,"SentryDSN":"","SentryDebug":false,"SentryEnvironment":"","SentryLevel":null,"SentryRelease":"linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f","SentrySampleRate":1,"SentryServerName":"","Speed":0,"TCPNumbers":null,"TCPRanges":null,"TCPs":null,"TLSAddr":"","TLSMinVersion":"tls1.2","Timeout":90000000000,"TimeoutOnUnidirectionalTraffic":false,"Users":null,"Version":""}
Fri Dec  9 18:41:46 CST 2022 WRN working on -allowAnyClient mode, because no user is configured
Fri Dec  9 18:41:46 CST 2022 INF Listening addr=:8080
Fri Dec  9 18:41:46 CST 2022 INF acceptLoop started addr=[::]:8080
```

- User

```shell
# curl http://id1.example.com:8081/status
{"status": "ok", "version":"linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f"}
```

## Benchmark

Stress test through wrk. This project is compared with frp. The intranet service points to the test page of running
nginx locally. The test results are as follows:

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

## Compile

### Get the code

```shell
git clone <url>
cd <folder>
```

### Download Dependencies

You can either download webrtc from official or from mirror:

#### 1. Download webrtc from official

```shell
mkdir -p dep/_google-webrtc
cd dep/_google-webrtc
git clone https://webrtc.googlesource.com/src
```

And then follow the steps to check out the build toolchain and many
dependencies: <https://webrtc.googlesource.com/src/+/main/docs/native-code/development/>

#### 2. Download webrtc from mirror

WIP

### Build

Build it on linux:

```shell
make release
```

The compiled executable file is in the release directory.

## TODO

- Add web management capabilities
- Support for using QUIC protocol, BBR congestion algorithm
- Support for configuring P2P connections to forward data to multiple services
- Authentication function supports public and private keys

## Contribution Guidelines

Contributions to this project are very welcome. Here are some guidelines and suggestions to help you get involved in the project.

### Contributing Code

If you want to contribute to the project, the best way is to submit code. Before submitting code, please ensure that you have downloaded and familiarized yourself with the project code repository, and that your code adheres to the following guidelines:

- The code should be as concise as possible, and easy to maintain and expand.
- The code should follow the naming convention agreed by the project to ensure the consistency of the code.
- The code should follow the code style guide of the project, and you can refer to the existing code in the project code library.

If you want to submit code to the project, you can do so by following these steps:

- Fork the project on GitHub.
- Clone your forked project locally.
- Make your modifications and improvements locally.
- Perform tests to ensure that any changes have no impact.
- Commit your changes and create a pull request.

### Code Quality

We attach great importance to the quality of the code, so the code you submit should meet the following requirements:

- Code should be fully tested to ensure its correctness and stability.
- Code should follow good design principles and best practices.
- The code should conform as closely as possible to the relevant requirements of your submitted code contribution.

### Submit Information

Before committing code, please ensure that you provide a meaningful and detailed commit message. This helps us better understand your code contribution and merge it more quickly.

Submission information should include the following:

- Describe the purpose or reason for this code contribution.
- Describe the content or changes of this code contribution.
- (Optional) Describe the test methods or results of this code contribution.

The submission information should be clear and consistent with the submission information agreement of the project code base.

### Problem Reporting

If you encounter problems with the project, or find bugs, please submit an issue report to us. Before submitting an issue report, please ensure that you have thoroughly investigated and experimented with the issue and include as much of the following information as possible:

- Describe the symptoms and manifestations of the problem.
- Describe the scenario and conditions under which the problem occurred.
- Describe contextual information or any relevant background information.
- Information describing your desired behavior.
- (Optional) Provide relevant screenshots or error messages.

Issue reports should be clear and follow the issue reporting conventions of the project's codebase.

### Feature Request

If you want to add new functionality or features to the project, you are welcome to submit a feature request to us. Before submitting a feature request, please make sure you understand the history and current state of the project, and provide as much of the following information as possible:

- Describe the functionality or features you would like to add.
- Describe the purpose and purpose of this function or feature.
- (Optional) Provide relevant implementation ideas or suggestions.

Feature requests should be clear and follow the feature request conventions of the project's codebase.

### Thanks for your contribution

Finally, thank you for your contribution to this project. We welcome contributions in all forms, including but not limited to code contributions, issue reports, feature requests, documentation writing, etc. We believe that with your help, this project will become more perfect and stronger.

## Contributors

Many thanks to the following people who have contributed to this project:

- [zhiyi](https://github.com/vyloy)
- [jianti](https://github.com/FH0)
