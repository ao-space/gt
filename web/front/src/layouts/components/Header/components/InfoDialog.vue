<template>
  <el-dialog
    v-model="dialogVisible"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    title="User Setting"
    width="500px"
    draggable
  >
    <InfoForm v-if="dialogVisible" />
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import InfoForm from "./InfoForm.vue";

const dialogVisible = ref(false);
const openDialog = () => {
  dialogVisible.value = true;
};

defineExpose({ openDialog });
</script>
