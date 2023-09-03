<template>
  <div class="card">
    <el-row>
      <el-card>
        <v-chart ref="pie" class="echarts" :option="chartOptions" />
      </el-card>
    </el-row>
    <el-row>
      <el-card>
        <template #header>
          <div class="card_header">External Connection</div>
        </template>
        <el-table ref="tableRef" :data="connection" highlight-current-row stripe style="width: 100%">
          <el-table-column type="index"></el-table-column>
          <el-table-column prop="family" label="Family" :formatter="formatFamily"></el-table-column>
          <el-table-column prop="type" label="Type" min-width="180" :formatter="formatType"></el-table-column>
          <el-table-column
            prop="localaddr"
            label="Local Address"
            min-width="180"
            :formatter="formatAddress('localaddr')"
            :filters="localAddrFilterOptions"
            :filter-method="filterAddr"
          ></el-table-column>
          <el-table-column
            prop="remoteaddr"
            label="Remote Address"
            min-width="180"
            :formatter="formatAddress('remoteaddr')"
            :filters="remoteAddrFilterOptions"
            :filter-method="filterAddr"
          ></el-table-column>
          <el-table-column prop="status" label="Status"></el-table-column>
        </el-table>
      </el-card>
    </el-row>
  </div>
</template>
<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref, shallowRef, watch } from "vue";
import { Connection } from "@/api/interface";
import { type TableColumnCtx, type TableInstance } from "element-plus";
import { use } from "echarts/core";
import { PieChart } from "echarts/charts";
import { PolarComponent, TitleComponent, TooltipComponent, LegendComponent } from "echarts/components";
import type { ComposeOption } from "echarts/core";
import type { PieSeriesOption } from "echarts/charts";
import type { TitleComponentOption, TooltipComponentOption, LegendComponentOption } from "echarts/components";
import type { ECharts } from "echarts";
import { CanvasRenderer } from "echarts/renderers";

use([CanvasRenderer, PolarComponent, TitleComponent, TooltipComponent, LegendComponent, PieChart]);

type EChartsOption = ComposeOption<TitleComponentOption | TooltipComponentOption | LegendComponentOption | PieSeriesOption>;
const chartOptions = ref<EChartsOption>({
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

let timer: NodeJS.Timeout | null = null;
function startActions() {
  stopActions();

  let dataIndex = -1;
  // const dataLen = chartOptions.value?.series?.[0]?.data?.length || 0;
  let dataLen = 0;
  const series = chartOptions.value?.series;
  if (Array.isArray(series)) {
    for (const item of series) {
      if (item.type === "pie") {
        dataLen = item.data?.length || 0;
        break;
      }
    }
  }
  if (!pie.value || dataLen === 0) {
    return;
  }

  timer = setInterval(() => {
    if (!pie.value) {
      stopActions();
      return;
    }
    dispatchAction("downplay", dataIndex);
    dataIndex = (dataIndex + 1) % dataLen;
    dispatchAction("highlight", dataIndex);
    dispatchAction("showTip", dataIndex);
  }, 1000);
}
function dispatchAction(type: string, dataIndex: number) {
  pie.value?.dispatchAction({
    type,
    seriesIndex: 0,
    dataIndex
  });
}
function stopActions() {
  if (timer !== null) {
    clearInterval(timer);
    timer = null;
  }
}

const connection = reactive<Connection.Connection[]>([]);
const pool = ref<Connection.Pool>({});

const tableRef = ref<TableInstance>();
const remoteAddrFilterOptions = reactive<{ text: string; value: string }[]>([]);
watch(connection, newVal => {
  const uniqueRemoteAddrs = [...new Set(newVal.map(item => item.remoteaddr.ip))];
  remoteAddrFilterOptions.splice(0, remoteAddrFilterOptions.length, ...uniqueRemoteAddrs.map(ip => ({ text: ip, value: ip })));
});

const localAddrFilterOptions = reactive<{ text: string; value: string }[]>([]);
watch(connection, newVal => {
  const uniqueLocalAddrs = [...new Set(newVal.map(item => item.localaddr.ip))];
  localAddrFilterOptions.splice(0, localAddrFilterOptions.length, ...uniqueLocalAddrs.map(ip => ({ text: ip, value: ip })));
});

const filterAddr = (value: string, row: Connection.Connection, column: TableColumnCtx<Connection.Connection>) => {
  const property = column["property"];
  if (property === "remoteaddr" || property === "localaddr") {
    return row[property].ip === value;
  }
  return false;
};

const formatFamily = (row: Connection.Connection) => {
  switch (row.family) {
    case 1:
      return "Unix";
    case 2:
      return "IPv4";
    case 10:
      return "IPv6";
    default:
      return "Unknown";
  }
};
const formatAddress = (type: "localaddr" | "remoteaddr") => (row: Connection.Connection) => {
  const addr = row[type];
  return `${addr.ip}:${addr.port}`;
};
const formatType = (row: Connection.Connection) => {
  switch (row.type) {
    case 1:
      return "SOCK_STREAM";
    case 2:
      return "SOCK_DGRAM";
    case 3:
      return "SOCK_RAW";
    default:
      return "Unknown";
  }
};
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

onMounted(() => {
  connection.splice(
    0,
    connection.length,
    ...[
      {
        family: 2,
        type: 1,
        localaddr: {
          ip: "127.0.0.1",
          port: 8000
        },
        remoteaddr: {
          ip: "0.0.0.0",
          port: 0
        },
        status: "LISTEN"
      },
      {
        family: 2,
        type: 1,
        localaddr: {
          ip: "127.0.0.1",
          port: 8000
        },
        remoteaddr: {
          ip: "127.0.0.1",
          port: 35950
        },
        status: "ESTABLISHED"
      },
      {
        family: 2,
        type: 1,
        localaddr: {
          ip: "172.17.60.128",
          port: 57436
        },
        remoteaddr: {
          ip: "34.120.195.249",
          port: 443
        },
        status: "SYN_SENT"
      }
    ]
  );
  pool.value = {
    1: 2,
    10: 2,
    2: 2,
    3: 2,
    4: 2,
    5: 1,
    6: 2,
    7: 2,
    8: 1,
    9: 1
  };
  const pieChartData = transformPoolToPieChartData(pool.value);
  // chartOptions.value?.series?[0].data = pieChartData;
  (chartOptions.value?.series as PieSeriesOption[])[0].data = pieChartData;
  startActions();
});
onUnmounted(() => {
  stopActions();
});
</script>

<style scoped lang="scss">
@import "./index.scss";
</style>
