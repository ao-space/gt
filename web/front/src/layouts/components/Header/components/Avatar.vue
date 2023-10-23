<template>
  <el-dropdown trigger="click">
    <div class="username">{{ username }}</div>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item id="userSetting" @click="openDialog('infoRef')">
          <el-icon><User /></el-icon>User Setting
        </el-dropdown-item>
        <el-dropdown-item @click="logout">
          <el-icon><SwitchButton /></el-icon>Log out
        </el-dropdown-item>
        <el-dropdown-item divided @click="restart">
          <el-icon><Refresh /></el-icon>Restart System
        </el-dropdown-item>
        <el-dropdown-item @click="shutdown">
          <el-icon><SwitchButton /></el-icon>Shutdown System
        </el-dropdown-item>
        <el-dropdown-item @click="kill">
          <el-icon><SwitchButton /></el-icon>Terminate System
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
  <!-- infoDialog -->
  <InfoDialog ref="infoRef"></InfoDialog>
</template>

<script setup lang="ts">
import { LOGIN_URL } from "@/config";
import { useRouter } from "vue-router";
import { useUserStore } from "@/stores/modules/user";
import { ElMessageBox, ElMessage } from "element-plus";
import { computed, ref } from "vue";
import { restartServerApi, stopServerApi, killServerApi } from "@/api/modules/server";
import InfoDialog from "./InfoDialog.vue";

const router = useRouter();
const userStore = useUserStore();

const username = computed(() => userStore.userInfo.name);
const clearToken = () => {
  userStore.setToken("");
  router.replace(LOGIN_URL);
};

const infoRef = ref<InstanceType<typeof InfoDialog> | null>(null);
const openDialog = (ref: string) => {
  if (ref == "infoRef") infoRef.value?.openDialog();
};

const logout = () => {
  ElMessageBox.confirm("Are you sure to log out?", "Tips", {
    confirmButtonText: "Confirm",
    cancelButtonText: "Cancel",
    type: "warning"
  }).then(async () => {
    clearToken();
    ElMessage.success("Logout success");
  });
};
const restart = () => {
  ElMessageBox.confirm("Are you sure to restart the system?", "Tips", {
    confirmButtonText: "Confirm",
    cancelButtonText: "Cancel",
    type: "warning"
  })
    .then(async () => {
      try {
        await restartServerApi();
        ElMessage.success("restart success!");
        window.close();
      } catch (e) {
        if (e instanceof Error) {
          ElMessage.error(e.message);
        } else {
          ElMessage.error("restart failed");
        }
      }
    })
    .catch(() => {
      ElMessage.info("restart canceled");
    });
};
const shutdown = () => {
  ElMessageBox.confirm("Are you sure to shutdown the system?", "Tips", {
    confirmButtonText: "Confirm",
    cancelButtonText: "Cancel",
    type: "warning"
  })
    .then(async () => {
      try {
        await stopServerApi();
        clearToken();
        ElMessage.success("shutdown success");
        window.close();
      } catch (e) {
        if (e instanceof Error) {
          ElMessage.error(e.message);
        } else {
          ElMessage.error("shutdown failed");
        }
      }
    })
    .catch(() => {
      ElMessage.info("shutdown canceled");
    });
};
const kill = () => {
  ElMessageBox.confirm("Are you sure to kill the system?", "Tips", {
    confirmButtonText: "Confirm",
    cancelButtonText: "Cancel",
    type: "warning"
  })
    .then(async () => {
      try {
        await killServerApi();
        clearToken();
        ElMessage.success("kill success");
        window.close();
      } catch (e) {
        if (e instanceof Error) {
          ElMessage.error(e.message);
        } else {
          ElMessage.error("kill failed");
        }
      }
    })
    .catch(() => {
      ElMessage.info("kill canceled");
    });
};
</script>

<style scoped lang="scss">
.username {
  overflow: hidden;
  font-size: 16px;
  color: var(--el-header-text-color);
  cursor: pointer;
}
</style>
