<template>
  <Anchor :tab-list="tabList">
    <template v-for="tab in staticTabs" :key="tab.uuid" #[tab.uuid]>
      <component :is="tab.component" :ref="tab.ref" :setting="tab.setting" @update:setting="tab.updateSetting" />
    </template>
    <!-- <GeneralSetting :setting="generalSetting" />
    <NetworkSetting :setting="netWorkSetting" />
    <SecuritySetting :setting="securitySetting" />
    <ConnectionSetting :setting="connectionsSetting" />
    <APISetting :setting="apiSetting" />
    <SentrySetting :setting="sentrySetting" />
    <LogSetting :setting="logSetting" /> -->
    <!-- <UserSetting :index="0" :is-last="true" :setting="userSetting" /> -->
    <template v-for="(tab, index) in dynamicTabs" :key="tab.uuid" #[tab.uuid]>
      <component
        :is="tab.component"
        :ref="userSettingRefs[index]"
        :index="index"
        :is-last="tab.isLast"
        :setting="tab.setting"
        @add-user="addUser"
        @remove-user="removeUser(index)"
        @update:setting="tab.updateSetting"
      />
    </template>
    <el-button type="primary" @click="submit">Submit</el-button>
  </Anchor>
</template>

<script setup lang="ts" name="ServerConfigForm">
import { Ref, markRaw, reactive, ref } from "vue";
import { ServerConfig } from "./interface";
import { v4 as uuidv4 } from "uuid";
import yaml from "js-yaml";
import Anchor, { Tab } from "@/components/Anchor/index.vue";
import GeneralSetting from "./components/GeneralSetting.vue";
import NetworkSetting from "./components/NetworkSetting.vue";
import SecuritySetting from "./components/SecuritySetting.vue";
import ConnectionSetting from "./components/ConnectionSetting.vue";
import APISetting from "./components/APISetting.vue";
import UserSetting from "./components/UserSetting.vue";

//TODO: move location
import SentrySetting from "@/components/ClientConfigForm/components/SentrySetting.vue";
import LogSetting from "../ClientConfigForm/components/LogSetting.vue";
import { ElMessage, ElMessageBox } from "element-plus";

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
  }
});
const userList = reactive<ServerConfig.UserSetting[]>([
  {
    ID: Object.keys(users)[0],
    Secret: users[Object.keys(users)[0]].Secret,
    TCPs: users[Object.keys(users)[0]].TCPs,
    Host: users[Object.keys(users)[0]].Host,
    Speed: users[Object.keys(users)[0]].Speed,
    Connections: users[Object.keys(users)[0]].Connections
  }
]);

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

// can ?
const updateGeneralSetting = (newSetting: ServerConfig.GeneralSetting) => {
  console.log("updateGeneralSetting", newSetting);
  Object.assign(generalSetting, newSetting);
};
const updateNetworkSetting = (newSetting: ServerConfig.NetworkSetting) => {
  console.log("updateNetworkSetting", newSetting);
  Object.assign(netWorkSetting, newSetting);
};
const updateSecuritySetting = (newSetting: ServerConfig.SecuritySetting) => {
  console.log("updateSecuritySetting", newSetting);
  Object.assign(securitySetting, newSetting);
};
const updateConnectionSetting = (newSetting: ServerConfig.ConnectionSetting) => {
  console.log("updateConnectionSetting", newSetting);
  Object.assign(connectionsSetting, newSetting);
};
const updateAPISetting = (newSetting: ServerConfig.APISetting) => {
  console.log("updateAPISetting", newSetting);
  Object.assign(apiSetting, newSetting);
};
const updateSentrySetting = (newSetting: ServerConfig.SentrySetting) => {
  console.log("updateSentrySetting", newSetting);
  Object.assign(sentrySetting, newSetting);
};
const updateLogSetting = (newSetting: ServerConfig.LogSetting) => {
  console.log("updateLogSetting", newSetting);
  Object.assign(logSetting, newSetting);
};

const generalSettingRef = ref<InstanceType<typeof GeneralSetting> | null>(null);
const networkSettingRef = ref<InstanceType<typeof NetworkSetting> | null>(null);
const securitySettingRef = ref<InstanceType<typeof SecuritySetting> | null>(null);
const connectionSettingRef = ref<InstanceType<typeof ConnectionSetting> | null>(null);
const apiSettingRef = ref<InstanceType<typeof APISetting> | null>(null);
const sentrySettingRef = ref<InstanceType<typeof SentrySetting> | null>(null);
const logSettingRef = ref<InstanceType<typeof LogSetting> | null>(null);
const userSettingRef = ref<InstanceType<typeof UserSetting> | null>(null);
const userSettingRefs = reactive<Ref<InstanceType<typeof UserSetting> | null>[]>([userSettingRef]);

