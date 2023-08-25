import { Config } from "@/api/interface";
import { ClientConfig } from "@/components/ClientConfigForm/interface";
export const mapClientGeneralSetting = (data: Config.Client.ResConfig): ClientConfig.GeneralSetting => ({
  ID: data.config.ID,
  Secret: data.config.Secret,
  ReconnectDelay: data.config.ReconnectDelay,
  Remote: data.config.Remote,
  RemoteSTUN: data.config.RemoteSTUN,
  RemoteAPI: data.config.RemoteAPI,
  RemoteCert: data.config.RemoteCert,
  RemoteCertInsecure: data.config.RemoteCertInsecure,
  RemoteConnections: data.config.RemoteConnections,
  RemoteIdleConnections: data.config.RemoteIdleConnections,
  RemoteTimeout: data.config.RemoteTimeout
});
export const mapClientSentrySetting = (data: Config.Client.ResConfig): ClientConfig.SentrySetting => ({
  SentryDSN: data.config.SentryDSN,
  SentryLevel: data.config.SentryLevel,
  SentrySampleRate: data.config.SentrySampleRate,
  SentryRelease: data.config.SentryRelease,
  SentryEnvironment: data.config.SentryEnvironment,
  SentryServerName: data.config.SentryServerName,
  SentryDebug: data.config.SentryDebug
});
export const mapClientWebRTCSetting = (data: Config.Client.ResConfig): ClientConfig.WebRTCSetting => ({
  WebRTCConnectionIdleTimeout: data.config.WebRTCConnectionIdleTimeout,
  WebRTCLogLevel: data.config.WebRTCLogLevel,
  WebRTCMinPort: data.config.WebRTCMinPort,
  WebRTCMaxPort: data.config.WebRTCMaxPort
});
export const mapClientTCPForwardSetting = (data: Config.Client.ResConfig): ClientConfig.TCPForwardSetting => ({
  TCPForwardAddr: data.config.TCPForwardAddr,
  TCPForwardHostPrefix: data.config.TCPForwardHostPrefix,
  TCPForwardConnections: data.config.TCPForwardConnections
});
export const mapClientLogSetting = (data: Config.Client.ResConfig): ClientConfig.LogSetting => ({
  LogFile: data.config.LogFile,
  LogFileMaxSize: data.config.LogFileMaxSize,
  LogFileMaxCount: data.config.LogFileMaxCount,
  LogLevel: data.config.LogLevel
});
export const mapClientServices = (data: Config.Client.ResConfig) =>
  data.config.Services.map(service => ({
    HostPrefix: service.HostPrefix,
    RemoteTCPPort: service.RemoteTCPPort,
    RemoteTCPRandom: service.RemoteTCPRandom,
    LocalURL: service.LocalURL.Host,
    LocalTimeout: service.LocalTimeout,
    UseLocalAsHTTPHost: service.UseLocalAsHTTPHost
  }));
