<!-- TODO: use form to have rule -->
<template>
  <el-form ref="ruleFormRef" :model="options" :rules="rules">
    <!-- General Setting -->
    <div class="card content-box">
      <!-- <div class="card content-box"> -->
      <!-- ID And Secret -->
      <el-descriptions :column="2" :border="true">
        <template #title> General Setting </template>
        <!-- </el-descriptions> -->
        <!-- <el-descriptions :column="2" :border="true"> -->
        <el-descriptions-item>
          <template #label>
            ID
            <UsageTooltip :usage-text="ClientConfig.usage['ID']" />
          </template>
          <!-- <el-form-item prop="ID"> -->
          <el-input v-model="options.ID"></el-input>
          <!-- </el-form-item> -->
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            Secret
            <UsageTooltip :usage-text="ClientConfig.usage['Secret']" />
          </template>
          <el-input v-model="options.Secret" type="password" show-password></el-input>
        </el-descriptions-item>
        <!-- </el-descriptions> -->
        <!-- TODO: fix the time duration -->
        <!-- ReconnectDelay And  RemoteTimeout-->
        <!-- <el-descriptions :column="2" :border="true"> -->
        <el-descriptions-item>
          <template #label>
            ReconnectDelay
            <UsageTooltip :usage-text="ClientConfig.usage['ReconnectDelay']" />
          </template>
          <el-form-item prop="ReconnectDelay">
            <el-input v-model="options.ReconnectDelay"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            RemoteTimeout
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteTimeout']" />
          </template>
          <el-form-item prop="RemoteTimeout">
            <el-input v-model="options.RemoteTimeout"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- </el-descriptions> -->
        <!-- Remote Setting -->
        <!-- <el-descriptions :column="2" :border="true"> -->
        <el-descriptions-item>
          <template #label>
            Remote
            <UsageTooltip :usage-text="ClientConfig.usage['Remote']" />
          </template>
          <el-input v-model="inputRemote">
            <template #prepend>
              <el-select v-model="selectRemote">
                <!-- <el-select v-model="selectRemote" style="position: absolute; z-index: 1; width: auto"> -->
                <el-option v-for="option in remoteOptions" :key="option" :label="option" :value="option" />
              </el-select>
            </template>
          </el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            RemoteSTUN
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteSTUN']" />
          </template>
          <el-input v-model="options.RemoteSTUN"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            RemoteAPI
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteAPI']" />
          </template>
          <el-input v-model="options.RemoteAPI"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            RemoteCert
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteCert']" />
          </template>
          <el-input v-model="options.RemoteCert"></el-input>
        </el-descriptions-item>
        <!-- </el-descriptions> -->
        <!-- <el-descriptions :column="2" :border="true"> -->
        <el-descriptions-item>
          <template #label>
            RemoteCertInsecure
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteCertInsecure']" />
          </template>
          <el-switch v-model="options.RemoteCertInsecure" active-text="true" inactive-text="false" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            RemoteConnections
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteConnections']" />
          </template>
          <el-input-number v-model="options.RemoteConnections" :min="1" :max="10" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            RemoteIdleConnections
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteIdleConnections']" />
          </template>
          <el-input-number v-model="options.RemoteIdleConnections" :min="0" :max="options.RemoteConnections" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            Version
            <UsageTooltip :usage-text="ClientConfig.usage['Version']" />
          </template>
          <el-switch v-model="options.Version" active-text="true" inactive-text="false" />
        </el-descriptions-item>
      </el-descriptions>
    </div>
    <!-- Sentry Setting -->
    <div class="card content-box">
      <!-- Sentry -->
      <el-descriptions :column="2" :border="true">
        <template #title> Sentry Setting </template>
        <el-descriptions-item>
          <template #label>
            SentryDSN
            <UsageTooltip :usage-text="ClientConfig.usage['SentryDSN']" />
          </template>
          <el-input v-model="options.SentryDSN"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            SentryServerName
            <UsageTooltip :usage-text="ClientConfig.usage['SentryServerName']" />
          </template>
          <el-input v-model="options.SentryServerName"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            SentryLevel
            <UsageTooltip :usage-text="ClientConfig.usage['SentryLevel']" />
          </template>
          <el-checkbox-group v-model="options.SentryLevel">
            <el-checkbox v-for="option in SentryLevelOptions" :key="option" :label="option" />
          </el-checkbox-group>
        </el-descriptions-item>
      </el-descriptions>
      <el-descriptions :column="1" :border="true">
        <el-descriptions-item>
          <template #label>
            SentrySampleRate
            <UsageTooltip :usage-text="ClientConfig.usage['SentrySampleRate']" />
          </template>
          <el-slider style="width: 400px" v-model="options.SentrySampleRate" :step="0.1" :min="0" :max="1" show-input></el-slider>
        </el-descriptions-item>
      </el-descriptions>
      <el-descriptions :column="2" :border="true">
        <el-descriptions-item>
          <template #label>
            SentryRelease
            <UsageTooltip :usage-text="ClientConfig.usage['SentryRelease']" />
          </template>
          <el-input v-model="options.SentryRelease"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            SentryEnvironment
            <UsageTooltip :usage-text="ClientConfig.usage['SentryEnvironment']" />
          </template>
          <el-input v-model="options.SentryEnvironment"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            SentryDebug
            <UsageTooltip :usage-text="ClientConfig.usage['SentryDebug']" />
          </template>
          <el-switch v-model="options.SentryDebug" active-text="true" inactive-text="false" />
        </el-descriptions-item>
      </el-descriptions>
    </div>
    <!-- WebRTC Setting -->
    <div class="card content-box">
      <!-- WebRTC -->
      <el-descriptions :column="2" :border="true">
        <template #title> WebRTC Setting </template>
        <el-descriptions-item>
          <template #label>
            WebRTCConnectionIdleTimeout
            <UsageTooltip :usage-text="ClientConfig.usage['WebRTCConnectionIdleTimeout']" />
          </template>
          <el-form-item prop="WebRTCConnectionIdleTimeout">
            <el-input v-model="options.WebRTCConnectionIdleTimeout"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            WebRTCLogLevel
            <UsageTooltip :usage-text="ClientConfig.usage['WebRTCLogLevel']" />
          </template>
          <el-select v-model="options.WebRTCLogLevel" placeholder="Select log level">
            <el-option v-for="option in WebRTCLogLevelOptions" :key="option" :label="option" :value="option" />
          </el-select>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            WebRTCMinPort
            <UsageTooltip :usage-text="ClientConfig.usage['WebRTCMinPort']" />
          </template>
          <el-input-number v-model="options.WebRTCMinPort" :min="0" :max="65535" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            WebRTCMaxPort
            <UsageTooltip :usage-text="ClientConfig.usage['WebRTCMaxPort']" />
          </template>
          <el-input-number v-model="options.WebRTCMaxPort" :min="0" :max="65535" />
        </el-descriptions-item>
      </el-descriptions>
    </div>
    <!-- TCPForward Setting -->
    <div class="card content-box">
      <!-- TCP -->
      <el-descriptions :column="2" :border="true">
        <template #title> TCPForward Setting</template>
        <el-descriptions-item>
          <template #label>
            TcpForwardAddr
            <UsageTooltip :usage-text="ClientConfig.usage['TCPForwardAddr']" />
          </template>
          <el-input v-model="options.TCPForwardAddr"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            TcpForwardHostPrefix
            <UsageTooltip :usage-text="ClientConfig.usage['TCPForwardHostPrefix']" />
          </template>
          <el-input v-model="options.TCPForwardHostPrefix"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            TcpForwardConnections
            <UsageTooltip :usage-text="ClientConfig.usage['TCPForwardConnections']" />
          </template>
          <el-input-number v-model="options.TCPForwardConnections" :min="1" :max="10" />
        </el-descriptions-item>
      </el-descriptions>
    </div>
    <!-- Log Setting -->
    <div class="card content-box">
      <!-- Log -->
      <el-descriptions :column="2" :border="true">
        <template #title> Log Setting </template>
        <el-descriptions-item>
          <template #label>
            LogFile
            <UsageTooltip :usage-text="ClientConfig.usage['LogFile']" />
          </template>
          <el-input v-model="options.LogFile"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            LogFileMaxSize
            <UsageTooltip :usage-text="ClientConfig.usage['LogFileMaxSize']" />
          </template>
          <el-input-number v-model="options.LogFileMaxSize" :min="0" :max="100" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            LogFileMaxCount
            <UsageTooltip :usage-text="ClientConfig.usage['LogFileMaxCount']" />
          </template>
          <el-input-number v-model="options.LogFileMaxCount" :min="0" :max="100" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            LogLevel
            <UsageTooltip :usage-text="ClientConfig.usage['LogLevel']" />
          </template>
          <el-select v-model="options.LogLevel" placeholder="Select log level">
            <el-option label="trace" value="trace"></el-option>
            <el-option label="debug" value="debug"></el-option>
            <el-option label="info" value="info"></el-option>
            <el-option label="warn" value="warn"></el-option>
            <el-option label="error" value="error"></el-option>
            <el-option label="fatal" value="fatal"></el-option>
            <el-option label="panic" value="panic"></el-option>
            <el-option label="disable" value="disable"></el-option>
          </el-select>
        </el-descriptions-item>
      </el-descriptions>
    </div>
    <!-- Service Setting -->
    <div class="card content-box" v-for="(service, index) in services" :key="index">
      <el-descriptions :column="2" :border="true">
        <template #title> Service {{ index + 1 }} Setting </template>
        <template #extra>
          <el-button v-if="index === services.length - 1" type="primary" @click="addService">Add Service</el-button>
          <el-button type="danger" @click="removeService(index)">Delete</el-button>
        </template>
        <el-descriptions-item>
          <template #label>
            HostPrefix
            <UsageTooltip :usage-text="ClientConfig.usage['HostPrefix']" />
          </template>
          <el-input v-model="service.HostPrefix"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            RemoteTCPPort
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteTCPPort']" />
          </template>
          <el-input-number v-model="service.RemoteTCPPort" :min="0" :max="65535" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            RemoteTCPRandom
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteTCPRandom']" />
          </template>
          <el-switch v-model="service.RemoteTCPRandom" active-text="true" inactive-text="false" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            LocalURL
            <UsageTooltip :usage-text="ClientConfig.usage['LocalURL']" />
          </template>
          <el-input v-model="service.LocalURL"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            LocalTimeout
            <UsageTooltip :usage-text="ClientConfig.usage['LocalTimeout']" />
          </template>
          <el-form-item prop="LocalTimeout">
            <el-input v-model="service.LocalTimeout"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            UseLocalAsHTTPHost
            <UsageTooltip :usage-text="ClientConfig.usage['UseLocalAsHTTPHost']" />
          </template>
          <el-switch v-model="service.UseLocalAsHTTPHost" active-text="true" inactive-text="false" />
        </el-descriptions-item>
      </el-descriptions>
    </div>
    <el-button type="primary" @click="onSubmit"> Submit</el-button>
  </el-form>
