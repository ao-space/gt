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
            <el-select v-model="localSetting.TLSMinVersion" placeholder="Select TLSMinVersion">
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
            STUNLogLevel
            <UsageTooltip :usage-text="ServerConfig.usage['STUNLogLevel']" />
          </template>
          <el-select v-model="localSetting.STUNLogLevel" placeholder="Select STUN log level">
            <el-option label="trace" value="trace"></el-option>
            <el-option label="debug" value="debug"></el-option>
            <el-option label="info" value="info"></el-option>
            <el-option label="warn" value="warn"></el-option>
            <el-option label="error" value="error"></el-option>
            <el-option label="disable" value="disable"></el-option>
          </el-select>
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
        <el-descriptions-item>
          <template #label>
            MaxHandShakeOptions
            <UsageTooltip :usage-text="ServerConfig.usage['LogFileMaxCount']" />
          </template>
          <el-input-number v-model="localSetting.MaxHandShakeOptions" :min="0" />
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-form>
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
const NetworkSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.NetworkSetting>>({
  Addr: [{ validator: validatorAddr, trigger: "blur" }],
  TLSAddr: [{ validator: validatorAddr, trigger: "blur" }],
  STUNAddr: [{ validator: validatorAddr, trigger: "blur" }],
  SNIAddr: [{ validator: validatorAddr, trigger: "blur" }]
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

<style scoped lang="scss">
@import "../index.scss";
</style>
