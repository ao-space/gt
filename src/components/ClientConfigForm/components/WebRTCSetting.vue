<template>
  <!-- WebRTC Setting -->
  <el-form ref="WebRTCSettingRef" :model="localSetting" :rules="rules">
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
            <el-input v-model="localSetting.WebRTCConnectionIdleTimeout"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            WebRTCLogLevel
            <UsageTooltip :usage-text="ClientConfig.usage['WebRTCLogLevel']" />
          </template>
          <el-select v-model="localSetting.WebRTCLogLevel" placeholder="Select log level">
            <el-option v-for="option in WebRTCLogLevelOptions" :key="option" :label="option" :value="option" />
          </el-select>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            WebRTCMinPort
            <UsageTooltip :usage-text="ClientConfig.usage['WebRTCMinPort']" />
          </template>
          <el-input-number v-model="localSetting.WebRTCMinPort" :min="0" :max="65535" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            WebRTCMaxPort
            <UsageTooltip :usage-text="ClientConfig.usage['WebRTCMaxPort']" />
          </template>
          <el-input-number v-model="localSetting.WebRTCMaxPort" :min="0" :max="65535" />
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-form>
</template>
<script setup name="WebRTCSetting" lang="ts">
import { validatorTimeFormat } from "@/utils/eleValidate";
import type { FormInstance, FormRules } from "element-plus";
import { reactive, ref, watchEffect } from "vue";
import { ClientConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";

const WebRTCLogLevelOptions = ["verbose", "info", "warning", "error"];

interface WebRTCSettingProps {
  setting: ClientConfig.WebRTCSetting;
}
const props = withDefaults(defineProps<WebRTCSettingProps>(), {
  setting: () => ClientConfig.defaultWebRTCSetting
});

const localSetting = reactive<ClientConfig.WebRTCSetting>({ ...props.setting });
watchEffect(() => {
  Object.assign(localSetting, props.setting);
});

const WebRTCSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ClientConfig.WebRTCSetting>>({
  WebRTCConnectionIdleTimeout: [{ validator: validatorTimeFormat, trigger: "blur" }]
});

const emit = defineEmits(["update:setting"]);
watchEffect(() => {
  emit("update:setting", localSetting);
});

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (WebRTCSettingRef.value) {
      WebRTCSettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("WebRTC Setting validation failed, please check your input"));
        }
      });
    } else {
      reject(new Error("WebRTC Setting is not ready"));
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
