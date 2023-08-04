<template>
  <div id="target" class="card">
    <el-tabs type="border-card" v-model="activeName" editable @edit="handleTabsEdit">
      <el-tab-pane v-for="(tab, index) in tabList" :key="index" :name="tab.name">
        <template #label>
          <el-input
            v-if="tab.isEditing"
            v-model="tab.title"
            :ref="inputRef => (inputRefs[tab.name] = inputRef)"
            @blur="finishEditing(tab)"
            @keyup.enter="finishEditing(tab)"
            @keydown.delete.stop
          />
          <span v-else @dblclick="startEditing(tab)">{{ tab.title }}</span>
        </template>
        <ClientConfigForm />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts" name="client">
import ClientConfigForm from "@/components/ClientConfigForm/index.vue";
import { ElMessage, TabPaneName } from "element-plus";
// import { active } from "sortablejs";
import { nextTick, reactive, ref } from "vue";
const activeName = ref("client1");
const tabList = ref([
  {
    title: "客户端配置",
    name: "client1",
    isEditing: false
  }
]);
const defaultNewTitle = "新建标签";
const inputRefs: Record<string, any> = reactive({});
const startEditing = async (tab: { title: string; name: string; isEditing: boolean }) => {
  tab.isEditing = true;
  await nextTick();
  inputRefs[tab.name].select();
};
const finishEditing = (tab: { title: string; isEditing: boolean }) => {
  tab.isEditing = false;
  if (tab.title.trim() === "") {
    tab.title = defaultNewTitle;
  }
};

// note that no connected with the content??
const handleTabsEdit = async (targetName: TabPaneName | undefined, action: "remove" | "add") => {
  if (action === "add") {
    let newTabName = `client${tabList.value.length + 1}`;
    tabList.value.push({
      title: defaultNewTitle,
      name: newTabName,
      isEditing: true
    });
    activeName.value = newTabName;
    await nextTick();
    inputRefs[newTabName].select();
  } else if (action === "remove") {
    console.log("remove");
    if (tabList.value.length === 1) {
      ElMessage.warning("至少保留一个标签页");
      return;
    }
    let index = tabList.value.findIndex(tab => tab.name === targetName);
    tabList.value.splice(index, 1);
    //rename the tab
    for (let i = 0; i < tabList.value.length; i++) {
      tabList.value[i].name = `client${i + 1}`;
    }
    const nextTab = tabList.value[index] || tabList.value[index - 1];
    if (nextTab) {
      activeName.value = nextTab.name;
    }
  }
};
</script>
