<template>
  <el-form ref="NetworkSettingRef" :model="localSetting" :rules="rules">
    <div class="card content-box">
      <!-- Network Setting -->
      <el-descriptions :column="2" :border="true">
        <template #title> {{ $t("sconfig.NetworkSetting") }} </template>
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.Addr") }}
            <UsageTooltip :usage-text="$t('susage.Addr')" />
          </template>
          <el-form-item prop="Addr">
            <el-input v-model="localSetting.Addr" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.TLSAddr") }}
            <UsageTooltip :usage-text="$t('susage.TLSAddr')" />
          </template>
          <el-form-item prop="TLSAddr">
            <el-input v-model="localSetting.TLSAddr" />
          </el-form-item>
        </el-descriptions-item>
      </el-descriptions>
      <el-row style="width: 100%">
        <el-collapse style="width: 100%">
          <el-collapse-item>
            <template #title>
              <el-text size="large" style="width: 100%">
                {{ $t("sconfig.DetailSettings") }}
              </el-text>
            </template>
            <el-descriptions :column="2" :border="true">
              <el-descriptions-item>
                <template #label>
                  {{ $t("sconfig.TLSMinVersion") }}
                  <UsageTooltip :usage-text="$t('susage.TLSMinVersion')" />
                </template>
                <el-form-item prop="TLSMinVersion">
                  <el-select v-model="localSetting.TLSMinVersion" :placeholder="$t('sconfig.SelectTLSMin')">
                    <el-option label="tls1.1" value="tls1.1" />
                    <el-option label="tls1.2" value="tls1.2" />
                    <el-option label="tls1.3" value="tls1.3" />
                  </el-select>
                </el-form-item>
              </el-descriptions-item>
              <el-descriptions-item>
                <template #label>
                  {{ $t("sconfig.STUNAddr") }}
                  <UsageTooltip :usage-text="$t('susage.STUNAddr')" />
                </template>
                <el-form-item prop="STUNAddr">
                  <el-input v-model="localSetting.STUNAddr" />
                </el-form-item>
              </el-descriptions-item>
              <el-descriptions-item>
                <template #label>
                  {{ $t("sconfig.STUNLogLevel") }}
                  <UsageTooltip :usage-text="$t('susage.STUNLogLevel')" />
                </template>
                <el-select v-model="localSetting.STUNLogLevel" :placeholder="$t('sconfig.SelectSTUNLogLevel')">
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
                  {{ $t("sconfig.SNIAddr") }}
                  <UsageTooltip :usage-text="$t('susage.SNIAddr')" />
                </template>
                <el-form-item prop="SNIAddr">
                  <el-input v-model="localSetting.SNIAddr" />
                </el-form-item>
              </el-descriptions-item>
              <el-descriptions-item>
                <template #label>
                  {{ $t("sconfig.HTTPMUXHeader") }}
                  <UsageTooltip :usage-text="$t('susage.HTTPMUXHeader')" />
                </template>
                <el-form-item prop="HTTPMUXHeader">
                  <el-input v-model="localSetting.HTTPMUXHeader" />
                </el-form-item>
              </el-descriptions-item>
              <el-descriptions-item>
                <template #label>
                  {{ $t("sconfig.MaxHandShakeOptions") }}
                  <UsageTooltip :usage-text="$t('susage.MaxHandShakeOptions')" />
                </template>
                <el-input-number v-model="localSetting.MaxHandShakeOptions" :min="0" />
              </el-descriptions-item>
            </el-descriptions>
          </el-collapse-item>
        </el-collapse>
      </el-row>
    </div>
  </el-form>
</template>
<script setup name="NetworkSetting" lang="ts">
import { FormInstance, FormRules } from "element-plus";
import { reactive, ref, watchEffect } from "vue";
import { ServerConfig } from "../interface";
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { validatorAddr } from "@/utils/eleValidate";
import i18n from "@/languages";

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
          reject(new Error(i18n.global.t("serror.NetworkSettingValidationFailed")));
        }
      });
    } else {
      reject(new Error(i18n.global.t("serror.NetworkSettingNotReady")));
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
