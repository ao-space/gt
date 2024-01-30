<template>
  <el-tooltip effect="dark" placement="top-start">
    <template #content>{{ props.usageText }}</template>
    <el-icon>
      <QuestionFilled />
    </el-icon>
  </el-tooltip>
</template>

<script setup lang="ts" name="UsageTooltip">
const props = defineProps({
  usageText: String
});
</script>
