<template>
  <el-card>
    <el-form ref="tcpSettingRef" :model="form">
      <el-divider content-position="left">TCP Setting</el-divider>
      <el-table :data="form.tableData" table-layout="auto" style="width: 100%" show-overflow-tooltip highlight-current-row>
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
              <!-- <el-input v-if="tableData[scope.$index].isEdit" v-model="tableData[scope.$index].Number" /> -->
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
import { reactive, ref, watch } from "vue";
import { ElMessage, FormInstance, FormRules } from "element-plus";
import { validatorPositiveInteger, validatorRange } from "@/utils/eleValidate";

interface TCPSettingProps {
  setting: ServerConfig.TCP[];
}
const props = defineProps<TCPSettingProps>();
const localSetting = reactive<ServerConfig.TCP[]>([...props.setting]);

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
  tableData: localSetting.map(item => {
    return {
      Range: item.Range,
      Number: item.Number,
      isEdit: false
    };
  })
});

const emit = defineEmits(["update:setting"]);
watch(
  () => props.setting,
  newSetting => {
    localSetting.splice(0, localSetting.length, ...newSetting);
    form.tableData = localSetting.map(item => {
      return {
        Range: item.Range,
        Number: item.Number,
        isEdit: false
      };
    });
  },
  { deep: true }
);
watch(
  () => form.tableData,
  newTableData => {
    localSetting.splice(0, localSetting.length, ...newTableData.map(item => ({ Range: item.Range, Number: item.Number })));
    emit("update:setting", localSetting);
  },
  { deep: true }
);

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
