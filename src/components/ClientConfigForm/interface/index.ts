export namespace ClientConfig {
  export interface Config {
    Version: string;
    Services: Service[];
    Options: Options;
  }

  export interface Service {
    HostPrefix: string;
    RemoteTCPPort: number;
    RemoteTCPRandom: boolean; //NOTE: don't have null
    LocalURL: string;
    LocalTimeout: string;
    UseLocalAsHTTPHost: boolean;
  }

  export interface Options {
    Config: string;
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
    HostPrefix: string[];
    RemoteTCPPort: number[];
    RemoteTCPRandom: (boolean | null)[];
    Local: string[];
    LocalTimeout: string[];
    UseLocalAsHTTPHost: boolean[];
    SentryDSN: string;
    SentryLevel: string[];
    SentrySampleRate: number;
    SentryRelease: string;
    SentryEnvironment: string;
    SentryServerName: string;
    SentryDebug: boolean;
    WebRTCConnectionIdleTimeout: string;
    WebRTCLogLevel: string;
    WebRTCMinPort: number;
    WebRTCMaxPort: number;
    TCPForwardAddr: string;
    TCPForwardHostPrefix: string;
    TCPForwardConnections: number;
    LogFile: string;
    LogFileMaxSize: number;
    LogFileMaxCount: number;
    LogLevel: string;
    Version: boolean;
  }
}
