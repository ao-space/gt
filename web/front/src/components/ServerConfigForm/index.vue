<template>
  <el-row>
    <el-text class="setting_class">{{ $t("sconfig.BasicSettings") }}</el-text>
  </el-row>
  <Anchor :tab-list="tabList">
    <template v-for="tab in staticBasicTabs" :key="tab.uuid" #[tab.uuid]>
      <component
        :id="tab.name"
        :is="tab.component"
        :ref="(el: InstanceType<typeof tab.component> | null) => tab.ref = el"
        :setting="tab.setting"
        @update:setting="tab.updateSetting"
      />
    </template>
    <template v-for="(tab, index) in dynamicTabs" :key="tab.uuid" #[tab.uuid]>
      <component
        :id="tab.name"
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
  <el-row>
    <el-text class="setting_class">{{ $t("sconfig.OptionSettings") }}</el-text>
  </el-row>
  <Anchor :tab-list="tabList">
    <template v-for="tab in staticOptionTab" :key="tab.uuid" #[tab.uuid]>
      <component
        :id="tab.name"
        :is="tab.component"
        :ref="(el: InstanceType<typeof tab.component> | null) => tab.ref = el"
        :setting="tab.setting"
        @update:setting="tab.updateSetting"
      />
    </template>
  </Anchor>
  <el-button type="primary" @click="submit">{{ $t("sconfig.Submit") }}</el-button>
  <el-button type="primary" @click="getFromFile">{{ $t("sconfig.GetFromFile") }}</el-button>
</template>

<script setup lang="ts" name="ServerConfigForm">
import { Ref, markRaw, reactive, ref, watch, watchEffect, computed, onMounted } from "vue";
import { ServerConfig } from "./interface";
import { v4 as uuidv4 } from "uuid";
import Anchor, { Tab } from "@/components/Anchor/index.vue";
import GeneralSetting from "./components/GeneralSetting.vue";
import NetworkSetting from "./components/NetworkSetting.vue";
import SecuritySetting from "./components/SecuritySetting.vue";
import ConnectionSetting from "./components/ConnectionSetting.vue";
import APISetting from "./components/APISetting.vue";
import UserSetting from "./components/UserSetting.vue";
// Note that the following imports is in ClientConfigForm/components
import SentrySetting from "@/components/ClientConfigForm/components/SentrySetting.vue";
import LogSetting from "@/components/ClientConfigForm/components/LogSetting.vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { getServerConfigFromFileApi, saveServerConfigApi } from "@/api/modules/serverConfig";
import { Config } from "@/api/interface";
import {
  mapServerGeneralSetting,
  mapServerAPISetting,
  mapServerConnectionSetting,
  mapServerHostSetting,
  mapServerLogSetting,
  mapServerNetworkSetting,
  mapServerSecuritySetting,
  mapServerSentrySetting,
  mapServerTCPSetting,
  mapServerUserSetting
} from "@/utils/map";
import cloneDeep from "lodash/cloneDeep";
import i18n from "@/languages";
import { useMetadataStore } from "@/stores/modules/metadata";

//Global Setting
const tcps = reactive<ServerConfig.TCP[]>([]);
const host = reactive<ServerConfig.Host>({ ...ServerConfig.getDefaultHostSetting() });

//users is used for serverConfig
//userList is used for child component -- UserSetting
const users = reactive<Record<string, ServerConfig.User>>({});
const userList = reactive<ServerConfig.UserSetting[]>([{ ...ServerConfig.getDefaultUserSetting() }]);
const uuids = reactive<string[]>([uuidv4()]); // record the uuid of each user

// the sync from userList to users is happened in submit function,
// to avoid the id conflict
watch(
  () => users,
  newUsers => {
    userList.splice(0, userList.length);
    userSettingRefs.splice(0, userSettingRefs.length);
    uuids.splice(0, uuids.length);
    Object.keys(newUsers).forEach(key => {
      userList.push({
        ID: key,
        Secret: newUsers[key].Secret,
        TCPs: cloneDeep(newUsers[key].TCPs),
        TCPNumber: newUsers[key].TCPNumber,
        Host: cloneDeep(newUsers[key].Host),
        Speed: newUsers[key].Speed,
        Connections: newUsers[key].Connections
      });
      userSettingRefs.push(ref<InstanceType<typeof UserSetting> | null>(null));
      uuids.push(uuidv4());
    });
  },
  { deep: true }
);

