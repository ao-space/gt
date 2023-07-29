<template>
  <h1>Anchor</h1>
  <el-tabs tab-position="right" @tab-click="handleClick" style="height: 200px" class="demo-tabs">
    <el-tab-pane v-for="(tab, index) in tabList" :key="index" :label="tab.title" :name="tab.name"> </el-tab-pane>
  </el-tabs>
  <div class="scroll-content">
    <!-- <div v-for="(tab, index) in tabList" :key="index" :ref="tab.name" class="scroll-item"> -->
    <div v-for="(tab, index) in tabList" :key="index" :ref="el => setRef(el, tab)" class="scroll-item">
      <div>
        <h2>{{ tab.title }}</h2>
      </div>
      <div>
        <p v-for="i in 20" :key="i">{{ tab.title }} - {{ i }}</p>
        <!-- 这里是每个tab对应的内容 -->
      </div>
    </div>
  </div>
</template>

<script setup lang="ts" name="Anchor">
import { ref } from "vue";
const refs = ref({});
const setRef = (el, tab) => {
  refs.value[tab.name] = el;
};
const tabList = [
  {
    title: "First",
    name: "first"
  },
  {
    title: "Second",
    name: "second"
  },
  {
    title: "Third",
    name: "third"
  }
];
const handleClick = tab => {
  // if (refs.value[tab.name]) {
  //   refs.value[tab.name].scrollIntoView({ behavior: "smooth" });
  // }
  if (refs.value[tab.props.name]) {
    refs.value[tab.props.name].scrollIntoView({ behavior: "smooth" });
  }
};
// const handleClick = tab => {
//   refs.value[tab.name].scrollIntoView({ behavior: "smooth" });
//   // this.$refs[tab.refName][0].scrollIntoView({ behavior: "smooth" });
// };
</script>

<style scoped lang="scss">
.demo-tabs > .el-tabs__content {
  padding: 32px;
  font-size: 32px;
  font-weight: 600;
  color: #1890ff;
}
.el-tabs--right .el-tabs__content,
.el-tabs--left .el-tabs__content {
  height: 100%;
}
.el-tabs__active-bar {
  background-color: #1890ff;
}
</style>
