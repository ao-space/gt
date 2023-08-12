<template>
  <el-card>
    <el-form ref="hostSettingRef" :model="form">
      <el-divider content-position="left">Host Setting</el-divider>
      <el-descriptions :column="2" :border="true">
        <el-descriptions-item>
          <template #label>
            HostNumber
            <UsageTooltip :usage-text="ServerConfig.usage['HostNumber']" />
          </template>
          <el-form-item prop="Number" :rules="rules.Number">
            <el-input v-model="form.Number" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            WithID
            <UsageTooltip :usage-text="ServerConfig.usage['HostWithID']" />
          </template>
          <el-switch v-model="localSetting.WithID" />
        </el-descriptions-item>
      </el-descriptions>
      <el-table :data="form.tableData" table-layout="auto" show-overflow-tooltip highlight-current-row>
        <template #empty> Please Add A Host Regex</template>
        <el-table-column type="index"></el-table-column>
        <el-table-column prop="Regex">
          <template #header>
            <div>
              HostRegex
              <UsageTooltip :usage-text="ServerConfig.usage['HostRegex']" />
            </div>
          </template>
          <template #default="scope">
            <el-form-item :prop="`tableData[${scope.$index}].Regex`" :rules="rules.Regex">
              <el-input v-if="scope.row.isEdit" v-model="scope.row.Regex" />
              <span v-else>{{ scope.row.Regex }}</span>
            </el-form-item>
          </template>
        </el-table-column>
        <el-table-column fixed="right">
          <template #header>
            <div>Operation</div>
          </template>
          <template #default="scope">
            <el-button v-if="scope.row.isEdit" icon="Check" type="success" size="small" @click="finishEdit(scope.$index)">
              Done
            </el-button>
            <el-button v-else type="primary" icon="Edit" size="small" @click="editRow(scope.$index)">Edit</el-button>
            <el-button icon="Delete" type="danger" size="small" @click="deleteRow(scope.$index)" />
          </template>
        </el-table-column>
      </el-table>
      <el-button icon="Plus" style="width: 100%" @click="addRow">Add</el-button>
    </el-form>
  </el-card>
</template>
<script setup name="HostSetting" lang="ts">
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { ServerConfig } from "../interface";
import { reactive, ref, watch } from "vue";
import { ElMessage, FormInstance, FormRules } from "element-plus";

interface HostSettingProps {
  setting: ServerConfig.HostSetting;
}
const props = withDefaults(defineProps<HostSettingProps>(), {
  setting: () => ServerConfig.defaultHostSetting
});
let localSetting = reactive<ServerConfig.HostSetting>({ ...props.setting });

const hostSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.HostSetting>>({
  Number: [
    { required: true, message: "Please input host number", trigger: "blur" },
    { type: "number", message: "Please input a number", trigger: "blur" }
  ],
  Regex: [
    { required: true, message: "Please input host regex", trigger: "blur" },
    { type: "regexp", message: "Please input a valid regex", trigger: "blur" }
  ]
});

interface tableDataType {
  Regex: string;
  isEdit: boolean;
}
interface formType {
  tableData: tableDataType[];
  Number: number;
  WithID: boolean;
}
const form = reactive<formType>({
  Number: localSetting.Number,
  tableData: localSetting.Regex.map(item => ({ Regex: item, isEdit: false })),
  WithID: localSetting.WithID
});

const emit = defineEmits(["update:setting"]);
watch(
  () => props.setting,
  newSetting => {
    localSetting = { ...newSetting };

    form.Number = localSetting.Number;
    form.tableData = localSetting.Regex.map(item => ({ Regex: item, isEdit: false }));
    form.WithID = localSetting.WithID;
  },
  { deep: true }
);

watch(
  () => form,
  (newForm: formType) => {
    localSetting.Number = newForm.Number;
    localSetting.Regex = newForm.tableData.map(item => item.Regex);
    localSetting.WithID = newForm.WithID;
    emit("update:setting", localSetting);
  },
  { deep: true }
);

const addRow = () => {
  form.tableData.push({
    Regex: "",
    isEdit: true
  });
};
const editRow = (index: number) => {
  form.tableData[index].isEdit = true;
};
const finishEdit = async (index: number) => {
  if (hostSettingRef.value) {
    try {
      await hostSettingRef.value.validateField(`tableData[${index}].Regex`);
      form.tableData[index].isEdit = false;
    } catch (e) {
      ElMessage.error("Please check your host regex input");
    }
  }
};
const deleteRow = (index: number) => {
  form.tableData.splice(index, 1);
};

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (hostSettingRef.value) {
      hostSettingRef.value
        .validate()
        .then(() => {
          resolve();
        })
        .catch(() => {
          reject(new Error("Host Setting Validation Failed, Please Check Your Input"));
        });
    } else {
      reject(new Error("Host Setting is not ready"));
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
