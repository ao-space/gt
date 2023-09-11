export namespace ServerConfig {
  export interface Config extends Options {
    Users: Record<string, User>;
    TCPs: TCP[];
    Host: Host;
  }
  export interface TCP {
    Range: string;
    Number: number;
  }
  // export interface TCPInOptions {
  //   TCPRanges: string[];
  //   TCPNumbers: string[];
  // }
  export interface User {
    Secret: string;
    TCPs: TCP[];
    Speed: number;
    Connections: number;
    Host: Host;
    // temp: boolean;
  }
  export interface UserSetting extends User {
    ID: string;
  }

  export interface Host {
    Number: number;
    RegexStr: string[];
    WithID: boolean;
  }
  // export interface HostInOptions {
  //   HostNumber: number;
  //   HostRegex: string[];
  //   HostWithID: boolean;
  // }

  export interface GeneralSetting {
    UserPath: string;
    AuthAPI: string;
  }
  export interface GeneralSettingProps {
    UserPath: string;
    AuthAPI: string;
    TCPs: TCP[];
    Host: Host;
  }
  // export interface GeneralSetting extends HostInOptions, TCPInOptions {
  //   Users: string;
  //   AuthAPI: string;
  // }
  export interface NetworkSetting {
    Addr: string;
    TLSAddr: string;
    TLSMinVersion: string;
    STUNAddr: string;
    SNIAddr: string;
    HTTPMUXHeader: string;
  }
  export interface SecuritySetting {
    CertFile: string;
    KeyFile: string;
    AllowAnyClient: boolean;
  }
  export interface ConnectionSetting {
    Speed: number;
    Connections: number;
    ReconnectTimes: number;
    ReconnectDuration: string;
    Timeout: string;
    TimeoutOnUnidirectionalTraffic: boolean;
  }

  export interface APISetting {
    APIAddr: string;
    APICertFile: string;
    APIKeyFile: string;
    APITLSMinVersion: string;
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
  export interface LogSetting {
    LogFile: string;
    LogFileMaxSize: number;
    LogFileMaxCount: number;
    LogLevel: string;
  }
  export interface Options
    extends GeneralSetting,
      NetworkSetting,
      SecuritySetting,
      ConnectionSetting,
      APISetting,
      SentrySetting,
      LogSetting {
    // Config: string;
    // IDs: string[];
    // Secrets: string[];
    // Users: string;
    // AuthAPI: string;
    // TCPRanges: string[];
    // TCPNumbers: string[];
    // Version: boolean;
  }
  export const defaultHostSetting: Host = {
    Number: 0,
    RegexStr: [],
    WithID: false
  };
  export const defaultTCPSetting: TCP = {
    Range: "",
    Number: 0
    // PortRange: [],
    // usedPort: 0
  };
  export const defaultGeneralSetting: GeneralSetting = {
    UserPath: "",
    AuthAPI: ""
    // TCPRanges: ["12-123", "123-11"],
    // TCPNumbers: ["12", "1223"],
    // HostNumber: 23,
    // HostRegex: [".*", "http"],
    // HostWithID: false
  };
  export const defaultGeneralSettingProps: GeneralSettingProps = {
    UserPath: "",
    AuthAPI: "",
    TCPs: [],
    Host: defaultHostSetting
  };
  export const defaultNetworkSetting: NetworkSetting = {
    Addr: "",
    TLSAddr: "",
    TLSMinVersion: "",
    STUNAddr: "",
    SNIAddr: "",
    HTTPMUXHeader: ""
  };
  export const defaultSecuritySetting: SecuritySetting = {
    CertFile: "",
    KeyFile: "",
    AllowAnyClient: false
  };
  export const defaultConnectionSetting: ConnectionSetting = {
    Speed: 0,
    Connections: 0,
    ReconnectTimes: 0,
    ReconnectDuration: "",
    Timeout: "",
    TimeoutOnUnidirectionalTraffic: false
  };

  export const defaultAPISetting: APISetting = {
    APIAddr: "",
    APICertFile: "",
    APIKeyFile: "",
    APITLSMinVersion: ""
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
  export const defaultLogSetting: LogSetting = {
    LogFile: "",
    LogFileMaxSize: 0,
    LogFileMaxCount: 0,
    LogLevel: ""
  };
  export const defaultUserSetting: UserSetting = {
    ID: "",
    Secret: "",
    TCPs: [],
    Speed: 0,
    Connections: 0,
    Host: defaultHostSetting
  };
  export interface FormRef {
    validateForm: () => Promise<void>;
  }
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
    TCPNumbers: "The number of tcp ports allowed to be opened for each id",
    Speed: "The max number of bytes the client can transfer per second",
    Connections: "The max number of tunnel connections for a client",
    ReconnectTimes: "The max number of times the client fails to reconnect",
    ReconnectDuration: "The time that the client cannot connect after the number of failed reconnections reaches the max number",
    HostNumber: "The number of host-based services that the user can start",
    HostRegex: "The host prefix started by user must conform to one of these rules",
    HostWithID: "The prefix of host will become the form of id-host",
    HTTPMUXHeader: "The http multiplexing header to be used",
    Timeout: "The timeout of connections. Supports values like '30s', '5m'",
    TimeoutOnUnidirectionalTraffic: "Timeout will happens when traffic is unidirectional",
    APIAddr: "The address to listen on for internal api service. Supports values like: '8080', ':8080' or '0.0.0.0:8080'",
    APICertFile: "The api TLS certificate file path",
    APIKeyFile: "The path to key file",
    APITLSMinVersion: "The tls min version. Supports values: tls1.1, tls1.2, tls1.3",
    STUNAddr: "The address to listen on for STUN service. Supports values like: '3478', ':3478' or '0.0.0.0:3478'",
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
}
