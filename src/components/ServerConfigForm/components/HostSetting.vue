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
            <el-input-number v-model="form.Number" />
          </el-form-item>
        </el-descriptions-item>
        <el-descriptions-item>
          <template #label>
            WithID
            <UsageTooltip :usage-text="ServerConfig.usage['HostWithID']" />
          </template>
          <el-switch v-model="form.WithID" />
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
            <el-form-item :prop="`tableData[${scope.$index}].RegexStr`" :rules="rules.RegexStr">
              <el-input v-if="scope.row.isEdit" v-model="scope.row.RegexStr" />
              <span v-else>{{ scope.row.RegexStr }}</span>
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
import { reactive, ref, watch, watchEffect } from "vue";
import { ElMessage, FormInstance, FormRules } from "element-plus";

interface HostSettingProps {
  setting: ServerConfig.Host;
}
const props = withDefaults(defineProps<HostSettingProps>(), {
  setting: () => ServerConfig.getDefaultHostSetting()
});

const hostSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.Host>>({
  Number: [
    { required: true, message: "Please input host number", trigger: "blur" },
    { type: "number", message: "Please input a number", trigger: "blur" }
  ],
  RegexStr: [
    { required: true, message: "Please input host regex", trigger: "blur" },
    { type: "regexp", message: "Please input a valid regex", trigger: "blur" }
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
  tableData: props.setting.RegexStr ? props.setting.RegexStr.map(item => ({ RegexStr: item, isEdit: false })) : [],
  WithID: props.setting.WithID
});

const emit = defineEmits(["update:setting"]);
watch(
  () => props.setting,
  newSetting => {
    if (JSON.stringify(form) === JSON.stringify(newSetting)) return;
    form.Number = newSetting.Number;
    form.tableData = Array.isArray(newSetting.RegexStr)
      ? newSetting.RegexStr.map(item => ({ RegexStr: item, isEdit: false }))
      : [];
    form.WithID = newSetting.WithID;
  },
  { deep: true }
);

let prevFormState = JSON.stringify(form);
watchEffect(() => {
  const currentFormState = JSON.stringify(form);
  if (currentFormState !== prevFormState) {
    for (let item of form?.tableData) {
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

const addRow = () => {
  form.tableData.push({
    RegexStr: "",
    isEdit: true
  });
};
const editRow = (index: number) => {
  console.log("editRow:", index);
  form.tableData[index].isEdit = true;
};
const finishEdit = async (index: number) => {
  if (hostSettingRef.value) {
    try {
      await hostSettingRef.value.validateField(`tableData[${index}].RegexStr`);
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
    //check if there is any row is editing
    const allNotEdit = form.tableData.every(item => !item.isEdit);
    if (!allNotEdit) {
      reject(new Error("Please finish editing before submit"));
      return;
    }
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
