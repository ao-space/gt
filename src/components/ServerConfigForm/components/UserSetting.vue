<template>
  <el-form ref="userSettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title> User {{ index + 1 }} Setting </template>
        <template #extra>
          <el-button v-if="isLast" type="primary" @click="emit('addUser')">Add User</el-button>
          <el-button type="danger" @click="emit('removeUser', index)">Delete</el-button>
        </template>
        <el-descriptions-item>
          <template #label>
            ID
            <UsageTooltip :usage-text="ServerConfig.usage.user['ID']" />
          </template>
          <el-form-item prop="ID">
            <el-input v-model="localSetting.ID"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            Secret
            <UsageTooltip :usage-text="ServerConfig.usage.user['Secret']" />
          </template>
          <el-form-item prop="Secret">
            <el-input v-model="localSetting.Secret" type="password" show-password></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            Speed
            <UsageTooltip :usage-text="ServerConfig.usage.user['Speed']" />
          </template>
          <el-form-item prop="Speed">
            <el-input-number v-model="localSetting.Speed" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            Connections
            <UsageTooltip :usage-text="ServerConfig.usage.user['Connections']" />
          </template>
          <el-form-item prop="Connections">
            <el-input-number v-model="localSetting.Connections" />
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
<script setup name="UserSetting" lang="ts">
import { FormInstance, FormRules } from "element-plus";
import { reactive, ref, watch, watchEffect } from "vue";
import { ServerConfig } from "../interface";
import TCPSetting from "./TCPSetting.vue";
import HostSetting from "./HostSetting.vue";
import UsageTooltip from "@/components/UsageTooltip/index.vue";

interface UserSettingProps {
  setting: ServerConfig.UserSetting;
  index: number;
  isLast: boolean;
}
const props = withDefaults(defineProps<UserSettingProps>(), {
  setting: () => ServerConfig.getDefaultUserSetting()
});
const emit = defineEmits<{
  (e: "update:setting", index: number, setting: ServerConfig.UserSetting): void;
  (e: "removeUser", index: number): void;
  (e: "addUser"): void;
}>();
const localSetting = reactive<ServerConfig.UserSetting>({ ...props.setting });
const tcpSetting = reactive<ServerConfig.TCP[]>(localSetting.TCPs);
const hostSetting = reactive<ServerConfig.Host>(localSetting.Host);

watchEffect(() => {
  Object.assign(localSetting, props.setting);
  tcpSetting.splice(0, tcpSetting.length, ...localSetting.TCPs);
  hostSetting.Number = localSetting.Host.Number;
  hostSetting.RegexStr.splice(0, hostSetting.RegexStr.length, ...localSetting.Host.RegexStr);
  hostSetting.WithID = localSetting.Host.WithID;
});

const updateTCPSetting = (setting: ServerConfig.TCP[]) => {
  tcpSetting.splice(0, tcpSetting.length, ...setting);
};
const updateHostSetting = (setting: ServerConfig.Host) => {
  hostSetting.Number = setting.Number;
  hostSetting.RegexStr = setting.RegexStr;
  hostSetting.WithID = setting.WithID;
};

watch(
  () => tcpSetting,
  () => {
    localSetting.TCPs.splice(0, localSetting.TCPs.length, ...tcpSetting);
  },
  { deep: true }
);
watch(
  () => hostSetting,
  () => {
    localSetting.Host.Number = hostSetting.Number;
    localSetting.Host.RegexStr = hostSetting.RegexStr;
    localSetting.Host.WithID = hostSetting.WithID;
  },
  { deep: true }
);
watch(
  () => localSetting,
  () => {
    emit("update:setting", props.index, localSetting);
  },
  { deep: true }
);

const userSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.UserSetting>>({
  ID: [
    {
      required: true,
      message: "Please input ID",
      transform(value) {
        return value.trim();
      },
      trigger: "blur"
    }
  ],
  Secret: [{ required: true, message: "Please input Secret", trigger: "blur" }]
});

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (userSettingRef.value) {
      userSettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("User Setting validation failed, please check your input"));
        }
      });
    } else {
      reject(new Error("User Setting is not ready"));
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
