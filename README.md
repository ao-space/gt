# GT

English | [简体中文](./README_CN.md)

GT is an open source reverse proxy project that supports peer-to-peer direct connection (P2P) and Internet relay.

It has the following design features:

- Emphasis on privacy protection, minimize server-side packet analysis to ensure functionality while protecting privacy,
  such as TCP-based implementation, analyze only the HTTP protocol header of the first data packet for application layer
  HTTP protocol transmission, without any redundant analysis, and directly forward subsequent data.

- Emphasis on performance, code implementation tends to use higher performance designs, such as modifying standard
  libraries to implement designs that reduce memory allocation and copying.

- P2P connection functionality implemented based on WebRTC, support all platforms that support WebRTC, such as iOS,
  Android, browsers, etc.

Currently implemented main functions:

- Forward communication protocols based on TCP, such as HTTP(S), WebSocket(S), SSH, SMB
- WebRTC P2P connection
- Multi-user functionality
  - Support multiple user authentication methods: API service, local configuration
  - Each user has independent configuration
  - Limit user speed
  - Limit number of client connections
  - Refuse access for a period of time if authentication fails more than a certain number of times
- Communication between server and client uses TCP connection pool
- Maintain consistency between command line parameters and YAML configuration parameters
- Support log reporting to Sentry service

## Index

<!-- TOC -->

