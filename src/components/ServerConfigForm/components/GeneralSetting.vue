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
            <el-input v-model="localSetting.UserPath"></el-input>
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
      <el-row :gutter="10" style="width: 100%">
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
  setting: ServerConfig.GeneralSettingProps;
}

const props = withDefaults(defineProps<GeneralSettingProps>(), {
  setting: () => ServerConfig.getDefaultGeneralSettingProps()
});
const localSetting = reactive<ServerConfig.GeneralSettingProps>({
  ...props.setting,
  Host: {
    ...props.setting.Host,
    RegexStr: props.setting.Host.RegexStr || []
  }
});
const tcpSetting = reactive<ServerConfig.TCP[]>([...localSetting.TCPs]);
const hostSetting = reactive<ServerConfig.Host>({
  Number: localSetting.Host.Number,
  RegexStr: [...localSetting.Host.RegexStr],
  WithID: localSetting.Host.WithID
});

//Sync with parent: props.setting -> localSetting
watch(
  () => props.setting,
  newSetting => {
    Object.assign(localSetting, newSetting);
    tcpSetting.splice(0, tcpSetting.length, ...newSetting.TCPs);
    Object.assign(hostSetting, newSetting.Host);
  },
  { deep: true }
);

//Sync tcpSetting and hostSetting -> localSetting
watch(
  () => tcpSetting,
  newSetting => {
    localSetting.TCPs.splice(0, localSetting.TCPs.length, ...newSetting);
  },
  { deep: true }
);
watch(
  () => hostSetting,
  () => {
    localSetting.Host.Number = hostSetting.Number;
    localSetting.Host.RegexStr?.splice(0, localSetting.Host.RegexStr.length, ...hostSetting.RegexStr);
    localSetting.Host.WithID = hostSetting.WithID;
  },
  { deep: true }
);

const emit = defineEmits(["update:setting"]);

//Sync with parent: localSetting -> emit("update:setting")
watch(
  () => localSetting,
  () => {
    emit("update:setting", localSetting);
  },
  { deep: true }
);

//Form Related
const generalSettingRef = ref<FormInstance>();
const tcpSettingRef = ref<InstanceType<typeof TCPSetting> | null>(null);
const hostSettingRef = ref<InstanceType<typeof HostSetting> | null>(null);

const rules = reactive<FormRules>({});

//Sync with child
const updateTCPSetting = (setting: ServerConfig.TCP[]) => {
  tcpSetting.splice(0, tcpSetting.length, ...setting);
};
const updateHostSetting = (setting: ServerConfig.Host) => {
  if (JSON.stringify(hostSetting) === JSON.stringify(setting)) return;
  hostSetting.Number = setting.Number;
  hostSetting.RegexStr.splice(0, hostSetting.RegexStr.length, ...setting.RegexStr);
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
