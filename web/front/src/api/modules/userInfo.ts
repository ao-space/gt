import { Register } from "@/api/interface/index";
import http from "@/api";

export const changeInfoApi = (params: Register.ReqRegisterForm) => {
  return http.post<Register.ResRegister>(`/user/change`, params);
};

export const getInfoApi = () => {
  return http.get<Register.ReqRegisterForm>(`/user/info`);
};
