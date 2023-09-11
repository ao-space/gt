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
  setting: () => ServerConfig.defaultHostSetting
});
const localSetting = reactive<ServerConfig.Host>({ ...props.setting });

watchEffect(() => {
  console.log("props changed");
  console.log("-------------");
  Object.assign(localSetting, props.setting);
  console.log("-------------");
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
  Number: localSetting.Number,
  tableData: localSetting.RegexStr.map(item => ({ RegexStr: item, isEdit: false })),
  WithID: localSetting.WithID
});

const emit = defineEmits(["update:setting"]);
watch(
  () => localSetting,
  () => {
    console.log("localSetting change");
    console.log("-------------");
    form.Number = localSetting.Number;
    localSetting.RegexStr.forEach((item, index) => {
      form.tableData[index].RegexStr = item;
      form.tableData[index].isEdit = false;
    });
    form.WithID = localSetting.WithID;

    console.log(JSON.stringify(localSetting));
    console.log("-------------");
  },
  { immediate: true, deep: true }
);

function hasFormChanged(oldForm: formType, newForm: formType): boolean {
  debugger;
  if (oldForm.Number !== newForm.Number) return true;
  if (oldForm.WithID !== newForm.WithID) return true;
  if (oldForm.tableData.length !== newForm.tableData.length) return true;
  for (let i = 0; i < oldForm.tableData.length; i++) {
    if (oldForm.tableData[i].RegexStr !== newForm.tableData[i].RegexStr) return true;
  }
  return false;
}

watch(
  () => form,
  (newForm: formType, oldForm: formType) => {
    console.log("form change");
    console.log(JSON.stringify(newForm));
    console.log(JSON.stringify(oldForm));
    if (!hasFormChanged(oldForm, newForm)) return;
    console.log("-------------");
    localSetting.Number = newForm.Number;
    localSetting.RegexStr = newForm.tableData.map(item => item.RegexStr);
    localSetting.WithID = newForm.WithID;
    console.log("-------------");
    emit("update:setting", localSetting);
  },
  { immediate: true }
);

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
// TODO: isEdit should be false when validate failed
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
