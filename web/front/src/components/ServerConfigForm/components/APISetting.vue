<template>
  <el-form ref="APISettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title> {{ $t("sconfig.APISetting") }} </template>
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.APIAddr") }}
            <UsageTooltip :usage-text="$t('susage[\'APIAddr\']')" />
          </template>
          <el-form-item prop="APIAddr">
            <el-input v-model="localSetting.APIAddr" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.APITLSMinVersion") }}
            <UsageTooltip :usage-text="$t('susage[\'APITLSMinVersion\']')" />
          </template>
          <el-form-item prop="APITLSMinVersion">
            <el-select v-model="localSetting.APITLSMinVersion" :placeholder="$t('sconfig.SelectApiTLSMin')">
              <el-option label="tls1.1" value="tls1.1" />
              <el-option label="tls1.2" value="tls1.2" />
              <el-option label="tls1.3" value="tls1.3" />
            </el-select>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.APICertFile") }}
            <UsageTooltip :usage-text="$t('susage[\'APICertFile\']')" />
          </template>
          <el-form-item prop="APICertFile">
            <el-input v-model="localSetting.APICertFile" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.APIKeyFile") }}
            <UsageTooltip :usage-text="$t('susage[\'APIKeyFile\']')" />
          </template>
          <el-form-item prop="APIKeyFile">
            <el-input v-model="localSetting.APIKeyFile" />
          </el-form-item>
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-form>
</template>
<script setup name="APISetting" lang="ts">
import { FormInstance, FormRules } from "element-plus";
import { reactive, ref, watchEffect } from "vue";
import { ServerConfig } from "../interface";
import { validatorAddr } from "@/utils/eleValidate";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import i18n from "@/languages";

interface APISettingProps {
  setting: ServerConfig.APISetting;
}

const props = withDefaults(defineProps<APISettingProps>(), {
  setting: () => ServerConfig.defaultAPISetting
});
const localSetting = reactive<ServerConfig.APISetting>({ ...props.setting });

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
const APISettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.APISetting>>({
  APIAddr: [{ validator: validatorAddr, trigger: "blur" }]
});

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (APISettingRef.value) {
      APISettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error(i18n.global.t("serror.APISettingValidationFailed")));
        }
      });
    } else {
      reject(new Error(i18n.global.t("serror.APISettingNotReady")));
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
