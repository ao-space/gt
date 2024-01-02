<template>
  <el-form ref="infoFormRef" :model="infoForm" :rules="infoFormRules" size="large">
    <el-form-item prop="username">
      <el-input v-model="infoForm.username" :placeholder="$t('layout_header.Username')">
        <template #prefix>
          <el-icon>
            <user />
          </el-icon>
        </template>
      </el-input>
    </el-form-item>
    <el-form-item prop="password">
      <el-input v-model="infoForm.password" type="password" :placeholder="$t('layout_header.Password')" show-password>
        <template #prefix>
          <el-icon>
            <lock />
          </el-icon>
        </template>
      </el-input>
    </el-form-item>
    <div>
      <span style="padding-right: 1em; font-weight: bolder">{{ $t("layout_header.EnablePprof") }}:</span>
      <el-switch v-model="infoForm.enablePprof" active-text="true" inactive-text="false" />
    </div>
  </el-form>
  <div style="text-align: right">
    <el-button :icon="CircleClose" round size="large" @click="resetForm(infoFormRef)">{{ $t("layout_header.Reset") }}</el-button>
    <el-button :icon="UserFilled" round size="large" type="primary" @click="changeInfo">
      >{{ $t("layout_header.Change") }}
    </el-button>
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
import i18n from "@/languages";

const userStore = useUserStore();

type FormInstance = InstanceType<typeof ElForm>;
const infoFormRef = ref<FormInstance>();
const infoFormRules = reactive({
  username: [{ required: true, message: i18n.global.t("layout_header.UsernameRequired"), trigger: "blur" }],
  password: [{ required: true, message: i18n.global.t("layout_header.PasswordRequired"), trigger: "blur" }]
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
  ElMessageBox.confirm(i18n.global.t("layout_header.ChangeInfoWarning"), i18n.global.t("layout_header.Warning"), {
    confirmButtonText: i18n.global.t("layout_header.OK"),
    cancelButtonText: i18n.global.t("layout_header.Cancel"),
    type: "warning"
  })
    .then(async () => {
      try {
        await infoFormRef.value?.validate();
        const { data } = await changeInfoApi(infoForm);
        userStore.setToken(data.token);
        console.log(data.token);
        ElMessage.success(i18n.global.t("layout_header.ChangeInfoSuccess"));
      } catch (e) {
        if (e instanceof Error) {
          ElMessage.error(e.message);
        } else {
          ElMessage.error(i18n.global.t("layout_header.ChangeInfoFailure"));
        }
      }
    })
    .catch(() => {
      ElMessage.info(i18n.global.t("layout_header.CancelChangeInfo"));
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
