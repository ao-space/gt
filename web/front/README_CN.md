# GT-Admin

[English](README.md) | 简体中文

## 介绍 📖

**GT-Admin** 是一个为 **[GT](https://github.com/ao-space/gt)** 用户专门开发的 Web
界面。它采用了 [Geeker-Admin](https://github.com/HalseySpicy/Geeker-Admin)
的模板，并基于[gin 框架](https://github.com/gin-gonic/gin) 构建。这个界面不仅允许用户在视觉上更直观地进行各种配置操作，还为用户提供了对系统状态的
**可视化监测功能。**

## 目录

- [项目功能](#项目功能)
- [项目结构](#项目结构)
- [安装步骤](#安装步骤)
- [Web端界面介绍](#web端界面介绍)
  - [登录](#登录)
  - [初次引导](#初次引导)
  - [系统控制](#系统控制)
  - [系统状态监测](#系统状态监测)
  - [连接状态查看](#连接状态查看)
  - [配置界面](#配置界面)
  - [pprof界面](#pprof界面)
- [Web端使用教程](#web端使用教程)
  - [基础使用步骤](#基础使用步骤)
  - [示例](#示例)
    - [更改Web用户设置](#更改web用户设置)
    - [HTTP 内网穿透](#http-内网穿透)
    - [HTTPS 内网穿透](#https-内网穿透)
    - [HTTPS SNI 内网穿透](#https-sni-内网穿透)
    - [TLS 加密客户端服务器之间的通信](#tls-加密客户端服务器之间的通信)
    - [TCP 内网穿透](#tcp-内网穿透)
    - [客户端同时开启多个服务](#客户端同时开启多个服务)
- [Web配置说明](#Web配置说明)
- [前端开发者设置](#前端开发者设置)

## 项目功能

- 将 Web 资源集成到单一二进制文件中，消除了额外部署的需求
- 实时系统健康监控，包括操作系统、CPU、内存和硬盘使用情况
- 提供详细的连接状态信息，包括连接池和外部连接
- 直观的配置管理界面，用于查看、修改和保存设置
- 内置 pprof 功能，用于高级性能分析

## 项目结构

![Architecture](https://github.com/ao-space/gt/assets/134463404/1cdbbebf-e890-4e13-a742-ada50a23ca92)

## 安装步骤

<details>
    <summary>详细步骤</summary>

1. 获取项目
      ```shell
      git clone https://github.com/ao-space/gt.git
      ```
2. 构建项目
   - 此处的编译会生成两个可执行文件，分别是 gt-server 与 gt-client 的可执行文件
     - 如需单独编译 gt-server 或 gt-client 可以使用 `make release_server` 或 `make release_client` 命令
     - 这些可执行文件会被保存在 `release` 文件夹中
     ```shell
     cd gt
     make release # The compiled files will be in the release folder
     ``` 
3. 启动服务
    ```shell
    cd release
    # Choose the following command you need, if you don't specify the config file, the default config file path will be used.
    # if you don't know how to choose, the upper one is recommended.
    ./linux-amd64-server  # start gt-server, default config file is server.yaml, which is located in the same directory as the executable file
    ./linux-amd64-client  # start gt-client, default config file is client.yaml, which is located in the same directory as the executable file
   
    # Replace the content within the square brackets with your actual content
    ./linux-amd64-server -webAddr [webAddr] -config [path/to/server.yaml] -webCertFile [path/to/certFile] -webKeyFile [path/to/keyfile] # Start gt-server
    ./linux-amd64-client -webAddr [webAddr] -config [path/to/server.yaml] -webCertFile [path/to/certFile] -webKeyFile [path/to/keyFile] # Start gt-client
   ```
   - Web相关的命令行配置：
     - `webAddr`:
       - 作用：设定Web服务地址。
       - 默认：只在用户零配置启动（无命令行参数）时生效。
         - `gt-server`：`127.0.0.1:8000`
         - `gt-client`：`127.0.0.1:7000`
       - 说明：如零配置启动时，默认端口被占用，会自动选取其他有效端口。其余情况若不设置`webAddr`，GT-Admin的Web服务将不会启动。若用户明确指定了地址，系统不会尝试其他端口。

     - `config`:   
         - 作用：指定配置文件的存储路径。   
         - 默认：   
           - `gt-server`：与可执行文件同目录的`server.yaml`   
           - `gt-client`：与可执行文件同目录的`client.yaml`   
   
     - `webCertFile` / `webKeyFile`:   
       - 作用：用于启动HTTPS服务。   
       - 默认：两者为空，不开启HTTPS服务。   
       - 说明：用户可以设置为"auto"使用自签发TLS，或指定路径使用用户提供的证书。

4. 清理(可选)

- 使用 `make clean` 命令将移除所有生成的文件。
- 执行 `make clean_web` 将移除所有与 web 相关的生成文件（如 node_modules、dist 等）。注意，执行此步骤会同时移除必要的依赖包，因此后续构建将需要重新安装依赖。
- 执行 `make clean_dist` 将移除所有生成的 dist 文件夹。该操作适用于 release 命令执行完成后，因为所有必要的静态文件已经被集成到二进制文件中。

</details>

## Web端界面介绍

<details>
     <summary>详细介绍</summary>

### 登录

- 根据配置文件中设置的 `admin` 和 `password` 完成登录操作。初次使用时，系统会自动生成一个 **tempKey**，用于后续的身份验证。
  ![Login](https://github.com/ao-space/gt/assets/134463404/8d543e1f-6af3-4e6f-b726-7215a5f7a04c)

### 初次引导

- 首次登录后，系统将介绍如何使用并引导你进行基本设置。系统会**自动生成**初始的 Web 用户名和密码，建议你**手动更改**这些信息以确保安全。
  ![Guide](https://github.com/ao-space/gt/assets/134463404/cb106ab4-f6bd-44a7-bf19-1fbb3f7e6af7)

### 系统控制

- 点击右上角的 **“GT-Admin”** 后有工具栏

  - **User Setting**：在这里你可以更改 Web 用户名和密码，以及选择是否启用 pprof 服务。
  - **Log Out**：清理用户信息并退出
  - **Restart System**：重启整个服务器

  **Note**: 下面的操作都会关闭整个系统，后续还需服务需要手动启动，**请谨慎操作**

  - **Shutdown System**：关闭系统
  - **Terminate System**： 中断系统

  ![ToolBar](https://github.com/ao-space/gt/assets/134463404/b0bbe91f-1351-4a31-878a-3a24906b0bc8)
  ![UserSetting](https://github.com/ao-space/gt/assets/134463404/19d509ff-cb68-49a0-b2cd-10830929493a)

### 系统状态监测

- 提供系统信息、DISK、CPU、RAM 信息查看
  ![DashBoard](https://github.com/ao-space/gt/assets/134463404/61e72873-7ba1-4ddf-b408-0f7597b4c336)

### 连接状态查看

- Server 端提供连接**信息**查看
  ![ServerConnection](https://github.com/ao-space/gt/assets/134463404/e4cec0dd-0e3d-4c54-9faf-08d64f2398ff)
- Client 端提供连接**状态**查看
  ![ClientConnection](https://github.com/ao-space/gt/assets/134463404/dfb3eaf6-5090-435c-9a0c-ebc464023447)

### 配置界面

- 通用配置修改流程

  1. 初始进入时会提示是否载入配置文件中的信息。
  2. 用户根据自身需要配置有关设置，有关设置的详细信息可以在 **"?"** 处查看。侧边的导航栏帮助用户高效跳转到相关内容上。
  3. 用户配置完成后点击 **Submit** 按钮，将有关的配置信息重写进配置文件中，（若一开始未指定配置文件，则会保存在与gt-server(
     client)的编译文件同处一个文件夹）
  4. 用户可以多次进行修改有关配置并进行保存。

  - 基本操作（操作栏在最后，可点击侧边最后内容进行跳转）：
    - **GetFromFile**：获取配置文件中的内容，若启动时未使用`-config` （即没有设置配置文件)，系统将使用默认配置文件路径。若配置文件不存在,那么会执行
      **GetFromRunning**的操作 （**推荐直接启动或者只使用命令行来配置`-config`**，因为后续的 Restart
      操作都是重新运行初始的的启动命令，只有初始启用时，没有配置除了配置文件之外的其他配置，才可保证后续操作符合用户预期,因为命令行的优先级更高）
    - **GetFromRunning**：获取正在运行着的配置信息。

- gt-server端启用配置

  - **用户保存修改后**，可以通过**Restart System** (位于系统控制栏中) 来进行新配置的启用（该操作会启用一个新的进程）。
  - **General Setting** 处设置的**TCP Setting** 与 **Host Setting** 均是 **全局**设置，**精细化**设置请在下面的**User
    Setting**处设置
    ![ServerConfig](https://github.com/ao-space/gt/assets/134463404/c6283fde-ce51-42a0-8bdb-025cece3de34)

- gt-client端启用配置

  - **用户保存修改后**，可以使用**Reload Services**来保持原有进程的同时，重启 Services 服务（前提是只更改了
    Services），但是如果更改了 Options 字段的内容（即非 Services 部分内容），则要启用该配置服务就只能通过**Restart System**
    来重启整个进程来实现配置的更改。
    ![ClientConfig](https://github.com/ao-space/gt/assets/134463404/a22d0a72-56c7-49be-82f2-a7c5420f127a)

### pprof界面

- 性能检测界面
  ![pprof](https://github.com/ao-space/gt/assets/134463404/0240a223-9476-49ac-bede-fedc239401b2)

</details>

## Web端使用教程

### 基础使用步骤

#### 1. 启用服务

- 执行 `./linux-amd64-server` 或 `./linux-amd64-client` 命令，将自动打开对应的Web界面
  - 服务端默认Web地址：`127.0.0.1:8000`
  - 客户端默认Web地址：`127.0.0.1:7000`
- 若需后续访问，请直接在浏览器中输入对应的Web地址

#### 2. 登录系统（初次登录无此步骤）

- 初次登录时，系统会随机为用户分配Web登录的用户名和密码，并附带token以绕过登录
- 输入配置文件中的`Admin`作为`Username`，并输入对应的`Password`进行登录。

#### 3. 用户设置

- 登录后，点击页面右上角的 **“GT-Admin”**，将展开系统控制栏，在系统控制栏中，选择**User Setting**进行用户设置。
  - 可设置的内容包括：下次登录的`Username`和`Password`，以及是否启用`pprof`性能检测功能。
- **强烈建议**：初次登录后，立即进行用户设置，以便下次登录。
- 若半小时内未进行相关设置，系统分配的登录信息将失效。此时，需重新启动服务以获取新的登录信息。

#### 4. 配置GT项目

- 若仅配置了GT项目但未修改User Setting，系统将在保存GT配置时同时保存当前的Web设置。
- 用户需记住此时Web用户设置（在User Setting中查看），或在下次登录时查看配置文件获取Web登录信息。配置文件默认位置与可执行文件相同。

#### 5. 保存并启用配置

- 完成配置后，点击配置界面的`Submit`按钮或User Setting的`Change`按钮保存配置。
- 在系统控制栏中点击`Restart System`重启系统，新的配置将在重启后生效。
  - Web配置逻辑：用户的配置信息将被保存在配置文件中，系统在下次重启时会加载这些配置。
  - 对于Client，若仅更改了Service部分配置，可在配置界面下方点击`Reload Services`在当前进程中更新配置。

#### 6.注意事项

- 用户一次登录的有效期是30分钟

### 示例

#### 更改Web用户设置

<details>
    <summary>详细步骤</summary>

1. 进入主界面后，点击页面右上角的 **“GT-Admin”** 后，出现系统控制栏
2. 点击`User Setting`后就出现设置用户信息的有关内容
   - 配置信息介绍：
     - Username 与 Password：即用户下次用来登录的账号设置
     - Enablepprof：是否启用pprof的性能检测功能
3. 用户进行相关配置后，点击`Change`按钮，即可将有关配置写入配置文件中
   ![Web User Setting](https://github.com/ao-space/gt/assets/134463404/decb7cae-f022-4c54-ad2c-c1881bda7306)

</details>

#### HTTP 内网穿透

<details>
    <summary>详细步骤</summary>

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:8080
  来访问内网服务器上 80 端口服务的网页。

1. 配置服务端（公网服务器）
   - 配置NetWork Setting： 设置Addr: 8080
   - 配置User Setting：设置 ID：id1，Secret： secret1
     ![HTTP Server](https://github.com/ao-space/gt/assets/134463404/b3d8b5a8-479b-44fa-bab4-5cefbff35832)

2. 配置客户端（内网服务器）
   - 配置General Setting： ID： id1， Secret：secret1，Remote ：tcp://id1.example.com:8080
   - 配置Service Setting：LocalURL: http://127.0.0.1:80
     ![HTTP Client](https://github.com/ao-space/gt/assets/134463404/89ca0b20-5dcf-46d2-a899-81094eee3b81)

</details>

#### HTTPS 内网穿透

<details>
    <summary>详细步骤</summary>

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 <https://id1.example.com>
  来访问内网服务器上 80 端口提供的 HTTP 网页。

1. 配置服务端（公网服务器）
   - 配置NetWork Setting： 设置 TLSAddr: 443，
   - 配置Security Setting: 设置 CertFile：/root/openssl_crt/tls.crt , KeyFile: /root/openssl_crt/tls.key
   - 配置User Setting： 设置 ID：id1，Secret： secret1
     ![HTTPS Server](https://github.com/ao-space/gt/assets/134463404/33f3e296-140c-4124-9626-1900dc28b369)

2. 配置客户端（内网服务器），因为使用了自签名证书，所以使用了 `remoteCertInsecure` 选项，其它情况禁止使用此选项（中间人攻击导致加密内容被解密）
   - 配置General Setting： ID：id1, Secret: secret1, Remote:  tls://id1.example.com , RemoteCertInsecure: true
   - 配置Service Setting:  LocalURL: http://127.0.0.1
     ![HTTPS Client](https://github.com/ao-space/gt/assets/134463404/f906237c-aea2-4127-ac8e-12a8764ca85b)

</details>

#### HTTPS SNI 内网穿透

<details>
    <summary>详细步骤</summary>

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 <https://id1.example.com>
  来访问内网服务器上 443 端口提供的 HTTPS 网页。

1. 配置服务端（公网服务器）
   - 配置NetWork Setting： 设置 Addr: 8080， SNIAdr: 443
   - 配置User Setting： 设置 ID：id1，Secret： secret1
     ![SNI Server](https://github.com/ao-space/gt/assets/134463404/b015d244-b5d8-42a5-9c99-9dbe6d5212c6)

2. 配置客户端（内网服务器）
   - 配置General Setting： ID：id1, Secret: secret1, Remote:  tcp://id1.example.com:8080
   - 配置Service Setting:  LocalURL: https://127.0.0.1
     ![SNI_Client](https://github.com/ao-space/gt/assets/134463404/4583323e-e2e3-443b-91ac-c3722d43438b)

</details>

#### TLS 加密客户端服务器之间的通信

<details>
    <summary>详细步骤</summary>

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:8080
  来访问内网服务器上 80 端口服务的网页。同时用 TLS 加密客户端与服务端之间的通信。

1. 配置服务端（公网服务器）
   - 配置NetWork Setting： 设置 Addr: 8080，TLSAdr: 443
   - 配置Security Setting: 设置 CertFile：/root/openssl_crt/tls.crt , KeyFile: /root/openssl_crt/tls.key
   - 配置User Setting： 设置 ID：id1，Secret： secret1
     ![TLS Server](https://github.com/ao-space/gt/assets/134463404/bb9121be-6e5e-49ec-be6e-0766c3e58f74)

2. 配置客户端（内网服务器），因为使用了自签名证书，所以使用了 `remoteCertInsecure` 选项，其它情况禁止使用此选项（中间人攻击导致加密内容被解密）
   - 配置General Setting： ID：id1, Secret: secret1, Remote:  tls://id1.example.com，RemoteCertInsecure: true
   - 配置Service Setting:  LocalURL: http://127.0.0.1:80
     ![TLS Client](https://github.com/ao-space/gt/assets/134463404/1bb27531-e92e-4250-8a78-f7b7c342410c)

</details>

#### TCP 内网穿透

<details>
    <summary>详细步骤</summary>

- 需求：有一台内网服务器和一台公网服务器，id1.example.com 解析到公网服务器的地址。希望通过访问 id1.example.com:2222
  来访问内网服务器上 22 端口上的 SSH 服务，如果服务端 2222 端口不可以，则由服务端选择一个随机端口。

1. 配置服务端（公网服务器）
   - 配置NetWork Setting： 设置 Addr: 8080
   - 配置User Setting： 设置 ID：id1，Secret： secret1，TCPNumber：1，TCPRanges：1024-65535
     ![TCP Server](https://github.com/ao-space/gt/assets/134463404/7f4da122-9e41-42a8-9f2c-b2dee98d96fe)

2. 配置客户端（内网服务器）
   - 配置General Setting： ID：id1, Secret: secret1, Remote: tcp://id1.example.com:8080，
   - 配置Service Setting：LocalURL：tcp://127.0.0.1:22， RemoteTCPPort：2222， RemoteTCPRandom： true
     ![TCP Client](https://github.com/ao-space/gt/assets/134463404/889ff532-443f-4feb-a6d5-c8e698bd29ee)

</details>

#### 客户端同时开启多个服务

<details>
    <summary>详细步骤</summary>

- 需求：有一台内网服务器和一台公网服务器，id1-1.example.com 和 id1-2.example.com 解析到公网服务器的地址。希望通过访问
  id1-1.example.com:8080 来访问内网服务器上 80 端口上的服务，希望通过访问 id1-2.example.com:8080 来访问内网服务器上
  8080端口上的服务，希望通过访问 id1-1.example.com:2222 来访问内网服务器上 2222 端口上的服务，希望通过访问
  id1-1.example.com:2223 来访问内网服务器上 2223 端口上的服务。同时服务端限制客户端的 hostPrefix 只能由纯数字或纯字母组成。

1. 配置服务端（公网服务器）
   - 配置NetWork Setting： 设置 Addr: 8080
   - 配置User Setting：设置 ID：id1，Secret：
     secret1，TCPNumber：2，TCPRanges：1024-65535，HostNumber：2，WithID：true，HostRegex：`^[0-9]+$`、 `^[a-zA-Z]+$`
     ![Multiple Server](https://github.com/ao-space/gt/assets/134463404/7d4dd44c-5d9b-4d8e-a35f-88743fbd68b2)

2. 配置客户端（内网服务器）
   - 配置General Setting： ID：id1, Secret: secret1, Remote: tcp://id1.example.com:8080，
   - 配置Service 1 Setting：HostPrefix:1，LocalURL：http://127.0.0.1:80， UseLocalAdHTTPHost：true
   - 配置Service 2 Setting：HostPrefix:2，LocalURL：http://127.0.0.1:8080， UseLocalAdHTTPHost：true
   - 配置Service 3 Setting：LocalURL：tcp://127.0.0.1:2222， RemoteTCPPort: 2222
   - 配置Service 4 Setting：LocalURL：tcp://127.0.0.1:2223， RemoteTCPPort: 2223
     ![Multiple Client](https://github.com/ao-space/gt/assets/134463404/7130769c-fd06-4012-8b22-ceb5586999a7)

</details>

## Web配置说明

- 所有的Web配置均一同写在与gt-server(gt-client)的同一个配置文件中
- 默认设置
  - gt-server 的 Web 地址默认为 127.0.0.1:8000
  - gt-client 的 Web 地址默认为 127.0.0.1:7000

<details>
   <summary>example</summary>

   ```yaml
   #example
   options:
     # Web Setting (Optional)
     # Web Address 
     webAddr: localhost:8000
     
     # HTTPS Settings 
     # If both are set to "auto", a self-signed certificate will be used
     webCertFile: path/to/certFile # Path to your certificate file
     webKeyFile: path/to/keyFile   # Path to your private key file
     
     # JWT Settings
     # Used to sign the JWT token. If not set, it will be generated automatically.
     # Token Validity Period: 6 hours
     signingKey: signature

     # Admin Credentials
     # Used for logging into the web page
     admin: username        # Admin username
     password: password      # Admin password

     # Performance Profiling 
     # Requires 'webAddr' to be set to a valid value
     pprof: true 
   ```

</details>

## 前端开发者设置

- src/api/modules/login.ts 的 getAuthMenuListApi 中注释上面的代码，可以实现无需开启后端得到路由权限
  ![code1](https://github.com/ao-space/gt/assets/134463404/56eec78a-1e6e-4018-9231-2c3b0529c777)

- src/routers/index.ts 的 router.beforeEach 函数中可以注释这行代码来实现绕过用户登录，来进行其他界面的跳转以及测试
  ![code2](https://github.com/ao-space/gt/assets/134463404/8f74e1a3-5893-4601-afdc-9603b6521308)

  进行上述操作后，即可实现在前端界面开发时，无需开启后端。

- 启动前端服务
  - a. 更改 proxy 设置（先检查下述文件，更改**PROXY**设置为对应的 web 后端 url，与 yaml
    配置文件中保持一致，此处与上面的example保持一致,故设为8000）
    ```ts
    //.env.development
    VITE_PROXY = [["/api", "http://localhost:8000"]];
    ```
  - b. 启动 web 服务
    ```shell
    npm run dev
    ```
