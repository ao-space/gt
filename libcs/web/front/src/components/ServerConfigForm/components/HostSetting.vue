<template>
  <el-card>
    <el-form ref="hostSettingRef" :model="form">
      <el-divider content-position="left">{{ $t("sconfig.HostSetting") }}</el-divider>
      <el-descriptions :column="2" :border="true">
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.HostNumber") }}
            <UsageTooltip :usage-text="$t('susage[\'HostNumber\']')" />
          </template>
          <el-form-item prop="Number" :rules="rules.Number">
            <el-input-number v-model="form.Number" :min="0" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            {{ $t("sconfig.WithID") }}
            <UsageTooltip :usage-text="$t('susage[\'HostWithID\']')" />
          </template>
          <el-switch v-model="form.WithID" />
        </el-descriptions-item>
      </el-descriptions>
      <el-table :data="form.tableData" table-layout="auto" show-overflow-tooltip highlight-current-row>
        <template #empty>{{ $t("sconfig.AddHostRegex") }}</template>
        <el-table-column type="index"></el-table-column>
        <el-table-column prop="Regex">
          <template #header>
            <div>
              {{ $t("sconfig.HostRegex") }}
              <UsageTooltip :usage-text="$t('susage[\'HostRegex\']')" />
            </div>
          </template>
          <template #default="scope">
            <el-form-item :prop="`tableData[${scope.$index}].RegexStr`" :rules="rules.RegexStr">
              <el-input v-if="scope.row.isEdit" v-model="scope.row.RegexStr" />
              <span v-else>{{ scope.row.RegexStr }}</span>
            </el-form-item>
          </template>
        </el-table-column>
        <el-table-column fixed="right">
          <template #header>
            <div>{{ $t("sconfig.Operation") }}</div>
          </template>
          <template #default="scope">
            <el-button v-if="scope.row.isEdit" icon="Check" type="success" size="small" @click="finishEdit(scope.$index)">
              {{ $t("sconfig.Done") }}
            </el-button>
            <el-button v-else type="primary" icon="Edit" size="small" @click="editRow(scope.$index)">Edit</el-button>
            <el-button icon="Delete" type="danger" size="small" @click="deleteRow(scope.$index)" />
          </template>
        </el-table-column>
      </el-table>
      <el-button icon="Plus" style="width: 100%" @click="addRow">{{ $t("sconfig.Add") }}</el-button>
    </el-form>
  </el-card>
</template>
<script setup name="HostSetting" lang="ts">
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { ServerConfig } from "../interface";
import { reactive, ref, watch, watchEffect } from "vue";
import { ElMessage, FormInstance, FormRules } from "element-plus";
import i18n from "@/languages";

interface HostSettingProps {
  setting: ServerConfig.Host;
}

const props = withDefaults(defineProps<HostSettingProps>(), {
  setting: () => ServerConfig.getDefaultHostSetting()
});

//Form Related
const hostSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.Host>>({
  Number: [
    { required: true, message: i18n.global.t("serror.PleaseInputHostNumber"), trigger: "blur" },
    { type: "number", message: i18n.global.t("serror.PleaseInputANumber"), trigger: "blur" }
  ],
  RegexStr: [
    { required: true, message: i18n.global.t("serror.PleaseInputHostRegex"), trigger: "blur" },
    { type: "regexp", message: i18n.global.t("serror.PleaseInputAValidRegex"), trigger: "blur" }
  ]
});

interface tableDataType {
  RegexStr: string;
  isEdit: boolean;
}
interface formType {
  tableData: tableDataType[];
  Number: number;
  WithID: boolean;
}

const form = reactive<formType>({
  Number: props.setting.Number,
  tableData: props.setting.RegexStr.map(item => ({ RegexStr: item, isEdit: false })),
  WithID: props.setting.WithID
});

// 同步到父组件: props.setting -> form
watch(
  () => props.setting,
  newSetting => {
    if (isSettingEqual(form, newSetting)) return;
    form.Number = newSetting.Number;
    if (newSetting.RegexStr) {
      form.tableData = newSetting.RegexStr.map(item => ({ RegexStr: item, isEdit: false }));
    } else {
      form.tableData = [];
    }
    form.WithID = newSetting.WithID;
  },
  { deep: true }
);

function isSettingEqual(form: formType, setting: ServerConfig.Host): boolean {
  return (
    form.Number === setting.Number &&
    form.WithID === setting.WithID &&
    form.tableData.length === setting.RegexStr.length &&
    form.tableData.every((item, index) => item.RegexStr === setting.RegexStr[index])
  );
}

const emit = defineEmits(["update:setting"]);
let prevFormState = JSON.stringify(form);

// 同步到父组件: form -> emit("update:setting")
watchEffect(() => {
  const currentFormState = JSON.stringify(form);
  if (currentFormState !== prevFormState) {
    for (let item of form.tableData) {
      if (item.isEdit) {
        return;
      }
    }
    emit("update:setting", {
      Number: form.Number,
      RegexStr: form.tableData.map(item => item.RegexStr),
      WithID: form.WithID
    });
    prevFormState = currentFormState;
  }
});

// 表格相关
const addRow = () => {
  form.tableData.push({
    RegexStr: "",
    isEdit: true
  });
};
const editRow = (index: number) => {
  form.tableData[index].isEdit = true;
};
const finishEdit = async (index: number) => {
  if (hostSettingRef.value) {
    try {
      await hostSettingRef.value.validateField(`tableData[${index}].RegexStr`);
      form.tableData[index].isEdit = false;
    } catch (e) {
      ElMessage.error(i18n.global.t("serror.PleaseCheckYourHostRegexInput"));
    }
  }
};
const deleteRow = (index: number) => {
  form.tableData.splice(index, 1);
};

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    // 检查是否有任何行正在编辑
    const allNotEdit = form.tableData.every(item => !item.isEdit);
    if (!allNotEdit) {
      reject(new Error(i18n.global.t("serror.PleaseFinishEditingBeforeSubmit")));
      return;
    }
    if (hostSettingRef.value) {
      hostSettingRef.value
        .validate()
        .then(() => {
          resolve();
        })
        .catch(() => {
          reject(new Error(i18n.global.t("serror.HostSettingValidationFailed")));
        });
    } else {
      reject(new Error(i18n.global.t("serror.HostSettingNotReady")));
    }
  });
};
defineExpose({
  validateForm
});
</script>

<style scoped lang="scss">
@import "../index.scss";
.el-card {
  margin: 20px 0;
}
:deep(.el-form-item__error) {
  z-index: 100;
}
:deep(.el-descriptions__cell.el-descriptions__label) {
  width: 30%;
}
.el-form {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}
:deep(.el-table__cell) {
  font-size: 16px;
  text-align: center;
}
:deep(.el-form-item__content) {
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
}
</style>
