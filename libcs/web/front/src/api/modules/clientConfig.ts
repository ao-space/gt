import http from "@/api";
import { Config } from "@/api/interface/index";
import { ClientConfig, ClientConfigBackend, transToBackendConfig } from "@/components/ClientConfigForm/interface";

export const getClientConfigFromFileApi = () => {
  return http.get<Config.Client.ResConfigBackend>(`/config/file`);
};

export const saveClientConfigApi = (data: ClientConfig.Config) => {
  return http.post<ClientConfigBackend.Config>(`/config/save`, transToBackendConfig(data));
};
