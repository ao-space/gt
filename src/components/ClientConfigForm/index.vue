<template>
  <div :is-loaded="dataLoaded">
    <Anchor :tab-list="tabList">
      <template v-for="tab in staticTabs" :key="tab.uuid" #[tab.uuid]>
        <component :is="tab.component" :ref="tab.ref" :setting="tab.setting" @update:setting="tab.updateSetting" />
      </template>

      <template v-for="(tab, index) in dynamicTabs" :key="tab.uuid" #[tab.uuid]>
        <component
          :is="tab.component"
          :ref="serviceSettingRefs[index]"
          :index="index"
          :is-last="tab.isLast"
          :setting="tab.setting"
          @add-service="addService"
          @remove-service="removeService(index)"
          @update:setting="tab.updateSetting"
        />
      </template>
    </Anchor>
    <el-button type="primary" @click="onSubmit"> Submit</el-button>
  </div>
</template>

<script setup lang="ts" name="ClientConfigForm">
import { ElMessage, ElMessageBox } from "element-plus";
import { markRaw, Ref, reactive, ref, watchEffect, onBeforeMount } from "vue";
import { ClientConfig } from "./interface";
import yaml from "js-yaml";
import axios from "axios";
import Anchor, { Tab } from "@/components/Anchor/index.vue";
import GeneralSetting from "./components/GeneralSetting.vue";
import SentrySetting from "./components/SentrySetting.vue";
import WebRTCSetting from "./components/WebRTCSetting.vue";
import TCPForwardSetting from "./components/TCPForwardSetting.vue";
import LogSetting from "./components/LogSetting.vue";
import ServiceSetting from "./components/ServiceSetting.vue";
import { getRunningClientConfigApi } from "@/api/modules/clientConfig";
import { v4 as uuidv4 } from "uuid";
import { Config } from "@/api/interface";
import {
  mapClientGeneralSetting,
  mapClientSentrySetting,
  mapClientLogSetting,
  mapClientTCPForwardSetting,
  mapClientWebRTCSetting
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

  // HostPrefix: [],
  // RemoteTCPPort: [],
  // RemoteTCPRandom: [],
  // Local: [],
  // LocalTimeout: [],
  // UseLocalAsHTTPHost: []
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
const addService = () => {
  services.push({ ...ClientConfig.defaultServiceSetting });
  serviceSettingRefs.push(ref<InstanceType<typeof ServiceSetting> | null>(null));
  const uuid = uuidv4();
  console.log(uuid);
  tabList.push({
    title: `Service ${services.length} Setting`,
    name: `Service${services.length}Setting`,
    uuid: uuid
  });

  dynamicTabs[dynamicTabs.length - 1].isLast = false;
  dynamicTabs.push({
    title: `Service ${services.length} Setting`,
    name: `Service${services.length}Setting`,
    uuid: uuid,
    component: markRaw(ServiceSetting),
    setting: services[services.length - 1],
    updateSetting: updateServiceSetting,
    index: services.length - 1,
    isLast: true
  });
};
const removeService = (index: number) => {
  console.log("index " + index);
  if (services.length === 1) {
    ElMessage.warning("至少需要一个服务");
    return;
  } else {
    services.splice(index, 1);
    serviceSettingRefs.splice(index, 1);
    //delete the related tablist
    let tabListIndex = tabList.findIndex(tab => tab.title === `Service ${index + 1} Setting`);

    tabList.splice(tabListIndex, 1);
    //update the name of the remaining tablist
    for (let i = index; i < services.length; i++) {
      tabList[tabListIndex + i - index].name = `Service${i + 1}Setting`;
      tabList[tabListIndex + i - index].title = `Service ${i + 1} Setting`;
    }

    dynamicTabs.splice(index, 1);
    //update the index of the remaining dynamicTabs
    for (let i = index; i < dynamicTabs.length - 1; i++) {
      dynamicTabs[i].index = i;
      dynamicTabs[i].name = `Service${i + 1}Setting`;
      dynamicTabs[i].title = `Service ${i + 1} Setting`;
    }
    dynamicTabs[dynamicTabs.length - 1].isLast = true;
  }
};

const clientConfig = reactive<ClientConfig.Config>({
  Version: "1",
  Services: services,
  ...options
  // Options: options
});
watchEffect(() => {
  Object.assign(clientConfig, {
    ...options,
    Services: services
  });
});

const updateGeneralSetting = (newSetting: ClientConfig.GeneralSetting) => {
  console.log("updateGeneralSetting");
  console.log(newSetting);
  Object.assign(generalSetting, newSetting);
};
const updateSentrySetting = (newSetting: ClientConfig.SentrySetting) => {
  console.log("updateSentrySetting");
  console.log(newSetting);
  Object.assign(sentrySetting, newSetting);
};
const updateWebRTCSetting = (newSetting: ClientConfig.WebRTCSetting) => {
  console.log("updateWebRTCSetting");
  console.log(newSetting);
  Object.assign(webRTCSetting, newSetting);
};
const updateTCPForwardSetting = (newSetting: ClientConfig.TCPForwardSetting) => {
  console.log("updateTCPForwardSetting");
  console.log(newSetting);
  Object.assign(tcpForwardSetting, newSetting);
};
const updateLogSetting = (newSetting: ClientConfig.LogSetting) => {
  console.log("updateLogSetting");
  console.log(newSetting);
  Object.assign(logSetting, newSetting);
};

const updateServiceSetting = (index: number, newSetting: ClientConfig.Service) => {
  //update the service setting of the corresponding index
  if (0 <= index && index < services.length) {
    console.log("updateServiceSetting");
    console.log(index);
    console.log(newSetting);
    Object.assign(services[index], newSetting);
  } else {
    //ignore the delete operation
  }
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
  component: any; //Can't set Component type?
  ref: string; // Note that: the ref must be a string, not a ref object, but need to be the same name of the ref object
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
  component: any; //Can't set Component type?
  setting: T;
  updateSetting: (index: number, newSetting: T) => void;
  index: number;
  isLast: boolean;
}
const dynamicTabs = reactive<dynamicTabType<ClientConfig.Service>[]>([
  {
    title: "Service 1 Setting",
    name: "Service1Setting",
    uuid: uuidv4(),
    component: markRaw(ServiceSetting),
    setting: services[0],
    updateSetting: updateServiceSetting,
    index: 0,
    isLast: true
  }
]);
const tabList = reactive<Tab[]>([
  ...staticTabs.map(tab => ({ title: tab.title, name: tab.name, uuid: tab.uuid })),
  ...dynamicTabs.map(tab => ({ title: tab.title, name: tab.name, uuid: tab.uuid }))
]);
watchEffect(() => {
  tabList.splice(staticTabs.length);
  dynamicTabs.forEach(tab => {
    tabList.push({
      title: tab.title,
      name: tab.name,
      uuid: tab.uuid
    });
  });
});
// TODO: must input and trim
const validateAllForms = (formRefs: Array<Ref<ClientConfig.FormRef | null>>) => {
  return Promise.all(formRefs.map(formRef => formRef.value?.validateForm()));
};
// TODO: api update
const onSubmit = async () => {
  console.log("onSubmit");
  console.log([...services]);
  console.log(services);
  console.log(tabList);
  console.log(dynamicTabs);
  console.log(serviceSettingRefs);

  const json1 = JSON.stringify(clientConfig);
  console.log(json1);
  const yamlData = yaml.dump(clientConfig);
  console.log(yamlData);
  ElMessageBox.confirm("确定要发送以下数据吗？\n" + json1, "确认发送", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
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
        const response = await axios.post("/api/config/client", yamlData, {
          headers: {
            "Content-Type": "application/x-yaml;charset=utf-8"
          }
        });
        console.log(response);
        ElMessage.success("发送成功");
      } catch (e) {
        console.log(e);
        if (e instanceof Error) {
          ElMessage.error(e.message);
        } else {
          ElMessage.error("发送失败");
        }
      }
    })
    .catch(() => {
      ElMessage.info("已取消发送");
    });
  // const { data } = await clientConfigApi({ ...clientConfig });
  // console.log(data);
};
const updateData = (data: Config.Client.ResConfig) => {
  console.log(mapClientGeneralSetting(data));
  console.log(generalSetting);
  Object.assign(generalSetting, mapClientGeneralSetting(data));
  Object.assign(sentrySetting, mapClientSentrySetting(data));
  Object.assign(webRTCSetting, mapClientWebRTCSetting(data));
  Object.assign(tcpForwardSetting, mapClientTCPForwardSetting(data));
  Object.assign(logSetting, mapClientLogSetting(data));
  Object.assign(options, data.config.Config);
  services.splice(0, services.length, ...data.config.Services);
  serviceSettingRefs.splice(0, serviceSettingRefs.length);
  dynamicTabs.splice(0, dynamicTabs.length);
  services.forEach((service, index) => {
    const uuid = uuidv4();
    tabList.push({
      title: `Service ${index + 1} Setting`,
      name: `Service${index + 1}Setting`,
      uuid: uuid
    });
    dynamicTabs.push({
      title: `Service ${index + 1} Setting`,
      name: `Service${index + 1}Setting`,
      uuid: uuid,
      component: markRaw(ServiceSetting),
      setting: service,
      updateSetting: updateServiceSetting,
      index: index,
      isLast: index == services.length - 1
    });
    serviceSettingRefs.push(ref<InstanceType<typeof ServiceSetting> | null>(null));
  });
};
const dataLoaded = ref(false);

const reload = async () => {
  const { data } = await getRunningClientConfigApi();

  console.log("--------------------------------------");
  console.log(data);
  // console.log(clientConfig);
  // console.log(options);
  // console.log(services);
  // console.log(tabList);
  updateData(data);
  dataLoaded.value = true;
  console.log("--------------------------------------");
};

onBeforeMount(() => {
  reload();
});
</script>

<style scoped lang="scss">
@import "./index.scss";
.el-form-item {
  display: contents; // To reduce interference from the form style, only its validation function needs to be used.
}
</style>
