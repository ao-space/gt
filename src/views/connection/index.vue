<template>
  <div class="card">
    <!-- Client Pool -->
    <el-row v-if="poolForClient">
      <el-card>
        <v-chart ref="pie" class="echarts" :option="chartOptions" />
      </el-card>
    </el-row>

    <!-- Server Pool -->
    <el-row v-if="poolForServer.length != 0">
      <el-card>
        <template #header>
          <div class="card_header">Server Pool Info</div>
        </template>
        <ConnectionTable :table-data="poolForServer" :show-i-d="true" />
      </el-card>
    </el-row>

    <!-- External Connection -->
    <el-row>
      <el-card>
        <template #header>
          <div class="card_header">External Connection</div>
        </template>
        <ConnectionTable :table-data="connection" :show-i-d="false" />
      </el-card>
    </el-row>
  </div>
</template>
<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref, shallowRef } from "vue";
import { Connection } from "@/api/interface";
import ConnectionTable from "@/components/ConnectionTable/index.vue";
import { use } from "echarts/core";
import { PieChart } from "echarts/charts";
import { PolarComponent, TitleComponent, TooltipComponent, LegendComponent } from "echarts/components";
import type { ComposeOption } from "echarts/core";
import type { PieSeriesOption } from "echarts/charts";
import type { TitleComponentOption, TooltipComponentOption, LegendComponentOption } from "echarts/components";
import type { ECharts } from "echarts";
import { CanvasRenderer } from "echarts/renderers";
import { getConnectionApi } from "@/api/modules/connection";

use([CanvasRenderer, PolarComponent, TitleComponent, TooltipComponent, LegendComponent, PieChart]);

type EChartsOption = ComposeOption<TitleComponentOption | TooltipComponentOption | LegendComponentOption | PieSeriesOption>;
const chartOptions = reactive<EChartsOption>({
  textStyle: {
    fontFamily: 'Inter, "Helvetica Neue", Arial, sans-serif',
    fontWeight: 300
  },
  title: {
    text: "Connection Pool Status",
    left: "center"
  },
  tooltip: {
    trigger: "item",
    formatter: "{a} <br/>{b} : {c} ({d}%)"
  },
  legend: {
    orient: "vertical",
    left: "left",
    data: ["Running", "Idle", "Wait"]
  },
  series: [
    {
      name: "Status",
      type: "pie",
      radius: "55%",
      center: ["50%", "60%"],
      data: [],
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: "rgba(0, 0, 0, 0.5)"
        }
      }
    }
  ]
});
const pie = shallowRef<ECharts | null>(null);

const connection = reactive<Connection.Connection[]>([]);
const poolForClient = ref<Connection.Pool>();
const poolForServer = reactive<Connection.Connection[]>([]);

function transformPoolToPieChartData(pool: Connection.Pool) {
  const statusCount: Record<string, number> = {};
  for (const key in pool) {
    const status = Connection.StatusMap[pool[key]];
    if (!statusCount[status]) {
      statusCount[status] = 0;
    }
    statusCount[status]++;
  }
  return Object.keys(statusCount).map(status => ({ name: status, value: statusCount[status] }));
}

const isFirstDataLoaded = ref(true);
const reload = async () => {
  const { data } = await getConnectionApi();
  updateConnectionData(data.external);
  updateClientPoolData(data.clientPool);
  updateServerPoolData(data.serverPool);
};

const updateConnectionData = (externalData: Connection.Connection[]) => {
  connection.splice(0, connection.length, ...externalData);
};
const updateClientPoolData = (clientPoolData: Connection.Pool | undefined) => {
  if (clientPoolData) {
    poolForClient.value = clientPoolData;
    const pieChartData = transformPoolToPieChartData(poolForClient.value);
    (chartOptions.series as PieSeriesOption[])[0].data = pieChartData;
    (chartOptions.legend as LegendComponentOption).data = pieChartData.map(item => item.name);
    if (isFirstDataLoaded.value) {
      isFirstDataLoaded.value = false;
      startChartSwitchTimer();
    }
  } else if (poolForClient.value) {
    poolForClient.value = {};
  }
};
const updateServerPoolData = (serverPoolData: Connection.Connection[] | undefined) => {
  if (serverPoolData) {
    poolForServer.splice(0, poolForServer.length, ...serverPoolData);
  } else {
    poolForServer.splice(0, poolForServer.length);
  }
};

const timers = new Set<NodeJS.Timeout>();
function dispatchAction(type: string, dataIndex: number) {
  pie.value?.dispatchAction({
    type,
    seriesIndex: 0,
    dataIndex
  });
}
function startDataFetchTimer() {
  const dataFetchTimer = setInterval(async () => {
    await reload();
  }, 5000);
  timers.add(dataFetchTimer);
}
function startChartSwitchTimer() {
  let dataIndex = -1;
  let dataLen = 0;
  const series = chartOptions?.series;
  if (Array.isArray(series)) {
    for (const item of series) {
      if (item.type === "pie") {
        dataLen = item.data?.length || 0;
        break;
      }
    }
  }
  const chartSwitchTimer = setInterval(() => {
    if (!pie.value || dataLen === 0) {
      return;
    }
    dispatchAction("downplay", dataIndex);
    dataIndex = (dataIndex + 1) % dataLen;
    dispatchAction("highlight", dataIndex);
    dispatchAction("showTip", dataIndex);
  }, 1000);
  timers.add(chartSwitchTimer);
}
function clearAllTimers() {
  for (const timer of timers) {
    clearInterval(timer);
  }
  timers.clear();
}

onMounted(() => {
  reload();
  startDataFetchTimer();
});
onUnmounted(() => {
  clearAllTimers();
});
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
