<template>
  <Anchor :tab-list="tabList">
    <template v-for="tab in staticTabs" :key="tab.uuid" #[tab.uuid]>
      <component :is="tab.component" :ref="tab.ref" :setting="tab.setting" @update:setting="tab.updateSetting" />
    </template>
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
  </Anchor>
  <el-button type="primary" @click="submit">Submit</el-button>
  <el-button type="primary" @click="getFromFile">Get From File</el-button>
  <el-button type="primary" @click="getFromRunning">Get From Running</el-button>
  <el-button type="primary" @click="save">Save</el-button>
  <el-button type="primary" @click="test">Test</el-button>
</template>

<script setup lang="ts" name="ServerConfigForm">
import { Ref, markRaw, reactive, ref, watch } from "vue";
import { ServerConfig } from "./interface";
import { v4 as uuidv4 } from "uuid";
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
import { getRunningServerConfigApi, getServerConfigFromFileApi, saveServerConfigApi } from "@/api/modules/serverConfig";
import { Config } from "@/api/interface";

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

// the sync from userList to users is happened in submit function to avoid the id conflict
watch(
  users,
  newUsers => {
    console.log("users change");
    userList.splice(0, userList.length);
    Object.keys(newUsers).forEach(key => {
      userList.push({
        ID: key,
        Secret: newUsers[key].Secret,
        TCPs: newUsers[key].TCPs,
        Host: newUsers[key].Host,
        Speed: newUsers[key].Speed,
        Connections: newUsers[key].Connections
      });
    });
  },
  { deep: true }
);

const generalSetting = reactive<ServerConfig.GeneralSetting>({
  UserPath: "",
  AuthAPI: ""
});
const generalSettingProps = reactive<ServerConfig.GeneralSettingProps>({
  ...generalSetting,
  TCPs: tcps,
  Host: host
});

watch(
  () => generalSetting,
  newSetting => {
    console.log("generalSetting change");
    Object.assign(generalSettingProps, newSetting);
  },
  { deep: true }
);
watch(
  () => generalSettingProps,
  newSetting => {
    console.log("generalSettingProps change");
    generalSetting.UserPath = newSetting.UserPath;
    generalSetting.AuthAPI = newSetting.AuthAPI;
  },
  { deep: true }
);

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
  Users: users,
  TCPs: tcps,
  Host: host,
  ...options
});

const updateGeneralSetting = (newSetting: ServerConfig.GeneralSettingProps) => {
  console.log("updateGeneralSetting", newSetting);
  Object.assign(generalSettingProps, newSetting);
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
  // userList.push({ ...ServerConfig.defaultUserSetting, ID: "id" + (userList.length + 1).toString() });
  userList.push({ ...ServerConfig.defaultUserSetting });
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
    ElMessage.warning("Can't delete the last user.");
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

const updateUserSetting = (index: number, newSetting: ServerConfig.UserSetting) => {
  if (0 <= index && index < userList.length) {
    // const oldId = Object.keys(users)[index];
    // const newId = newSetting.ID;

    // if (oldId !== newId) {
    //   delete users[oldId];
    // }

    // users[newId] = {
    //   Secret: newSetting.Secret,
    //   TCPs: newSetting.TCPs,
    //   Host: newSetting.Host,
    //   Speed: newSetting.Speed,
    //   Connections: newSetting.Connections
    // };

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
    setting: generalSettingProps,
    updateSetting: updateGeneralSetting
  } as staticTabsType<ServerConfig.GeneralSettingProps>,
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
const updateData = (data: Config.Server.ResConfig) => {
  console.log(data);
};

const submit = () => {
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
          //check userList id conflict
          const ids = userList.map(user => user.ID);
          const set = new Set(ids);
          if (set.size !== ids.length) {
            ElMessage.error("ID conflict in user setting, please check.");
            return;
          }
          //empty the users
          for (const key of Object.keys(users)) {
            delete users[key];
          }
          //update the users
          userList.forEach(user => {
            users[user.ID] = {
              Secret: user.Secret,
              TCPs: user.TCPs,
              Host: user.Host,
              Speed: user.Speed,
              Connections: user.Connections
            };
          });
          ElMessage.success("Submit success");
        })
        .catch(() => {
          ElMessage.error("Submit failed");
        });
    })
    .catch(() => {
      console.log("cancel submit");
    });
};
const getFromFile = async () => {
  const runningConfig = await getServerConfigFromFileApi();
  console.log(runningConfig);
};
const getFromRunning = async () => {
  const runningConfig = await getRunningServerConfigApi();
  console.log(runningConfig);
};
const save = async () => {
  await saveServerConfigApi(serverConfig);
};
const test = () => {
  console.log(updateData);
};
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
