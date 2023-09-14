import { createRouter, createWebHashHistory } from "vue-router";
import { useUserStore } from "@/stores/modules/user";
import { useAuthStore } from "@/stores/modules/auth";
import { LOGIN_URL, ROUTER_WHITE_LIST } from "@/config";
import { initDynamicRouter } from "@/routers/modules/dynamicRouter";
import { staticRouter, errorRouter } from "@/routers/modules/staticRouter";
import NProgress from "@/config/nprogress";

/**
 * @description 📚 Brief introduction to router configuration parameters
 * @param path ==> Route menu access path
 * @param name ==> Route name (corresponds to the page component name, can be used as KeepAlive cache identifier && button permission filtering)
 * @param redirect ==> Route redirection address
 * @param component ==> View file path
 * @param meta ==> Route menu meta information
 * @param meta.icon ==> Icon corresponding to the menu and breadcrumb
 * @param meta.title ==> Route title (used as document.title or menu name)
 * @param meta.activeMenu ==> The menu to be highlighted when the current route is a detail page
 * @param meta.isLink ==> Access address filled in when the route is an external link
 * @param meta.isHide ==> Whether to hide in the menu (usually detail pages of lists need to be hidden)
 * @param meta.isFull ==> Whether the menu is full screen (example: data screen page)
 * @param meta.isAffix ==> Whether the menu is pinned in the tab page (the homepage is usually a pinned item)
 * @param meta.isKeepAlive ==> Whether the current route is cached
 * */
const router = createRouter({
  history: createWebHashHistory(),
  routes: [...staticRouter, ...errorRouter],
  strict: false,
  scrollBehavior: () => ({ left: 0, top: 0 })
});

/**
 * @description Route interception beforeEach
 * */
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore();
  const authStore = useAuthStore();

  // 1.Start NProgress
  NProgress.start();

  // 2.Dynamically set the title
  const title = import.meta.env.VITE_GLOB_APP_TITLE;
  document.title = to.meta.title ? `${to.meta.title} - ${title}` : title;

  // 3.Check if visiting the login page, if there's a Token stay on the current page, if not reset the route to the login page
  if (to.path.toLocaleLowerCase() === LOGIN_URL) {
    if (userStore.token) return next(from.fullPath);
    resetRouter();
    return next();
  }

  // 4.Check if the visited page is in the route whitelist (static routes), if it exists, let it pass directly
  if (ROUTER_WHITE_LIST.includes(to.path)) return next();

  // 5.Check if there's a Token, if not redirect to the login page
  if (!userStore.token) return next({ path: LOGIN_URL, replace: true });

  // 6.If there's no menu list, request the menu list again and add dynamic routes
  if (!authStore.authMenuListGet.length) {
    await initDynamicRouter();
    return next({ ...to, replace: true });
  }

  // 7.Normal page access
  next();
});

/**
 * @description Reset router
 * */
export const resetRouter = () => {
  const authStore = useAuthStore();
  authStore.flatMenuListGet.forEach(route => {
    const { name } = route;
    if (name && router.hasRoute(name)) router.removeRoute(name);
  });
};

/**
 * @description Route jump error
 * */
router.onError(error => {
  NProgress.done();
  console.warn("Route Error", error.message);
});

/**
 * @description Route interception afterEach
 * */
router.afterEach(() => {
  NProgress.done();
});

export default router;
