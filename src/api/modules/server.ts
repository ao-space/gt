import http from "@/api";
import { Server } from "@/api/interface/index";

export const getServerInfoApi = () => {
  return http.get<Server.ResServerInfo>(`/server/info`, {}, { noLoading: true });
};

export const reloadServicesApi = () => {
  return http.put(`/server/reload`);
};

export const restartServerApi = () => {
  return http.put(`/server/restart`);
};