</template>

<script setup lang="ts" name="ClientConfigForm">
import { ElMessage, ElMessageBox, FormInstance, FormRules } from "element-plus";
import { computed, reactive, ref } from "vue";
import { ClientConfig } from "./interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
// import http from "@/api";
import yaml from "js-yaml";
import axios from "axios";

const inputRemote = ref("");
const selectRemote = ref("tcp://");
const remoteOptions = ["tcp://", "tls://"];
const SentryLevelOptions = ["trace", "debug", "info", "warn", "error", "fatal", "panic"];
const WebRTCLogLevelOptions = ["verbose", "info", "warning", "error"];
const RemoteURL = computed(() => {
  return selectRemote.value + inputRemote.value;
});
const options = reactive<ClientConfig.Options>({
  Config: "",
  ID: "",
  Secret: "",
  ReconnectDelay: "",
  Remote: RemoteURL.value,
  RemoteSTUN: "",
  RemoteAPI: "",
  RemoteCert: "",
  RemoteCertInsecure: false,
  RemoteConnections: 1,
  RemoteIdleConnections: 0,
  RemoteTimeout: "",

  SentryDSN: "",
  SentryLevel: ["error", "fatal", "panic"],
  SentrySampleRate: 0,
  SentryRelease: "",
  SentryEnvironment: "",
  SentryServerName: "",
  SentryDebug: false,

  WebRTCConnectionIdleTimeout: "",
  WebRTCLogLevel: "",
  WebRTCMinPort: 0,
  WebRTCMaxPort: 0,

  TCPForwardAddr: "",
  TCPForwardHostPrefix: "",
  TCPForwardConnections: 0,

  LogFile: "",
  LogFileMaxSize: 0,
  LogFileMaxCount: 0,
  LogLevel: "",
  HostPrefix: [],
  RemoteTCPPort: [],
  RemoteTCPRandom: [],
  Local: [],
  LocalTimeout: [],
  UseLocalAsHTTPHost: [],
  Version: false
});
const initialService: ClientConfig.Service = {
  HostPrefix: "",
  RemoteTCPPort: 0,
  RemoteTCPRandom: false,
  LocalURL: "",
  LocalTimeout: "",
  UseLocalAsHTTPHost: false
};

