<template>
  <!-- General Setting -->
  <el-form ref="generalSettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title>General Setting</template>
        <!-- Users -->
        <el-descriptions-item>
          <template #label>
            Users
            <UsageTooltip :usage-text="ServerConfig.usage['Users']" />
          </template>
          <el-form-item prop="Users">
            <el-input v-model="localSetting.Users"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- AuthAPI -->
        <el-descriptions-item>
          <template #label>
            AuthAPI
            <UsageTooltip :usage-text="ServerConfig.usage['AuthAPI']" />
          </template>
          <el-form-item prop="AuthAPI">
            <el-input v-model="localSetting.AuthAPI"></el-input>
          </el-form-item>
        </el-descriptions-item>
      </el-descriptions>
      <TCPSetting ref="tcpSettingRef" :setting="tcpSetting" @update:setting="updateTCPSetting" />
    </div>
  </el-form>
</template>

<script setup lang="ts" name="GeneralSetting">
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { ServerConfig } from "../interface";
import { reactive, ref, watchEffect } from "vue";
import { FormInstance, FormRules } from "element-plus";
import TCPSetting from "./TCPSetting.vue";

interface GeneralSettingProps {
  setting: ServerConfig.GeneralSetting;
}
const props = withDefaults(defineProps<GeneralSettingProps>(), {
  setting: () => ServerConfig.defaultGeneralSetting
});
const localSetting = reactive<ServerConfig.GeneralSetting>({ ...props.setting });

const generalSettingRef = ref<FormInstance>();
const rules = reactive<FormRules>({});
const emit = defineEmits(["update:setting"]);
watchEffect(() => {
  emit("update:setting", localSetting);
});
let tcpSetting = reactive<ServerConfig.TCP[]>([
  {
    Range: "1-100",
    Number: 100
  },
  {
    Range: "101-200",
    Number: 100
  }
]);
const tcpSettingRef = ref<InstanceType<typeof TCPSetting> | null>(null);
// TODO: assign的局限性不能覆盖
const updateTCPSetting = (setting: ServerConfig.TCP[]) => {
  console.log("updateTCPSetting");
  console.log(setting);
  tcpSetting = setting; //in case of recursive of updateTCPSetting
};

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    //TODO: 子组件校验
    tcpSettingRef.value?.validateForm().then(() => {
      console.log("tcpSettingRef.value?.validateForm() success");
    });
    if (generalSettingRef.value) {
      generalSettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("General Setting validation failed, please check your input"));
        }
      });
    } else {
      reject(new Error("General Setting is not ready"));
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
