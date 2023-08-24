<template>
  <!-- Service Setting -->
  <el-form ref="serviceSettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title> Service {{ index + 1 }} Setting </template>
        <template #extra>
          <!-- <el-button v-if="index === services.length - 1" type="primary" @click="addService">Add Service</el-button> -->
          <el-button v-if="isLast" type="primary" @click="emit('addService')">Add Service</el-button>
          <el-button type="danger" @click="emit('removeService', index)">Delete</el-button>
        </template>
        <el-descriptions-item>
          <template #label>
            HostPrefix
            <UsageTooltip :usage-text="ClientConfig.usage['HostPrefix']" />
          </template>
          <el-input v-model="localSetting.HostPrefix"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            RemoteTCPPort
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteTCPPort']" />
          </template>
          <el-input-number v-model="localSetting.RemoteTCPPort" :min="0" :max="65535" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            RemoteTCPRandom
            <UsageTooltip :usage-text="ClientConfig.usage['RemoteTCPRandom']" />
          </template>
          <el-switch v-model="localSetting.RemoteTCPRandom" active-text="true" inactive-text="false" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            LocalURL
            <UsageTooltip :usage-text="ClientConfig.usage['LocalURL']" />
          </template>
          <el-input v-model="localSetting.LocalURL"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            LocalTimeout
            <UsageTooltip :usage-text="ClientConfig.usage['LocalTimeout']" />
          </template>
          <el-form-item prop="LocalTimeout">
            <el-input v-model="localSetting.LocalTimeout"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            UseLocalAsHTTPHost
            <UsageTooltip :usage-text="ClientConfig.usage['UseLocalAsHTTPHost']" />
          </template>
          <el-switch v-model="localSetting.UseLocalAsHTTPHost" active-text="true" inactive-text="false" />
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-form>
</template>
<script setup name="ServiceSetting" lang="ts">
import { reactive, ref, watchEffect } from "vue";
import { ClientConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { FormInstance, FormRules } from "element-plus";
import { validatorTimeFormat } from "@/utils/eleValidate";

interface ServiceSettingProps {
  setting: ClientConfig.Service;
  index: number;
  isLast: boolean;
}
const props = withDefaults(defineProps<ServiceSettingProps>(), {
  setting: () => ClientConfig.defaultServiceSetting
});
const emit = defineEmits<{
  (e: "update:setting", index: number, setting: ClientConfig.Service): void;
  (e: "removeService", index: number): void;
  (e: "addService"): void;
}>();
const localSetting = reactive<ClientConfig.Service>({ ...props.setting });
watchEffect(() => {
  Object.assign(localSetting, props.setting);
});

const serviceSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ClientConfig.Service>>({
  LocalTimeout: [{ validator: validatorTimeFormat, trigger: "blur" }]
});
watchEffect(() => {
  emit("update:setting", props.index, localSetting);
});
const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (serviceSettingRef.value) {
      serviceSettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("Service Setting validation failed, please check your input"));
        }
      });
    } else {
      reject(new Error("Service Setting is not ready"));
    }
  });
};

defineExpose({
  validateForm
});
</script>

<style lang="scss" scoped>
@import "../index.scss";
</style>
