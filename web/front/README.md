# GT-Admin

English | [简体中文](README_CN.md)

## Introduction 📖

**GT-Admin** is a Web interface specifically developed for **[GT](https://github.com/ao-space/gt)** users. It adopts the
template of [Geeker-Admin](https://github.com/HalseySpicy/Geeker-Admin) and is built upon
the [gin framework](https://github.com/gin-gonic/gin). This interface not only allows users to perform various
configuration operations more intuitively visually, but also provides users with **visualization monitoring capabilities** for system status.

## Table of Contents

- [Features](#features)
- [Project Structure](#project-structure)
- [Installation Steps](#installation-steps)
- [Web Interface Introduction](#web-interface-introduction)
  - [Login](#login)
  - [Initial Guidance](#initial-guidance)
  - [System Control](#system-control)
  - [System Status Monitoring](#system-status-monitoring)
  - [Connection Status Overview](#connection-status-overview)
  - [Configuration Interface](#configuration-interface)
  - [pprof Interface](#pprof-interface)
- [Web Usage Tutorial](#web-usage-tutorial)
  - [Basic Usage Steps](#basic-usage-steps)
  - [Example](#example)
    - [Change Web User Settings](#change-web-user-settings)
    - [HTTP Internal Penetration](#http-internal-penetration)
    - [HTTPS Internal Penetration](#https-internal-penetration)
    - [HTTPS SNI Internal Penetration](#https-sni-internal-penetration)
    - [Encrypt Client-Server Communication with TLS](#encrypt-client-server-communication-with-tls)
    - [TCP Internal Penetration](#tcp-internal-penetration)
    - [Client Start Multiple Services Simultaneously](#client-start-multiple-services-simultaneously)
- [Web Configuration Instructions](#web-configuration-instructions)
- [Frontend Developer Settings](#frontend-developer-settings)

## Features

- **Integrates Web resources into a single binary file**, eliminating the need for additional deployments
- **Offers real-time system health monitoring**, including metrics on the operating system, CPU, memory, and disk usage
- **Provides detailed connection status information**, covering both the connection pool and external connections
- Features an **intuitive configuration management interface** for viewing, modifying, and saving settings
- Incorporates built-in **pprof** for advanced performance analysis

## Project Structure

![Architecture](https://github.com/ao-space/gt/assets/134463404/1cdbbebf-e890-4e13-a742-ada50a23ca92)

## Installation Steps

<details>
    <summary>Detailed Steps</summary>

1. Clone the Project
    ```shell
    git clone https://github.com/ao-space/gt.git
    ```
2. Build the Project
   - Compiling here produces two executable files: gt-server and gt-client.
     - To compile gt-server or gt-client individually, you can use `make release_server` or `make release_client`
       commands.
     - These executables are saved in the `release` folder.

     ```shell
     cd gt
     make release # The compiled files will be in the release folder
     ```
     
3. Start the Service
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
   - Web-related Command Line Configuration:
     - `webAddr`:
       - Purpose: Set the Web service address.
       - Default: Effective only when the user launches with zero configuration (no command-line parameters).
         - gt-server: 127.0.0.1:8000
         - gt-client: 127.0.0.1:7000
       - Note: During a zero-configuration launch, if the default port is occupied, another available port will be automatically selected. In other situations, if webAddr is not set, the GT-Admin Web service will not start. If a specific address is provided by the user, the system won't attempt other ports.
     - `config`:
       - Purpose: Specify the storage path for the configuration file.
       - Default:
         - `gt-server`: `server.yaml` located in the same directory as the executable.
         - `gt-client`: `client.yaml` located in the same directory as the executable.
     - `webCertFile`/`webKeyFile`:
       - Purpose: Used to initiate the HTTPS service.
       - Default: Both are unset, and the HTTPS service is not enabled.
       - Note: Users can set to "auto" to use self-signed TLS or specify paths to use certificates provided by the user.

4. Cleanup (Optional)
   - Running `make clean` will remove all generated files.
   - Running `make clean_web` will remove all web-related generated files (such as node_modules, dist, etc.). Note that
     this will also remove essential dependencies, requiring them to be reinstalled for future builds.
   - Running `make clean_dist` will remove all generated dist folders. This step is recommended after executing the
     release command, as all the required static files would have already been embedded into the binary.

</details>

## Web Interface Introduction

<details>
    <summary>Click to Expand</summary>

### Login

- To log in, use the `admin` and `password` as set in the configuration file. For first-time users, the system will
  automatically generate a **tempKey** to bypass the login process. 
  ![Login](https://github.com/ao-space/gt/assets/134463404/8d543e1f-6af3-4e6f-b726-7215a5f7a04c)

### Initial Guidance

- Upon your first login, the system will guide you through its usage and initial settings. The system will
  **auto-generate** the initial `admin` and `password`; it's recommended to **manually update** this information for
  security reasons.
  ![Guide](https://github.com/ao-space/gt/assets/134463404/cb106ab4-f6bd-44a7-bf19-1fbb3f7e6af7)

### System Control

- Click on **"GT-Admin"** in the upper-right corner of the interface to reveal the toolbar options:
  - **User Setting**: Here, you can modify the Web username and password, as well as decide whether to enable the
    pprof service.
  - **Log Out**: Clear user information and exit the system.
  - **Restart System**: This will restart the entire server.

  **Note**: The following operations will shut down the entire system, and manual restart will be required afterward. *
  *Please proceed with caution**

  - **Shutdown System**: This will shut down the system.
  - **Terminate System**: This will immediately terminate the system.

  ![ToolBar](https://github.com/ao-space/gt/assets/134463404/b0bbe91f-1351-4a31-878a-3a24906b0bc8)
  ![UserSetting](https://github.com/ao-space/gt/assets/134463404/19d509ff-cb68-49a0-b2cd-10830929493a)

### System Status Monitoring

- Provides system information, and views for DISK, CPU, RAM.
  ![DashBoard](https://github.com/ao-space/gt/assets/134463404/61e72873-7ba1-4ddf-b408-0f7597b4c336)

### Connection Status Overview

- The Server side provides a view for connection **information**.
  ![ServerConnection](https://github.com/ao-space/gt/assets/134463404/e4cec0dd-0e3d-4c54-9faf-08d64f2398ff)
- The Client side provides a view for connection **status**.
  ![ClientConnection](https://github.com/ao-space/gt/assets/134463404/dfb3eaf6-5090-435c-9a0c-ebc464023447)

### Configuration Interface

- General configuration modification workflow:

  1. Initially, a prompt will ask if you want to load information from the configuration file.
  2. Users can set configurations according to their needs. Detailed information about each setting can be viewed by
     clicking on the **"?"**. The sidebar navigation aids users in quickly accessing related content.
  3. After configuring, users can click the **Submit** button to overwrite the configuration file with the new
     information. (If no initial configuration file is specified, it will save in the same folder as the gt-server(
     client) compilation file.)
  4. Users can modify and save the configuration multiple times.

  - Basic operations (action bar at the end, click the last content on the sidebar for navigation):
    - **GetFromFile**: Retrieves content from the configuration file. If `-config` is not used during startup (i.e.,
      no configuration file is specified), the system will use the default configuration file path. If the
      configuration file does not exist, the GetFromRunning operation will be executed. **It's recommended to either
      start directly or only use the `-config` option in the command line** because subsequent "Restart" operations
      will
      re-execute the initial startup command. Only when initially started without any configurations other than the
      configuration file can we ensure that subsequent operations meet user expectations, as command line
      configurations have higher priority.

    - **GetFromRunning**: Retrieve configuration from the currently running service.

- Server configuration activation:

  - **After users save their changes**, they can activate the new configuration by using **Restart System** (this
    action
    will start a new process).

  - The **TCP Setting** and **Host Setting** set in the **General Setting** are **global** configurations. For more *
    *detailed** settings, please configure them in the **User Setting** section below.
    ![ServerConfig](https://github.com/ao-space/gt/assets/134463404/c6283fde-ce51-42a0-8bdb-025cece3de34)

- Client configuration activation:

  - **After users save their changes**, they can use **Reload Services** to keep the existing process running while
    restarting the Services service (provided only the Services were changed). However, if content in the Options
    field
    was changed (i.e., parts other than Services), to activate the new configuration, you'll need to use **Restart
    System** to restart the entire process.
    ![ClientConfig](https://github.com/ao-space/gt/assets/134463404/a22d0a72-56c7-49be-82f2-a7c5420f127a)

### pprof Interface

- Performance monitoring interface.
  ![pprof](https://github.com/ao-space/gt/assets/134463404/0240a223-9476-49ac-bede-fedc239401b2)

</details>

## Web Usage Tutorial

### Basic Usage Steps

#### 1. Start the Service

- Run the `./linux-amd64-server` or `./linux-amd64-client` command, which will automatically start the Web interface.
  - Default Web address for gt-server: `127.0.0.1:8000`
  - Default Web address for gt-client: `127.0.0.1:7000`
- For subsequent access, please directly enter the corresponding Web address in the browser.

#### 2. Login(This step is skipped for the first-time login)

- On the first login, the system will randomly assign a Web login username and password for the user, accompanied by a
  token to bypass the login.
- Enter `Admin` from the configuration file as the `Username` and input the corresponding `Password` to log in.

#### 3. User Setting

- After logging in, click on **“GT-Admin”** in the top right corner of the page to expand the system control bar. In the
  control bar, select **User Setting** to adjust user settings.
  - Available settings include: Username and Password for the next login, and whether to enable the pprof performance
    monitoring feature.
- **Strongly Recommended**: After the first login, immediately adjust the user settings for easier subsequent logins.
- If no settings are made within half an hour, the system-assigned login information will expire. In this case, you need to
  restart the service to obtain new login information.

#### 4. Configure GT Project

- If you've only set up the GT project and haven't modified the User Setting, the system will save the current Web
  settings when the GT configuration is saved.
- Users need to remember the Web user settings at this time (viewable in User Setting) or check the configuration file
  for Web login information during the next login. The default location of the configuration file is the same as the
  executable file.

#### 5. Save and Apply Configuration

- After completing the configuration, click the `Submit` button on the configuration interface or the `Change` button in
  User Setting to save the configuration.
- Click `Restart System` in the system control bar to restart the system. The new configuration will take effect after
  the
  restart.
  - Web configuration logic: User configuration information will be saved in the configuration file, and the system
    will load these configurations upon the next restart.
  - For the Client, if only the Service configuration has been changed, you can click `Reload Services` below the
    configuration interface to update the configuration in the current process.

#### 6. Notes

- The validity period of a user's single login is 30 minutes.

### Example

#### Change Web User Settings

<details>
   <summary>Detailed Steps</summary>

1. After entering the main interface, click on **“GT-Admin”** in the top right corner, and the system control bar will
   appear.
2. Click on `User Setting` to access the user information settings.
   - Configuration information includes:
     - Username and Password: The account settings for the user's next login.
     - Enablepprof: Whether to enable the pprof performance monitoring feature.
3. After making the necessary adjustments, click the `Change` button to write the relevant configurations to the
   configuration file.
   ![Web User Setting](https://github.com/ao-space/gt/assets/134463404/decb7cae-f022-4c54-ad2c-c1881bda7306)

</details>

#### HTTP Internal Penetration

<details>
   <summary>Detailed Steps</summary>
- Requirement: There's an intranet server and a public network server. id1.example.com resolves to the public network server's address. The goal is to access the web page of the service on port 80 of the intranet server by visiting id1.example.com:8080.

1. Configure the server (public network server):
   - Configure NetWork Setting: Set Addr: 8080
   - Configure User Setting: Set ID: id1, Secret: secret1
     ![HTTP Server](https://github.com/ao-space/gt/assets/134463404/b3d8b5a8-479b-44fa-bab4-5cefbff35832)

2. Configure the client (intranet server):
   - Configure General Setting: ID: id1, Secret: secret1, Remote: tcp://id1.example.com:8080
   - Configure Service Setting: LocalURL: http://127.0.0.1:80
     ![HTTP Client](https://github.com/ao-space/gt/assets/134463404/89ca0b20-5dcf-46d2-a899-81094eee3b81)

</details>

#### HTTPS Internal Penetration

<details>
    <summary>Detailed Steps</summary>

- Requirement: There's an intranet server and a public network server. id1.example.com resolves to the public network
  server's address. The goal is to access the HTTP web page provided by the service on port 80 of the intranet server by
  visiting https://id1.example.com.

1. Configure the server (public network server):
   - Configure NetWork Setting: Set TLSAddr: 443
   - Configure Security Setting: Set CertFile: /root/openssl_crt/tls.crt, KeyFile: /root/openssl_crt/tls.key
   - Configure User Setting: Set ID: id1, Secret: secret1   
     ![HTTPS Server](https://github.com/ao-space/gt/assets/134463404/33f3e296-140c-4124-9626-1900dc28b369)

2. Configure the client (intranet server). Since a self-signed certificate is used, the `remoteCertInsecure` option is
   used. This option should not be used in other scenarios (to prevent man-in-the-middle attacks that decrypt encrypted
   content):
   - Configure General Setting: ID: id1, Secret: secret1, Remote: tls://id1.example.com, RemoteCertInsecure: true
   - Configure Service Setting: LocalURL: http://127.0.0.1
     ![HTTPS Client](https://github.com/ao-space/gt/assets/134463404/f906237c-aea2-4127-ac8e-12a8764ca85b)

</details>

#### HTTPS SNI Internal Penetration

<details>
    <summary>Detailed Steps</summary>

- Requirement: There's an intranet server and a public network server. id1.example.com resolves to the public network
  server's address. The goal is to access the HTTPS web page provided by the service on port 443 of the intranet server
  by visiting https://id1.example.com.

1. Configure the server (public network server):
   - Configure NetWork Setting: Set Addr: 8080, SNIAdr: 443
   - Configure User Setting: Set ID: id1, Secret: secret1
     ![SNI Server](https://github.com/ao-space/gt/assets/134463404/b015d244-b5d8-42a5-9c99-9dbe6d5212c6)

2. Configure the client (intranet server):
   - Configure General Setting: ID: id1, Secret: secret1, Remote: tcp://id1.example.com:8080
   - Configure Service Setting: LocalURL: https://127.0.0.1
     ![SNI_Client](https://github.com/ao-space/gt/assets/134463404/4583323e-e2e3-443b-91ac-c3722d43438b)
   
</details>

#### Encrypt Client-Server Communication with TLS

<details>
    <summary>Detailed Steps</summary>

- Requirement: There's an intranet server and a public network server. id1.example.com resolves to the public network
  server's address. The goal is to access the web page of the service on port 80 of the intranet server by visiting
  id1.example.com:8080. Additionally, TLS is used to encrypt communication between the client and the server.

1. Configure the server (public network server):
   - Configure NetWork Setting: Set Addr: 8080, TLSAdr: 443
   - Configure Security Setting: Set CertFile: /root/openssl_crt/tls.crt, KeyFile: /root/openssl_crt/tls.key
   - Configure User Setting: Set ID: id1, Secret: secret1
     ![TLS Server](https://github.com/ao-space/gt/assets/134463404/bb9121be-6e5e-49ec-be6e-0766c3e58f74)

2. Configure the client (intranet server). Since a self-signed certificate is used, the `remoteCertInsecure` option is
   used. This option should not be used in other scenarios (to prevent man-in-the-middle attacks that decrypt encrypted
   content):
   - Configure General Setting: ID: id1, Secret: secret1, Remote: tls://id1.example.com, RemoteCertInsecure: true
   - Configure Service Setting: LocalURL: http://127.0.0.1:80
     ![TLS Client](https://github.com/ao-space/gt/assets/134463404/1bb27531-e92e-4250-8a78-f7b7c342410c)

</details>

#### TCP Internal Penetration

<details>
    <summary>Detailed Steps</summary>

- Requirement: There's an intranet server and a public network server. id1.example.com resolves to the public network
  server's address. The goal is to access the SSH service on port 22 of the intranet server by visiting id1.example.com:
  2222.If port 2222 on the server is not available, the server will choose a random port.

1. Configure the server (public network server):
   - Configure NetWork Setting: Set Addr: 8080
   - Configure User Setting: Set ID: id1, Secret: secret1, TCPNumber: 1, TCPRanges: 1024-65535
     ![TCP Server](https://github.com/ao-space/gt/assets/134463404/7f4da122-9e41-42a8-9f2c-b2dee98d96fe)

2. Configure the client (intranet server):
   - Configure General Setting: ID: id1, Secret: secret1, Remote: tcp://id1.example.com:8080
   - Configure Service Setting: LocalURL: tcp://127.0.0.1:22, RemoteTCPPort: 2222
     ![TCP Client](https://github.com/ao-space/gt/assets/134463404/889ff532-443f-4feb-a6d5-c8e698bd29ee)

</details>

#### Client Start Multiple Services Simultaneously

<details>
    <summary>Detailed Steps</summary>

- Requirement: There's an intranet server and a public network server. id1-1.example.com and id1-2.example.com resolve
  to the public network server's address. The goal is to access the service on port 80 of the intranet server by
  visiting id1-1.example.com:8080, access the service on port 8080 of the intranet server by visiting id1-2.example.com:
  8080, access the service on port 2222 of the intranet server by visiting id1-1.example.com:2222, and access the
  service on port 2223 of the intranet server by visiting id1-1.example.com:2223. Additionally, the server restricts the
  client's hostPrefix to be composed only of pure numbers or pure letters.

1. Configure the server (public network server):
   - Configure NetWork Setting: Set Addr: 8080
   - Configure User Setting: Set ID: id1, Secret: secret1, TCPNumber: 2, TCPRanges: 1024-65535, HostNumber: 2, WithID:
     true, HostRegex: ^[0-9]+$, ^[a-zA-Z]+$
     ![Multiple Server](https://github.com/ao-space/gt/assets/134463404/7d4dd44c-5d9b-4d8e-a35f-88743fbd68b2)

2. Configure the client (intranet server):
   - Configure General Setting: ID: id1, Secret: secret1, Remote: tcp://id1.example.com:8080
   - Configure Service 1 Setting: HostPrefix: 1, LocalURL: http://127.0.0.1:80, UseLocalAdHTTPHost: true
   - Configure Service 2 Setting: HostPrefix: 2, LocalURL: http://127.0.0.1:8080, UseLocalAdHTTPHost: true
   - Configure Service 3 Setting: LocalURL: tcp://127.0.0.1:2222, RemoteTCPPort: 2222
   - Configure Service 4 Setting: LocalURL: tcp://127.0.0.1:2223, RemoteTCPPort: 2223   
    ![Multiple Client](https://github.com/ao-space/gt/assets/134463404/7130769c-fd06-4012-8b22-ceb5586999a7)

</details>

## Web Configuration Instructions

- All web configurations related to gt-server(gt-client) are stored in the same configuration file.
- Default Settings 
  - Default Web address for gt-server: 127.0.0.1:8000
  - Default Web address for gt-client: 127.0.0.1:7000

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

## Frontend Developer Settings

- In `src/api/modules/login.ts`, by commenting out the code in the `getAuthMenuListApi`, you can get route permissions
  without starting the backend.
  ![code1](https://github.com/ao-space/gt/assets/134463404/56eec78a-1e6e-4018-9231-2c3b0529c777)

- In `src/routers/index.ts`, you can comment out the line in the `router.beforeEach` function to bypass user login,
  allowing navigation and testing of other interfaces.
  ![code2](https://github.com/ao-space/gt/assets/134463404/8f74e1a3-5893-4601-afdc-9603b6521308)

  After following the above steps, you can develop the frontend interface without needing to start the backend.

- Start the frontend development server:
  - a. Modify the proxy settings (First, review the following file and change the **PROXY** setting to match the web
    backend URL, ensuring it aligns with the YAML configuration file. In this case, it's set to 8000 to be consistent
    with the example above.)
    ```ts
    //.env.development
    VITE_PROXY = [["/api", "http://localhost:8000"]];
    ```
  - b. Start the development server
    ```shell
    npm run dev
    ```
