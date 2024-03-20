[简体中文](./README.md) | English

<h1 align="center">GT</h1>

<div align="center">

A fast WebSocket(s)/HTTP(s)/TCP relay proxy with WebRTC P2P supports.

[![GitHub (pre-)release](https://img.shields.io/github/release/ao-space/gt/all.svg)](https://github.com/ao-space/gt/releases) [![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/ao-space/gt/.github%2Fworkflows%2Fcontainer.yml)](https://github.com/ao-space/gt/actions) [![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/ao-space/gt/total)](https://github.com/ao-space/gt/releases) [![GitHub issues](https://img.shields.io/github/issues/ao-space/gt.svg)](https://github.com/ao-space/gt/issues) [![GitHub closed issues](https://img.shields.io/github/issues-closed/ao-space/gt.svg)](https://github.com/ao-space/gt/issues?q=is%3Aissue+is%3Aclosed) [![GitHub](https://img.shields.io/github/license/ao-space/gt.svg)](./LICENSE)

</div>

**Key Design Features:**

- **Stability**
  - Highly available, can upgrade versions on the fly without worrying about connection loss, service interruptions, or downtime.
  - During reload, forwarding is maintained successfully. During stress testing, reload resulted in 0 errors.
  - Process supervision ensures that worker processes automatically restart if they crash.

- **Performance**
  - Aiming for higher performance while maintaining cross-platform compatibility, employing more efficient technical solutions.
  - Minimize memory allocation to ease the burden on the garbage collector: use resource pools; instantiate a Value in LoadOrStore only during the Store operation.
  - Minimize memory copying: Reader uses Peek and Discard instead of Read.
  - Avoid system calls: Virtual Listener and Conn forward request data to in-process API services.
  - Utilize appropriate concurrency techniques for different concurrency scenarios.

- **Usability**
  - Supports web-based configuration management.
  - Zero-parameter startup initiates web configuration setup.
  - Supports loading configuration file directories, allowing simultaneous startup of multiple servers and clients.
  - Clients can point to multiple services.
  - Server supports multi-user functionality.
  - Clients intelligently choose the communication protocol with the server based on network conditions.

- **Privacy Protection**
  - The server's port reuse feature is based on the characteristic position of the protocol target. For instance, when forwarding HTTP protocol at the application layer, it is based on the TCP data stream, targeting only the first packet's HTTP protocol header for forwarding, then directly forwarding subsequent data.
  - Does not log sensitive information.
  - Supports HTTPS SNI for end-to-end encryption forwarding.

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

## Download

### Github

Choose the appropriate version to download from [https://github.com/ao-space/gt/releases](https://github.com/ao-space/gt/releases).

### Docker Container

More container image information can be found at [https://github.com/ao-space/gt/pkgs/container/gt](https://github.com/ao-space/gt/pkgs/container/gt).

```shell
docker pull ghcr.io/ao-space/gt:server-dev

docker pull ghcr.io/ao-space/gt:client-dev
```

## Usage

```shell
gt --help
Fast WebSocket(s)/HTTP(s)/TCP relay proxy with WebRTC P2P supports.

Usage: gt [OPTIONS] [COMMAND]

Commands:
  server  Run GT Server
  client  Run GT Client
  help    Print this message or the help of the given subcommand(s)

Options:
  -c, --config <CONFIG>
          Path to the config file or the directory containing the config files

  -s, --signal <SIGNAL>
          Send signal to the running GT processes

          Possible values:
          - reload:  Send reload signal
          - restart: Send restart signal
          - stop:    Send stop signal

  -h, --help
          Print help (see a summary with '-h')

  -V, --version
          Print version
```

## Configuration File

Configuration files can be edited and generated through the web management backend.

### Server

Run with default configuration, after which you can obtain the web management backend address from the logs, open it with a browser, and edit the configuration items:

```shell
gt server
```

Run with a specified configuration file:

```shell
gt server -c ./config.yml
```

### Client

Run with default configuration, after which you can obtain the web management backend address from the logs, open it with a browser, and edit the configuration items:

```shell
gt client
```

Run with a specified configuration file:

```shell
gt client -c ./config.yml
```

### Batch Startup

Batch startup by specifying the configuration file directory:

```shell
gt -c ./conf.d
```

## Performance Testing

### Group 1 (MacOS Environment + Nginx Test)

Stress test comparison between this project and frp using wrk, with the internal network service pointing to the local nginx test page, results as follows:

```text
Model Name: MacBook Pro
Model Identifier: MacBookPro17,1
Chip: Apple M1
Total Number of Cores: 8 (4 performance and 4 efficiency)
Memory: 16 GB
```

#### GT Benchmark

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

#### frp dev branch 42745a3

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

### Group 2 (Ubuntu Environment + Nginx Test)

Stress test comparison between this project and frp using wrk, with the internal network service pointing to the local nginx test page, results as follows:

```text
System: Ubuntu 22.04
Chip: Intel i9-12900
Total Number of Cores: 16 (8 performance and 8 efficiency)
Memory: 32 GB
```

#### GT-TCP

```shell
$ ./release/linux-amd64-server -addr 12080 -id id1 -secret secret1
$ ./release/linux-amd64-client -local http://127.0.0.1:80 -remote tcp://id1.example.com:12080 -id id1 -secret secret1

$ wrk -c 100 -d 30s -t 10 http://id1.example.com:12080
Running 30s test @ http://id1.example.com:12080
  10 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   558.51us    2.05ms  71.54ms   99.03%
    Req/Sec    24.29k     2.28k   49.07k    95.74%
  7264421 requests in 30.10s, 5.81GB read
Requests/sec: 241344.46
Transfer/sec:    197.70MB
```

#### GT-QUIC

```shell
$ ./release/linux-amd64-server -addr 12080 -quicAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
$ ./release/linux-amd64-client -local http://127.0.0.1:80 -remote quic://id1.example.com:443 -remoteCertInsecure -id id1 -secret secret1

$ wrk -c 100 -d 30s -t 10 http://id1.example.com:12080
Running 30s test @ http://id1.example.com:12080
  10 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   826.65us    1.14ms  66.29ms   98.68%
    Req/Sec    12.91k     1.36k   23.53k    79.43%
  3864241 requests in 30.10s, 3.09GB read
Requests/sec: 128380.49
Transfer/sec:    105.16MB
```

#### frp v0.52.1

```shell
$ ./frps -c ./frps.toml
$ ./frpc -c ./frpc.toml

$ wrk -c 100 -d 30s -t 10 http://id1.example.com:12080/
Running 30s test @ http://id1.example.com:12080/
  10 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.49ms    8.27ms 154.62ms   92.43%
    Req/Sec     4.02k     2.08k    7.51k    53.21%
  1203236 requests in 30.08s, 0.93GB read
Requests/sec:  40003.03
Transfer/sec:     31.82MB
```

### Group 3 (Ubuntu Environment + Short Request Test)

Stress test comparison between this project and frp using wrk, where each request only returns a field response of less than 10 bytes, simulating HTTP short requests, results as follows:

#### GT-TCP

```shell
$ ./release/linux-amd64-server -addr 12080 -id id1 -secret secret1
$ ./release/linux-amd64-client -local http://127.0.0.1:80 -remote tcp://id1.example.com:12080 -id id1 -secret secret1

$ wrk -c 100 -d 30s -t 10 http://id1.example.com:12080/
Running 30s test @ http://id1.example.com:12080/
  10 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.55ms   13.48ms 220.23ms   95.31%
    Req/Sec     5.23k     2.11k   12.40k    76.10%
  1557980 requests in 30.06s, 191.67MB read
Requests/sec:  51822.69
Transfer/sec:      6.38MB
```

#### GT-QUIC

```shell
$ ./release/linux-amd64-server -addr 12080 -quicAddr 443 -certFile /root/openssl_crt/tls.crt -keyFile /root/openssl_crt/tls.key -id id1 -secret secret1
$ ./release/linux-amd64-client -local http://127.0.0.1:80 -remote quic://id1.example.com:443 -remoteCertInsecure -id id1 -secret secret1

$ wrk -c 100 -d 30s -t 10 http://id1.example.com:12080/
Running 30s test @ http://id1.example.com:12080/
  10 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.84ms    6.75ms 168.93ms   98.47%
    Req/Sec     9.33k     2.13k   22.86k    78.54%
  2787908 requests in 30.10s, 342.98MB read
Requests/sec:  92622.63
Transfer/sec:     11.39MB
```

#### frp v0.52.1

```shell
$ ./frps -c ./frps.toml
$ ./frpc -c ./frpc.toml

$ wrk -c 100 -d 30s -t 10 http://id1.example.com:12080/
Running 30s test @ http://id1.example.com:12080/
  10 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     2.95ms    3.74ms 136.09ms   91.10%
    Req/Sec     4.16k     1.22k   12.86k    87.85%
  1243103 requests in 30.07s, 152.93MB read
Requests/sec:  41334.52
Transfer/sec:      5.09MB
```

## Community

Welcome to join the [Slack](https://slack.ao.space) channel for discussions.

## Contributors

Thank you to the following developers for their contributions to the project:

- [zhiyi](https://github.com/vyloy)
- [jianti](https://github.com/FH0)
- [huwf5](https://github.com/huwf5)
- [AdachiAndShimamura](https://github.com/AdachiAndShimamura)
- [DrakenLibra](https://github.com/DrakenLibra)
