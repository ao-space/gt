<template>
  <div class="home card" v-if="state">
    <el-row :gutter="15" class="system_state">
      <el-col :span="12">
        <el-card v-if="state.os" class="card_item">
          <template #header>
            <div class="card_header">{{ $t("view_home.Runtime") }}</div>
          </template>
          <div>
            <el-row :gutter="10">
              <el-col :span="12">{{ $t("view_home.os") }}:</el-col>
              <el-col :span="12">{{ state.os.goos }}</el-col>
            </el-row>
            <el-row :gutter="10">
              <el-col :span="12">{{ $t("view_home.cpu_nums") }}:</el-col>
              <el-col :span="12">{{ state.os.numCpu }}</el-col>
            </el-row>
            <el-row :gutter="10">
              <el-col :span="12">{{ $t("view_home.compiler") }}:</el-col>
              <el-col :span="12">{{ state.os.compiler }}</el-col>
            </el-row>
            <el-row :gutter="10">
              <el-col :span="12">{{ $t("view_home.go_version") }}:</el-col>
              <el-col :span="12">{{ state.os.goVersion }}</el-col>
            </el-row>
            <el-row :gutter="10">
              <el-col :span="12">{{ $t("view_home.goroutine_nums") }}:</el-col>
              <el-col :span="12">{{ state.os.numGoroutine }}</el-col>
            </el-row>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card v-if="state.disk" class="card_item">
          <template #header>
            <div class="card_header">{{ $t("view_home.Disk") }}</div>
          </template>
          <div>
            <el-row :gutter="10">
              <el-col :span="12">
                <el-row :gutter="10">
                  <el-col :span="12">{{ $t("view_home.Total") }} (MB)</el-col>
                  <el-col :span="12">{{ state.disk.totalMb }}</el-col>
                </el-row>
                <el-row :gutter="10">
                  <el-col :span="12">{{ $t("view_home.Used") }} (MB)</el-col>
                  <el-col :span="12">{{ state.disk.usedMb }}</el-col>
                </el-row>
                <el-row :gutter="10">
                  <el-col :span="12">{{ $t("view_home.Total") }} (GB)</el-col>
                  <el-col :span="12">{{ state.disk.totalGb }}</el-col>
                </el-row>
                <el-row :gutter="10">
                  <el-col :span="12">{{ $t("view_home.Used") }} (GB)</el-col>
                  <el-col :span="12">{{ state.disk.usedGb }}</el-col>
                </el-row>
              </el-col>
              <el-col :span="12">
                <el-progress type="dashboard" :percentage="state.disk.usedPercent" :color="colors">
                  <template #default="{ percentage }">
                    <span class="percentage-value">{{ percentage }}%</span>
                    <span class="percentage-label">{{ $t("view_home.Used") }}</span>
                  </template>
                </el-progress>
              </el-col>
            </el-row>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="15" class="system_state">
      <el-col :span="12">
        <el-card v-if="state.cpu" class="card_item" :body-style="{ height: '180px', 'overflow-y': 'scroll' }">
          <template #header>
            <div class="card_header">{{ $t("view_home.CPU") }}</div>
          </template>
          <div>
            <el-row :gutter="10">
              <el-col :span="12">{{ $t("view_home.Core_Number") }}:</el-col>
              <el-col :span="12">{{ state.cpu.cores }}</el-col>
            </el-row>
            <el-row v-for="(item, index) in state.cpu.cpus" :key="index" :gutter="10">
              <el-col :span="12">{{ $t("view_home.core") }} {{ index }}:</el-col>
              <el-col :span="12"><el-progress type="line" :percentage="+item.toFixed(0)" :color="colors" /></el-col>
            </el-row>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card v-if="state.ram" class="card_item">
          <template #header>
            <div class="card_header">{{ $t("view_home.Ram") }}</div>
          </template>
          <div>
            <el-row :gutter="10">
              <el-col :span="12">
                <el-row :gutter="10">
                  <el-col :span="12">{{ $t("view_home.Total") }} (MB)</el-col>
                  <el-col :span="12">{{ state.ram.totalMb }}</el-col>
                </el-row>
                <el-row :gutter="10">
                  <el-col :span="12">{{ $t("view_home.Used") }} (MB)</el-col>
                  <el-col :span="12">{{ state.ram.usedMb }}</el-col>
                </el-row>
                <el-row :gutter="10">
                  <el-col :span="12">{{ $t("view_home.Total") }} (GB)</el-col>
                  <el-col :span="12">{{ (state.ram.totalMb / 1024).toFixed(2) }}</el-col>
                </el-row>
                <el-row :gutter="10">
                  <el-col :span="12">{{ $t("view_home.Used") }} (GB)</el-col>
                  <el-col :span="12">{{ (state.ram.usedMb / 1024).toFixed(2) }}</el-col>
                </el-row>
              </el-col>
              <el-col :span="12">
                <el-progress type="dashboard" :percentage="state.ram.usedPercent" :color="colors">
                  <template #default="{ percentage }">
                    <span class="percentage-value">{{ percentage }}%</span>
                    <span class="percentage-label">{{ $t("view_home.Used") }}</span>
                  </template>
                </el-progress>
              </el-col>
            </el-row>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
  <div class="home card" v-else>
    <img class="home-bg" src="@/assets/images/welcome.png" alt="welcome" />
  </div>
</template>

<script setup lang="ts" name="home">
import { ref, onUnmounted } from "vue";
import { getServerInfoApi } from "@/api/modules/server";
import { Server } from "@/api/interface";

const state = ref<Server.SystemState | null>();
const colors = ref([
  { color: "#5cb87a", percentage: 40 },
  { color: "#e6a23c", percentage: 70 },
  { color: "#f56c6c", percentage: 100 }
]);

const reload = async () => {
  const { data } = await getServerInfoApi();
  state.value = data.serverInfo;
};

reload();

const timer = setInterval(() => {
  reload();
}, 1000 * 3);

onUnmounted(() => {
  clearInterval(timer);
});
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
