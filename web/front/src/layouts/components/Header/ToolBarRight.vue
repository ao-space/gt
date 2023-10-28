<template>
  <div class="tool-bar-ri">
    <div class="header-icon">
      <Guide id="guide" />
      <AssemblySize id="assemblySize" />
      <ThemeSetting id="themeSetting" />
      <Fullscreen id="fullscreen" />
    </div>
    <div id="avatar">
      <Avatar />
    </div>
  </div>
</template>

<script setup lang="ts">
import Guide from "./components/Guide.vue";
import AssemblySize from "./components/AssemblySize.vue";
import ThemeSetting from "./components/ThemeSetting.vue";
import Fullscreen from "./components/Fullscreen.vue";
import Avatar from "./components/Avatar.vue";
</script>

<style scoped lang="scss">
.tool-bar-ri {
  display: flex;
  align-items: center;
  justify-content: center;
  padding-right: 25px;
  .header-icon {
    display: flex;
    align-items: center;
    & > * {
      margin-left: 21px;
      color: var(--el-header-text-color);
    }

    margin-right: 20px;
  }
}
</style>
