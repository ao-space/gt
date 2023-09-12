<template>
  <el-form ref="APISettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title> API Setting </template>
        <el-descriptions-item>
          <template #label>
            APIAddr
            <UsageTooltip :usage-text="ServerConfig.usage['APIAddr']" />
          </template>
          <el-form-item prop="APIAddr">
            <el-input v-model="localSetting.APIAddr" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            APITLSMinVersion
            <UsageTooltip :usage-text="ServerConfig.usage['APITLSMinVersion']" />
          </template>
          <el-form-item prop="APITLSMinVersion">
            <el-select v-model="localSetting.APITLSMinVersion" placeholder="Select APITLSMinVersion">
              <el-option label="tls1.1" value="tls1.1" />
              <el-option label="tls1.2" value="tls1.2" />
              <el-option label="tls1.3" value="tls1.3" />
            </el-select>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            APICertFile
            <UsageTooltip :usage-text="ServerConfig.usage['APICertFile']" />
          </template>
          <el-form-item prop="APICertFile">
            <el-input v-model="localSetting.APICertFile" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            APIKeyFile
            <UsageTooltip :usage-text="ServerConfig.usage['APIKeyFile']" />
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

interface APISettingProps {
  setting: ServerConfig.APISetting;
}
const props = withDefaults(defineProps<APISettingProps>(), {
  setting: () => ServerConfig.defaultAPISetting
});
const localSetting = reactive<ServerConfig.APISetting>({ ...props.setting });
watchEffect(() => {
  Object.assign(localSetting, props.setting);
});

const APISettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.APISetting>>({
  APIAddr: [{ validator: validatorAddr, trigger: "bur" }]
});

const emit = defineEmits(["update:setting"]);
watchEffect(() => {
  emit("update:setting", localSetting);
});
const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (APISettingRef.value) {
      APISettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("API Setting validation failed, please check your input!"));
        }
      });
    } else {
      reject(new Error("API Setting is not ready!"));
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
