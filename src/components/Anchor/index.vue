<template>
  <el-tabs tab-position="right" @tab-click="handleClick">
    <el-tab-pane v-for="(tab, index) in tabList" :key="index" :label="tab.title" :name="tab.name"> </el-tab-pane>
  </el-tabs>
  <div v-for="(tab, index) in tabList" :key="index" :ref="el => (tabRefs[tab.name] = el as HTMLDivElement)">
    <slot :name="tab.name"></slot>
  </div>

  <!-- <div class="scroll-content">
    <div v-for="(tab, index) in tabList" :key="index" :ref="el => setRef(el, tab)" class="scroll-item">
      <div>
        <h2>{{ tab.title }}</h2>
      </div>
      <div>
        <p v-for="i in 20" :key="i">{{ tab.title }} - {{ i }}</p>
      </div>
    </div>
  </div> -->
</template>

<script setup lang="ts" name="Anchor">
import { PropType, Ref, ref } from "vue";
import type { TabsPaneContext } from "element-plus";
interface Tab {
  title: string;
  name: string;
}
// const tabRefs = ref<Record<string, HTMLElement | null>>({});
// const tabRefs = ref({});
defineProps({
  tabList: {
    type: Array as PropType<Tab[]>,
    default: () => []
  }
});

const tabRefs: Ref<{ [key: string]: HTMLDivElement | null }> = ref({});
// const setRef = (el: Element | ComponentPublicInstance | null, tab: Tab) => {
//   if (el) {
//     refs.value[tab.name] = el as HTMLDivElement;
//   }
// };
const handleClick = (pane: TabsPaneContext) => {
  if (pane.props.name !== undefined && tabRefs.value[pane.props.name.toString()]) {
    tabRefs.value[pane.props.name.toString()]!.scrollIntoView({ behavior: "smooth" });
  }
};
// const handleClick = (tab: { props: { name: string } }) => {
//   // if (refs.value[tab.name]) {
//   //   refs.value[tab.name].scrollIntoView({ behavior: "smooth" });
//   // }
//   if (refs.value[tab.props.name]) {
//     refs.value[tab.props.name]!.scrollIntoView({ behavior: "smooth" });
//   }
// };
// const handleClick = tab => {
//   refs.value[tab.name].scrollIntoView({ behavior: "smooth" });
//   // this.$refs[tab.refName][0].scrollIntoView({ behavior: "smooth" });
// };
</script>

<style scoped lang="scss">
// .el-tabs--right.el-tabs__header.is-right {
//   background-color: yellow !important;
// }
.el-tabs.el-tabs--right {
  position: relative;

  // width: 0;

  // height: 0;
  // background-color: red;
  .el-tabs__header.is-right {
    // position: absolute;

    // width: 200px;
    background-color: yellow !important;
    .el-tabs__nav-wrap.is-right {
      .el-tabs__nav.is-right {
        .el-tabs__active-bar.is-right {
          // width: 0;
          // height: 0;
          background-color: green !important;
        }
        .el-tabs__item.is-right {
          // width: 0;

          // height: 0;
          background-color: green !important;
        }
      }

      background-color: green !important;
    }
  }
}

// .demo-tabs > .el-tabs__content {
//   padding: 32px;
//   font-size: 32px;
//   font-weight: 600;
//   color: #1890ff;
// }
// .el-tabs--right .el-tabs__content,
// .el-tabs--left .el-tabs__content {
//   height: 100%;
// }
// .el-tabs__active-bar {
//   background-color: #1890ff;
// }
</style>