//generalSetting is used for serverConfig
//generalSettingProps is used for child component -- GeneralSetting
const generalSetting = reactive<ServerConfig.GeneralSetting>({ ...ServerConfig.defaultGeneralSetting });
const generalSettingProps = reactive<ServerConfig.GeneralSettingProps>({
  ...generalSetting,
  //use shallow clone to avoid sync in the current component
  TCPs: tcps, //global Setting
  Host: host //global Setting
});

//Sync generalSetting with generalSettingProps
watch(
  () => generalSetting,
  newSetting => {
    Object.assign(generalSettingProps, newSetting);
  },
  { deep: true }
);
watch(
  () => generalSettingProps,
  newSetting => {
    generalSetting.UserPath = newSetting.UserPath;
    generalSetting.AuthAPI = newSetting.AuthAPI;
    generalSetting.TCPNumber = newSetting.TCPNumber;
  },
  { deep: true }
);

const netWorkSetting = reactive<ServerConfig.NetworkSetting>({ ...ServerConfig.defaultNetworkSetting });
const securitySetting = reactive<ServerConfig.SecuritySetting>({ ...ServerConfig.defaultSecuritySetting });
const connectionsSetting = reactive<ServerConfig.ConnectionSetting>({ ...ServerConfig.defaultConnectionSetting });
const apiSetting = reactive<ServerConfig.APISetting>({ ...ServerConfig.defaultAPISetting });
const sentrySetting = reactive<ServerConfig.SentrySetting>({ ...ServerConfig.defaultSentrySetting });
const logSetting = reactive<ServerConfig.LogSetting>({ ...ServerConfig.defaultLogSetting });
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

//Sync options with other settings
watchEffect(() => {
  Object.assign(options, {
    ...generalSetting,
    ...netWorkSetting,
    ...connectionsSetting,
    ...apiSetting,
    ...securitySetting,
    ...logSetting,
    ...sentrySetting
  });
});

//Sync: users, tcps, host, options -> serverConfig
watchEffect(() => {
  Object.assign(serverConfig, {
    Users: users,
    TCPs: tcps,
    Host: host,
    ...options
  });
});

//Sync with child component
const updateGeneralSetting = (newSetting: ServerConfig.GeneralSettingProps) => {
  console.log("updateGeneralSetting");
  generalSettingProps.UserPath = newSetting.UserPath;
  generalSettingProps.AuthAPI = newSetting.AuthAPI;
  generalSettingProps.TCPNumber = newSetting.TCPNumber;
  tcps.splice(0, tcps.length, ...newSetting.TCPs);
  host.Number = newSetting.Host.Number;
  host.RegexStr.splice(0, host.RegexStr.length, ...newSetting.Host.RegexStr);
  host.WithID = newSetting.Host.WithID;
  // Object.assign(generalSettingProps, newSetting);
};
const updateNetworkSetting = (newSetting: ServerConfig.NetworkSetting) => {
  Object.assign(netWorkSetting, newSetting);
};
const updateSecuritySetting = (newSetting: ServerConfig.SecuritySetting) => {
  Object.assign(securitySetting, newSetting);
};
const updateConnectionSetting = (newSetting: ServerConfig.ConnectionSetting) => {
  Object.assign(connectionsSetting, newSetting);
};
const updateAPISetting = (newSetting: ServerConfig.APISetting) => {
  Object.assign(apiSetting, newSetting);
};
const updateSentrySetting = (newSetting: ServerConfig.SentrySetting) => {
  Object.assign(sentrySetting, newSetting);
};
const updateLogSetting = (newSetting: ServerConfig.LogSetting) => {
  Object.assign(logSetting, newSetting);
};

