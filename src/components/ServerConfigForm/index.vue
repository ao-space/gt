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
import { getRunningServerConfigApi, getServerConfigFromFileApi, saveServerConfigApi } from "@/api/modules/serverConfig";
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

//Global Setting
const tcps = reactive<ServerConfig.TCP[]>([]);
const host = reactive<ServerConfig.Host>({ ...ServerConfig.getDefaultHostSetting() });

//users is used for serverConfig
//userList is used for child component -- UserSetting
const users = reactive<Record<string, ServerConfig.User>>({});
const userList = reactive<ServerConfig.UserSetting[]>([{ ...ServerConfig.getDefaultUserSetting() }]);

// the sync from userList to users is happened in submit function,
// to avoid the id conflict
watch(
  users,
  newUsers => {
    console.log("users change", newUsers);
    userList.splice(0, userList.length);
    userSettingRefs.splice(0, userSettingRefs.length);
    Object.keys(newUsers).forEach(key => {
      console.log("process user key", key, newUsers[key]);
      userList.push({
        ID: key,
        Secret: newUsers[key].Secret,
        TCPs: newUsers[key].TCPs,
        Host: newUsers[key].Host,
        Speed: newUsers[key].Speed,
        Connections: newUsers[key].Connections
      });
      userSettingRefs.push(ref<InstanceType<typeof UserSetting> | null>(null));
    });
    console.log("updated userList", userList);
  },
  { deep: true }
);

//generalSetting is used for serverConfig
//generalSettingProps is used for child component -- GeneralSetting
const generalSetting = reactive<ServerConfig.GeneralSetting>({ ...ServerConfig.defaultGeneralSetting });
const generalSettingProps = reactive<ServerConfig.GeneralSettingProps>({
  ...generalSetting,
  TCPs: tcps, //global Setting
  Host: host //global Setting
});

//Sync generalSetting with generalSettingProps
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
  Object.assign(generalSettingProps, newSetting);
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
};
const removeUser = (index: number) => {
  if (userList.length === 1) {
    ElMessage.warning("Can't delete the last user.");
    return;
  }
  userList.splice(index, 1);
  userSettingRefs.splice(index, 1);
};

const updateUserSetting = (index: number, newSetting: ServerConfig.UserSetting) => {
  console.log("updateUserSetting", index, newSetting);
  Object.assign(userList[index], newSetting);
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

const dynamicTabs = computed<dynamicTabType<ServerConfig.UserSetting>[]>(() => {
  return userList.map((user, index) => ({
    title: `User ${index + 1} Setting`,
    name: `User${index + 1}Setting`,
    uuid: uuidv4(),
    component: markRaw(UserSetting),
    setting: userList[index],
    updateSetting: updateUserSetting,
    index: index,
    isLast: index === userList.length - 1
  }));
});

const tabList = computed<Tab[]>(() => [
  ...staticTabs.map(tab => ({ title: tab.title, name: tab.name, uuid: tab.uuid })),
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
  tcps.splice(0, tcps.length, ...mapServerTCPSetting(data));
  Object.assign(host, mapServerHostSetting(data));
  Object.keys(users).forEach(key => {
    delete users[key];
  });
  Object.assign(users, mapServerUserSetting(data));
};

const checkIDConflict = () => {
  const ids = userList.map(user => user.ID);
  const uniqueIDs = new Set(ids);
  if (uniqueIDs.size !== ids.length) {
    throw new Error("ID conflict in user setting, please check.");
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
      Host: user.Host,
      Speed: user.Speed,
      Connections: user.Connections
    };
  });
};

//submit the configuration to save in file
const submit = () => {
  ElMessageBox.confirm("Make sure you want to save the configuration file", "Save The Configuration", {
    confirmButtonText: "Confirm",
    cancelButtonText: "Cancel",
    type: "info"
  })
    .then(async () => {
      try {
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
        ElMessage.success("Submit success");
      } catch (e) {
        if (e instanceof Error) {
          ElMessage.error(e.message);
        } else {
          ElMessage.error("Failed to save the configuration file!");
        }
      }
    })
    .catch(() => {
      console.log("Cancel Submit Operation!");
    });
};

//get the configuration
const getFromFile = async () => {
  ElMessageBox.confirm(
    "Make sure you want to get the configuration from file, if you fail to get from file, it will get from the running system. NOTE: please make sure the change you made is saved, or it will be discarded.",
    "Get Configuration From File",
    {
      confirmButtonText: "Confirm",
      cancelButtonText: "Cancel",
      type: "info"
    }
  ).then(async () => {
    try {
      const { data } = await getServerConfigFromFileApi();
      updateData(data);
      ElMessage.success("Get from file success");
    } catch (e) {
      if (e instanceof Error) {
        ElMessage.error(e.message);
      } else {
        ElMessage.error("Failed to get from file!");
      }
    }
  });
};
const getFromRunning = async () => {
  ElMessageBox.confirm(
    "Make sure you want to get the configuration from running system. NOTE: please make sure the change you made is saved, or it will be discarded.",
    "Get Configuration From Running System",
    {
      confirmButtonText: "Confirm",
      cancelButtonText: "Cancel",
      type: "info"
    }
  ).then(async () => {
    try {
      const { data } = await getRunningServerConfigApi();
      updateData(data);
      ElMessage.success("Get from running system success");
    } catch (e) {
      if (e instanceof Error) {
        ElMessage.error(e.message);
      } else {
        ElMessage.error("Failed to get from running system!");
      }
    }
  });
};

onMounted(() => {
  getFromFile();
});
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
