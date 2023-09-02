<template>
  <Anchor :tab-list="tabList">
    <template v-for="tab in staticTabs" :key="tab.uuid" #[tab.uuid]>
      <component :is="tab.component" :ref="tab.ref" :setting="tab.setting" @update:setting="tab.updateSetting" />
    </template>

    <template v-for="tab in dynamicTabs" :key="tab.uuid" #[tab.uuid]>
      <component
        :is="tab.component"
        :ref="serviceSettingRefs[tab.index]"
        :index="tab.index"
        :is-last="tab.isLast"
        :setting="tab.setting"
        @add-service="addService"
        @remove-service="removeService(tab.index)"
        @update:setting="tab.updateSetting"
      />
    </template>
  </Anchor>
  <el-button type="primary" @click="onSubmit"> Submit</el-button>
  <el-button type="primary" @click="reload"> Reload</el-button>
  <el-button type="primary" @click="test">Test</el-button>
</template>

<script setup lang="ts" name="ClientConfigForm">
import { ElMessage, ElMessageBox } from "element-plus";
import { markRaw, Ref, reactive, ref, watchEffect, onMounted, computed, nextTick } from "vue";
import { ClientConfig } from "./interface";
import Anchor, { Tab } from "@/components/Anchor/index.vue";
import GeneralSetting from "./components/GeneralSetting.vue";
import SentrySetting from "./components/SentrySetting.vue";
import WebRTCSetting from "./components/WebRTCSetting.vue";
import TCPForwardSetting from "./components/TCPForwardSetting.vue";
import LogSetting from "./components/LogSetting.vue";
import ServiceSetting from "./components/ServiceSetting.vue";
import { getClientConfigFromFileApi, saveClientConfigApi } from "@/api/modules/clientConfig";
import { getConnectionApi } from "@/api/modules/connection";
import { v4 as uuidv4 } from "uuid";
import yaml from "js-yaml";
import { Config } from "@/api/interface";
import {
  mapClientGeneralSetting,
  mapClientSentrySetting,
  mapClientLogSetting,
  mapClientTCPForwardSetting,
  mapClientWebRTCSetting,
  mapClientServices
} from "@/utils/map";

const generalSetting = reactive<ClientConfig.GeneralSetting>({ ...ClientConfig.defaultGeneralSetting });
const sentrySetting = reactive<ClientConfig.SentrySetting>({ ...ClientConfig.defaultSentrySetting });
const webRTCSetting = reactive<ClientConfig.WebRTCSetting>({ ...ClientConfig.defaultWebRTCSetting });
const tcpForwardSetting = reactive<ClientConfig.TCPForwardSetting>({ ...ClientConfig.defaultTCPForwardSetting });
const logSetting = reactive<ClientConfig.LogSetting>({ ...ClientConfig.defaultLogSetting });

const options = reactive<ClientConfig.Options>({
  Config: "",

  ...generalSetting,
  ...sentrySetting,
  ...webRTCSetting,
  ...tcpForwardSetting,
  ...logSetting
});

watchEffect(() => {
  Object.assign(options, {
    ...generalSetting,
    ...sentrySetting,
    ...webRTCSetting,
    ...tcpForwardSetting,
    ...logSetting
  });
});

let services = reactive<ClientConfig.Service[]>([{ ...ClientConfig.defaultServiceSetting }]);

const adjustView = () => {
  //Need the el-main element to be set to overflow: auto
  const elMain = document.querySelector(".el-main");
  if (elMain) {
    const currentScrollPosition = elMain.scrollTop;
    nextTick(() => {
      elMain.scrollTop = currentScrollPosition;
    });
  }
};

const addService = () => {
  console.log("addService");
  services.push({ ...ClientConfig.defaultServiceSetting });
  serviceSettingRefs.push(ref<InstanceType<typeof ServiceSetting> | null>(null));
  adjustView();
};
const removeService = (index: number) => {
  console.log("removeService");
  console.log("index " + index);
  if (services.length === 1) {
    ElMessage.warning("At least one service is required!");
    return;
  } else {
    services.splice(index, 1);
    serviceSettingRefs.splice(index, 1);
    console.log("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!");
    adjustView();
  }
};

const clientConfig = reactive<ClientConfig.Config>({
  Version: "1",
  Services: services,
  ...options
});
watchEffect(() => {
  Object.assign(clientConfig, {
    ...options
  });
});

const updateGeneralSetting = (newSetting: ClientConfig.GeneralSetting) => {
  Object.assign(generalSetting, newSetting);
};
const updateSentrySetting = (newSetting: ClientConfig.SentrySetting) => {
  Object.assign(sentrySetting, newSetting);
};
const updateWebRTCSetting = (newSetting: ClientConfig.WebRTCSetting) => {
  Object.assign(webRTCSetting, newSetting);
};
const updateTCPForwardSetting = (newSetting: ClientConfig.TCPForwardSetting) => {
  Object.assign(tcpForwardSetting, newSetting);
};
const updateLogSetting = (newSetting: ClientConfig.LogSetting) => {
  Object.assign(logSetting, newSetting);
};

const updateServiceSetting = (index: number, newSetting: ClientConfig.Service) => {
  //update the service setting with corresponding index
  Object.assign(services[index], newSetting);
};

