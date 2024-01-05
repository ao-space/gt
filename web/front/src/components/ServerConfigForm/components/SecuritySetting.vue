<template>
  <!-- Security Setting -->
  <el-form ref="SecuritySettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title> {{ $t("sconfig.SecuritySetting") }}</template>
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.CertFile") }}
            <UsageTooltip :usage-text="$t('susage[\'CertFile\']')" />
          </template>
          <el-form-item prop="CertFile">
            <el-input v-model="localSetting.CertFile" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.KeyFile") }}
            <UsageTooltip :usage-text="$t('susage[\'KeyFile\']')" />
          </template>
          <el-form-item prop="KeyFile">
            <el-input v-model="localSetting.KeyFile" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.AllowAnyClient") }}
            <UsageTooltip :usage-text="$t('susage[\'AllowAnyClient\']')" />
          </template>
          <el-form-item prop="AllowAnyClient">
            <el-switch v-model="localSetting.AllowAnyClient" />
          </el-form-item>
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-form>
</template>
<script setup name="SecuritySetting" lang="ts">
import { FormInstance, FormRules } from "element-plus";
import { reactive, ref, watchEffect } from "vue";
import { ServerConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import i18n from "@/languages";

interface SecuritySettingProps {
  setting: ServerConfig.SecuritySetting;
}
const props = withDefaults(defineProps<SecuritySettingProps>(), {
  setting: () => ServerConfig.defaultSecuritySetting
});
const localSetting = reactive<ServerConfig.SecuritySetting>({ ...props.setting });

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
const SecuritySettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.SecuritySetting>>({});

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (SecuritySettingRef.value) {
      SecuritySettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error(i18n.global.t("serror.SecuritySettingValidationFailed")));
        }
      });
    } else {
      reject(new Error(i18n.global.t("serror.SecuritySettingNotReady")));
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
