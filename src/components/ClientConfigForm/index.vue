<template>
  <Anchor :tab-list="tabList">
    <!-- <template v-for="tab in staticTabs" :key="tab.uuid" #[`tab.uuid`]>
      <component :is="tab.component" :ref="tab.ref" :setting="tab.setting" @update:setting="updateSetting"
    /></template>

    <template v-for="tab in dynamicTabs" :key="tab.uuid" #[`tab.uuid`]>
      <component
        :is="tab.component"
        :ref="tab.ref"
        :index="tab.index"
        :is-last="tab.isLast"
        :setting="tab.setting"
        @add-service="addService"
        @remove-service="removeService(tab.index)"
        @update:setting="tab.updateSetting"
    /></template> -->
    <template #GeneralSetting>
      <GeneralSetting ref="generalSettingRef" :setting="generalSetting" @update:setting="updateGeneralSetting" />
    </template>
    <template #SentrySetting>
      <SentrySetting ref="sentrySettingRef" :setting="sentrySetting" @update:setting="updateSentrySetting" />
    </template>
    <template #WebRTCSetting>
      <WebRTCSetting ref="webRTCSettingRef" :setting="webRTCSetting" @update:setting="updateWebRTCSetting" />
    </template>
    <template #TCPForwardSetting>
      <TCPForwardSetting ref="tcpForwardSetting" :setting="tcpForwardSetting" @update:setting="updateTCPForwardSetting" />
    </template>
    <template #LogSetting>
      <LogSetting ref="logSettingRef" :setting="logSetting" @update:setting="updateLogSetting" />
    </template>
    <!-- <template v-for="(service, index) in services" :key="index" #="service.HostPrefix"> -->
    <!-- <template #Service1Setting> -->
    <!-- <template v-for="(service, index) in services" :key="index" #[`Service${index+1}Setting`]> -->

    <template v-for="(service, index) in services" :key="serviceKeys[index]" #[`Service${index+1}Setting`]>
      <ServiceSetting
        :ref="serviceSettingRefs[index]"
        :index="index"
        :is-last="index === services.length - 1"
        :setting="services[index]"
        @add-service="addService"
        @remove-service="removeService(index)"
        @update:setting="updateServiceSetting"
      />
    </template>
  </Anchor>
  <el-button type="primary" @click="onSubmit"> Submit</el-button>
</template>

<script setup lang="ts" name="ClientConfigForm">
import { ElMessage, ElMessageBox } from "element-plus";
import { Ref, reactive, ref } from "vue";
import { ClientConfig } from "./interface";
// import http from "@/api";
import yaml from "js-yaml";
import axios from "axios";
import Anchor, { Tab } from "@/components/Anchor/index.vue";
import GeneralSetting from "./components/GeneralSetting.vue";
import SentrySetting from "./components/SentrySetting.vue";
import WebRTCSetting from "./components/WebRTCSetting.vue";
import TCPForwardSetting from "./components/TCPForwardSetting.vue";
import LogSetting from "./components/LogSetting.vue";
import ServiceSetting from "./components/ServiceSetting.vue";
import { v4 as uuidv4 } from "uuid";
const tabList = reactive<Tab[]>([
  {
    title: "General Setting",
    name: "GeneralSetting",
    uuid: uuidv4()
  },
  {
    title: "Sentry Setting",
    name: "SentrySetting",
    uuid: uuidv4()
  },
  {
    title: "WebRTC Setting",
    name: "WebRTCSetting",
    uuid: uuidv4()
  },
  {
    title: "TCPForward Setting",
    name: "TCPForwardSetting",
    uuid: uuidv4()
  },
  {
    title: "Log Setting",
    name: "LogSetting",
    uuid: uuidv4()
  },
  {
    title: "Service 1 Setting",
    name: "Service1Setting",
    uuid: uuidv4()
  }
]);

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
  ...logSetting,

  HostPrefix: [],
  RemoteTCPPort: [],
  RemoteTCPRandom: [],
  Local: [],
  LocalTimeout: [],
  UseLocalAsHTTPHost: []
});

