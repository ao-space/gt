<template>
  <el-icon :size="20" class="toolBar-icon" style="cursor: pointer" @click="guide">
    <Promotion />
  </el-icon>
</template>

<script setup lang="ts" name="guide">
import { onMounted } from "vue";
import { driver } from "driver.js";
import "driver.js/dist/driver.css";
import i18n from "@/languages";

const guide = () => {
  const driverObj = driver({
    ...config,
    steps
  });
  driverObj.drive();
};

const config = {
  allowClose: false,
  doneBtnText: i18n.global.t("layout_header.DoneBtnText"),
  closeBtnText: i18n.global.t("layout_header.CloseBtnText"),
  nextBtnText: i18n.global.t("layout_header.NextBtnText"),
  prevBtnText: i18n.global.t("layout_header.PrevBtnText")
};
const steps = [
  {
    element: "#collapseIcon",
    popover: {
      title: i18n.global.t("layout_header.CollapseIconTitle"),
      description: i18n.global.t("layout_header.CollapseIconDescription"),
      position: "right"
    }
  },
  {
    element: "#breadcrumb",
    popover: {
      title: i18n.global.t("layout_header.BreadcrumbTitle"),
      description: i18n.global.t("layout_header.BreadcrumbDescription"),
      position: "right"
    }
  },
  {
    element: "#guide",
    popover: {
      title: i18n.global.t("layout_header.GuideTitle"),
      description: i18n.global.t("layout_header.GuideDescription"),
      position: "left"
    }
  },
  {
    element: "#assemblySize",
    popover: {
      title: i18n.global.t("layout_header.AssemblySizeTitle"),
      description: i18n.global.t("layout_header.AssemblySizeDescription"),
      position: "left"
    }
  },
  {
    element: "#themeSetting",
    popover: {
      title: i18n.global.t("layout_header.ThemeSettingTitle"),
      description: i18n.global.t("layout_header.ThemeSettingDescription"),
      position: "left"
    }
  },
  {
    element: "#fullscreen",
    popover: {
      title: i18n.global.t("layout_header.FullScreenTitle"),
      description: i18n.global.t("layout_header.FullScreenDescription"),
      position: "left"
    }
  },
  {
    element: "#avatar",
    popover: {
      title: i18n.global.t("layout_header.UserTitle"),
      description: i18n.global.t("layout_header.UserDescription"),
      position: "left"
    }
  }
];

onMounted(() => {
  if (!localStorage.getItem("guide")) {
    guide();
    // Mark that the guide has been shown
    localStorage.setItem("guide", "true");
  }
});
</script>
