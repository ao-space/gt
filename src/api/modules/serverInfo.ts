import http from "@/api";
import { Server } from "@/api/interface/index";

export const getServerInfoApi = () => {
  return http.get<Server.ResServerInfo>(`/server/info`, {}, { noLoading: true });
};
