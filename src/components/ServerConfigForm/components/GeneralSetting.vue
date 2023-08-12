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
      <el-row style="width: 100%">
        <el-col :span="12">
          <TCPSetting ref="tcpSettingRef" :setting="tcpSetting" @update:setting="updateTCPSetting" />
        </el-col>
        <el-col :span="12">
          <HostSetting ref="hostSettingRef" :setting="hostSetting" @update:setting="updateHostSetting" />
        </el-col>
      </el-row>
    </div>
  </el-form>
</template>

<script setup lang="ts" name="GeneralSetting">
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { ServerConfig } from "../interface";
import { reactive, ref, watch } from "vue";
import { FormInstance, FormRules } from "element-plus";
import TCPSetting from "./TCPSetting.vue";
import HostSetting from "./HostSetting.vue";

interface GeneralSettingProps {
  setting: ServerConfig.GeneralSetting;
}
const props = withDefaults(defineProps<GeneralSettingProps>(), {
  setting: () => ServerConfig.defaultGeneralSetting
});
const localSetting = reactive<ServerConfig.GeneralSetting>({ ...props.setting });
const emit = defineEmits(["update:setting"]);
// Note: the component TCPSetting and HostSetting
// need the type of TCP and Host respectively
// instead of TCPInOptions and HostInOptions

//switch TCPInOptions and HostInOptions to TCP and Host
const tcpSetting = reactive<ServerConfig.TCP[]>(
  localSetting.TCPRanges.map((range, index) => ({
    Range: range,
    Number: parseInt(localSetting.TCPNumbers[index])
  }))
);
const hostSetting = reactive<ServerConfig.Host>({
  Number: localSetting.HostNumber,
  RegexStr: localSetting.HostRegex,
  WithID: localSetting.HostWithID
});

//make sure the consistency
watch(
  () => tcpSetting,
  newSetting => {
    console.log("tcpSetting change");
    localSetting.TCPRanges = newSetting.map(tcp => tcp.Range);
    localSetting.TCPNumbers = newSetting.map(tcp => tcp.Number.toString());
  },
  { deep: true }
);
watch(
  () => hostSetting,
  newSetting => {
    console.log("hostSetting change");
    localSetting.HostNumber = newSetting.Number;
    localSetting.HostRegex = newSetting.RegexStr;
    localSetting.HostWithID = newSetting.WithID;
  },
  { deep: true }
);

//inform parent component to update setting
watch(
  () => localSetting,
  () => {
    emit("update:setting", localSetting);
    console.log("update:setting");
  },
  { deep: true }
);

const generalSettingRef = ref<FormInstance>();
const rules = reactive<FormRules>({});

const tcpSettingRef = ref<InstanceType<typeof TCPSetting> | null>(null);
const hostSettingRef = ref<InstanceType<typeof HostSetting> | null>(null);

const updateTCPSetting = (setting: ServerConfig.TCP[]) => {
  console.log("updateTCPSetting");
  tcpSetting.splice(0, tcpSetting.length, ...setting);
};
const updateHostSetting = (setting: ServerConfig.Host) => {
  console.log("updateHostSetting");
  hostSetting.Number = setting.Number;
  hostSetting.RegexStr = setting.RegexStr;
  hostSetting.WithID = setting.WithID;
};

const validateForm = (): Promise<void> => {
  const validations = [
    tcpSettingRef.value?.validateForm(),
    hostSettingRef.value?.validateForm(),
    new Promise<void>((resolve, reject) => {
      if (generalSettingRef.value) {
        generalSettingRef.value.validate(valid => {
          if (valid) {
            resolve();
          } else {
            reject(new Error("General Setting validation failed, please check your input!"));
          }
        });
      } else {
        reject(new Error("General Setting is not ready!"));
      }
    })
  ];
  return Promise.all(validations).then(() => {
    console.log("General Setting validation passed!");
  });
};

defineExpose({
  validateForm
});
</script>
<style scoped lang="scss">
@import "../index.scss";
.el-row {
  margin-bottom: 20px;
}
.el-row:last-child {
  margin-bottom: 0;
}
.el-col {
  border-radius: 4px;
}
</style>
