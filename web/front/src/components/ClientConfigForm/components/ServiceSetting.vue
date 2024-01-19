<template>
  <!-- Service Setting -->
  <el-form ref="serviceSettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <el-descriptions :column="2" :border="true">
        <template #title>{{ $t("cconfig.Service") }} {{ index + 1 }} {{ $t("cconfig.Setting") }} </template>
        <template #extra>
          <el-button v-if="isLast" type="primary" @click="emit('addService')">{{ $t("cconfig.AddService") }}</el-button>
          <el-button type="danger" @click="emit('removeService', props.index)">{{ $t("cconfig.Delete") }}</el-button>
        </template>
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.HostPrefix") }}
            <UsageTooltip :usage-text="$t('cusage[\'HostPrefix\']')" />
          </template>
          <el-form-item prop="HostPrefix">
            <el-input v-model="localSetting.HostPrefix"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.RemoteTCPPort") }}
            <UsageTooltip :usage-text="$t('cusage[\'RemoteTCPPort\']')" />
          </template>
          <el-input-number v-model="localSetting.RemoteTCPPort" :min="0" :max="65535" />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.RemoteTCPRandom") }}
            <UsageTooltip :usage-text="$t('cusage[\'RemoteTCPRandom\']')" />
          </template>
          <el-switch
            v-model="localSetting.RemoteTCPRandom"
            :active-text="$t('cconfig.true')"
            :inactive-text="$t('cconfig.false')"
          />
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.LocalURL") }}
            <UsageTooltip :usage-text="$t('cusage[\'LocalURL\']')" />
          </template>
          <el-form-item prop="LocalURL">
            <el-input v-model="localSetting.LocalURL"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.LocalTimeout") }}
            <UsageTooltip :usage-text="$t('cusage[\'LocalTimeout\']')" />
          </template>
          <el-form-item prop="LocalTimeout">
            <el-input v-model="localSetting.LocalTimeout"></el-input>
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("cconfig.UseLocalAsHTTPHost") }}
            <UsageTooltip :usage-text="$t('cusage[\'UseLocalAsHTTPHost\']')" />
          </template>
          <el-switch
            v-model="localSetting.UseLocalAsHTTPHost"
            :active-text="$t('cconfig.true')"
            :inactive-text="$t('cconfig.false')"
          />
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-form>
</template>
<script setup name="ServiceSetting" lang="ts">
import { reactive, ref, watch, watchEffect } from "vue";
import { ClientConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { FormInstance, FormRules } from "element-plus";
import { validatorLocalURL, validatorTimeFormat } from "@/utils/eleValidate";
import i18n from "@/languages";

interface ServiceSettingProps {
  setting: ClientConfig.Service;
  index: number;
  isLast: boolean;
}

const props = withDefaults(defineProps<ServiceSettingProps>(), {
  setting: () => ClientConfig.defaultServiceSetting
});
const localSetting = reactive<ClientConfig.Service>({ ...props.setting });

const emit = defineEmits<{
  (e: "update:setting", index: number, setting: ClientConfig.Service): void;
  (e: "removeService", index: number): void;
  (e: "addService"): void;
}>();

//Sync with parent: props.setting -> localSetting
watchEffect(() => {
  Object.assign(localSetting, props.setting);
});

//Sync with parent: localSetting -> emit("update:setting")
watch(
  () => localSetting,
  () => {
    emit("update:setting", props.index, localSetting);
  },
  { deep: true }
);

//Form Related
const serviceSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ClientConfig.Service>>({
  LocalURL: [
    { validator: validatorLocalURL, trigger: "blur" },
    { required: true, message: i18n.global.t("cerror.LocalURLIsRequired"), trigger: "blur" }
  ],
  LocalTimeout: [{ validator: validatorTimeFormat, trigger: "blur" }]
});

const checkTCPSetting = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (localSetting.LocalURL?.startsWith("tcp://")) {
      if (!localSetting.RemoteTCPPort && !localSetting.RemoteTCPRandom) {
        reject(new Error(i18n.global.t("cerror.RemoteTCPPortOrRandomRequired")));
      }
    }
    resolve();
  });
};

const validateForm = (): Promise<void> => {
  const validations = [
    checkTCPSetting(),
    new Promise<void>((resolve, reject) => {
      if (serviceSettingRef.value) {
        serviceSettingRef.value.validate(valid => {
          if (valid) {
            resolve();
          } else {
            reject(new Error(i18n.global.t("cerror.ServiceSettingValidationFailedCheckInput")));
          }
        });
      } else {
        reject(new Error(i18n.global.t("cerror.ServiceSettingNotReady")));
      }
    })
  ];
  return Promise.all(validations).then(() => {
    console.log(i18n.global.t("cerror.ServiceSettingValidationPassed"));
  });
};

defineExpose({
  validateForm
});
</script>

<style lang="scss" scoped>
@import "../index.scss";
</style>
