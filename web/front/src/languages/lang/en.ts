export namespace en {
  export const usage = {
    // General Setting

    Config: "The config file path to load",
    Addr: "The address to listen on. Supports values like: '80', ':80' or '0.0.0.0:80'",
    TLSAddr: "The address for tls to listen on. Supports values like: '443', ':443' or '0.0.0.0:443'",
    TLSMinVersion: "The tls min version. Supports values: tls1.1, tls1.2, tls1.3",
    CertFile: "The path to cert file",
    KeyFile: "The path to key file",

    IDs: "The user id",
    Secrets: "The secret for user id",
    Users: "The users yaml file to load",
    AuthAPI: "The API to authenticate user with id and secret",
    AllowAnyClient: "Allow any client to connect to the server",
    TCPRanges: "The tcp port range, like 1024-65535",
    TCPNumber: "The number of tcp ports allowed to be opened for each id",
    Speed: "The max number of bytes the client can transfer per second",
    Connections: "The max number of tunnel connections for a client",
    ReconnectTimes: "The max number of times the client fails to reconnect",
    ReconnectDuration: "The time that the client cannot connect after the number of failed reconnections reaches the max number",
    HostNumber: "The number of host-based services that the user can start",
    HostRegex: "The host prefix started by user must conform to one of these rules",
    HostWithID: "The prefix of host will become the form of id-host",

    HTTPMUXHeader: "The http multiplexing header to be used",
    MaxHandShakeOptions: "The max number of hand shake options",

    Timeout: "The timeout of connections. Supports values like '30s', '5m'",
    TimeoutOnUnidirectionalTraffic: "Timeout will happens when traffic is unidirectional",

    APIAddr: "The address to listen on for internal api service. Supports values like: '8080', ':8080' or '0.0.0.0:8080'",
    APICertFile: "The api TLS certificate file path",
    APIKeyFile: "The path to key file",
    APITLSMinVersion: "The tls min version. Supports values: tls1.1, tls1.2, tls1.3",

    STUNAddr: "The address to listen on for STUN service. Supports values like: '3478', ':3478' or '0.0.0.0:3478'",
    STUNLogLevel: "Log level: trace, debug, info, warn, error, disable",

    SNIAddr:
      "The address to listen on for raw tls proxy. Host comes from Server Name Indication. Supports values like: '443', ':443' or '0.0.0.0:443'",

    SentryDSN: "Sentry DSN to use",
    SentryLevel: 'Sentry levels: trace, debug, info, warn, error, fatal, panic (default ["error", "fatal", "panic"])',
    SentrySampleRate: "Sentry sample rate for event submission: [0.0 - 1.0]",
    SentryRelease: "Sentry release to be sent with events",
    SentryEnvironment: "Sentry environment to be sent with events",
    SentryServerName: "Sentry server name to be reported",
    SentryDebug: "Sentry debug mode, the debug information is printed to help you understand what sentry is doing",

    LogFile: "Path to save the log file",
    LogFileMaxSize: "Max size of the log files",
    LogFileMaxCount: "Max count of the log files",
    LogLevel: "Log level: trace, debug, info, warn, error, fatal, panic, disable",
    Version: "Show the version of this program",

    tcp: {
      Range: "The tcp port range",
      Number: "The tcp port number",
      PortRange: "The tcp port range",
      usedPort: "The used tcp port"
    },
    user: {
      ID: "The user id", //for the mapping key
      Secret: "The user secret",
      TCPs: "The user tcp ports",
      Speed: "The user speed limit in bytes",
      Connections: "The user max connections",
      Host: "The user host",
      temp: "The user temp"
    },
    host: {
      Number: "The host number",
      RegexStr: "The host regex string",
      Regex: "The host regex",
      WithID: "The host with id",
      usedHost: "The used host"
    }
  };
  export const cusage = {
    // General Setting
    Config: "The config file path to load",
    ID: "The unique id used to connect to server. Now it's the prefix of the domain.",
    Secret: "The secret used to verify the id",
    ReconnectDelay: "The delay before reconnect. Supports values like '30s', '5m'",
    Remote: "The remote server url. Supports tcp:// and tls://, default tcp://",
    RemoteSTUN: "The remote STUN server address",
    RemoteAPI: "The API to get remote server url",
    RemoteCert: "The path to remote cert",
    RemoteCertInsecure: "Accept self-signed SSL certs from remote",
    RemoteConnections: "The max number of server connections in the pool. Valid value is 1 to 10",
    RemoteIdleConnections: "The number of idle server connections kept in the pool",
    RemoteTimeout: "The timeout of remote connections. Supports values like '30s', '5m'",
    Version: "Show the version of this program",

    // Service Setting
    HostPrefix: "The server will recognize this host prefix and forward data to local",
    RemoteTCPPort: "The TCP port that the remote server will open",
    RemoteTCPRandom: "Whether to choose a random port by the remote server",
    Local: "The local service url",
    LocalURL: "The local service url",
    LocalTimeout: "The timeout of local connections. Supports values like '30s', '5m'",
    UseLocalAsHTTPHost: "Use the local host as host",

    // Sentry Setting
    SentryDSN: "Sentry DSN to use",
    SentryLevel: 'Sentry levels: trace, debug, info, warn, error, fatal, panic (default ["error", "fatal", "panic"])',
    SentrySampleRate: "Sentry sample rate for event submission: [0.0 - 1.0]",
    SentryRelease: "Sentry release to be sent with events",
    SentryEnvironment: "Sentry environment to be sent with events",
    SentryServerName: "Sentry server name to be reported",
    SentryDebug: "Sentry debug mode, the debug information is printed to help you understand what sentry is doing",

    // WebRTC Setting
    WebRTCConnectionIdleTimeout: "The timeout of WebRTC connection. Supports values like '30s', '5m'",
    WebRTCLogLevel: "WebRTC log level: verbose, info, warning, error",
    WebRTCMinPort: "The min port of WebRTC peer connection",
    WebRTCMaxPort: "The max port of WebRTC peer connection",

    // TCP Forward Setting
    TCPForwardAddr: "The address of TCP forward",
    TCPForwardHostPrefix: "The host prefix of TCP forward",
    TCPForwardConnections: "The max number of TCP forward peer connections in the pool. Valid value is 1 to 10",

    // Log Setting
    LogFile: "Path to save the log file",
    LogFileMaxSize: "Max size of the log files",
    LogFileMaxCount: "Max count of the log files",
    LogLevel: "Log level: trace, debug, info, warn, error, fatal, panic, disable",
    SelectLogLevel: "Select log level"
  };
  export const sconfig = {
    APIAddr: "APIAddr",
    APITLSMinVersion: "APITLSMinVersion",
    APICertFile: "APICertFile",
    APIKeyFile: "APIKeyFile",
    Speed: "Speed",
    Connections: "Connections",
    ReconnectTimes: "ReconnectTimes",
    ReconnectDuration: "ReconnectDuration",
    Timeout: "Timeout",
    TimeoutOnUnidirectionalTraffic: "TimeoutOnUnidirectionalTraffic",
    Users: "Users",
    AuthAPI: "AuthAPI",
    TCPNumber: "TCPNumber",
    HostNumber: "HostNumber",
    WithID: "WithID",
    HostRegex: "HostRegex",
    Done: "Done",
    Addr: "Addr",
    TLSAddr: "TLSAddr",
    TLSMinVersion: "TLSMinVersion",
    STUNAddr: "STUNAddr",
    STUNLogLevel: "STUNLogLevel",
    SNIAddr: "SNIAddr",
    HTTPMUXHeader: "HTTPMUXHeader",
    MaxHandShakeOptions: "MaxHandShakeOptions",
    CertFile: "CertFile",
    KeyFile: "KeyFile",
    AllowAnyClient: "AllowAnyClient",
    TCPRanges: "TCPRanges",
    Operation: "Operation",
    Add: "Add",
    ID: "ID",
    Secret: "Secret",
    Edit: "Edit",
    Delete: "Delete",
    AddUser: "AddUser",
    ConnectionSetting: "Connection Setting",
    GeneralSetting: "General Setting",
    SecuritySetting: "Security Setting",
    NetworkSetting: "Network Setting",
    APISetting: "API Setting",
    HostSetting: "Host Setting",
    TCPSetting: "TCP Setting",
    LogSetting: "Log Setting",
    SentrySetting: "Sentry Setting",
    User: "User",
    Setting: "Setting",
    Submit: "Submit",
    GetFromFile: "GetFromFile",
    GetFromRunning: "GetFromRunning",
    AddService: "Add Service",
    DetailSettings: "DetailSettings",
    //sentences
    AddTcpRanges: "Please Add TCP Ranges and Numbers",
    AddHostRegex: "Please Add A Host Regex",
    SelectApiTLSMin: "Select APITLSMinVersion",
    SelectTLSMin: "Select TLSMinVersion",
    SelectSTUNLogLevel: "Select STUN log level",

    IDConflictError: "ID conflict in user setting, please check.",
    SaveConfigConfirm: "Make sure you want to save the configuration file",
    SaveConfigTitle: "Save The Configuration",
    SaveConfigConfirmBtn: "Confirm",
    SaveConfigCancelBtn: "Cancel",
    SubmitSuccess: "Submit success",
    FailedToSaveConfig: "Failed to save the configuration file!",

    GetFromFileConfirm:
      "Make sure you want to get the configuration from file, if you fail to get from file, it will get from the running system. NOTE: please make sure the change you made is saved, or it will be discarded.",
    GetFromFileTitle: "Get Configuration From File",
    GetFromFileConfirmBtn: "Confirm",
    GetFromFileCancelBtn: "Cancel",
    GetFromFileSuccess: "Get from file success",
    FailedToGetFromFile: "Failed to get from file!",

    GetFromRunningConfirm:
      "Make sure you want to get the configuration from running system. NOTE: please make sure the change you made is saved, or it will be discarded.",
    GetFromRunningTitle: "Get Configuration From Running System",
    GetFromRunningConfirmBtn: "Confirm",
    GetFromRunningCancelBtn: "Cancel",
    GetFromRunningSuccess: "Get from running system success",
    FailedToGetFromRunning: "Failed to get from running system!"
  };
  export const cconfig = {
    ID: "ID",
    ReconnectDelay: "ReconnectDelay",
    RemoteTimeout: "RemoteTimeout",
    Remote: "Remote",
    RemoteSTUN: "RemoteSTUN",
    RemoteAPI: "RemoteAPI",
    RemoteCert: "RemoteCert",
    RemoteCertInsecure: "RemoteCertInsecure",
    RemoteConnections: "RemoteConnections",
    RemoteIdleConnections: "RemoteIdleConnections",
    LogFile: "LogFile",
    LogFileMaxSize: "LogFileMaxSize",
    LogFileMaxCount: "LogFileMaxCount",
    LogLevel: "LogLevel",
    SentryDSN: "SentryDSN",
    SentryServerName: "SentryServerName",
    SentryLevel: "SentryLevel",
    SentrySampleRate: "SentrySampleRate",
    SentryRelease: "SentryRelease",
    SentryEnvironment: "SentryEnvironment",
    SentryDebug: "SentryDebug",
    HostPrefix: "HostPrefix",
    RemoteTCPPort: "RemoteTCPPort",
    RemoteTCPRandom: "RemoteTCPRandom",
    LocalURL: "LocalURL",
    LocalTimeout: "LocalTimeout",
    UseLocalAsHTTPHost: "UseLocalAsHTTPHost",
    TcpForwardAddr: "TcpForwardAddr",
    TcpForwardHostPrefix: "TcpForwardHostPrefix",
    TcpForwardConnections: "TcpForwardConnections",
    WebRTCConnectionIdleTimeout: "WebRTCConnectionIdleTimeout",
    WebRTCLogLevel: "WebRTCLogLevel",
    WebRTCMinPort: "WebRTCMinPort",
    WebRTCMaxPort: "WebRTCMaxPort",
    GeneralSetting: "General Setting",
    LogSetting: "Log Setting",
    SentrySetting: "Sentry Setting",
    ServiceSetting: "Service Setting",
    TCPForwardSetting: "TcpForward Setting",
    WebRTCSetting: "WebRTC Setting",
    Submit: "Submit",
    GetFromFile: "GetFromFile",
    GetFromRunning: "GetFromRunning",
    ReloadServices: "Reload Services",
    AddService: "Add Service",
    Delete: "Delete",
    Secret: "Secret",
    Service: "Service",
    Setting: "Setting",
    //sentences
    SaveConfigConfirm: "Make sure you want to save the configuration to file.",
    SaveConfigTitle: "Save The Configuration",
    SaveConfigConfirmBtn: "Confirm",
    SaveConfigCancelBtn: "Cancel",
    OperationSuccess: "Operation Success!",
    FailedOperation: "Failed operation!",

    GetFromFileConfirm:
      "Make sure you want to get the configuration from file, if you fail to get from file, it will get from the running system. NOTE: please make sure the change you made is saved, or it will be discarded.",
    GetFromFileTitle: "Get Configuration From File",
    GetFromFileConfirmBtn: "Confirm",
    GetFromFileCancelBtn: "Cancel",

    GetFromRunningConfirm:
      "Make sure you want to get the configuration from running system. NOTE: please make sure the change you made is saved, or it will be discarded.",
    GetFromRunningTitle: "Get Configuration From Running System",
    GetFromRunningConfirmBtn: "Confirm",
    GetFromRunningCancelBtn: "Cancel",

    ReloadServicesConfirm:
      "You need to make sure that the changes you make only happen in the services section, and make sure it has been saved, or the system won't reload the services.",
    ReloadServicesTitle: "Reload Services",
    ReloadServicesConfirmBtn: "Confirm",
    ReloadServicesCancelBtn: "Cancel",
    InconsistentOptionsWarning: "The options you changed are not consistent with the running system!"
  };
  export const view_home = {
    Runtime: "Runtime",
    Used: "Used",
    Ram: "Ram",
    Total: "Total",
    Core_Number: "physical number of cores",
    CPU: "CPU",
    Disk: "Disk",
    os: "os",
    cpu_num: "CPU nums",
    compiler: "compiler",
    go_version: "go version",
    goroutine_nums: "goroutine nums",
    core: "core"
  };
  export const view_login = {
    Username: "Username",
    Password: "Password",
    Login: "Login",
    Reset: "Reset"
  };
  export const view_connection = {
    Server_Pool_Info: "Server Pool Info",
    External_Connection: "External Connection"
  };
  export const layout_header = {
    UserSetting: "User Setting",
    Logout: "Log out",
    RestartSystem: "Restart System",
    ShutdownSystem: "Shutdown System",
    TerminateSystem: "Terminate System",
    Username: "Username",
    Password: "Password",
    Login: "Login",
    Reset: "Reset",
    EnablePprof: "EnablePprof",
    large: "large",
    default: "default",
    small: "small",
    DoneBtnText: "Finish",
    CloseBtnText: "Close",
    NextBtnText: "Next",
    PrevBtnText: "Previous",
    CollapseIconTitle: "Collapse Icon",
    CollapseIconDescription: "Toggle the sidebar open or closed.",
    BreadcrumbTitle: "Breadcrumb",
    BreadcrumbDescription: "Indicate the current page location",
    GuideTitle: "Guide",
    GuideDescription: "Guide the user to use the system",
    AssemblySizeTitle: "Switch Assembly Size",
    AssemblySizeDescription: "Adjust the system's display size.",
    ThemeSettingTitle: "Setting theme",
    ThemeSettingDescription: "Customize the system's theme.",
    FullScreenTitle: "Full Screen",
    FullScreenDescription: "Enter or exit full-screen mode.",
    UserTitle: "User",
    UserDescription:
      "Click here to open the System Settings.<br/> Upon the first launch, the system " +
      "automatically generates a random username and password for you. <strong>We strongly recommend updating these details " +
      "within 30 minutes</strong> to ensure smooth future logins.",
    UsernameRequired: "Please enter username",
    PasswordRequired: "Please enter password",
    ChangeInfoWarning:
      "Are you sure you want to change your account information? If you want to apply this new change please restart the system!",
    Warning: "Warning",
    OK: "OK",
    Cancel: "Cancel",
    ChangeInfoSuccess: "Change account information success",
    ChangeInfoFailure: "Failed to change account information",
    CancelChangeInfo: "Cancel change account information"
  };
  export const layout_tabs = {
    Refresh: "Refresh",
    Maximize: "Maximize",
    CloseCurrentTab: "Close Current Tab",
    CloseOtherTabs: "Close Other Tabs",
    CloseAllTabs: "Close All Tabs",
    More: "More"
  };
  export const layout_theme = {
    InvertedAsideColor: "Inverted Aside Color",
    Theme: "Theme",
    SwitchAside: "Switch Aside color to Dark mode",
    ThemeColor: "Theme Color",
    DarkMode: "Dark Mode",
    GreyMode: "Grey Mode",
    ColorAccessibilityMode: "Color Accessibility Mode",
    UISettings: "UI Settings",
    CollapseMenu: "Collapse Menu",
    Breadcrumb: "Breadcrumb",
    BreadcrumbIcon: "Breadcrumb Icon",
    Tab: "Tab",
    TabIcon: "Tab Icon",
    Footer: "Footer"
  };
  export const connection_table = {
    ID: "ID",
    Family: "Family",
    Type: "Type",
    LocalAddress: "Local Address",
    RemoteAddress: "Remote Address",
    Status: "Status"
  };
  export const result = {
    RequestFailed: "Request failed! Please try again later.",
    LoginExpired: "Login expired! Please log in again.",
    NoPermission: "You do not have permission to access this resource.",
    ResourceNotFound: "The resource you are trying to access does not exist!",
    InvalidRequestMethod: "Invalid request method! Please try again later.",
    RequestTimedOut: "Request timed out! Please try again later.",
    InternalServerError: "Internal server error.",
    BadGateway: "Bad gateway.",
    ServiceUnavailable: "Service is currently unavailable. Please try again later.",
    GatewayTimeout: "Gateway timeout. The server took too long to respond.",
    UnexpectedError: "An unexpected error occurred. Please try again."
  };
}
