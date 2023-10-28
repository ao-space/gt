<template>
  <!-- General Setting -->
  <el-form ref="generalSettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title> General Setting </template>
        <!-- ID -->
        <el-descriptions-item>
          <template #label>
            ID
            <UsageTooltip :usage-text="ClientConfig.usage['ID']" />
          </template>
          <el-form-item prop="ID">
            <el-input v-model="localSetting.ID"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- Secret -->
        <el-descriptions-item>
          <template #label>
            Secret
            <UsageTooltip :usage-text="ClientConfig.usage['Secret']" />
          </template>
          <el-form-item prop="Secret">
            <el-input v-model="localSetting.Secret" type="password" show-password></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- ReconnectDelay -->
        <el-descriptions-item>
          <template #label>
            ReconnectDelay
            <UsageTooltip :usage-text="ClientConfig.usage['ReconnectDelay']" />
          </template>
          <el-form-item prop="ReconnectDelay">
            <el-input v-model="localSetting.ReconnectDelay"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- RemoteTimeout -->
        <el-descriptions-item>
          <template #label>
            RemoteTimeout
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteTimeout']" />
          </template>
          <el-form-item prop="RemoteTimeout">
            <el-input v-model="localSetting.RemoteTimeout"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- Remote -->
        <el-descriptions-item>
          <template #label>
            Remote
            <UsageTooltip :usage-text="ClientConfig.usage['Remote']" />
          </template>
          <el-form-item prop="Remote">
            <el-input v-model="localSetting.Remote" />
          </el-form-item>
        </el-descriptions-item>
        <!-- RemoteSTUN -->
        <el-descriptions-item>
          <template #label>
            RemoteSTUN
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteSTUN']" />
          </template>
          <el-form-item prop="RemoteSTUN">
            <el-input v-model="localSetting.RemoteSTUN"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- RemoteAPI -->
        <el-descriptions-item>
          <template #label>
            RemoteAPI
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteAPI']" />
          </template>
          <el-form-item prop="RemoteAPI">
            <el-input v-model="localSetting.RemoteAPI"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- RemoteCert -->
        <el-descriptions-item>
          <template #label>
            RemoteCert
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteCert']" />
          </template>
          <el-form-item prop="RemoteCert">
            <el-input v-model="localSetting.RemoteCert"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- RemoteCertInsecure -->
        <el-descriptions-item>
          <template #label>
            RemoteCertInsecure
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteCertInsecure']" />
          </template>
          <el-switch v-model="localSetting.RemoteCertInsecure" active-text="true" inactive-text="false" />
        </el-descriptions-item>
        <!-- RemoteConnections -->
        <el-descriptions-item>
          <template #label>
            RemoteConnections
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteConnections']" />
          </template>
          <el-form-item prop="RemoteConnections">
            <el-input-number v-model="localSetting.RemoteConnections" :min="1" :max="10" />
          </el-form-item>
        </el-descriptions-item>
        <!-- RemoteIdleConnections -->
        <el-descriptions-item>
          <template #label>
            RemoteIdleConnections
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteIdleConnections']" />
          </template>
          <el-form-item prop="RemoteIdleConnections">
            <el-input-number v-model="localSetting.RemoteIdleConnections" :min="0" :max="localSetting.RemoteConnections" />
          </el-form-item>
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-form>
</template>

<script setup name="GeneralSetting" lang="ts">
import { reactive, ref, watchEffect } from "vue";
import { ClientConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import type { FormInstance, FormRules } from "element-plus";
import { validatorTimeFormat } from "@/utils/eleValidate";

interface GeneralSettingProps {
  setting: ClientConfig.GeneralSetting;
}

const props = withDefaults(defineProps<GeneralSettingProps>(), {
  setting: () => ClientConfig.defaultGeneralSetting
});
const localSetting = reactive<ClientConfig.GeneralSetting>({ ...props.setting });

//Sync with parent: props.setting -> localSetting
watchEffect(() => {
  Object.assign(localSetting, props.setting);
});

const emit = defineEmits(["update:setting"]);
//Sync with parent: localSetting -> emit("update:setting")
watchEffect(() => {
  emit("update:setting", localSetting);
});

//Form Related
const generalSettingRef = ref<FormInstance>();
const validatorRemoteIdleConnections = (rule: any, value: number, callback: any) => {
  if (value < 0 || value > localSetting.RemoteConnections) {
    callback(new Error("Please input RemoteIdleConnections between 0 and RemoteConnections"));
  } else {
    callback();
  }
};
const validatorRemote = (rule: any, value: any, callback: any) => {
  console.log("Calling validatorRemote");
  if (!value) {
    callback();
  } else if (value.startsWith("tls://") || value.startsWith("tcp://")) {
    console.log("Valid remote format");
    callback();
  } else {
    console.log("Invalid remote format");
    callback(new Error("Please enter a valid Remote format (tcp:// or tls://)"));
  }
};
const validatorRemoteAPI = (rule: any, value: any, callback: any) => {
  console.log("Calling validatorRemoteAPI");
  if (!value) {
    callback();
  } else if (value.startsWith("http://") || value.startsWith("https://")) {
    console.log("Valid remoteAPI format");
    callback();
  } else {
    console.log("Invalid remoteAPI format");
    callback(new Error("Please enter a valid RemoteAPI format (http:// or https://)"));
  }
};
const rules = reactive<FormRules<ClientConfig.GeneralSetting>>({
  ID: [{ required: true, message: "Please input ID", trigger: "blur" }],
  Secret: [{ message: "Please input Secret", trigger: "blur" }],
  ReconnectDelay: [{ validator: validatorTimeFormat, trigger: "blur" }],
  RemoteTimeout: [{ validator: validatorTimeFormat, trigger: "blur" }],
  Remote: [{ validator: validatorRemote, trigger: "blur" }],
  RemoteAPI: [{ validator: validatorRemoteAPI, trigger: "blur" }],
  RemoteConnections: [
    { type: "number", message: "Please input RemoteConnections", trigger: "blur" },
    {
      type: "number",
      min: 1,
      max: 10,
      message: "Please input RemoteConnections between 1 and 10",
      trigger: "blur"
    }
  ],
  RemoteIdleConnections: [
    { type: "number", message: "Please input RemoteIdleConnections", trigger: "blur" },
    {
      validator: validatorRemoteIdleConnections,
      trigger: "change"
    }
  ]
});

const checkRemoteSetting = (): Promise<void> => {
  return new Promise<void>((resolve, reject) => {
    const isRemoteEmpty = !localSetting.Remote?.trim();
    const isRemoteAPIEmpty = !localSetting.RemoteAPI?.trim();

    if (isRemoteEmpty && isRemoteAPIEmpty) {
      reject(new Error("Please input Remote or RemoteAPI"));
    } else {
      resolve();
    }
  });
};

const validateForm = (): Promise<void> => {
  const validations = [
    checkRemoteSetting(),
    new Promise<void>((resolve, reject) => {
      if (generalSettingRef.value) {
        generalSettingRef.value.validate(valid => {
          if (valid) {
            console.log("General Setting validation passed!");
            resolve();
          } else {
            console.log("General Setting validation failed!");
            reject(new Error("General Setting validation failed, please check your input"));
          }
        });
      } else {
        reject(new Error("General Setting is not ready"));
      }
    })
  ];
  return Promise.all(validations).then(
    () => {
      console.log("General Setting validation passed!");
      return Promise.resolve();
    },
    error => {
      return Promise.reject(error);
    }
  );
};
defineExpose({
  validateForm
});
</script>

<style scoped lang="scss">
@import "../index.scss";
</style>