const generalSettingRef = ref<InstanceType<typeof GeneralSetting> | null>(null);
const sentrySettingRef = ref<InstanceType<typeof SentrySetting> | null>(null);
const webRTCSettingRef = ref<InstanceType<typeof WebRTCSetting> | null>(null);
const tcpForwardSettingRef = ref<InstanceType<typeof TCPForwardSetting> | null>(null);
const logSettingRef = ref<InstanceType<typeof LogSetting> | null>(null);
const serviceSettingRef = ref<InstanceType<typeof ServiceSetting> | null>(null);
const serviceSettingRefs = reactive<Ref<InstanceType<typeof ServiceSetting> | null>[]>([serviceSettingRef]);

interface staticTabType<T> {
  title: string;
  name: string;
  uuid: string;
  component: any;
  // Note that: the ref must be a string, not a ref object,
  // but need to be the same name of the ref object.
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
  } as staticTabType<ClientConfig.GeneralSetting>,
  {
    title: "Sentry Setting",
    name: "SentrySetting",
    uuid: uuidv4(),
    component: markRaw(SentrySetting),
    ref: "sentrySettingRef",
    setting: sentrySetting,
    updateSetting: updateSentrySetting
  } as staticTabType<ClientConfig.SentrySetting>,
  {
    title: "WebRTC Setting",
    name: "WebRTCSetting",
    uuid: uuidv4(),
    component: markRaw(WebRTCSetting),
    ref: "webRTCSettingRef",
    setting: webRTCSetting,
    updateSetting: updateWebRTCSetting
  } as staticTabType<ClientConfig.WebRTCSetting>,
  {
    title: "TCPForward Setting",
    name: "TCPForwardSetting",
    uuid: uuidv4(),
    component: markRaw(TCPForwardSetting),
    ref: "tcpForwardSettingRef",
    setting: tcpForwardSetting,
    updateSetting: updateTCPForwardSetting
  } as staticTabType<ClientConfig.TCPForwardSetting>,
  {
    title: "Log Setting",
    name: "LogSetting",
    uuid: uuidv4(),
    component: markRaw(LogSetting),
    ref: "logSettingRef",
    setting: logSetting,
    updateSetting: updateLogSetting
  } as staticTabType<ClientConfig.LogSetting>
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
const dynamicTabs = computed<dynamicTabType<ClientConfig.Service>[]>(() => {
  return services.map((service, index) => ({
    title: `Service ${index + 1} Setting`,
    name: `Service${index + 1}Setting`,
    uuid: uuidv4(),
    component: markRaw(ServiceSetting),
    setting: service,
    updateSetting: updateServiceSetting,
    index: index,
    isLast: index == services.length - 1
  }));
});
const tabList = computed<Tab[]>(() => [
  ...staticTabs.map(tab => ({ title: tab.title, name: tab.name, uuid: tab.uuid })),
  ...dynamicTabs.value.map(tab => ({ title: tab.title, name: tab.name, uuid: tab.uuid }))
]);
// TODO: must input and trim
const validateAllForms = (formRefs: Array<Ref<ClientConfig.FormRef | null>>) => {
  return Promise.all(formRefs.map(formRef => formRef.value?.validateForm()));
};
// TODO: api update
const onSubmit = async () => {
  console.log("onSubmit");
  const json1 = JSON.stringify(clientConfig);
  const yamlString = yaml.dump(clientConfig);
  console.log(json1);
  console.log(yamlString);
  ElMessageBox.confirm("Make sure you want to save the config setting to file.", "Save The Config Setting", {
    confirmButtonText: "Confirm",
    cancelButtonText: "Cancel",
    type: "info"
  })
    .then(async () => {
      try {
        await validateAllForms([
          generalSettingRef,
          sentrySettingRef,
          webRTCSettingRef,
          tcpForwardSettingRef,
          logSettingRef,
          ...serviceSettingRefs
        ]);
        const response = await saveClientConfigApi(clientConfig);
        console.log(response);
        ElMessage.success("Save Success!");
      } catch (e) {
        console.log(e);
        if (e instanceof Error) {
          ElMessage.error(e.message);
        } else {
          ElMessage.error("Failed to save!");
        }
      }
    })
    .catch(() => {
      ElMessage.info("已取消发送");
    });
};
const updateData = (data: Config.Client.ResConfig) => {
  console.log("--------------------------------------");
  console.log("updateData");
  Object.assign(generalSetting, mapClientGeneralSetting(data));
  Object.assign(sentrySetting, mapClientSentrySetting(data));
  Object.assign(webRTCSetting, mapClientWebRTCSetting(data));
  Object.assign(tcpForwardSetting, mapClientTCPForwardSetting(data));
  Object.assign(logSetting, mapClientLogSetting(data));
  options.Config = data.config.Config;
  services.splice(0, services.length, ...mapClientServices(data));
  serviceSettingRefs.splice(0, serviceSettingRefs.length);
  console.log("--------------------------------------");
};

const reload = async () => {
  const { data } = await getClientConfigFromFileApi();
  console.log("--------------------------------------");
  console.log(data);
  console.log(JSON.stringify(services));
  updateData(data);
  console.log(JSON.stringify(services));
  console.log("--------------------------------------");
};

onMounted(() => {
  reload();
});

const test = async () => {
  const { data } = await getConnectionApi();
  console.log(data);
};
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
