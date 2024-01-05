<template>
  <!-- Sentry Setting -->
  <el-form ref="sentrySettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title> {{ $t("cconfig.SentrySetting") }} </template>
        <!-- SentryDSN -->
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.SentryDSN") }}
            <UsageTooltip :usage-text="$t('cusage[\'SentryDSN\']')" />
          </template>
          <el-form-item prop="SentryDSN">
            <el-input v-model="localSetting.SentryDSN"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- SentryServerName -->
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.SentryServerName") }}
            <UsageTooltip :usage-text="$t('cusage[\'SentryServerName\']')" />
          </template>
          <el-form-item prop="SentryServerName">
            <el-input v-model="localSetting.SentryServerName"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- SentryLevel -->
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.SentryLevel") }}
            <UsageTooltip :usage-text="$t('cusage[\'SentryLevel\']')" />
          </template>
          <el-checkbox-group v-model="localSetting.SentryLevel">
            <el-checkbox v-for="option in SentryLevelOptions" :key="option" :label="option" />
          </el-checkbox-group>
        </el-descriptions-item>
      </el-descriptions>
      <el-row style="width: 100%">
        <el-collapse style="width: 100%">
          <el-collapse-item name="1">
            <template #title>
              <el-text style="width: 100%" size="large">{{ $t("cconfig.DetailSettings") }}</el-text>
            </template>
            <!-- SentrySampleRate -->
            <el-descriptions :column="1" :border="true">
              <el-descriptions-item>
                <template #label>
                  {{ $t("cconfig.SentrySampleRate") }}
                  <UsageTooltip :usage-text="$t('cusage[\'SentrySampleRate\']')" />
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
                  {{ $t("cconfig.SentryRelease") }}
                  <UsageTooltip :usage-text="$t('cusage[\'SentryRelease\']')" />
                </template>
                <el-form-item prop="SentryRelease">
                  <el-input v-model="localSetting.SentryRelease"></el-input>
                </el-form-item>
              </el-descriptions-item>
              <!-- SentryEnvironment -->
              <el-descriptions-item>
                <template #label>
                  {{ $t("cconfig.SentryEnvironment") }}
                  <UsageTooltip :usage-text="$t('cusage[\'SentryEnvironment\']')" />
                </template>
                <el-form-item prop="SentryEnvironment">
                  <el-input v-model="localSetting.SentryEnvironment"></el-input>
                </el-form-item>
              </el-descriptions-item>
              <!-- SentryDebug -->
              <el-descriptions-item>
                <template #label>
                  {{ $t("cconfig.SentryDebug") }}
                  <UsageTooltip :usage-text="$t('cusage[\'SentryDebug\']')" />
                </template>
                <el-switch v-model="localSetting.SentryDebug" active-text="true" inactive-text="false" />
              </el-descriptions-item>
            </el-descriptions>
          </el-collapse-item>
        </el-collapse>
      </el-row>
    </div>
  </el-form>
</template>
<script setup name="SentrySetting" lang="ts">
import type { FormInstance, FormRules } from "element-plus";
import { reactive, ref, watchEffect } from "vue";
import { ClientConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import i18n from "@/languages";

interface SentrySettingProps {
  setting: ClientConfig.SentrySetting;
}

const props = withDefaults(defineProps<SentrySettingProps>(), {
  setting: () => ClientConfig.defaultSentrySetting
});
const localSetting = reactive<ClientConfig.SentrySetting>({ ...props.setting });

//Sync with parent: props.setting -> localSetting
watchEffect(() => {
  Object.assign(localSetting, props.setting);
});

//Form Related
const SentryLevelOptions = ["trace", "debug", "info", "warn", "error", "fatal", "panic"];
const sentrySettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ClientConfig.SentrySetting>>({});

const emit = defineEmits(["update:setting"]);
//Sync with parent: localSetting -> emit("update:setting")
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
          reject(new Error(i18n.global.t("sconfig.SentrySettingValidationFailedCheckInput")));
        }
      });
    } else {
      reject(new Error(i18n.global.t("sconfig.SentrySettingNotReady")));
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
