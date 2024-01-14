<template>
  <el-select v-model="lang" class="m-2" placeholder="Select" size="small">
    <el-option v-for="item in langs" :key="item.value" :label="item.label" :value="item.value" />
  </el-select>
</template>
<script setup lang="ts">
import { useMetadataStore } from "@/stores/modules/metadata";
import { ref, watch } from "vue";
let status = useMetadataStore();
let lang = ref(status.language);
watch(lang, status.changeLangStatus);
let langs = [
  {
    value: "zh",
    label: "中文"
  },
  {
    value: "en",
    label: "English"
  }
];
</script>

<style scoped lang="scss"></style>