//From Related
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
  userList.push({ ...ServerConfig.getDefaultUserSetting() });
  userSettingRefs.push(ref<InstanceType<typeof UserSetting> | null>(null));
  uuids.push(uuidv4());
};
const removeUser = (index: number) => {
  if (userList.length === 1) {
    ElMessage.warning("Can't delete the last user.");
    return;
  }
  userList.splice(index, 1);
  userSettingRefs.splice(index, 1);
  uuids.splice(index, 1);
};

const updateUserSetting = (index: number, newSetting: ServerConfig.UserSetting) => {
  userList[index].Connections = newSetting.Connections;
  userList[index].Host = newSetting.Host;
  userList[index].ID = newSetting.ID;
  userList[index].Secret = newSetting.Secret;
  userList[index].Speed = newSetting.Speed;
  userList[index].TCPs.splice(0, userList[index].TCPs.length, ...newSetting.TCPs);
  userList[index].TCPNumber = newSetting.TCPNumber;
};

interface staticTabsType<T> {
  title: string;
  name: string;
  uuid: string;
  component: any;
  ref: Ref;
  setting: T;
  updateSetting: (newSetting: T) => void;
}
const staticBasicTabs = reactive([
  {
    title: i18n.global.t("sconfig.GeneralSetting"),
    name: "GeneralSetting",
    uuid: uuidv4(),
    component: markRaw(GeneralSetting),
    ref: generalSettingRef,
    setting: generalSettingProps,
    updateSetting: updateGeneralSetting
  } as staticTabsType<ServerConfig.GeneralSettingProps>,
  {
    title: i18n.global.t("sconfig.NetworkSetting"),
    name: "NetworkSetting",
    uuid: uuidv4(),
    component: markRaw(NetworkSetting),
    ref: networkSettingRef,
    setting: netWorkSetting,
    updateSetting: updateNetworkSetting
  } as staticTabsType<ServerConfig.NetworkSetting>,
  {
    title: i18n.global.t("sconfig.SecuritySetting"),
    name: "SecuritySetting",
    uuid: uuidv4(),
    component: markRaw(SecuritySetting),
    ref: securitySettingRef,
    setting: securitySetting,
    updateSetting: updateSecuritySetting
  } as staticTabsType<ServerConfig.SecuritySetting>
]);
const staticOptionTab = reactive([
  {
    title: i18n.global.t("sconfig.ConnectionSetting"),
    name: "ConnectionSetting",
    uuid: uuidv4(),
    component: markRaw(ConnectionSetting),
    ref: connectionSettingRef,
    setting: connectionsSetting,
    updateSetting: updateConnectionSetting
  } as staticTabsType<ServerConfig.ConnectionSetting>,
  {
    title: i18n.global.t("sconfig.APISetting"),
    name: "APISetting",
    uuid: uuidv4(),
    component: markRaw(APISetting),
    ref: apiSettingRef,
    setting: apiSetting,
    updateSetting: updateAPISetting
  } as staticTabsType<ServerConfig.APISetting>,
  {
    title: i18n.global.t("sconfig.SentrySetting"),
    name: "SentrySetting",
    uuid: uuidv4(),
    component: markRaw(SentrySetting),
    ref: sentrySettingRef,
    setting: sentrySetting,
    updateSetting: updateSentrySetting
  } as staticTabsType<ServerConfig.SentrySetting>,
  {
    title: i18n.global.t("sconfig.LogSetting"),
    name: "LogSetting",
    uuid: uuidv4(),
    component: markRaw(LogSetting),
    ref: logSettingRef,
    setting: logSetting,
    updateSetting: updateLogSetting
  } as staticTabsType<ServerConfig.LogSetting>
]);
let metadataStore = useMetadataStore();
metadataStore.$subscribe(() => {
  staticBasicTabs.map(value => {
    value.title = i18n.global.t("sconfig." + value.name);
  });
});
metadataStore.$subscribe(() => {
  staticOptionTab.map(value => {
    value.title = i18n.global.t("sconfig." + value.name);
  });
});
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

const dynamicTabs = computed<dynamicTabType<ServerConfig.UserSetting>[]>(() => {
  return userList.map((user, index) => ({
    title: i18n.global.t("sconfig.User") + `${index + 1}` + i18n.global.t("sconfig.Setting"),
    name: `User${index + 1}Setting`,
    uuid: uuids[index],
    component: markRaw(UserSetting),
    setting: userList[index],
    updateSetting: updateUserSetting,
    index: index,
    isLast: index === userList.length - 1
  }));
});

