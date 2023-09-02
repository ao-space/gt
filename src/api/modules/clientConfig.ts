import http from "@/api";
import { Config } from "@/api/interface/index";
import { ClientConfig } from "@/components/ClientConfigForm/interface";

export const getRunningClientConfigApi = () => {
  return http.get<Config.Client.ResConfig>(`/config/running`);
};

export const getClientConfigFromFileApi = () => {
  return http.get<Config.Client.ResConfig>(`/config/file`);
};

export const saveClientConfigApi = (data: ClientConfig.Config) => {
  return http.post<ClientConfig.Config>(`/config/save`, data);
};
