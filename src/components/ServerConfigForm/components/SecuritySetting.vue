<template>
  <!-- Security Setting -->
  <el-form ref="SecuritySettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title> Security Setting </template>
        <el-descriptions-item>
          <template #label>
            CertFile
            <UsageTooltip :usage-text="ServerConfig.usage['CertFile']" />
          </template>
          <el-form-item prop="CertFile">
            <el-input v-model="localSetting.CertFile" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            KeyFile
            <UsageTooltip :usage-text="ServerConfig.usage['KeyFile']" />
          </template>
          <el-form-item prop="KeyFile">
            <el-input v-model="localSetting.KeyFile" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            AllowAnyClient
            <UsageTooltip :usage-text="ServerConfig.usage['AllowAnyClient']" />
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

interface SecuritySettingProps {
  setting: ServerConfig.SecuritySetting;
}
const props = withDefaults(defineProps<SecuritySettingProps>(), {
  setting: () => ServerConfig.defaultSecuritySetting
});
const localSetting = reactive<ServerConfig.SecuritySetting>({ ...props.setting });
watchEffect(() => {
  Object.assign(localSetting, props.setting);
});

const SecuritySettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.SecuritySetting>>({});

const emit = defineEmits(["update:setting"]);
watchEffect(() => {
  emit("update:setting", localSetting);
});
const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (SecuritySettingRef.value) {
      SecuritySettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("Security Setting validation failed, please check your input!"));
        }
      });
    } else {
      reject(new Error("Security Setting is not ready!"));
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
