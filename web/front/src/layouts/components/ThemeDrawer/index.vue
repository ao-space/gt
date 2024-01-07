<template>
  <el-drawer v-model="drawerVisible" :title="$t('layout_theme.LayoutSetting')" size="290px">
    <!-- Layout Style -->
    <el-divider class="divider" content-position="center">
      <el-icon><Notification /></el-icon>
      {{ $t("layout_theme.Layout") }}
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
        {{ $t("layout_theme.InvertedAsideColor") }}
        <el-tooltip effect="dark" :content="$t('layout_theme.SwitchAside')" placement="top">
          <el-icon><QuestionFilled /></el-icon>
        </el-tooltip>
      </span>
      <el-switch v-model="asideInverted" @change="setAsideTheme" />
    </div>
    <!-- Global Theme -->
    <el-divider class="divider" content-position="center">
      <el-icon><ColdDrink /></el-icon>
      {{ $t("layout_theme.Theme") }}
    </el-divider>
    <div class="theme-item">
      <span>{{ $t("layout_theme.ThemeColor") }}</span>
      <el-color-picker v-model="primary" :predefine="colorList" @change="changePrimary" />
    </div>
    <div class="theme-item">
      <span>{{ $t("layout_theme.DarkMode") }}</span>
      <SwitchDark />
    </div>
    <div class="theme-item">
      <span>{{ $t("layout_theme.GreyMode") }}</span>
      <el-switch v-model="isGrey" @change="changeGreyOrWeak('grey', !!$event)" />
    </div>
    <div class="theme-item mb40">
      <span>{{ $t("layout_theme.ColorAccessibilityMode") }}</span>
      <el-switch v-model="isWeak" @change="changeGreyOrWeak('weak', !!$event)" />
    </div>

    <!-- UI Settings -->
    <el-divider class="divider" content-position="center">
      <el-icon><Setting /></el-icon>
      {{ $t("layout_theme.UISettings") }}
    </el-divider>
    <div class="theme-item">
      <span>{{ $t("layout_theme.CollapseMenu") }}</span>
      <el-switch v-model="isCollapse" />
    </div>
    <div class="theme-item">
      <span>{{ $t("layout_theme.Breadcrumb") }}</span>
      <el-switch v-model="breadcrumb" />
    </div>
    <div class="theme-item">
      <span>{{ $t("layout_theme.BreadcrumbIcon") }}</span>
      <el-switch v-model="breadcrumbIcon" />
    </div>
    <div class="theme-item">
      <span>{{ $t("layout_theme.Tab") }}</span>
      <el-switch v-model="tabs" />
    </div>
    <div class="theme-item">
      <span>{{ $t("layout_theme.TabIcon") }}</span>
      <el-switch v-model="tabsIcon" />
    </div>
    <div class="theme-item">
      <span>{{ $t("layout_theme.Footer") }}</span>
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

const globalStore = useGlobalStore();
const { layout, primary, isGrey, isWeak, asideInverted, isCollapse, breadcrumb, breadcrumbIcon, tabs, tabsIcon, footer } =
  storeToRefs(globalStore);

// preset color list
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

// Set the layout mode
const setLayout = (val: LayoutType) => {
  globalStore.setGlobalState("layout", val);
  setAsideTheme();
};

// Open theme settings
const drawerVisible = ref(false);
mittBus.on("openThemeDrawer", () => (drawerVisible.value = true));
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
