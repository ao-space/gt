<template>
  <el-form ref="NetworkSettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <!-- Network Setting -->
      <el-descriptions :column="2" :border="true">
        <template #title> Network Setting </template>
        <el-descriptions-item>
          <template #label>
            Addr
            <UsageTooltip :usage-text="ServerConfig.usage['Addr']" />
          </template>
          <el-form-item prop="Addr">
            <el-input v-model="localSetting.Addr" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            TLSAddr
            <UsageTooltip :usage-text="ServerConfig.usage['TLSAddr']" />
          </template>
          <el-form-item prop="TLSAddr">
            <el-input v-model="localSetting.TLSAddr" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            TLSMinVersion
            <UsageTooltip :usage-text="ServerConfig.usage['TLSMinVersion']" />
          </template>
          <el-form-item prop="TLSMinVersion">
            <el-select v-model="localSetting.TLSMinVersion" placeholder="Please select TLSMinVersion">
              <el-option label="tls1.1" value="tls1.1" />
              <el-option label="tls1.2" value="tls1.2" />
              <el-option label="tls1.3" value="tls1.3" />
            </el-select>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            STUNAddr
            <UsageTooltip :usage-text="ServerConfig.usage['STUNAddr']" />
          </template>
          <el-form-item prop="STUNAddr">
            <el-input v-model="localSetting.STUNAddr" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            SNIAddr
            <UsageTooltip :usage-text="ServerConfig.usage['SNIAddr']" />
          </template>
          <el-form-item prop="SNIAddr">
            <el-input v-model="localSetting.SNIAddr" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            HTTPMUXHeader
            <UsageTooltip :usage-text="ServerConfig.usage['HTTPMUXHeader']" />
          </template>
          <el-form-item prop="HTTPMUXHeader">
            <el-input v-model="localSetting.HTTPMUXHeader" />
          </el-form-item>
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-form>
  <el-button type="primary" @click="validateForm">Validate</el-button>
</template>
<script setup name="NetworkSetting" lang="ts">
import { FormInstance, FormRules } from "element-plus";
import { reactive, ref, watchEffect } from "vue";
import { ServerConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { validatorAddr } from "@/utils/eleValidate";

interface NetworkSettingProps {
  setting: ServerConfig.NetworkSetting;
}
const props = withDefaults(defineProps<NetworkSettingProps>(), {
  setting: () => ServerConfig.defaultNetworkSetting
});
const localSetting = reactive<ServerConfig.NetworkSetting>({ ...props.setting });

const NetworkSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.NetworkSetting>>({
  Addr: [{ validator: validatorAddr, trigger: "blur" }],
  TLSAddr: [{ validator: validatorAddr, trigger: "blur" }],
  STUNAddr: [{ validator: validatorAddr, trigger: "blur" }],
  SNIAddr: [{ validator: validatorAddr, trigger: "blur" }]
});

const emit = defineEmits(["update:setting"]);
watchEffect(() => {
  emit("update:setting", localSetting);
});
const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (NetworkSettingRef.value) {
      NetworkSettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("Network Setting validation failed, please check your input!"));
        }
      });
    } else {
      reject(new Error("Network Setting is not ready!"));
    }
  });
};
defineExpose({
  validateForm
});
</script>

<style lang="scss">
@import "../index.scss";
</style>