const tabList = computed<Tab[]>(() => [
  ...staticBasicTabs.map(tab => ({ title: tab.title, name: tab.name, uuid: tab.uuid })),
  ...staticOptionTab.map(tab => ({ title: tab.title, name: tab.name, uuid: tab.uuid })),
  ...dynamicTabs.value.map(tab => ({ title: tab.title, name: tab.name, uuid: tab.uuid }))
]);

const validateAllForms = (formRefs: Array<Ref<ServerConfig.FormRef | null>>) => {
  return Promise.all(formRefs.map(formRef => formRef.value?.validateForm()));
};

//update the data with the response from server
const updateData = (data: Config.Server.ResConfig) => {
  Object.assign(generalSetting, mapServerGeneralSetting(data));
  Object.assign(netWorkSetting, mapServerNetworkSetting(data));
  Object.assign(connectionsSetting, mapServerConnectionSetting(data));
  Object.assign(apiSetting, mapServerAPISetting(data));
  Object.assign(securitySetting, mapServerSecuritySetting(data));
  Object.assign(logSetting, mapServerLogSetting(data));
  Object.assign(sentrySetting, mapServerSentrySetting(data));
  Object.assign(host, mapServerHostSetting(data));
  tcps.splice(0, tcps.length, ...mapServerTCPSetting(data));
  clearUsers();
  Object.assign(users, mapServerUserSetting(data));
};

const checkIDConflict = () => {
  const ids = userList.map(user => user.ID);
  const uniqueIDs = new Set(ids);
  if (uniqueIDs.size !== ids.length) {
    throw new Error(i18n.global.t("sconfig.IDConflictError"));
  }
};

const clearUsers = () => {
  for (const key of Object.keys(users)) {
    delete users[key];
  }
};

const updateUsersFormUserList = () => {
  userList.forEach(user => {
    users[user.ID] = {
      Secret: user.Secret,
      TCPs: user.TCPs,
      TCPNumber: user.TCPNumber,
      Host: user.Host,
      Speed: user.Speed,
      Connections: user.Connections
    };
  });
};

const submit = async () => {
  try {
    await ElMessageBox.confirm(i18n.global.t("sconfig.SaveConfigConfirm"), i18n.global.t("sconfig.SaveConfigTitle"), {
      confirmButtonText: i18n.global.t("sconfig.SaveConfigConfirmBtn"),
      cancelButtonText: i18n.global.t("sconfig.SaveConfigCancelBtn"),
      type: "info"
    });
    await validateAllForms([
      generalSettingRef,
      networkSettingRef,
      securitySettingRef,
      connectionSettingRef,
      apiSettingRef,
      sentrySettingRef,
      logSettingRef,
      ...userSettingRefs
    ]);
    checkIDConflict();
    clearUsers();
    updateUsersFormUserList();
    await saveServerConfigApi(serverConfig);
    ElMessage.success(i18n.global.t("sconfig.SubmitSuccess"));
  } catch (e) {
    if (e instanceof Error) {
      ElMessage.error(e.message);
    } else {
      ElMessage.error(i18n.global.t("sconfig.FailedToSaveConfig"));
    }
  }
};

const getFromFile = async () => {
  try {
    await ElMessageBox.confirm(i18n.global.t("sconfig.GetFromFileConfirm"), i18n.global.t("sconfig.GetFromFileTitle"), {
      confirmButtonText: i18n.global.t("sconfig.GetFromFileConfirmBtn"),
      cancelButtonText: i18n.global.t("sconfig.GetFromFileCancelBtn"),
      type: "info"
    });
    const { data } = await getServerConfigFromFileApi();
    updateData(data);
    ElMessage.success(i18n.global.t("sconfig.GetFromFileSuccess"));
  } catch (e) {
    if (e instanceof Error) {
      ElMessage.error(e.message);
    } else {
      ElMessage.error(i18n.global.t("sconfig.FailedToGetFromFile"));
    }
  }
};

onMounted(() => {
  getFromFile();
});
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
