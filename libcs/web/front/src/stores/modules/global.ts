import { defineStore } from "pinia";
import { GlobalState } from "@/stores/interface";
import { DEFAULT_PRIMARY } from "@/config";
import piniaPersistConfig from "@/config/piniaPersist";

export const useGlobalStore = defineStore({
  id: "GT-global",
  // After modifying the default values, need to clear the localStorage data
  state: (): GlobalState => ({
    // Layout mode
    layout: "classic",
    // Size of the element components
    assemblySize: "default",
    // Whether the current page is in full screen
    maximize: false,
    // Theme color
    primary: DEFAULT_PRIMARY,
    // Dark mode
    isDark: false,
    // Gray mode
    isGrey: false,
    // Color weak mode
    isWeak: false,
    // Invert the sidebar
    asideInverted: false,
    // Collapse the menu
    isCollapse: false,
    // Breadcrumb navigation
    breadcrumb: true,
    // Breadcrumb navigation icon
    breadcrumbIcon: true,
    // Tabs
    tabs: true,
    // Tabs icon
    tabsIcon: true,
    // Footer
    footer: true
  }),
  getters: {},
  actions: {
    // Set GlobalState
    setGlobalState(...args: ObjToKeyValArray<GlobalState>) {
      this.$patch({ [args[0]]: args[1] });
    }
  },
  persist: piniaPersistConfig("GT-global", localStorage)
});
