<template>
  <!-- Log Setting -->
  <el-form ref="LogSettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <!-- Log -->
      <el-descriptions :column="2" :border="true">
        <template #title> Log Setting </template>
        <el-descriptions-item>
          <template #label>
            LogFile
            <UsageTooltip :usage-text="ClientConfig.usage['LogFile']" />
          </template>
          <el-input v-model="localSetting.LogFile"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            LogFileMaxSize
            <UsageTooltip :usage-text="ClientConfig.usage['LogFileMaxSize']" />
          </template>
          <el-input-number v-model="localSetting.LogFileMaxSize" :min="0" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            LogFileMaxCount
            <UsageTooltip :usage-text="ClientConfig.usage['LogFileMaxCount']" />
          </template>
          <el-input-number v-model="localSetting.LogFileMaxCount" :min="0" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            LogLevel
            <UsageTooltip :usage-text="ClientConfig.usage['LogLevel']" />
          </template>
          <el-select v-model="localSetting.LogLevel" placeholder="Select log level">
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
  </el-form>
</template>
<script setup name="LogSetting" lang="ts">
import type { FormInstance, FormRules } from "element-plus";
import { reactive, ref, watchEffect } from "vue";
import { ClientConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
interface LogSettingProps {
  setting: ClientConfig.LogSetting;
}
const props = withDefaults(defineProps<LogSettingProps>(), {
  setting: () => ClientConfig.defaultLogSetting
});

const localSetting = reactive<ClientConfig.LogSetting>({ ...props.setting });
watchEffect(() => {
  Object.assign(localSetting, props.setting);
});

const LogSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ClientConfig.LogSetting>>({});

const emit = defineEmits(["update:setting"]);
watchEffect(() => {
  emit("update:setting", localSetting);
});

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (LogSettingRef.value) {
      LogSettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("Log Setting validation failed, please check your input"));
        }
      });
    } else {
      reject(new Error("Log Setting is not ready"));
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
