<template>
  <el-drawer v-model="drawerVisible" title="Layout Setting" size="290px">
    <!-- Layout Style -->
    <el-divider class="divider" content-position="center">
      <el-icon><Notification /></el-icon>
      Layout
    </el-divider>
    <div class="layout-box">
      <el-tooltip effect="dark" content="classic" placement="top" :show-after="200">
        <div :class="['layout-item layout-classic', { 'is-active': layout == 'classic' }]" @click="setLayout('classic')">
          <div class="layout-dark"></div>
          <div class="layout-container">
            <div class="layout-light"></div>
            <div class="layout-content"></div>
          </div>
          <el-icon v-if="layout == 'classic'">
            <CircleCheckFilled />
          </el-icon>
        </div>
      </el-tooltip>
    </div>
    <div class="theme-item">
      <span>
        Inverted Aside Color
        <el-tooltip effect="dark" content="Switch Aside color to Dark mode" placement="top">
          <el-icon><QuestionFilled /></el-icon>
        </el-tooltip>
      </span>
      <el-switch v-model="asideInverted" @change="setAsideTheme" />
    </div>
    <!-- Global Theme -->
    <el-divider class="divider" content-position="center">
      <el-icon><ColdDrink /></el-icon>
      Theme
    </el-divider>
    <div class="theme-item">
      <span>Theme Color</span>
      <el-color-picker v-model="primary" :predefine="colorList" @change="changePrimary" />
    </div>
    <div class="theme-item">
      <span>Dark Mode</span>
      <SwitchDark />
    </div>
    <div class="theme-item">
      <span>Grey Mode</span>
      <el-switch v-model="isGrey" @change="changeGreyOrWeak('grey', !!$event)" />
    </div>
    <div class="theme-item mb40">
      <span>Color Accessibility Mode</span>
      <el-switch v-model="isWeak" @change="changeGreyOrWeak('weak', !!$event)" />
    </div>

    <!-- UI Settings -->
    <el-divider class="divider" content-position="center">
      <el-icon><Setting /></el-icon>
      UI Settings
    </el-divider>
    <div class="theme-item">
      <span>Collapse Menu</span>
      <el-switch v-model="isCollapse" />
    </div>
    <div class="theme-item">
      <span>Breadcrumb</span>
      <el-switch v-model="breadcrumb" />
    </div>
    <div class="theme-item">
      <span>Breadcrumb Icon</span>
      <el-switch v-model="breadcrumbIcon" />
    </div>
    <div class="theme-item">
      <span>Tab</span>
      <el-switch v-model="tabs" />
    </div>
    <div class="theme-item">
      <span>Tab Icon</span>
      <el-switch v-model="tabsIcon" />
    </div>
    <div class="theme-item">
      <span>Footer</span>
      <el-switch v-model="footer" />
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { storeToRefs } from "pinia";
import { useTheme } from "@/hooks/useTheme";
import { useGlobalStore } from "@/stores/modules/global";
import { LayoutType } from "@/stores/interface";
import { DEFAULT_PRIMARY } from "@/config";
import mittBus from "@/utils/mittBus";
import SwitchDark from "@/components/SwitchDark/index.vue";

const { changePrimary, changeGreyOrWeak, setAsideTheme } = useTheme();
// const { changePrimary, changeGreyOrWeak, setAsideTheme, setHeaderTheme } = useTheme();

const globalStore = useGlobalStore();
const {
  layout,
  primary,
  isGrey,
  isWeak,
  asideInverted,
  // headerInverted,
  isCollapse,
  breadcrumb,
  breadcrumbIcon,
  tabs,
  tabsIcon,
  footer
} = storeToRefs(globalStore);

// 预定义主题颜色
const colorList = [
  DEFAULT_PRIMARY,
  "#daa96e",
  "#0c819f",
  "#409eff",
  "#27ae60",
  "#ff5c93",
  "#e74c3c",
  "#fd726d",
  "#f39c12",
  "#9b59b6"
];

// 设置布局方式
const setLayout = (val: LayoutType) => {
  globalStore.setGlobalState("layout", val);
  setAsideTheme();
};

// 打开主题设置
const drawerVisible = ref(false);
mittBus.on("openThemeDrawer", () => (drawerVisible.value = true));
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
