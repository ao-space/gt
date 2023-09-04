export namespace ClientConfig {
  export interface Config extends Options {
    // Version: string;
    Services?: Service[];
  }

  export interface GeneralSetting {
    ID: string;
    Secret: string;
    ReconnectDelay: string;
    Remote: string;
    RemoteSTUN: string;
    RemoteAPI: string;
    RemoteCert: string;
    RemoteCertInsecure: boolean;
    RemoteConnections: number;
    RemoteIdleConnections: number;
    RemoteTimeout: string;
  }
  export interface SentrySetting {
    SentryDSN: string;
    SentryLevel: string[];
    SentrySampleRate: number;
    SentryRelease: string;
    SentryEnvironment: string;
    SentryServerName: string;
    SentryDebug: boolean;
  }
  export interface WebRTCSetting {
    WebRTCConnectionIdleTimeout: string;
    WebRTCLogLevel: string;
    WebRTCMinPort: number;
    WebRTCMaxPort: number;
  }
  export interface TCPForwardSetting {
    TCPForwardAddr: string;
    TCPForwardHostPrefix: string;
    TCPForwardConnections: number;
  }
  export interface LogSetting {
    LogFile: string;
    LogFileMaxSize: number;
    LogFileMaxCount: number;
    LogLevel: string;
  }
  export interface Service {
    HostPrefix: string;
    RemoteTCPPort: number;
    RemoteTCPRandom: boolean;
    LocalURL: string;
    LocalTimeout: string;
    UseLocalAsHTTPHost: boolean;
  }
  export interface Options extends GeneralSetting, SentrySetting, WebRTCSetting, TCPForwardSetting, LogSetting {
    Config: string;
    // HostPrefix: string[];
    // RemoteTCPPort: number[];
    // RemoteTCPRandom: (boolean | null)[];
    // Local: string[];
    // LocalTimeout: string[];
    // UseLocalAsHTTPHost: boolean[];
  }

  export const defaultGeneralSetting: GeneralSetting = {
    ID: "",
    Secret: "",
    ReconnectDelay: "",
    RemoteTimeout: "",
    Remote: "",
    RemoteSTUN: "",
    RemoteAPI: "",
    RemoteCert: "",
    RemoteCertInsecure: false,
    RemoteConnections: 1,
    RemoteIdleConnections: 0
  };
  export const defaultSentrySetting: SentrySetting = {
    SentryDSN: "",
    SentryLevel: ["error", "fatal", "panic"],
    SentrySampleRate: 0,
    SentryRelease: "",
    SentryEnvironment: "",
    SentryServerName: "",
    SentryDebug: false
  };
  export const defaultWebRTCSetting: WebRTCSetting = {
    WebRTCConnectionIdleTimeout: "",
    WebRTCLogLevel: "",
    WebRTCMinPort: 0,
    WebRTCMaxPort: 0
  };
  export const defaultTCPForwardSetting: TCPForwardSetting = {
    TCPForwardAddr: "",
    TCPForwardHostPrefix: "",
    TCPForwardConnections: 0
  };
  export const defaultLogSetting: LogSetting = {
    LogFile: "",
    LogFileMaxSize: 0,
    LogFileMaxCount: 0,
    LogLevel: ""
  };
  export const defaultServiceSetting: Service = {
    HostPrefix: "",
    RemoteTCPPort: 0,
    RemoteTCPRandom: false,
    LocalURL: "",
    LocalTimeout: "",
    UseLocalAsHTTPHost: false
  };

  export const usage = {
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
    LogLevel: "Log level: trace, debug, info, warn, error, fatal, panic, disable"
  };

  export interface FormRef {
    validateForm: () => Promise<void>;
  }
  export interface RuleForm {
    ReconnectDelay: string;
    RemoteTimeout: string;
    LocalTimeout: string;
    WebRTCConnectionIdleTimeout: string;
  }
}
