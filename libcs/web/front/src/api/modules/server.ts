import http from "@/api";
import { Server } from "@/api/interface/index";

export const getServerInfoApi = () => {
  return http.get<Server.ResServerInfo>(`/server/info`, {}, { noLoading: true });
};

// currently use for gt-client to control Server's behavior
export const reloadServicesApi = () => {
  return http.put(`/server/reload`);
};

export const restartServerApi = () => {
  return http.put(`/server/restart`);
};

export const stopServerApi = () => {
  return http.put(`/server/stop`);
};

export const killServerApi = () => {
  return http.put(`/server/kill`);
};
