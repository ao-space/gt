<template>
  <el-dropdown trigger="click" @command="setAssemblySize">
    <i :class="'iconfont icon-contentright'" class="toolBar-icon"></i>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="item in assemblySizeList"
          :key="item.value"
          :command="item.value"
          :disabled="assemblySize === item.value"
        >
          {{ item.label }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useGlobalStore } from "@/stores/modules/global";
import { AssemblySizeType } from "@/stores/interface";
import i18n from "@/languages";

const globalStore = useGlobalStore();
const assemblySize = computed(() => globalStore.assemblySize);

const assemblySizeList = [
  { label: "default", value: i18n.global.t("layout_header.default") },
  { label: "large", value: i18n.global.t("layout_header.large") },
  { label: "small", value: i18n.global.t("layout_header.small") }
];

const setAssemblySize = (item: AssemblySizeType) => {
  if (assemblySize.value === item) return;
  globalStore.setGlobalState("assemblySize", item);
};
</script>
