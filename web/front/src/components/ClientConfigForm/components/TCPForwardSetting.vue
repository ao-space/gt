<template>
  <el-form ref="TCPForwardSettingRef" :model="localSetting" :rules="rules">
    <!-- TCPForward Setting -->
    <div class="card content-box">
      <!-- TCP -->
      <el-descriptions :column="2" :border="true">
        <template #title>{{ $t("cconfig.TCPForwardSetting") }} </template>
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.TcpForwardAddr") }}
            <UsageTooltip :usage-text="ClientConfig.usage['TCPForwardAddr']" />
          </template>
          <el-form-item prop="TCPForwardAddr">
            <el-input v-model="localSetting.TCPForwardAddr"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.TcpForwardHostPrefix") }}
            <UsageTooltip :usage-text="ClientConfig.usage['TCPForwardHostPrefix']" />
          </template>
          <el-form-item prop="TCPForwardHostPrefix">
            <el-input v-model="localSetting.TCPForwardHostPrefix"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.TcpForwardConnections") }}
            <UsageTooltip :usage-text="ClientConfig.usage['TCPForwardConnections']" />
          </template>
          <el-input-number v-model="localSetting.TCPForwardConnections" :min="1" :max="10" />
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-form>
</template>
<script setup name="TCPForwardSetting" lang="ts">
import { reactive, ref, watchEffect } from "vue";
import { ClientConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { FormInstance, FormRules } from "element-plus";
import { validatorAddr } from "@/utils/eleValidate";

interface TCPForwardSettingProps {
  setting: ClientConfig.TCPForwardSetting;
}

const props = withDefaults(defineProps<TCPForwardSettingProps>(), {
  setting: () => ClientConfig.defaultTCPForwardSetting
});
const localSetting = reactive<ClientConfig.TCPForwardSetting>({ ...props.setting });

//Sync with parent: props.setting -> localSetting
watchEffect(() => {
  Object.assign(localSetting, props.setting);
});

//Sync with parent: localSetting -> emit("update:setting")
const emit = defineEmits(["update:setting"]);
watchEffect(() => {
  emit("update:setting", localSetting);
});

//Form Related
const TCPForwardSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ClientConfig.TCPForwardSetting>>({
  TCPForwardAddr: [{ validator: validatorAddr, trigger: "blur" }]
});

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (TCPForwardSettingRef.value) {
      TCPForwardSettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("TCPForward Setting validation failed, please check your input"));
        }
      });
    } else {
      reject(new Error("TCPForward Setting is not ready"));
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
