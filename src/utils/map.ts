import { Config, Connection } from "@/api/interface";
import { ClientConfig } from "@/components/ClientConfigForm/interface";
export const mapClientGeneralSetting = (data: Config.Client.ResConfig): ClientConfig.GeneralSetting => ({
  ID: data.config.ID,
  Secret: data.config.Secret,
  ReconnectDelay: humanizeDuration(data.config.ReconnectDelay),
  Remote: data.config.Remote,
  RemoteSTUN: data.config.RemoteSTUN,
  RemoteAPI: data.config.RemoteAPI,
  RemoteCert: data.config.RemoteCert,
  RemoteCertInsecure: data.config.RemoteCertInsecure,
  RemoteConnections: data.config.RemoteConnections,
  RemoteIdleConnections: data.config.RemoteIdleConnections,
  RemoteTimeout: humanizeDuration(data.config.RemoteTimeout)
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
  WebRTCConnectionIdleTimeout: humanizeDuration(data.config.WebRTCConnectionIdleTimeout),
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
export const mapClientServices = (data: Config.Client.ResConfig) => {
  if (!data.config.Services) {
    return [];
  } else {
    return data.config.Services.map(service => ({
      HostPrefix: service.HostPrefix,
      RemoteTCPPort: service.RemoteTCPPort,
      RemoteTCPRandom: service.RemoteTCPRandom,
      LocalURL: service.LocalURL,
      LocalTimeout: humanizeDuration(service.LocalTimeout),
      UseLocalAsHTTPHost: service.UseLocalAsHTTPHost
    }));
  }
};
const humanizeDuration = (value: string): string => {
  if (!value) return "";
  const regex = /^(?:\d+(?:ns|µ?s|ms|[smh]))+$/;
  if (regex.test(value)) return value;
  const units = [
    { label: "h", value: 60 * 60 * 1_000_000_000 },
    { label: "m", value: 60 * 1_000_000_000 },
    { label: "s", value: 1_000_000_000 },
    { label: "ms", value: 1_000_000 },
    { label: "µs", value: 1_000 },
    { label: "ns", value: 1 }
  ];
  let remaining = parseInt(value, 10);
  let result = "";
  for (const uint of units) {
    if (remaining >= uint.value) {
      const count = Math.floor(remaining / uint.value);
      result += `${count}${uint.label}`;
      remaining -= count * uint.value;
    }
  }
  return result;
};

export const convertToStatus = (statusCode: Connection.Status): string => {
  return Connection.StatusMap[statusCode];
};
