<template>
  <el-card>
    <el-form ref="tcpSettingRef" :model="form">
      <el-divider content-position="left">{{ $t("sconfig.TCPSetting") }}</el-divider>
      <el-table :data="form.tableData" table-layout="auto" show-overflow-tooltip highlight-current-row>
        <template #empty> {{ $t("sconfig.AddTcpRanges") }} </template>
        <el-table-column type="index"></el-table-column>
        <el-table-column prop="Range">
          <template #header>
            <div>
              {{ $t("sconfig.TCPRanges") }}
              <UsageTooltip :usage-text="$t('susage[\'TCPRanges\']')" />
            </div>
          </template>
          <template #default="scope">
            <el-form-item :prop="`tableData[${scope.$index}].Range`" :rules="rules.Range">
              <el-input v-if="scope.row.isEdit" v-model="scope.row.Range" />
              <span v-else>{{ scope.row.Range }}</span>
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
            <el-button v-else type="primary" icon="Edit" size="small" @click="editRow(scope.$index)">{{
              $t("sconfig.Edit")
            }}</el-button>
            <el-button icon="Delete" type="danger" size="small" @click="deleteRow(scope.$index)" />
          </template>
        </el-table-column>
      </el-table>
      <el-button icon="Plus" style="width: 100%" @click="addRow">{{ $t("sconfig.Add") }}</el-button>
    </el-form>
  </el-card>
</template>
<script setup lang="ts" name="TCPSetting">
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { ServerConfig } from "../interface";
import { reactive, ref, watch, watchEffect } from "vue";
import { ElMessage, FormInstance, FormRules } from "element-plus";
import { validatorRange } from "@/utils/eleValidate";
import i18n from "@/languages";

interface TCPSettingProps {
  setting: ServerConfig.TCP[];
}
const props = defineProps<TCPSettingProps>();

//Form Related
const tcpSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.TCP>>({
  Range: [
    { required: true, validator: validatorRange, message: i18n.global.t("serror.PleaseInputValidTCPRange"), trigger: "blur" }
  ]
});

interface tableDataType {
  Range: string;
  isEdit: boolean;
}

const form = reactive<{ tableData: tableDataType[] }>({
  tableData: props.setting.map(({ Range }) => ({ Range, isEdit: false }))
});

// Sync with parent: props.setting -> form.tableData
watch(
  () => props.setting,
  newSetting => {
    if (isSettingEqual(form.tableData, newSetting)) return;
    form.tableData.splice(0, form.tableData.length, ...newSetting.map(({ Range }) => ({ Range, isEdit: false })));
  },
  { deep: true }
);
function isSettingEqual(tableData: tableDataType[], setting: ServerConfig.TCP[]): boolean {
  if (tableData.length !== setting.length) return false;
  for (let i = 0; i < tableData.length; i++) {
    if (tableData[i].Range !== setting[i].Range) return false;
  }
  return true;
}

const emit = defineEmits(["update:setting"]);
let prevFormState = JSON.stringify(form);
// Sync with parent: form.tableData -> emit("update:setting")
watchEffect(() => {
  const currentFormState = JSON.stringify(form);
  if (currentFormState !== prevFormState) {
    for (let item of form.tableData) {
      if (item.isEdit) {
        return;
      }
    }
    emit(
      "update:setting",
      form.tableData.map(item => ({ Range: item.Range }))
    );
    prevFormState = currentFormState;
  }
});

// Table Related
const addRow = () => {
  form.tableData.push({
    Range: "",
    isEdit: true
  });
};
const editRow = (index: number) => {
  form.tableData[index].isEdit = true;
};
const finishEdit = async (index: number) => {
  if (tcpSettingRef.value) {
    try {
      await tcpSettingRef.value.validateField(`tableData[${index}].Range`);
      form.tableData[index].isEdit = false;
    } catch (error) {
      ElMessage.error(i18n.global.t("serror.PleaseCheckYourInput"));
    }
  }
};
const deleteRow = (index: number) => {
  form.tableData.splice(index, 1);
};

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    const allNotEdit = form.tableData.every(item => !item.isEdit);
    if (!allNotEdit) {
      reject(new Error(i18n.global.t("serror.PleaseFinishEditingBeforeSubmit")));
      return;
    }
    if (tcpSettingRef.value) {
      tcpSettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error(i18n.global.t("serror.TCPSettingFormValidateFailed")));
        }
      });
    } else {
      reject(new Error(i18n.global.t("serror.TCPSettingNotReady")));
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
