import router from "@/routers/index";
import { LOGIN_URL } from "@/config";
import { RouteRecordRaw } from "vue-router";
import { ElNotification } from "element-plus";
import { useUserStore } from "@/stores/modules/user";
import { useAuthStore } from "@/stores/modules/auth";

// Import all vue files from the views directory
// ignore the login page and the login form component for static import
const modules = import.meta.glob(["@/views/**/*.vue", "!@/views/login/index.vue", "!@/views/login/components/LoginForm.vue"]);

/**
 * @description Initialize dynamic routes
 */
export const initDynamicRouter = async () => {
  const userStore = useUserStore();
  const authStore = useAuthStore();

  try {
    // 1. Retrieve the menu list
    await authStore.getAuthMenuList();

    // 2. Check if the current user has menu permissions
    if (!authStore.authMenuListGet.length) {
      ElNotification({
        title: "Have no permission to access",
        message: "Current account has no menu permission, please contact the system administrator!",
        type: "warning",
        duration: 3000
      });
      userStore.setToken("");
      router.replace(LOGIN_URL);
      return Promise.reject("No permission");
    }

    // 3. Add dynamic routes
    authStore.flatMenuListGet.forEach(item => {
      item.children && delete item.children;
      if (item.component && typeof item.component == "string") {
        item.component = modules["/src/views" + item.component + ".vue"];
      }
      if (item.meta.isFull) {
        router.addRoute(item as unknown as RouteRecordRaw);
      } else {
        router.addRoute("layout", item as unknown as RouteRecordRaw);
      }
    });
  } catch (error) {
    // If there's an error fetching the buttons or menu, redirect to the login page
    userStore.setToken("");
    router.replace(LOGIN_URL);
    return Promise.reject(error);
  }
};
