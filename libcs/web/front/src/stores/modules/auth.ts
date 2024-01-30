import { defineStore } from "pinia";
import { AuthState } from "@/stores/interface";
import { getAuthMenuListApi } from "@/api/modules/login";
import { getFlatMenuList, getShowMenuList, getAllBreadcrumbList } from "@/utils";

export const useAuthStore = defineStore({
  id: "GT-auth",
  state: (): AuthState => ({
    // Menu permission list
    authMenuList: []
  }),
  getters: {
    // Menu permission list ==> The menu here has not been processed
    authMenuListGet: state => state.authMenuList,
    // Menu permission list ==> Left menu bar rendering, you need to remove isHide == true
    showMenuListGet: state => getShowMenuList(state.authMenuList),
    // Menu permission list ==> One-dimensional array menu after flattening, mainly used to add dynamic routes
    flatMenuListGet: state => getFlatMenuList(state.authMenuList),
    // Menu permission list ==> All breadcrumb navigation lists after recursive processing
    breadcrumbListGet: state => getAllBreadcrumbList(state.authMenuList)
  },
  actions: {
    // Get AuthMenuList
    async getAuthMenuList() {
      const { data } = await getAuthMenuListApi();
      this.authMenuList = data;
    }
  }
});
