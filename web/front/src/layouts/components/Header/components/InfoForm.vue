<template>
  <el-form ref="infoFormRef" :model="infoForm" :rules="infoFormRules" size="large">
    <el-form-item prop="username">
      <el-input v-model="infoForm.username" placeholder="Username:">
        <template #prefix>
          <el-icon>
            <user />
          </el-icon>
        </template>
      </el-input>
    </el-form-item>
    <el-form-item prop="password">
      <el-input v-model="infoForm.password" type="password" placeholder="Password:" show-password>
        <template #prefix>
          <el-icon>
            <lock />
          </el-icon>
        </template>
      </el-input>
    </el-form-item>
    <div>
      <span style="padding-right: 1em; font-weight: bolder">EnablePprof:</span>
      <el-switch v-model="infoForm.enablePprof" active-text="true" inactive-text="false" />
    </div>
  </el-form>
  <div style="text-align: right">
    <el-button :icon="CircleClose" round size="large" @click="resetForm(infoFormRef)">Reset</el-button>
    <el-button :icon="UserFilled" round size="large" type="primary" @click="changeInfo"> Change </el-button>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { Register } from "@/api/interface";
import { ElMessageBox, ElMessage } from "element-plus";
import { changeInfoApi, getInfoApi } from "@/api/modules/userInfo";
import { useUserStore } from "@/stores/modules/user";
import { CircleClose, UserFilled } from "@element-plus/icons-vue";
import type { ElForm } from "element-plus";

const userStore = useUserStore();

type FormInstance = InstanceType<typeof ElForm>;
const infoFormRef = ref<FormInstance>();
const infoFormRules = reactive({
  username: [{ required: true, message: "Please enter username", trigger: "blur" }],
  password: [{ required: true, message: "Please enter password", trigger: "blur" }]
});

const infoForm = reactive<Register.ReqRegisterForm>({
  username: "",
  password: "",
  enablePprof: false
});

const update = async () => {
  await getInfoApi().then(({ data }) => {
    infoForm.username = data.username;
    infoForm.password = data.password;
    infoForm.enablePprof = data.enablePprof;
  });
};

const changeInfo = async () => {
  ElMessageBox.confirm(
    "Are you sure you want to change your account information? If you want to apply this new change please restart the system!",
    "Warning",
    {
      confirmButtonText: "OK",
      cancelButtonText: "Cancel",
      type: "warning"
    }
  )
    .then(async () => {
      try {
        await infoFormRef.value?.validate();
        const { data } = await changeInfoApi(infoForm);
        userStore.setToken(data.token);
        console.log(data.token);
        ElMessage.success("Change account information success");
      } catch (e) {
        if (e instanceof Error) {
          ElMessage.error(e.message);
        } else {
          ElMessage.error("Failed to change account information");
        }
      }
    })
    .catch(() => {
      ElMessage.info("Cancel change account information");
    });
};

// resetForm
const resetForm = (formEl: FormInstance | undefined) => {
  if (!formEl) return;
  formEl.resetFields();
};

onMounted(() => {
  update();
});
</script>

<style scoped lang="scss"></style>
