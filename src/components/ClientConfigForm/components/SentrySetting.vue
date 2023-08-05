<template>
  <!-- Sentry Setting -->
  <el-form ref="sentrySettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title> Sentry Setting </template>
        <!-- SentryDSN -->
        <el-descriptions-item>
          <template #label>
            SentryDSN
            <UsageTooltip :usage-text="ClientConfig.usage['SentryDSN']" />
          </template>
          <el-input v-model="localSetting.SentryDSN"></el-input>
        </el-descriptions-item>
        <!-- SentryServerName -->
        <el-descriptions-item>
          <template #label>
            SentryServerName
            <UsageTooltip :usage-text="ClientConfig.usage['SentryServerName']" />
          </template>
          <el-input v-model="localSetting.SentryServerName"></el-input>
        </el-descriptions-item>
        <!-- SentryLevel -->
        <el-descriptions-item>
          <template #label>
            SentryLevel
            <UsageTooltip :usage-text="ClientConfig.usage['SentryLevel']" />
          </template>
          <el-checkbox-group v-model="localSetting.SentryLevel">
            <el-checkbox v-for="option in SentryLevelOptions" :key="option" :label="option" />
          </el-checkbox-group>
        </el-descriptions-item>
      </el-descriptions>
      <!-- SentrySampleRate -->
      <el-descriptions :column="1" :border="true">
        <el-descriptions-item>
          <template #label>
            SentrySampleRate
            <UsageTooltip :usage-text="ClientConfig.usage['SentrySampleRate']" />
          </template>
          <el-slider
            style="width: 400px"
            v-model="localSetting.SentrySampleRate"
            :step="0.1"
            :min="0"
            :max="1"
            show-input
          ></el-slider>
        </el-descriptions-item>
      </el-descriptions>
      <el-descriptions :column="2" :border="true">
        <!-- SentryRelease -->
        <el-descriptions-item>
          <template #label>
            SentryRelease
            <UsageTooltip :usage-text="ClientConfig.usage['SentryRelease']" />
          </template>
          <el-input v-model="localSetting.SentryRelease"></el-input>
        </el-descriptions-item>
        <!-- SentryEnvironment -->
        <el-descriptions-item>
          <template #label>
            SentryEnvironment
            <UsageTooltip :usage-text="ClientConfig.usage['SentryEnvironment']" />
          </template>
          <el-input v-model="localSetting.SentryEnvironment"></el-input>
        </el-descriptions-item>
        <!-- SentryDebug -->
        <el-descriptions-item>
          <template #label>
            SentryDebug
            <UsageTooltip :usage-text="ClientConfig.usage['SentryDebug']" />
          </template>
          <el-switch v-model="localSetting.SentryDebug" active-text="true" inactive-text="false" />
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-form>
</template>
<script setup name="SentrySetting" lang="ts">
import type { FormInstance, FormRules } from "element-plus";
import { reactive, ref, watchEffect } from "vue";
import { ClientConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";

interface SentrySettingProps {
  setting: ClientConfig.SentrySetting;
}

const SentryLevelOptions = ["trace", "debug", "info", "warn", "error", "fatal", "panic"];

const props = withDefaults(defineProps<SentrySettingProps>(), {
  setting: () => ClientConfig.defaultSentrySetting
});
const localSetting = reactive<ClientConfig.SentrySetting>({ ...props.setting });

const sentrySettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ClientConfig.SentrySetting>>({});
const emit = defineEmits(["update:setting"]);
watchEffect(() => {
  emit("update:setting", localSetting);
});

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (sentrySettingRef.value) {
      sentrySettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("SentrySetting validation failed, please check your input"));
        }
      });
    } else {
      reject(new Error("SentrySetting is not ready"));
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
