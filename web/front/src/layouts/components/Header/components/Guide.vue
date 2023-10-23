<template>
  <el-icon :size="20" class="toolBar-icon" style="cursor: pointer" @click="guide">
    <Promotion />
  </el-icon>
</template>

<script setup lang="ts" name="guide">
import { onMounted } from "vue";
import { driver } from "driver.js";
import "driver.js/dist/driver.css";

const guide = () => {
  const driverObj = driver({
    ...config,
    steps
  });
  driverObj.drive();
};

const config = {
  allowClose: false,
  doneBtnText: "Finish",
  closeBtnText: "Close",
  nextBtnText: "Next",
  prevBtnText: "Previous"
};
const steps = [
  {
    element: "#collapseIcon",
    popover: {
      title: "Collapse Icon",
      description: "Toggle the sidebar open or closed.",
      position: "right"
    }
  },
  {
    element: "#breadcrumb",
    popover: {
      title: "Breadcrumb",
      description: "Indicate the current page location",
      position: "right"
    }
  },
  {
    element: "#guide",
    popover: {
      title: "Guide",
      description: "Guide the user to use the system",
      position: "left"
    }
  },
  {
    element: "#assemblySize",
    popover: {
      title: "Switch Assembly Size",
      description: "Adjust the system's display size.",
      position: "left"
    }
  },
  {
    element: "#themeSetting",
    popover: {
      title: "Setting theme",
      description: "Customize the system's theme.",
      position: "left"
    }
  },
  {
    element: "#fullscreen",
    popover: {
      title: "Full Screen",
      description: "Enter or exit full-screen mode.",
      position: "left"
    }
  },
  {
    element: "#avatar",
    popover: {
      title: "User",
      description:
        "Click here to open the System Settings.<br/> Upon the first launch, the system automatically generates a random username and password for you. <strong>We strongly recommend updating these details within 30 minutes</strong> to ensure smooth future logins.",
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
