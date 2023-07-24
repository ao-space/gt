export namespace ServerConfig {
  export interface Config {
    Version: string;
    Users: Record<string, User>;
    TCPs: TCP[];
    Host: Host;
    Options: Options;
  }

  export interface Options {
    Config: string;
    Addr: string;
    TLSAddr: string;
    TLSMinVersion: string;
    CertFile: string;
    KeyFile: string;
    IDs: string[];
    Secrets: string[];
    Users: string;
    AuthAPI: string;
    AllowAnyClient: boolean;
    TCPRanges: string[];
    TCPNumbers: string[];
    Speed: number;
    Connections: number;
    ReconnectTimes: number;
    ReconnectDuration: string; // assuming this is a string representation of a duration
    HostNumber: number;
    HostRegex: string[];
    HostWithID: boolean;
    HTTPMUXHeader: string;
    Timeout: string; // assuming this is a string representation of a duration
    TimeoutOnUnidirectionalTraffic: boolean;
    APIAddr: string;
    APICertFile: string;
    APIKeyFile: string;
    APITLSMinVersion: string;
    STUNAddr: string;
    SNIAddr: string;
    SentryDSN: string;
    SentryLevel: string[];
    SentrySampleRate: number;
    SentryRelease: string;
    SentryEnvironment: string;
    SentryServerName: string;
    SentryDebug: boolean;
    LogFile: string;
    LogFileMaxSize: number;
    LogFileMaxCount: number;
    LogLevel: string;
    Version: boolean;
  }

  export interface TCP {
    Range: string;
    Number: number;
    PortRange: any; // assuming this is a custom type, replace with actual type if available
    usedPort: number;
  }

  export interface User {
    Secret: string;
    TCPs: TCP[];
    Speed: number;
    Connections: number;
    Host: Host;
    temp: boolean;
  }

  export interface Host {
    Number: number;
    RegexStr: string[];
    Regex: any[]; // assuming this is a custom type, replace with actual type if available
    WithID: boolean;
    usedHost: number;
  }
}