const services = reactive<ClientConfig.Service[]>([{ ...initialService }]);
const addService = () => {
  services.push({ ...initialService });
};
const removeService = (index: number) => {
  if (services.length === 1) {
    ElMessage.warning("至少需要一个服务");
    return;
  } else {
    services.splice(index, 1);
  }
};
const clientConfig = reactive<ClientConfig.Config>({
  Version: "1",
  Services: services,
  Options: options
});
const validatorTimeFormat = (rule: any, value: any, callback: any) => {
  console.log("Calling validatorTimeFormat");
  const regex = /^(\d+(ns|us|µs|ms|s|m|h))+$/;
  if (!value || regex.test(value)) {
    console.log("regex test passed");
    console.log(value);
    callback();
  } else {
    console.log("regex test failed");
    console.log(value);
    callback(new Error("Please enter a valid time format"));
  }
};

const ruleFormRef = ref<FormInstance>();
const rules = reactive<FormRules<ClientConfig.RuleForm>>({
  ReconnectDelay: [{ validator: validatorTimeFormat, trigger: "blur" }],
  RemoteTimeout: [{ validator: validatorTimeFormat, trigger: "blur" }],
  WebRTCConnectionIdleTimeout: [{ validator: validatorTimeFormat, trigger: "blur" }],
  LocalTimeout: [{ validator: validatorTimeFormat, trigger: "blur" }]
});
// const clientConfigApi = (params: ClientConfig.Config) => {
//   return http.post("/config/client", params);
// };

// TODO: api update
const onSubmit = async () => {
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