// TODO: 重构 + 删除 key + TabList title
let services = reactive<ClientConfig.Service[]>([{ ...ClientConfig.defaultServiceSetting }]);
const serviceKeys = [uuidv4()];

const addService = () => {
  services.push({ ...ClientConfig.defaultServiceSetting });
  const uuid = uuidv4();
  console.log(uuid);

  serviceKeys.push(uuid);
  serviceSettingRefs.push(ref<InstanceType<typeof ServiceSetting> | null>(null));
  tabList.push({
    title: `Service ${services.length} Setting`,
    name: `Service${services.length}Setting`,
    uuid: uuid
  });

  console.log(tabList);
};
const removeService = (index: number) => {
  console.log("index " + index);
  if (services.length === 1) {
    ElMessage.warning("至少需要一个服务");
    return;
  } else {
    services.splice(index, 1);
    serviceKeys.splice(index, 1);
    serviceSettingRefs.splice(index, 1);
    //delete the related tablist
    // let tabListIndex = tabList.findIndex(tab => tab.name === `Service${index + 1}Setting`);
    let tabListIndex = tabList.findIndex(tab => tab.title === `Service ${index + 1} Setting`);
    console.log("tabindex:" + tabListIndex);
    tabList.splice(tabListIndex, 1);
    //update the name of the remaining tablist
    for (let i = index; i < services.length; i++) {
      console.log("-------------------------");
      // tabList[tabListIndex + i - index].name = `Service${i + 1}Setting`;
      tabList[tabListIndex + i - index].name = `Service${i + 1}Setting`;
      tabList[tabListIndex + i - index].title = `Service ${i + 1} Setting`;
    }
    // services = [...services];
    // console.log(services);
    // console.log(serviceKeys);
  }
};
const clientConfig = reactive<ClientConfig.Config>({
  Version: "1",
  Services: services,
  Options: options
});

// TODO: 响应式 assign 合理性

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
// const serviceSettingRef = ref<InstanceType<typeof ServiceSetting> | null>(null);
// const serviceSettingRefs = ref<InstanceType<typeof ServiceSetting>[]>([]);
const serviceSettingRefs = Array(services.length)
  .fill(null)
  .map(() => ref<InstanceType<typeof ServiceSetting> | null>(null));

// const generalSettingRef = ref<ClientConfig.FormRef | null>(null);
// const sentrySettingRef = ref<ClientConfig.FormRef | null>(null);
// const webRTCSettingRef = ref<ClientConfig.FormRef | null>(null);
// const tcpForwardSettingRef = ref<ClientConfig.FormRef | null>(null);
// const logSettingRef = ref<ClientConfig.FormRef | null>(null);
// const serviceSettingRef = ref<ClientConfig.FormRef | null>(null);
// const serviceSettingRefs = ref<ClientConfig.FormRef[]>([]);
// const serviceSettingRefs = reactive<ClientConfig.FormRef[]>([]);

const validateAllForms = (formRefs: Array<Ref<ClientConfig.FormRef | null>>) => {
  return Promise.all(formRefs.map(formRef => formRef.value?.validateForm()));
};
// TODO: api update
const onSubmit = async () => {
  console.log("onSubmit");
  console.log(services);
  console.log(tabList);
  console.log(serviceKeys);
  return;
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
          logSettingRef
          // serviceSettingRef
          // ...serviceSettingRefs.map(ref => ref.validateForm())
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
        ElMessage.error("发送失败");
      }
    })
    .catch(() => {
      ElMessage.info("已取消发送");
    });
  // const { data } = await clientConfigApi({ ...clientConfig });
  // console.log(data);
};
</script>

<style scoped lang="scss">
@import "./index.scss";
.el-form-item {
  display: contents; // To reduce interference from the form style, only its validation function needs to be used.
}

// .input-with-prefix {
//   // width: calc(100% - 25px);

//   // position: relative;
//   // display: inline-block;
//   // background-color: black;
// }
</style>
