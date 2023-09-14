<template>
  <el-card>
    <el-form ref="tcpSettingRef" :model="form">
      <el-divider content-position="left">TCP Setting</el-divider>
      <el-table :data="form.tableData" table-layout="auto" show-overflow-tooltip highlight-current-row>
        <template #empty> Please Add TCP Ranges and Numbers </template>
        <el-table-column type="index"></el-table-column>
        <el-table-column prop="Range">
          <template #header>
            <div>
              TCPRanges
              <UsageTooltip :usage-text="ServerConfig.usage['TCPRanges']" />
            </div>
          </template>
          <template #default="scope">
            <el-form-item :prop="`tableData[${scope.$index}].Range`" :rules="rules.Range">
              <el-input v-if="scope.row.isEdit" v-model="scope.row.Range" />
              <span v-else>{{ scope.row.Range }}</span>
            </el-form-item>
          </template>
        </el-table-column>
        <el-table-column prop="Number">
          <template #header>
            <div>
              TCPNumbers
              <UsageTooltip :usage-text="ServerConfig.usage['TCPNumbers']" />
            </div>
          </template>
          <template #default="scope">
            <el-form-item :prop="`tableData[${scope.$index}].Number`" :rules="rules.Number">
              <el-input v-if="scope.row.isEdit" v-model.number="scope.row.Number" />
              <span v-else>{{ scope.row.Number }}</span>
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
<script setup lang="ts" name="TCPSetting">
import UsageTooltip from "@/components/UsageTooltip/index.vue";
import { ServerConfig } from "../interface";
import { reactive, ref, watch, watchEffect } from "vue";
import { ElMessage, FormInstance, FormRules } from "element-plus";
import { validatorPositiveInteger, validatorRange } from "@/utils/eleValidate";

interface TCPSettingProps {
  setting: ServerConfig.TCP[];
}
const props = defineProps<TCPSettingProps>();

//Form Related
const tcpSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.TCP>>({
  Range: [{ required: true, validator: validatorRange, message: "Please input a valid TCPRange", trigger: "blur" }],
  Number: [
    {
      required: true,
      type: "number",
      validator: validatorPositiveInteger,
      message: "Please input a valid TCPNumber",
      trigger: "blur"
    }
  ]
});

interface tableDataType {
  Range: string;
  Number: number;
  isEdit: boolean;
}

const form = reactive<{ tableData: tableDataType[] }>({
  tableData: props.setting.map(({ Range, Number }) => ({ Range, Number, isEdit: false }))
});

//Sync with parent: props.setting -> form.tableData
watch(
  () => props.setting,
  newSetting => {
    console.log("props.setting change");
    if (JSON.stringify(form.tableData) === JSON.stringify(newSetting)) return;
    form.tableData.splice(0, form.tableData.length, ...newSetting.map(({ Range, Number }) => ({ Range, Number, isEdit: false })));
  },
  { deep: true }
);

const emit = defineEmits(["update:setting"]);
let prevFormState = JSON.stringify(form);
//Sync with parent: form.tableData -> emit("update:setting")
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
      form.tableData.map(item => ({ Range: item.Range, Number: item.Number }))
    );
    prevFormState = currentFormState;
  }
});

//Table Related
const addRow = () => {
  form.tableData.push({
    Range: "",
    Number: 0,
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
      await tcpSettingRef.value.validateField(`tableData[${index}].Number`);
      form.tableData[index].isEdit = false;
    } catch (error) {
      ElMessage.error("Please check your input");
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
      reject(new Error("Please finish editing before submit"));
      return;
    }
    if (tcpSettingRef.value) {
      tcpSettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("TCP Setting form validate failed, please check your input"));
        }
      });
    } else {
      reject(new Error("TCP Setting is not ready"));
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
