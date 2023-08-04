<template>
  <el-tabs v-model="activeTab" tab-position="right">
    <el-tab-pane v-for="(tab, index) in tabs" :key="index" :label="tab.title" @click="tabClicked(index)"> </el-tab-pane>
  </el-tabs>
</template>

<script setup lang="ts">
const props = defineProps({
  tabs: {
    type: Array as () => Array<{ title: string }>,
    default: () => []
  }
});

const emit = defineEmits(["change", "input"]);

const activeTab = ref(0);

const tabClicked = (index: number) => {
  activeTab.value = index;
  emit("change", index);
  emit("input", index);
};
</script>
