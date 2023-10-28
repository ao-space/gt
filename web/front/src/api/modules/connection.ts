import http from "@/api";
import { Connection } from "@/api/interface/index";

export const getConnectionApi = () => {
  return http.get<Connection.ResConnection>(`/connection/list`);
};
