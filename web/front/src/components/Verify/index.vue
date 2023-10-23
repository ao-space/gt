<template>
  <div class="not-container">
    <h2>Verifying...</h2>
  </div>
</template>

<script setup lang="ts" name="verify">
import { useRoute, useRouter } from "vue-router";
import { ref, onMounted } from "vue";
import { useUserStore } from "@/stores/modules/user";
import { useTabsStore } from "@/stores/modules/tabs";
import { useKeepAliveStore } from "@/stores/modules/keepAlive";
import { verifyKeyApi } from "@/api/modules/login";
import { getTimeState } from "@/utils";
import { ElNotification } from "element-plus";
import { initDynamicRouter } from "@/routers/modules/dynamicRouter";
import { HOME_URL, LOGIN_URL } from "@/config";
import { showFullScreenLoading, tryHideFullScreenLoading } from "@/config/serviceLoading";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const tabsStore = useTabsStore();
const keepAliveStore = useKeepAliveStore();

const key = ref((route.query.key as string) || "");

onMounted(async () => {
  try {
    if (!key.value) {
      throw new Error("Key is empty");
    }
    showFullScreenLoading();
    const { data } = await verifyKeyApi({ key: key.value });
    userStore.setToken(data.token);

    await initDynamicRouter();

    tabsStore.closeMultipleTab();
    keepAliveStore.setKeepAliveName();

    router.push(HOME_URL);
    ElNotification({
      title: getTimeState(),
      message: "Welcome to use GT-Admin",
      type: "success",
      duration: 3000
    });
  } catch (error) {
    router.push(LOGIN_URL);
    ElNotification({
      message: "Verification failed, please try again",
      type: "error",
      duration: 3000
    });
  } finally {
    tryHideFullScreenLoading();
  }
});
</script>

<style scoped lang="scss"></style>
