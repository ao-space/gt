<template>
  <el-affix :offset="120">
    <el-tabs tab-position="right" @tab-click="handleClick">
      <!-- <el-tab-pane v-for="(tab, index) in tabList" :key="index" :label="tab.title" :name="tab.name"> </el-tab-pane> -->
      <el-tab-pane v-for="tab in tabList" :key="tab.uuid" :label="tab.title" :name="tab.uuid"> </el-tab-pane>
    </el-tabs>
  </el-affix>
  <div v-for="tab in tabList" :key="tab.uuid" :ref="el => (tabRefs[tab.uuid] = el as HTMLDivElement)">
    <!-- <div v-for="(tab, index) in tabList" :key="index" :ref="el => (tabRefs[tab.name] = el as HTMLDivElement)"> -->
    <slot :name="tab.uuid"></slot>
    <!-- <slot :name="tab.name"></slot> -->
    <!-- <slot :name="tab.title"></slot> -->
  </div>
</template>

<script setup lang="ts" name="Anchor">
import { PropType, Ref, ref } from "vue";
import type { TabsPaneContext } from "element-plus";
export interface Tab {
  title: string;
  name: string;
  uuid: string;
}
defineProps({
  tabList: {
    type: Array as PropType<Tab[]>,
    required: true,
    default: () => []
  }
});

const tabRefs: Ref<{ [key: string]: HTMLDivElement | null }> = ref({});
const handleClick = (pane: TabsPaneContext) => {
  console.log(pane);
  if (pane.props.name !== undefined && tabRefs.value[pane.props.name.toString()]) {
    tabRefs.value[pane.props.name.toString()]!.scrollIntoView({ behavior: "smooth" });
  }
};
</script>

<style scoped lang="scss">
$primary-color: #1890ff;
.el-tabs--right {
  background-color: red;
  :deep(.el-tabs__header.is-right) {
    // todo location
    position: fixed;
    top: 50%;
    right: 30px;
    z-index: 1;
    width: auto;
    height: auto;
    background-color: rgb(170 170 170 / 20%);
    transform: translateY(-50%);
  }
}
:deep(.el-tabs__item:hover) {
  font: bold;
  font-size: large;
  color: $primary-color;
}
:deep(.el-tabs__active-bar.is-right) {
  width: 3px;
  background-color: $primary-color;
}
:deep(.el-tabs__item.is-right.is-active) {
  color: $primary-color;
}
</style>