* [Working Principle](#working-principle)
* [Usage](#usage)
  * [Configuration File](#configuration-file)
  * [Server User Configuration](#server-user-configuration)
    * [Configure Users via Command Line](#configure-users-via-command-line)
    * [Configure Users via Users Configuration File](#configure-users-via-users-configuration-file)
    * [Configure Users via Config Configuration File](#configure-users-via-config-configuration-file)
    * [Allow All Clients](#allow-all-clients)
  * [Server TCP Configuration](#server-tcp-configuration)
    * [Configure TCP via Users Configuration File](#configure-tcp-via-users-configuration-file)
    * [Configure TCP via Config Configuration File](#configure-tcp-via-config-configuration-file)
  * [Command Line Parameters](#command-line-parameters)
    * [Internal HTTP Penetration](#internal-http-penetration)
    * [Internal HTTPS Penetration](#internal-https-penetration)
    * [Internal HTTPS SNI Penetration](#internal-https-sni-penetration)
    * [Encrypt Client-Server Communication with TLS](#encrypt-client-server-communication-with-tls)
    * [Internal TCP Penetration](#internal-tcp-penetration)
    * [Client Start Multiple Services Simultaneously](#client-start-multiple-services-simultaneously)
    * [Server API](#server-api)
* [Performance Test](#performance-test)
  * [GT benchmark](#gt-benchmark)
  * [frp dev branch 42745a3](#frp-dev-branch-42745a3)
* [Run](#run)
  * [Docker Container Run](#docker-container-run)
* [Compilation](#compilation)
  * [Compilation on Ubuntu/Debian](#compilation-on-ubuntudebian)
    * [Install Dependencies](#install-dependencies)
    * [Get Code and Compile](#get-code-and-compile)
      * [Obtain WebRTC from ISCAS Mirror and Compile GT](#obtain-webrtc-from-iscas-mirror-and-compile-gt)
      * [Obtain WebRTC from Official and Compile GT](#obtain-webrtc-from-official-and-compile-gt)
  * [Compile on Ubuntu/Debian via Docker](#compile-on-ubuntudebian-via-docker)
    * [Install Dependencies](#install-dependencies-1)
    * [Get Code and Compile](#get-code-and-compile-1)
      * [Obtain WebRTC from ISCAS Mirror and Compile GT](#obtain-webrtc-from-iscas-mirror-and-compile-gt-1)
      * [Obtain WebRTC from Official and Compile GT](#obtain-webrtc-from-official-and-compile-gt-1)
* [Roadmap](#roadmap)
* [Contribution Guide](#contribution-guide)
  * [Contribute Code](#contribute-code)
  * [Code Quality](#code-quality)
  * [Commit Messages](#commit-messages)
  * [Issue Reporting](#issue-reporting)
  * [Feature Requests](#feature-requests)
  * [Thank You for Your Contribution](#thank-you-for-your-contribution)
  * [Contributors](#contributors)

<!-- TOC -->

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

## Usage

### Configuration File

Configuration files use YAML format. Both clients and servers can use configuration files.

```shell
./release/linux-amd64-server -config server.yaml
./release/linux-amd64-client -config client.yaml
```

See the [server.yaml](example/config/server.yaml) file for a basic server configuration.
See the [client.yaml](example/config/client.yaml) file for a basic client configuration.

### Server User Configuration

The following four ways can be used simultaneously. Conflicts are resolved in order of decreasing priority from top to
bottom.

#### Configure Users via Command Line

The ith id matches the ith secret. The following two startup methods are equivalent:

```shell
./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -id id2 -secret secret2
```

```shell
./release/linux-amd64-server -addr 8080 -id id1 -id id2 -secret secret1 -secret secret2
```

#### Configure Users via Users Configuration File

```yaml
id3:
  secret: secret3
id1:
  secret: secret1-overwrite
```

#### Configure Users via Config Configuration File

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

#### Allow All Clients

Add `-allowAnyClient` to the server startup parameters to allow all clients to connect to the server without
configuration on the server side. But the `secret` of the first client connecting to the server with the same `id` will
be used as the correct `secret` and cannot be overwritten by the `secret` of subsequent clients connecting to the server
with the same `id` to ensure security.

### Server TCP Configuration

The following three ways can be used simultaneously. Priority: User > Global. User priority: users configuration file >
config configuration file. Global priority: command line > config configuration file. If TCP is not configured, it means
TCP functionality is not enabled.

#### Configure TCP via Users Configuration File

The users configuration file can configure TCP for individual users. The following configuration file indicates that
user id1 can open TCP ports of any number and any port, and user id2 does not have the permission to open TCP ports.

```yaml
id1:
  secret: secret1
  tcp:
    - range: 1-65535
id2:
  secret: secret2  
```

#### Configure TCP via Config Configuration File

The config configuration file can configure global TCP and TCP for individual users. The following configuration file
indicates that user id1 can open TCP ports of any number and between ports
10000 to 20000, and user id2 can open 1 TCP port between ports
50000 to 65535.

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

### Command Line Parameters

```shell
./release/linux-amd64-server -h
./release/linux-amd64-client -h
```

#### Internal HTTP Penetration

- Requirement: There is an internal network server and a public network server, and id1.example.com resolves to the
  address of the public network server. It is hoped to access the webpage service on port 80 of the internal network
  server through accessing id1.example.com:8080.

- Server (Public network server)

```shell
./release/linux-amd64-server -addr 8080 -id id1 -secret secret1
```

- Client (Internal network server)

```shell
./release/linux-amd64-client -local http://127.0.0.1:80 -remote tcp://id1.example.com:8080 -id id1 -secret secret1
```

#### Internal HTTPS Penetration

- Requirement: There is an internal network server and a public network server, and id1.example.com resolves to the
  address of the public network server. It is hoped to access the HTTP webpage provided on port 80 of the internal
  network server through accessing <https://id1.example.com>

- Server (Public network server)

```shell
./release/linux-amd64-server -addr "" -tlsAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
```

- Client (Internal network server), uses the `-remoteCertInsecure` option as a self-signed certificate is used,
  otherwise this option should not be used (encryption content is decrypted due to man-in-the-middle attack)

```shell
./release/linux-amd64-client -local http://127.0.0.1 -remote tls://id1.example.com -remoteCertInsecure -id id1 -secret secret1  
```

#### Internal HTTPS SNI Penetration

- Requirement: There is an internal network server and a public network server, and id1.example.com resolves to the
  address of the public network server. It is hoped to access the HTTPS webpage provided on port 443 of the internal
  network server through accessing <https://id1.example.com>

- Server (Public network server)

```shell
./release/linux-amd64-server -addr 8080 -sniAddr 443 -id id1 -secret secret1
```

- Client (Internal network server)

```shell 
./release/linux-amd64-client -local https://127.0.0.1 -remote tcp://id1.example.com:8080 -id id1 -secret secret1
```

#### Encrypt Client-Server Communication with TLS

- Requirement: There is an internal network server and a public network server, and id1.example.com resolves to the
  address of the public network server. It is hoped to access the webpage service on port 80 of the internal network
  server through accessing id1.example.com:8080. Meanwhile, use TLS to encrypt the communication between the client and
  server.

- Server (Public network server)

```shell
./release/linux-amd64-server -addr 8080 -tlsAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
```

- Client (Internal network server), uses the `-remoteCertInsecure` option as a self-signed certificate is used,
  otherwise this option should not be used (encryption content is decrypted due to man-in-the-middle attack)

```shell  
./release/linux-amd64-client -local http://127.0.0.1:80 -remote tls://id1.example.com -remoteCertInsecure -id id1 -secret secret1
```

#### Internal TCP Penetration

- Requirement: There is an internal network server and a public network server, and id1.example.com resolves to the
  address of the public network server. It is hoped to access the SSH service on port 22 of the internal network server
  through accessing id1.example.com:2222, and if the 2222 port is not available on the server side, the server selects a
  random port.

- Server (Public network server)

```shell
./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 1 -tcpRange 1024-65535
```

- Client (Internal network server)

```shell
./release/linux-amd64-client -local tcp://127.0.0.1:22 -remote tcp://id1.example.com:8080 -id id1 -secret secret1 -remoteTCPPort 2222 -remoteTCPRandom
```

#### Internal QUIC Penetration

- Requirements: There is an intranet server and a public network server, and id1.example.com resolves to the address of the public network server. Hopefully by accessing id1.example.com:8080
  To access the web page served by port 80 on the intranet server. Use QUIC to build a transport connection between the client and the server. QUIC uses TLS 1.3 for transport encryption. When the user also gives certFile
  and keyFile, use them for encrypted communication. Otherwise, keys and certificates are automatically generated using the ECDSA encryption algorithm.

- Server (public network server)

```shell
./release/linux-amd64-server -addr 8080 -quicAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
```

- Client (internal network server), because a self-signed certificate is used, the `-remoteCertInsecure` option is used. This option is prohibited from being used in other cases (man-in-the-middle attacks cause encrypted content to be decrypted

```shell
./release/linux-amd64-client -local http://127.0.0.1:80 -remote quic://id1.example.com:443 -remoteCertInsecure -id id1 -secret secret1
```

#### Client Start Multiple Services Simultaneously

- Requirement: There is an internal network server and a public network server, and id1-1.example.com and
  id1-2.example.com resolve to the address of the public network server. It is hoped to access the service on port 80 of
  the internal network server through accessing id1-1.example.com:8080, and to access the service on port 8080 of the
  internal network server through accessing id1-2.example.com:8080, and to access the service on port 2222 of the
  internal network server through accessing id1-1.example.com:2222, and to access the service on port 2223 of the
  internal network server through accessing id1-1.example.com:2223. At the same time, the server limits the client's
  hostPrefix to only contain digits or letters.

- Note: In this mode, the parameters corresponding to the client local (remoteTCPPort, hostPrefix, etc.) need to be
  placed between this local and the next local.

- Server (Public network server)

```shell
./release/linux-amd64-server -addr 8080 -id id1 -secret secret1 -tcpNumber 2 -tcpRange 1024-65535 -hostNumber 2 -hostWithID -hostRegex ^[0-9]+$ -hostRegex ^[a-zA-Z]+$
```

- Client (Internal network server)

```shell
./release/linux-amd64-client -remote tcp://id1.example.com:8080 -id id1 -secret secret1 \
   -local http://127.0.0.1:80 -useLocalAsHTTPHost -hostPrefix 1 \
   -local http://127.0.0.1:8080 -useLocalAsHTTPHost -hostPrefix 2 \
   -local tcp://127.0.0.1:2222 -remoteTCPPort 2222 \
   -local tcp://127.0.0.1:2223 -remoteTCPPort 2223
```

The above command line can also be started using a configuration file:

```shell
./release/linux-amd64-client -config client.yaml
```

client.yaml file content:

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

#### Server API

The server API detects service availability by simulating a client. The following example can help you better understand
this point, where id1.example.com resolves to the address of the public network server. HTTPS is used when apiCertFile
and apiKeyFile options are not empty, otherwise HTTP is used.

- Server (Public network server)

```shell
./release/linux-amd64-server -addr 8080 -apiAddr 8081
```

- User

```shell
# curl http://id1.example.com:8081/status
{"status": "ok", "version":"linux-amd64-server - 2022-12-09 05:20:24 - dev 88d322f"}
```

## Performance Test

Load testing was performed on this project and frp for comparison using wrk. The internal service points to a test page
running nginx locally, and the test results are as follows:

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

## Run

### Docker Container Run

More container image information can be obtained from <https://github.com/ao-space/gt/pkgs/container/gt>.

```shell
docker pull ghcr.io/ao-space/gt:server-dev

docker pull ghcr.io/ao-space/gt:client-dev
```

## Compilation

### Compilation on Ubuntu/Debian

#### Install Dependencies

```shell
apt-get update
apt-get install make git gn ninja-build python3 python3-pip libgtk-3-dev gcc-aarch64-linux-gnu g++-aarch64-linux-gnu gcc-x86-64-linux-gnu g++-x86-64-linux-gnu -y
```

#### Get Code and Compile

You can choose to obtain WebRTC from the mirror or official and compile GT:

##### Obtain WebRTC from ISCAS Mirror and Compile GT

1. Get code

     ```shell
     git clone <url>
     cd <folder>
     ```

2. Compile

     ```shell
     make release
     ```

   The executable files are in the release directory.

##### Obtain WebRTC from Official and Compile GT

1. Get code

     ```shell
     git clone <url> 
     cd <folder>
     ```

2. Obtain WebRTC from official

     ```shell
     mkdir -p dep/_google-webrtc
     cd dep/_google-webrtc
     git clone https://webrtc.googlesource.com/src
     ```

   Then follow the steps in [this link](https://webrtc.googlesource.com/src/+/main/docs/native-code/development/) to
   check out the build toolchain and many dependencies.

3. Compile

     ```shell
     WITH_OFFICIAL_WEBRTC=1 make release
     ```

   The executable files are in the release directory.

### Compile on Ubuntu/Debian via Docker

#### Install Dependencies

[Install Docker](https://docs.docker.com/engine/install/)

#### Get Code and Compile

You can choose to obtain WebRTC from the mirror or official and compile GT:

##### Obtain WebRTC from ISCAS Mirror and Compile GT

1. Get code

     ```shell
     git clone <url>
     cd <folder>
     ```

2. Compile

     ```shell
     make docker_release_linux_amd64 # docker_release_linux_arm64
     ```

   The executable files are in the release directory.

##### Obtain WebRTC from Official and Compile GT

1. Get code

     ```shell
     git clone <url>
     cd <folder>
     ``` 

2. Obtain WebRTC from official

     ```shell
     mkdir -p dep/_google-webrtc
     cd dep/_google-webrtc
     git clone https://webrtc.googlesource.com/src
     ```

   Then follow the steps in [this link](https://webrtc.googlesource.com/src/+/main/docs/native-code/development/) to
   check out the build toolchain and many dependencies.

3. Compile

     ```shell
     WITH_OFFICIAL_WEBRTC=1 make docker_release_linux_amd64 # docker_release_linux_arm64
     ```

   The executable files are in the release directory.

## Roadmap

- Add web management functionality
- Support QUIC protocol and BBR congestion algorithm
- Support configuring P2P connections to forward data to multiple services
- Authentication support for public and private keys

## Contribution Guide

We highly welcome contributions to this project. Here are some guiding principles and recommendations to help you get
involved:

### Contribute Code

The best way to contribute is by submitting code. Before submitting code, please ensure you have downloaded and are
familiar with the project codebase, and that your code follows the below guidelines:

- Code should be as clean and minimal as possible while being maintainable and extensible.
- Code should follow the project's naming conventions to ensure consistency.
- Code should follow the project's style guide by referencing existing code in the codebase.

To submit code to the project, you can:

- Fork the project on GitHub
- Clone your fork locally
- Make your changes/improvements locally
- Ensure any changes are tested without impacts
- Commit your changes and new pull request

### Code Quality

We place strong emphasis on code quality, so submitted code should meet the following requirements:

- Code should be thoroughly tested to ensure correctness and stability
- Code should follow good design principles and best practices
- Code should align with the requirements of your submitted contribution as much as possible

### Commit Messages

Before submitting code, ensure you provide meaningful and detailed commit messages. This helps us better understand your
contribution and merge it more quickly.

Commit messages should include:

- Purpose/reason for the code contribution
- What the code contribution includes/changes
- Optional: How to test the code contribution/results

Messages should be clear and follow conventions set in the project codebase.

### Issue Reporting

If you encounter issues or bugs in the project, feel free to submit issue reports. Before reporting, ensure you have
fully investigated and tested the issue, and provide:

- Description of observed behavior
- Context/conditions of when issue occurs
- Relevant background/contextual information
- Description of expected behavior
- Optional: Screenshots or error output

Reports should be clear and follow conventions set in the project codebase.

### Feature Requests

If you want to suggest adding new features or capabilities, feel free to submit feature requests. Before submitting,
ensure you understand the project history/status, and provide:

- Description of suggested feature/capability
- Purpose/intent of the feature
- Optional: Suggested implementation approach(es)

Requests should be clear and follow conventions set in the project codebase.

### Thank You for Your Contribution

Lastly, thank you for contributing to this project. We welcome all forms of contribution including but not limited to
code contribution, issue reporting, feature requests, documentation writing and more. We believe with your help, this
project will become more robust and powerful.

### Contributors

Thanks to the following individuals for contributing to the project:

- [zhiyi](https://github.com/vyloy)
- [jianti](https://github.com/FH0)