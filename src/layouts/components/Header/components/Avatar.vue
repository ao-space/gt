<template>
  <el-dropdown trigger="click">
    <div class="username">{{ username }}</div>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item divided @click="logout">
          <el-icon><SwitchButton /></el-icon>Login out
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { LOGIN_URL } from "@/config";
import { useRouter } from "vue-router";
import { useUserStore } from "@/stores/modules/user";
import { ElMessageBox, ElMessage } from "element-plus";
import { computed } from "vue";

const router = useRouter();
const userStore = useUserStore();

const username = computed(() => userStore.userInfo.name);
// 退出登录
const logout = () => {
  ElMessageBox.confirm("Are you sure to log out?", "Tips", {
    confirmButtonText: "Confirm",
    cancelButtonText: "Cancel",
    type: "warning"
  }).then(async () => {
    //clear token
    userStore.setToken("");
    router.replace(LOGIN_URL);
    ElMessage.success("Logout success");
  });
};
</script>

<style scoped lang="scss">
.username {
  margin-left: 20px;
  overflow: hidden;
  font-size: 15px;
  color: var(--el-header-text-color);
  cursor: pointer;
}
</style>
