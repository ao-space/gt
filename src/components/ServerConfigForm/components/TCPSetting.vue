<template>
  <el-card>
    <el-form ref="tcpSettingRef" :model="localSetting" :rules="rules">
      <el-divider content-position="left">TCP Setting </el-divider>
      <el-table :data="tableData" table-layout="auto" style="width: 100%" show-overflow-tooltip highlight-current-row>
        <el-table-column type="index"></el-table-column>
        <el-table-column prop="TCPRange">
          <template #header>
            <div>
              TCPRanges
              <UsageTooltip :usage-text="ServerConfig.usage['TCPRanges']" />
            </div>
          </template>
          <template #default="scope">
            <el-form-item prop="Range">
              <el-input v-if="scope.row.isEdit" v-model="scope.row.TCPRange" />
              <span v-else>{{ scope.row.TCPRange }}</span>
            </el-form-item>
          </template>
        </el-table-column>
        <el-table-column prop="TCPNumber">
          <template #header>
            <div>
              TCPNumbers
              <UsageTooltip :usage-text="ServerConfig.usage['TCPNumbers']" />
            </div>
          </template>
          <template #default="scope">
            <el-form-item prop="Number">
              <el-input v-if="scope.row.isEdit" v-model="scope.row.TCPNumber" />
              <span v-else>{{ scope.row.TCPNumber }}</span>
            </el-form-item>
          </template>
        </el-table-column>
        <el-table-column fixed="right">
          <template #header>
            <div>Operation</div>
          </template>
          <template #default="scope">
            <el-button v-if="scope.row.isEdit" type="success" size="small" @click="finishEdit(scope.$index)">Done</el-button>
            <el-button v-else type="primary" size="small" @click="editRow(scope.$index)">Edit</el-button>
            <el-button type="danger" size="small" @click="deleteRow(scope.$index)">Remove</el-button>
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
import { reactive, ref, watchEffect } from "vue";
import { FormInstance, FormRules } from "element-plus";
import { validatorRange } from "@/utils/eleValidate";

interface TCPSettingProps {
  setting: ServerConfig.TCP;
}
const props = withDefaults(defineProps<TCPSettingProps>(), {
  setting: () => ServerConfig.defaultTCPSetting
});

const localSetting = reactive<ServerConfig.TCP>({ ...props.setting });

const tcpSettingRef = ref<FormInstance>();
const rules = reactive<FormRules<ServerConfig.TCP>>({
  Range: [{ required: true, validator: validatorRange, message: "Please input a valid TCPRange", trigger: "blur" }],
  Number: [{ required: true, type: "number", message: "Please input a valid TCPNumber", trigger: "blur" }]
});

const emit = defineEmits(["update:setting"]);
watchEffect(() => {
  emit("update:setting", localSetting);
});

const tableData = reactive([
  {
    TCPRange: "1-100",
    TCPNumber: "100",
    isEdit: false
  },
  {
    TCPRange: "101-200",
    TCPNumber: "100",
    isEdit: false
  }
]);
const addRow = () => {
  console.log("localSetting", localSetting);
  tableData.push({
    TCPRange: "",
    TCPNumber: "",
    isEdit: true
  });
};
const editRow = (index: number) => {
  tableData[index].isEdit = true;
  console.log(index);
  console.log(tableData[index]);
};
const finishEdit = (index: number) => {
  tableData[index].isEdit = false;
  console.log(index);
  console.log(tableData[index]);
};
const deleteRow = (index: number) => {
  console.log(index);
  console.log(tableData[index]);
  tableData.splice(index, 1);
};

const validateForm = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (tcpSettingRef.value) {
      tcpSettingRef.value.validate(valid => {
        if (valid) {
          resolve();
        } else {
          reject(new Error("TCP Setting form validate failed, please check your input"));
        }
      });
    } else {
      reject(new Error("TCP Setting is not ready, please check your input"));
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
  width: 45%;
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
