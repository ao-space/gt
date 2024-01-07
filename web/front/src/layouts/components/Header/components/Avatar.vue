<template>
  <el-dropdown trigger="click">
    <div class="username">{{ username }}</div>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item id="userSetting" @click="openDialog('infoRef')">
          <el-icon><User /></el-icon>{{ $t("layout_header.UserSetting") }}
        </el-dropdown-item>
        <el-dropdown-item @click="logout">
          <el-icon><SwitchButton /></el-icon>{{ $t("layout_header.Logout") }}
        </el-dropdown-item>
        <el-dropdown-item divided @click="restart">
          <el-icon><Refresh /></el-icon>{{ $t("layout_header.RestartSystem") }}
        </el-dropdown-item>
        <el-dropdown-item @click="shutdown">
          <el-icon><SwitchButton /></el-icon>{{ $t("layout_header.ShutdownSystem") }}
        </el-dropdown-item>
        <el-dropdown-item @click="kill">
          <el-icon><SwitchButton /></el-icon>{{ $t("layout_header.TerminateSystem") }}
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
import i18n from "@/languages";

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
  ElMessageBox.confirm(i18n.global.t("layout_header.ConfirmLogout"), i18n.global.t("layout_header.Tips"), {
    confirmButtonText: i18n.global.t("layout_header.Confirm"),
    cancelButtonText: i18n.global.t("layout_header.Cancel"),
    type: "warning"
  }).then(async () => {
    clearToken();
    ElMessage.success(i18n.global.t("layout_header.LogoutSuccess"));
  });
};

const restart = () => {
  ElMessageBox.confirm(i18n.global.t("layout_header.ConfirmRestartSystem"), i18n.global.t("layout_header.Tips"), {
    confirmButtonText: i18n.global.t("layout_header.Confirm"),
    cancelButtonText: i18n.global.t("layout_header.Cancel"),
    type: "warning"
  })
    .then(async () => {
      try {
        await restartServerApi();
        ElMessage.success(i18n.global.t("layout_header.RestartSuccess"));
        window.close();
      } catch (e) {
        if (e instanceof Error) {
          ElMessage.error(e.message);
        } else {
          ElMessage.error(i18n.global.t("layout_header.RestartFailed"));
        }
      }
    })
    .catch(() => {
      ElMessage.info(i18n.global.t("layout_header.RestartCanceled"));
    });
};

const shutdown = () => {
  ElMessageBox.confirm(i18n.global.t("layout_header.ConfirmShutdownSystem"), i18n.global.t("layout_header.Tips"), {
    confirmButtonText: i18n.global.t("layout_header.Confirm"),
    cancelButtonText: i18n.global.t("layout_header.Cancel"),
    type: "warning"
  })
    .then(async () => {
      try {
        await stopServerApi();
        clearToken();
        ElMessage.success(i18n.global.t("layout_header.ShutdownSuccess"));
        window.close();
      } catch (e) {
        if (e instanceof Error) {
          ElMessage.error(e.message);
        } else {
          ElMessage.error(i18n.global.t("layout_header.ShutdownFailed"));
        }
      }
    })
    .catch(() => {
      ElMessage.info(i18n.global.t("layout_header.ShutdownCanceled"));
    });
};

const kill = () => {
  ElMessageBox.confirm(i18n.global.t("layout_header.ConfirmKillSystem"), i18n.global.t("layout_header.Tips"), {
    confirmButtonText: i18n.global.t("layout_header.Confirm"),
    cancelButtonText: i18n.global.t("layout_header.Cancel"),
    type: "warning"
  })
    .then(async () => {
      try {
        await killServerApi();
        clearToken();
        ElMessage.success(i18n.global.t("layout_header.KillSuccess"));
        window.close();
      } catch (e) {
        if (e instanceof Error) {
          ElMessage.error(e.message);
        } else {
          ElMessage.error(i18n.global.t("layout_header.KillFailed"));
        }
      }
    })
    .catch(() => {
      ElMessage.info(i18n.global.t("layout_header.KillCanceled"));
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
