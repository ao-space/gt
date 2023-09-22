# GT-Admin

English | [简体中文](README_CN.md)

## Introduction 📖

**GT-Admin** is a Web interface specifically developed for **[GT](https://github.com/ao-space/gt)** users. It adopts the
template of [Geeker-Admin](https://github.com/HalseySpicy/Geeker-Admin) and is built upon
the [gin framework](https://github.com/gin-gonic/gin). This interface not only allows users to perform various
configuration operations more intuitively visually, but also provides users with **visualization monitoring capabilities
** for system status.

## Table of Contents

- [Features](#features)
- [Project Structure](#project-structure)
- [Installation Steps](#installation-steps)
  - [GT-Server Setup](#gt-server-setup)
  - [GT-Client Setup](#gt-client-setup)
- [Web Interface User Guide](#web-interface-user-guide)
  - [Login](#login)
  - [System Control](#system-control)
  - [System Status Monitoring](#system-status-monitoring)
  - [Connection Status View](#connection-status-view)
  - [Configuration Interface](#configuration-interface)
  - [pprof Interface](#pprof-interface)
- [Frontend Developer Settings](#frontend-developer-settings)

## Features

- System status monitoring (Monitor OS, CPU, Memory, Disk)
- Connection status viewing (Connection pool connections, external connections)
- Configuration management features (View, Modify, Save)
- pprof Performance Analysis

## Project Structure

![Architecture](src/assets/images/Architecture.png)

## Installation Steps

### GT-Server Setup

<details>
    <summary>Detailed Steps</summary>

1. Clone the project
    ```shell
    git clone https://github.com/huwf5/gt.git
    ```
2. Compile the backend project
    ```shell
    cd gt
    git checkout -b first origin/first   ## Temporarily required
    make release_server # The compiled files will be in the release folder
    ```
3. Compile the frontend project
    ```shell
    cd web/front
    npm install
    ```
4. Write the web configuration file (**Please configure Web Setting in detail**, other configurations can be done on the
   web later. It's recommended to **save** it in the **release** folder).
    <details>
    <summary>server.yaml</summary>

      ```yaml
    #server.yaml
    options:
    # General Setting (MUST!) :To start the gt-server
    # You can change it later on the web page
      addr: 8080

    # Web Setting(Optional)
      # Whether to start the Web Server
      web: true
      #Set Web Address
      webAddr: localhost
      webPort: 7000
      # Use to sign the jwt token(Validity Period: 6 hours)
      signingKey: signature
      # Use to log in on the web page
      admin: server
      password: admin
      # Start the pprof services
      # 'web' prop must be set to true first
      pprof: true #(optional)
      ```

    </details>

5. Start the service

- Start the backend service (switch to the **release** folder)
  ```shell
  # Note: change the [] to your actual location
  ./linux-amd64-server -config [path/to/server.yaml]   # start gt-server
  ```
- Start the frontend service
  - a. Change the proxy settings (first check the file below, change the **PROXY** setting to the corresponding web
    backend URL, consistent with the yaml configuration file, which is 7000 in this case)
    ```ts
    //.env.development
    VITE_PROXY = [["/api", "http://localhost:7000"]];
    ```
  - b. Start the web service
    ```shell
    npm run dev
    ```

</details>

### GT-Client Setup

<details>
    <summary>Detailed Steps</summary>

1. Clone the project
    ```shell
    git clone https://github.com/huwf5/gt.git
    ```
2. Compile the backend project
    ```shell
    cd gt
    git checkout -b first origin/first   ## Temporarily required
    make release_client # The compiled files will be in the release folder
    ```
3. Compile the frontend project
    ```shell
    cd web/front
    npm install
    ```
4. Write the web configuration file (**Please configure Web Setting in detail**, other configurations can be done on the
   web later. It's recommended to **save** it in the **release** folder).
    <details>
    <summary>client.yaml</summary>

   ```yaml
   #client.yaml
   options:
   # General Setting (MUST!) : To start the gt-client
   # You can change it later on the web page
   id: id1
   remote: tcp://localhost:8080

   # Web Setting (Optional)
   # Whether to start the Web Server
   web: true
   # Set Web Address
   webAddr: localhost
   webPort: 8000
   # Used to sign the jwt token (Validity Period: 6 hours)
   signingKey: signature
   # Used to log in on the web page
   admin: client
   password: admin
   # Start the pprof services
   # 'web' property needs to be set to true first
   pprof: true #(optional)
    ```

    </details>

5. Start the service

- Start the backend service (switch to the **release** folder)
  ```shell
  # Note: change the [] to your actual location
  ./linux-amd64-client -config [path/to/client.yaml]   # start gt-client
  ```
- Start the frontend service
  - a. Change the proxy settings (first check the file below, change the **PROXY** setting to the corresponding web
    backend URL, consistent with the yaml configuration file, which is 8000 in this case)
    ```ts
    //.env.development
    VITE_PROXY = [["/api", "http://localhost:8000"]];
    ```
  - b. Start the web service
    ```shell
    npm run dev
    ```

</details>

## Web Interface User Guide

### Login

- Log in using the `admin` and `password` set in the configuration file.
  ![Login](src/assets/images/Login.png)

### System Control

- Click on **“GT-Admin”** in the top right corner to access the toolbar.

  - **Login out**: Clear user information and log out.
  - **Restart System**: Restart the entire server.

  **Note**: The following actions will shut down the entire system. If you want the services to continue running, you'll
  need to start them manually. **Proceed with caution**.

  - **Shutdown System**: Shut down the system.
  - **Terminate System**: Interrupt the system.

  ![ToolBar](src/assets/images/ToolBar.png)

### System Status Monitoring

- Provides system information, and views for DISK, CPU, RAM.
  ![DashBoard](src/assets/images/DashBoard.png)

### Connection Status View

- The Server side provides a view for connection **information**.
  ![ServerConnection](src/assets/images/ServerConnection.png)

- The Client side provides a view for connection **status**.
  ![ClientConnection](src/assets/images/ClientConnection.png)

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
    - **GetFromFile**: Retrieve content from the configuration file. If the service wasn't started with `-config` (
      meaning no configuration file was set), then it will perform the **GetFromRunning** action. (**It's recommended to
      start the service with `-config`**, because subsequent Restart actions will re-execute the initial startup
      command. Only if the initial startup specified a configuration file path can subsequent operations match user
      expectations.)
    - **GetFromRunning**: Retrieve configuration from the currently running service.

- Server configuration activation:

  - **After users save their changes**, they can activate the new configuration by using **Restart System** (this action
    will start a new process).

  - The **TCP Setting** and **Host Setting** set in the **General Setting** are **global** configurations. For more *
    *detailed** settings, please configure them in the **User Setting** section below.
    ![ServerConfig](src/assets/images/ServerConfig.png)

- Client configuration activation:

  - **After users save their changes**, they can use **Reload Services** to keep the existing process running while
    restarting the Services service (provided only the Services were changed). However, if content in the Options field
    was changed (i.e., parts other than Services), to activate the new configuration, you'll need to use **Restart
    System** to restart the entire process.
    ![ClientConfig](src/assets/images/ClientConfig.png)

### pprof Interface

- Performance monitoring interface.
  ![pprof](src/assets/images/pprof.png)

## Frontend Developer Settings

- In `src/api/modules/login.ts`, by commenting out the code in the `getAuthMenuListApi`, you can get route permissions
  without starting the backend.
  ![code1](src/assets/images/code1.png)

- In `src/routers/index.ts`, you can comment out the line in the `router.beforeEach` function to bypass user login,
  allowing navigation and testing of other interfaces.
  ![code2](src/assets/images/code2.png)

After following the above steps, you can develop the frontend interface without needing to start the backend.