const addUser = () => {
  userList.push({ ...ServerConfig.defaultUserSetting, ID: "id" + (userList.length + 1).toString() });
  userSettingRefs.push(ref<InstanceType<typeof UserSetting> | null>(null));
  const uuid = uuidv4();

  dynamicTabs[dynamicTabs.length - 1].isLast = false;
  dynamicTabs.push({
    title: `User ${userList.length} Setting`,
    name: `User${userList.length}Setting`,
    uuid: uuid,
    component: markRaw(UserSetting),
    setting: userList[userList.length - 1],
    updateSetting: updateUserSetting,
    index: userList.length - 1,
    isLast: true
  });

  tabList.push({
    title: `User ${userList.length} Setting`,
    name: `User${userList.length}Setting`,
    uuid: uuid
  });
};
const removeUser = (index: number) => {
  if (userList.length === 1) {
    ElMessage.warning("至少需要一个用户");
    return;
  }
  userList.splice(index, 1);
  userSettingRefs.splice(index, 1);

  dynamicTabs.splice(index, 1);
  for (let i = index; i < dynamicTabs.length - 1; i++) {
    dynamicTabs[i].title = `User ${i + 1} Setting`;
    dynamicTabs[i].name = `User${i + 1}Setting`;
    dynamicTabs[i].index = i;
  }
  dynamicTabs[dynamicTabs.length - 1].isLast = true;

  let tabListIndex = tabList.findIndex(item => item.title === `User ${index + 1} Setting`);
  tabList.splice(tabListIndex, 1);
  for (let i = index; i < userList.length; i++) {
    tabList[tabListIndex + i - index].title = `User ${i + 1} Setting`;
    tabList[tabListIndex + i - index].name = `User${i + 1}Setting`;
  }
};
//TODO : IDchange 先创建 多个id 有问题key值一样
const updateUserSetting = (index: number, newSetting: ServerConfig.UserSetting) => {
  if (0 <= index && index < userList.length) {
    const oldId = Object.keys(users)[index];
    const newId = newSetting.ID;

    if (oldId !== newId) {
      delete users[oldId];
    }

    users[newId] = {
      Secret: newSetting.Secret,
      TCPs: newSetting.TCPs,
      Host: newSetting.Host,
      Speed: newSetting.Speed,
      Connections: newSetting.Connections
    };

    userList.splice(index, 1, newSetting);
  } else {
    //ignore the delete operation
  }
};

interface staticTabsType<T> {
  title: string;
  name: string;
  uuid: string;
  component: any;
  ref: string;
  setting: T;
  updateSetting: (newSetting: T) => void;
}
const staticTabs = reactive([
  {
    title: "General Setting",
    name: "GeneralSetting",
    uuid: uuidv4(),
    component: markRaw(GeneralSetting),
    ref: "generalSettingRef",
    setting: generalSetting,
    updateSetting: updateGeneralSetting
  } as staticTabsType<ServerConfig.GeneralSetting>,
  {
    title: "Network Setting",
    name: "NetworkSetting",
    uuid: uuidv4(),
    component: markRaw(NetworkSetting),
    ref: "networkSettingRef",
    setting: netWorkSetting,
    updateSetting: updateNetworkSetting
  } as staticTabsType<ServerConfig.NetworkSetting>,
  {
    title: "Security Setting",
    name: "SecuritySetting",
    uuid: uuidv4(),
    component: markRaw(SecuritySetting),
    ref: "securitySettingRef",
    setting: securitySetting,
    updateSetting: updateSecuritySetting
  } as staticTabsType<ServerConfig.SecuritySetting>,
  {
    title: "Connection Setting",
    name: "ConnectionSetting",
    uuid: uuidv4(),
    component: markRaw(ConnectionSetting),
    ref: "connectionSettingRef",
    setting: connectionsSetting,
    updateSetting: updateConnectionSetting
  } as staticTabsType<ServerConfig.ConnectionSetting>,
  {
    title: "API Setting",
    name: "APISetting",
    uuid: uuidv4(),
    component: markRaw(APISetting),
    ref: "apiSettingRef",
    setting: apiSetting,
    updateSetting: updateAPISetting
  } as staticTabsType<ServerConfig.APISetting>,
  {
    title: "Sentry Setting",
    name: "SentrySetting",
    uuid: uuidv4(),
    component: markRaw(SentrySetting),
    ref: "sentrySettingRef",
    setting: sentrySetting,
    updateSetting: updateSentrySetting
  } as staticTabsType<ServerConfig.SentrySetting>,
  {
    title: "Log Setting",
    name: "LogSetting",
    uuid: uuidv4(),
    component: markRaw(LogSetting),
    ref: "logSettingRef",
    setting: logSetting,
    updateSetting: updateLogSetting
  } as staticTabsType<ServerConfig.LogSetting>
]);
interface dynamicTabType<T> {
  title: string;
  name: string;
  uuid: string;
  component: any;
  setting: T;
  updateSetting: (index: number, newSetting: T) => void;
  index: number;
  isLast: boolean;
}
const dynamicTabs = reactive<dynamicTabType<ServerConfig.UserSetting>[]>([
  {
    title: "User 1 Setting",
    name: "User1Setting",
    uuid: uuidv4(),
    component: markRaw(UserSetting),
    setting: userList[0],
    updateSetting: updateUserSetting,
    index: 0,
    isLast: true
  }
]);

const tabList = reactive<Tab[]>([
  ...staticTabs.map(tab => ({ title: tab.title, name: tab.name, uuid: tab.uuid })),
  ...dynamicTabs.map(tab => ({ title: tab.title, name: tab.name, uuid: tab.uuid }))
]);
const validateAllForms = (formRefs: Array<Ref<ServerConfig.FormRef | null>>) => {
  return Promise.all(formRefs.map(formRef => formRef.value?.validateForm()));
};

const submit = () => {
  console.log("submit");
  yaml.dump(serverConfig);
  ElMessageBox.confirm("确认提交吗？", "提示", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning"
  })
    .then(() => {
      validateAllForms([
        generalSettingRef,
        networkSettingRef,
        securitySettingRef,
        connectionSettingRef,
        apiSettingRef,
        sentrySettingRef,
        logSettingRef,
        ...userSettingRefs
      ])
        .then(() => {
          console.log("submit success");
          ElMessage.success("提交成功");
        })
        .catch(() => {
          console.log("submit fail");
          ElMessage.error("提交失败");
        });
    })
    .catch(() => {
      console.log("cancel submit");
    });
};
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
