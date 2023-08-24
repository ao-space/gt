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
          <el-input v-model="inputRemote">
            <template #prepend>
              <el-select v-model="selectRemote">
                <el-option v-for="option in remoteOptions" :key="option" :label="option" :value="option" />
              </el-select>
            </template>
          </el-input>
        </el-descriptions-item>
        <!-- RemoteSTUN -->
        <el-descriptions-item>
          <template #label>
            RemoteSTUN
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteSTUN']" />
          </template>
          <el-input v-model="localSetting.RemoteSTUN"></el-input>
        </el-descriptions-item>
        <!-- RemoteAPI -->
        <el-descriptions-item>
          <template #label>
            RemoteAPI
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteAPI']" />
          </template>
          <el-input v-model="localSetting.RemoteAPI"></el-input>
        </el-descriptions-item>
        <!-- RemoteCert -->
        <el-descriptions-item>
          <template #label>
            RemoteCert
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteCert']" />
          </template>
          <el-input v-model="localSetting.RemoteCert"></el-input>
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
        <!-- Version -->
        <!-- <el-descriptions-item>
          <template #label>
            Version
            <UsageTooltip :usage-text="ClientConfig.usage['Version']" />
          </template>
          <el-switch v-model="localSetting.Version" active-text="true" inactive-text="false" />
        </el-descriptions-item> -->
      </el-descriptions>
    </div>
  </el-form>
</template>

<script setup name="GeneralSetting" lang="ts">
import { computed, reactive, ref, watchEffect } from "vue";
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

const generalSettingRef = ref<FormInstance>();
const validatorRemoteIdleConnections = (rule: any, value: number, callback: any) => {
  if (value < 0 || value > localSetting.RemoteConnections) {
    callback(new Error("Please input RemoteIdleConnections between 0 and RemoteConnections"));
  } else {
    callback();
  }
};
const rules = reactive<FormRules<ClientConfig.GeneralSetting>>({
  ID: [
    {
      required: true,
      message: "Please input ID",
      transform(value) {
        return value.trim();
      },
      trigger: "blur"
    }
  ],
  Secret: [{ required: true, message: "Please input Secret", trigger: "blur" }],
  ReconnectDelay: [{ validator: validatorTimeFormat, trigger: "blur" }],
  RemoteTimeout: [{ validator: validatorTimeFormat, trigger: "blur" }],
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

const inputRemote = ref("");
const selectRemote = ref("tcp://");
const remoteOptions = ["tcp://", "tls://"];
const remote = computed(() => selectRemote.value + inputRemote.value);

const emit = defineEmits(["update:setting"]);
watchEffect(() => {
  localSetting.Remote = remote.value;
  emit("update:setting", localSetting);
});

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (generalSettingRef.value) {
      generalSettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("General Setting validation failed, please check your input"));
        }
      });
    } else {
      reject(new Error("General Setting is not ready"));
    }
  });
};

defineExpose({
  validateForm
});
</script>

<style scoped lang="scss">
@import "../index.scss";
</style>
