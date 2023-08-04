<template>
  <el-form ref="TCPForwardSettingRef" :model="localSetting" :rules="rules">
    <!-- TCPForward Setting -->
    <div class="card content-box">
      <!-- TCP -->
      <el-descriptions :column="2" :border="true">
        <template #title> TCPForward Setting</template>
        <el-descriptions-item>
          <template #label>
            TcpForwardAddr
            <UsageTooltip :usage-text="ClientConfig.usage['TCPForwardAddr']" />
          </template>
          <el-input v-model="localSetting.TCPForwardAddr"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            TcpForwardHostPrefix
            <UsageTooltip :usage-text="ClientConfig.usage['TCPForwardHostPrefix']" />
          </template>
          <el-input v-model="localSetting.TCPForwardHostPrefix"></el-input>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            TcpForwardConnections
            <UsageTooltip :usage-text="ClientConfig.usage['TCPForwardConnections']" />
          </template>
          <el-input-number v-model="localSetting.TCPForwardConnections" :min="1" :max="10" />
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-form>
</template>
<script setup name="TCPForwardSetting" lang="ts">
import { reactive, ref } from "vue";
import { ClientConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { FormInstance, FormRules } from "element-plus";

interface TCPForwardSettingProps {
  setting: ClientConfig.TCPForwardSetting;
}

const props = withDefaults(defineProps<TCPForwardSettingProps>(), {
  setting: () => ClientConfig.defaultTCPForwardSetting
});
const localSetting = reactive<ClientConfig.TCPForwardSetting>({ ...props.setting });

const TCPForwardSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ClientConfig.TCPForwardSetting>>({});
const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    TCPForwardSettingRef.value?.validate(valid => {
      if (valid) {
        resolve();
      } else {
        reject();
      }
    });
  });
};
defineExpose({
  validateForm
});
</script>
<style scoped lang="scss">
@import "../index.scss";
</style>
