# GT

[English](libcs/README.md) | 简体中文

GT 是一个高性能的 WebSocket(s)/HTTP(s)/TCP 中转代理。

具有以下设计特点：
- 多进程+多线程架构
  - reload 过程中，保持转发可以成功。压力测试过程中 reload，测试结果 0 error
  - 进程守护，工作进程崩溃自动重启
  - 加载配置文件目录，可以同时启动多个服务端，客户端
- 注重性能，在保持跨平台和功能稳定的前提下，会尝试采用性能更高的技术方案
  - 减少内存分配来减轻 GC 负担：资源池，LoadOrStore Store 时才创建 Value
  - 减少内存复制：Reader 使用 Peek 和 Discard 取代 Read
  - 避免系统调用：Virtual Listener、Conn 将请求数据转发到进程内 API 服务
  - 根据不同的并发场景，使用适当的并发技术
- 注重易用性
  - 客户端支持指向多个服务
  - 支持 yaml 配置文件和 Web 配置管理
  - 服务端支持多用户功能
  - 零参数配置启动，进入 Web 配置初始化
  - 客户端自动根据网络状况，智能选择与服务端的通信协议
- 注重隐私保护
  - 服务端的端口复用功能，是基于协议目标的特征位置来实现的。例如：应用层 HTTP 协议转发，基于 TCP 数据流，只定位获取第一个数据包的 HTTP 协议头的转发目标，然后将后续数据直接转发
  - 不打印敏感信息到日志
  - 支持 HTTPS SNI 端到端加密转发
- 基于 WebRTC 实现的点对点连接功能，支持所有支持 WebRTC 的平台，例如：iOS，Android，浏览器等。

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

## 下载

### Github

从 <https://github.com/ao-space/gt/releases> 选择合适的版本下载。

### Docker 容器

更多容器镜像信息可以从 <https://github.com/ao-space/gt/pkgs/container/gt> 获取。

```shell
docker pull ghcr.io/ao-space/gt:server-dev

docker pull ghcr.io/ao-space/gt:client-dev
```

## 用法

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
          Path to the config file or the directory contains the config files

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

### 配置文件

配置文件可以通过 Web 管理后台编辑生成，推荐直接使用 Web 管理后台编辑。

### 服务端

使用默认配置运行，运行后可从日志获取 Web 管理后台地址，用浏览器打开后，进行配置项编辑：

```shell
gt server
```

通过指定配置文件运行：

```shell
gt server -c ./config.yml
```

### 客户端

使用默认配置运行，运行后可从日志获取 Web 管理后台地址，用浏览器打开后，进行配置项编辑：

```shell
gt client
```

通过指定配置文件运行：

```shell
gt client -c ./config.yml
```

### 批量启动

通过指定配置文件目录批量启动：

```shell
gt -c ./conf.d
```

## 性能测试

### 第一组（MacOS环境+nginx测试）

通过 wrk 进行压力测试本项目与 frp 进行对比，内网服务指向在本地运行 nginx 的测试页面，测试结果如下：

```text
Model Name: MacBook Pro
Model Identifier: MacBookPro17,1
Chip: Apple M1
Total Number of Cores: 8 (4 performance and 4 efficiency)
Memory: 16 GB
```

#### GT benchmark

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

### 第二组（Ubuntu环境+nginx测试）

通过 wrk 进行压力测试本项目与 frp 进行对比，内网服务指向在本地运行 nginx 的测试页面，测试结果如下：

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

### 第三组(Ubuntu环境+short request测试)

通过 wrk 进行压力测试本项目与 frp 进行对比，每次请求只会返回小于10字节的字段回复，用于模拟HTTP short request，测试结果如下：

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

## 贡献者

感谢以下人员为项目做出的贡献：

- [zhiyi](https://github.com/vyloy)
- [jianti](https://github.com/FH0)
