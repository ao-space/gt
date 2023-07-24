<!-- TODO: use form to have rule -->
<template>
  <!-- General Setting -->
  <div class="card content-box">
    <!-- ID And Secret -->
    <el-descriptions :column="2" :border="true">
      <template #title> General Setting </template>
      <el-descriptions-item>
        <template #label> ID </template>
        <el-input v-model="options.ID"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label>Secret</template>
        <el-input v-model="options.Secret" type="password" show-password></el-input>
      </el-descriptions-item>
    </el-descriptions>
    <!-- TODO: fix the time duration -->
    <!-- ReconnectDelay And  RemoteTimeout-->
    <el-descriptions :column="2" :border="true">
      <el-descriptions-item>
        <template #label> ReconnectDelay </template>
        <el-input v-model="options.ReconnectDelay"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> RemoteTimeout</template>
        <el-input v-model="options.RemoteTimeout"></el-input>
      </el-descriptions-item>
    </el-descriptions>
    <!-- Remote Setting -->
    <el-descriptions :border="true">
      <el-descriptions-item>
        <template #label> Remote</template>
        <el-input v-model="inputRemote" placeholder="Please input">
          <template #prepend>
            <el-select v-model="selectRemote" style="width: 115px">
              <el-option v-for="option in remoteOptions" :key="option" :label="option" :value="option" />
            </el-select>
          </template>
        </el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label>RemoteSTUN </template>
        <el-input v-model="options.RemoteSTUN"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label>RemoteAPI </template>
        <el-input v-model="options.RemoteAPI"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label>RemoteCert </template>
        <el-input v-model="options.RemoteCert"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label>RemoteCertInsecure </template>
        <el-switch v-model="options.RemoteCertInsecure" active-text="true" inactive-text="false" />
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label>RemoteConnections </template>
        <el-input-number v-model="options.RemoteConnections" :min="1" :max="10" />
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label>RemoteIdleConnections </template>
        <el-input-number v-model="options.RemoteIdleConnections" :min="0" :max="options.RemoteConnections" />
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label>Version </template>
        <el-switch v-model="options.Version" active-text="true" inactive-text="false" />
      </el-descriptions-item>
    </el-descriptions>
  </div>
  <!-- Sentry Setting -->
  <div class="card content-box">
    <!-- Sentry -->
    <el-descriptions :border="true">
      <template #title> Sentry Setting </template>
      <el-descriptions-item>
        <template #label> SentryDSN </template>
        <el-input v-model="options.SentryDSN"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> SentryLevel</template>
        <el-checkbox-group v-model="options.SentryLevel">
          <el-checkbox label="trace"></el-checkbox>
          <el-checkbox label="debug"></el-checkbox>
          <el-checkbox label="info"></el-checkbox>
          <el-checkbox label="warn"></el-checkbox>
          <el-checkbox label="error"></el-checkbox>
          <el-checkbox label="fatal"></el-checkbox>
          <el-checkbox label="panic"></el-checkbox>
        </el-checkbox-group>
      </el-descriptions-item>
    </el-descriptions>
    <el-descriptions :border="true">
      <el-descriptions-item>
        <template #label> SentrySampleRate </template>
        <el-slider v-model="options.SentrySampleRate" :step="0.1" :min="0" :max="1" show-input></el-slider>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> SentryRelease </template>
        <el-input v-model="options.SentryRelease"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> SentryEnvironment </template>
        <el-input v-model="options.SentryEnvironment"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> SentryServerName </template>
        <el-input v-model="options.SentryServerName"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> SentryDebug </template>
        <el-switch v-model="options.SentryDebug" active-text="true" inactive-text="false" />
      </el-descriptions-item>
    </el-descriptions>
  </div>
  <!-- WebRTC Setting -->
  <div class="card content-box">
    <!-- WebRTC -->
    <el-descriptions :border="true">
      <template #title> WebRTC Setting </template>
      <el-descriptions-item>
        <template #label> WebRTCConnectionIdleTimeout </template>
        <el-input v-model="options.WebRTCConnectionIdleTimeout"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> WebRTCLogLevel</template>
        <!-- TODO:美化代码 -->
        <el-select v-model="options.WebRTCLogLevel" placeholder="Select log level">
          <el-option label="verbose" value="verbose"></el-option>
          <el-option label="info" value="info"></el-option>
          <el-option label="warning" value="warning"></el-option>
          <el-option label="error" value="error"></el-option>
        </el-select>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> WebRTCMinPort </template>
        <el-input-number v-model="options.WebRTCMinPort" :min="0" :max="65535" />
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> WebRTCMaxPort </template>
        <el-input-number v-model="options.WebRTCMaxPort" :min="0" :max="65535" />
      </el-descriptions-item>
    </el-descriptions>
  </div>
  <!-- TCP Setting -->
  <div class="card content-box">
    <!-- TCP -->
    <el-descriptions :border="true">
      <template #title> TCP Setting </template>
      <el-descriptions-item>
        <template #label> TcpForwardAddr </template>
        <el-input v-model="options.TCPForwardAddr"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> TcpForwardHostPrefix</template>
        <el-input v-model="options.TCPForwardHostPrefix"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> TcpForwardConnections </template>
        <el-input-number v-model="options.TCPForwardConnections" :min="1" :max="10" />
      </el-descriptions-item>
    </el-descriptions>
  </div>
  <!-- Log Setting -->
  <div class="card content-box">
    <!-- Log -->
    <el-descriptions :border="true">
      <template #title> Log Setting </template>
      <el-descriptions-item>
        <template #label> LogFile </template>
        <el-input v-model="options.LogFile"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> LogFileMaxSize</template>
        <el-input-number v-model="options.LogFileMaxSize" :min="0" :max="100" />
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> LogFileMaxCount </template>
        <el-input-number v-model="options.LogFileMaxCount" :min="0" :max="100" />
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> LogLevel </template>
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
    <el-descriptions :border="true">
      <template #title> Service {{ index + 1 }} Setting </template>
      <template #extra>
        <el-button v-if="index === services.length - 1" type="primary" @click="addService">Add Service</el-button>
        <el-button type="danger" @click="removeService(index)">Delete</el-button>
      </template>
      <el-descriptions-item>
        <template #label> HostPrefix </template>
        <el-input v-model="service.HostPrefix"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> RemoteTCPPort </template>
        <el-input-number v-model="service.RemoteTCPPort" :min="0" :max="65535" />
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> RemoteTCPRandom </template>
        <el-switch v-model="service.RemoteTCPRandom" active-text="true" inactive-text="false" />
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> LocalURL </template>
        <el-input v-model="service.LocalURL"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> LocalTimeout </template>
        <el-input v-model="service.LocalTimeout"></el-input>
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label> UseLocalAsHTTPHost </template>
        <el-switch v-model="service.UseLocalAsHTTPHost" active-text="true" inactive-text="false" />
      </el-descriptions-item>
    </el-descriptions>
  </div>
  <!-- <div class="card content-box">
    <el-descriptions :border="true">
      <template #title> Setting </template>
      <el-descriptions-item>
        <template #label> HostPrefix </template>
        <el-input v-model="services[0].HostPrefix"></el-input>
      </el-descriptions-item>
      <el-button type="primary" @click="addService">Add Service</el-button>
    </el-descriptions>
    <el-form>
      <el-form-item v-for="(srv, index) in services" :key="index" :label="'Service ' + (index + 1)" :prop="'srv' + index">
        <el-input v-model=""></el-input>
      </el-form-item>
    </el-form>
    <el-button type="primary" @click="addService">Add Service</el-button>
  </div> -->

  <!-- <div class="card content-box">
    <el-form :model="ClientConfig" label-width="140px">
      <el-form-item v-for="(value, key) in options" :key="key" :label="key" :prop="key">
        <el-input v-model="options[key]"></el-input>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="onSubmit"> Create </el-button>
        <el-button>Cancel</el-button>
      </el-form-item>
    </el-form>
  </div> -->
  <el-button type="primary" @click="onSubmit"> Submit</el-button>
</template>

<script setup lang="ts" name="ClientConfigForm">
import { ElMessage } from "element-plus";
import { computed, reactive, ref } from "vue";
import { ClientConfig } from "./interface";

const inputRemote = ref("");
const selectRemote = ref("tcp://");
const remoteOptions = ["tcp://", "tls://"];
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

const onSubmit = () => {
  ElMessage.success("提交的数据为 : " + JSON.stringify(clientConfig));
};
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
