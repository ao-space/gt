<template>
  <GeneralSetting :setting="generalSetting" />
  <NetworkSetting :setting="netWorkSetting" />
  <SecuritySetting :setting="securitySetting" />
  <el-button type="primary" @click="submit">Submit</el-button>
</template>

<script setup lang="ts" name="ServerConfigForm">
import { reactive } from "vue";
import yaml from "js-yaml";
import GeneralSetting from "./components/GeneralSetting.vue";
import NetworkSetting from "./components/NetworkSetting.vue";
import SecuritySetting from "./components/SecuritySetting.vue";
import { ServerConfig } from "./interface";
const tcps = reactive<ServerConfig.TCP[]>([
  {
    Range: "1-100",
    Number: 100
  },
  {
    Range: "2-200",
    Number: 150
  }
]);
const host = reactive<ServerConfig.Host>({
  Number: 1,
  RegexStr: [".*", "http.*", "tcp://"],
  WithID: false
});
const hostInOptions = reactive<ServerConfig.HostInOptions>({
  HostNumber: host.Number,
  HostRegex: host.RegexStr,
  HostWithID: host.WithID
});
const users = reactive<Record<string, ServerConfig.User>>({
  id1: {
    Secret: "secret1",
    TCPs: tcps,
    Host: host,
    Speed: 100,
    Connections: 100
  },
  id2: {
    Secret: "secret2",
    TCPs: tcps,
    Host: host,
    Speed: 100,
    Connections: 100
  }
});
const generalSetting = reactive<ServerConfig.GeneralSetting>({
  // ...ServerConfig.defaultGeneralSetting
  Users: "",
  AuthAPI: "",
  TCPRanges: tcps.map(item => item.Range),
  TCPNumbers: tcps.map(item => item.Number.toString()),
  ...hostInOptions
});
const netWorkSetting = reactive<ServerConfig.NetworkSetting>({
  Addr: "ao.space",
  TLSAddr: "asdf",
  TLSMinVersion: "tls1.1",
  STUNAddr: "23",
  SNIAddr: "31234",
  HTTPMUXHeader: "dsfasd"
});
const securitySetting = reactive<ServerConfig.SecuritySetting>({
  CertFile: "sdfa",
  KeyFile: "sdf",
  AllowAnyClient: false
});
const connectionsSetting = reactive<ServerConfig.ConnectionSetting>({
  Speed: 123,
  Connections: 324,
  ReconnectTimes: 23,
  ReconnectDuration: "323s",
  Timeout: "23s",
  TimeoutOnUnidirectionalTraffic: false
});
const apiSetting = reactive<ServerConfig.APISetting>({
  APIAddr: ":8080",
  APICertFile: "sdf",
  APIKeyFile: "dsaf",
  APITLSMinVersion: "tls1.1"
});
const sentrySetting = reactive<ServerConfig.SentrySetting>({
  SentryDSN: "sadf",
  SentryLevel: ["error", "fatal", "panic"],
  SentrySampleRate: 0.2,
  SentryRelease: "sdf",
  SentryEnvironment: "DSF",
  SentryServerName: "SAdf",
  SentryDebug: false
});
const logSetting = reactive<ServerConfig.LogSetting>({
  LogFile: "adsf",
  LogFileMaxSize: 23,
  LogFileMaxCount: 213,
  LogLevel: "trace"
});
const options = reactive<ServerConfig.Options>({
  ...generalSetting,
  ...netWorkSetting,
  ...connectionsSetting,
  ...apiSetting,
  ...securitySetting,
  ...logSetting,
  ...sentrySetting
});
const serverConfig = reactive<ServerConfig.Config>({
  Version: "",
  Users: users,
  TCPs: tcps,
  Host: host,
  Options: options
});
const submit = () => {
  yaml.dump(serverConfig);
};
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
