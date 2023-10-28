<template>
  <el-table ref="tableRef" :data="props.tableData" highlight-current-row stripe style="width: 100%">
    <el-table-column type="index"></el-table-column>
    <el-table-column v-if="showID" prop="id" label="ID" :filters="idFilterOptions" :filter-method="filterID"></el-table-column>
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
</template>

<script setup lang="ts" name="ConnectionTable">
import { type TableColumnCtx, type TableInstance } from "element-plus";
import { Connection } from "@/api/interface";
import { reactive, ref, watch } from "vue";

const props = defineProps<{
  tableData: Connection.Connection[];
  showID: boolean;
}>();

//Form Related
const tableRef = ref<TableInstance>();

const remoteAddrFilterOptions = reactive<{ text: string; value: string }[]>([]);
//Watch tableData to update filter options
watch(
  props.tableData,
  newVal => {
    const uniqueRemoteAddrs = [...new Set(newVal.map(item => item.remoteaddr.ip))];
    remoteAddrFilterOptions.splice(0, remoteAddrFilterOptions.length, ...uniqueRemoteAddrs.map(ip => ({ text: ip, value: ip })));
  },
  {
    immediate: true
  }
);

const localAddrFilterOptions = reactive<{ text: string; value: string }[]>([]);
//Watch tableData to update filter options
watch(
  props.tableData,
  newVal => {
    const uniqueLocalAddrs = [...new Set(newVal.map(item => item.localaddr.ip))];
    localAddrFilterOptions.splice(0, localAddrFilterOptions.length, ...uniqueLocalAddrs.map(ip => ({ text: ip, value: ip })));
  },
  {
    immediate: true
  }
);

const idFilterOptions = reactive<{ text: string; value: string }[]>([]);
//Watch tableData to update filter options
watch(
  props.tableData,
  newVal => {
    const uniqueIDs = [...new Set(newVal.map(item => item.id))];
    idFilterOptions.splice(0, idFilterOptions.length, ...uniqueIDs.map(id => ({ text: id as string, value: id as string })));
  },
  {
    immediate: true
  }
);

//Filter Methods
const filterID = (value: string, row: Connection.Connection) => {
  return row.id === value;
};
const filterAddr = (value: string, row: Connection.Connection, column: TableColumnCtx<Connection.Connection>) => {
  const property = column["property"];
  if (property === "remoteaddr" || property === "localaddr") {
    return row[property].ip === value;
  }
  return false;
};

//Formatter
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
</script>
