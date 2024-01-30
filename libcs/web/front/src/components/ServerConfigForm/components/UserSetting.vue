<template>
  <el-form ref="userSettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title> {{ $t("sconfig.User") }} {{ index + 1 }} {{ $t("sconfig.Setting") }} </template>
        <template #extra>
          <el-button v-if="isLast" type="primary" @click="emit('addUser')">{{ $t("sconfig.AddUser") }}</el-button>
          <el-button type="danger" @click="emit('removeUser', index)">{{ $t("sconfig.Delete") }}</el-button>
        </template>
        <!-- ID -->
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.ID") }}
            <UsageTooltip :usage-text="$t('susage.user[\'ID\']')" />
          </template>
          <el-form-item prop="ID">
            <el-input v-model="localSetting.ID"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <!-- Secret -->
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.Secret") }}
            <UsageTooltip :usage-text="$t('susage.user[\'Secret\']')" />
          </template>
          <el-form-item prop="Secret">
            <el-input v-model="localSetting.Secret" type="password" show-password></el-input>
          </el-form-item>
        </el-descriptions-item>
      </el-descriptions>
      <el-row style="width: 100%">
        <el-collapse style="width: 100%">
          <el-collapse-item name="1">
            <template #title>
              <el-text style="width: 100%" size="large">{{ $t("sconfig.DetailSettings") }} </el-text>
            </template>
            <el-descriptions :column="2" :border="true">
              <!-- Speed -->
              <el-descriptions-item>
                <template #label>
                  {{ $t("sconfig.Speed") }}
                  <UsageTooltip :usage-text="$t('susage.user[\'Speed\']')" />
                </template>
                <el-form-item prop="Speed">
                  <el-input-number v-model="localSetting.Speed" />
                </el-form-item>
              </el-descriptions-item>
              <!-- Connections -->
              <el-descriptions-item>
                <template #label>
                  {{ $t("sconfig.Connections") }}
                  <UsageTooltip :usage-text="$t('susage.user[\'Connections\']')" />
                </template>
                <el-form-item prop="Connections">
                  <el-input-number v-model="localSetting.Connections" />
                </el-form-item>
              </el-descriptions-item>
              <!-- TCP Number -->
              <el-descriptions-item>
                <template #label>
                  {{ $t("sconfig.TCPNumber") }}
                  <UsageTooltip :usage-text="$t('susage[\'TCPNumber\']')" />
                </template>
                <el-form-item prop="TCPNumber">
                  <el-input-number v-model="localSetting.TCPNumber" :min="0" />
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
          </el-collapse-item>
        </el-collapse>
      </el-row>
    </div>
  </el-form>
</template>
<script setup name="UserSetting" lang="ts">
import { FormInstance, FormRules } from "element-plus";
import { reactive, ref, watch } from "vue";
import { ServerConfig } from "../interface";
import TCPSetting from "./TCPSetting.vue";
import HostSetting from "./HostSetting.vue";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import cloneDeep from "lodash/cloneDeep";
import i18n from "@/languages";

const emit = defineEmits<{
  (e: "update:setting", index: number, setting: ServerConfig.UserSetting): void;
  (e: "removeUser", index: number): void;
  (e: "addUser"): void;
}>();
interface UserSettingProps {
  setting: ServerConfig.UserSetting;
  index: number;
  isLast: boolean;
}
const props = withDefaults(defineProps<UserSettingProps>(), {
  setting: () => ServerConfig.getDefaultUserSetting()
});
//use deep clone to avoid changing props
const localSetting = reactive<ServerConfig.UserSetting>(cloneDeep(props.setting));

//use shallow clone to avoid sync in the current component
const tcpSetting = reactive<ServerConfig.TCP[]>(localSetting.TCPs);
const hostSetting = reactive<ServerConfig.Host>(localSetting.Host);

//Sync with parent: props.setting -> localSetting (tcpSetting, hostSetting)
watch(
  () => props.setting,
  () => {
    localSetting.ID = props.setting.ID;
    localSetting.Secret = props.setting.Secret;
    localSetting.Speed = props.setting.Speed;
    localSetting.Connections = props.setting.Connections;
    localSetting.TCPNumber = props.setting.TCPNumber;
    tcpSetting.splice(0, tcpSetting.length, ...props.setting.TCPs);
    hostSetting.Number = props.setting.Host.Number;
    hostSetting.RegexStr.splice(0, hostSetting.RegexStr.length, ...props.setting.Host.RegexStr);
    hostSetting.WithID = props.setting.Host.WithID;
  }
);

//Sync with child: tcpSetting -> localSetting.TCPs
const updateTCPSetting = (setting: ServerConfig.TCP[]) => {
  tcpSetting.splice(0, tcpSetting.length, ...setting);
};
//Sync with child: hostSetting -> localSetting.Host
const updateHostSetting = (setting: ServerConfig.Host) => {
  hostSetting.Number = setting.Number;
  hostSetting.RegexStr = setting.RegexStr;
  hostSetting.WithID = setting.WithID;
};

//Sync with parent: localSetting(tcpSetting,hostSetting) -> emit("update:setting")
watch(
  () => localSetting,
  () => {
    emit("update:setting", props.index, localSetting);
  },
  { deep: true }
);

//Form Related
const userSettingRef = ref<FormInstance>();
const tcpSettingRef = ref<InstanceType<typeof TCPSetting> | null>(null);
const hostSettingRef = ref<InstanceType<typeof HostSetting> | null>(null);
const rules = reactive<FormRules<ServerConfig.UserSetting>>({
  ID: [
    {
      required: true,
      message: i18n.global.t("serror.PleaseInputID"),
      transform(value) {
        return value.trim();
      },
      trigger: "blur"
    }
  ],
  Secret: [{ required: true, message: i18n.global.t("serror.PleaseInputSecret"), trigger: "blur" }]
});

const validateForm = (): Promise<void> => {
  const validations = [
    tcpSettingRef.value?.validateForm(),
    hostSettingRef.value?.validateForm(),
    new Promise<void>((resolve, reject) => {
      if (userSettingRef.value) {
        userSettingRef.value.validate(valid => {
          if (valid) {
            resolve();
          } else {
            reject(new Error(i18n.global.t("serror.UserSettingValidationFailed")));
          }
        });
      } else {
        reject(new Error(i18n.global.t("serror.UserSettingNotReady")));
      }
    })
  ];
  return Promise.all(validations).then(() => {
    console.log(i18n.global.t("serror.UserSettingValidationPassed"));
  });
};

defineExpose({
  validateForm
});
</script>
<style scoped lang="scss">
@import "../index.scss";
</style>
