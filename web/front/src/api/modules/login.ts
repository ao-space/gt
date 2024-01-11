import { Login } from "@/api/interface/index";
import authMenuList from "@/assets/json/authMenuList.json";
import http from "@/api";

/**
 * @name Login
 */
// User Login in
export const loginApi = (params: Login.ReqLoginForm) => {
  return http.post<Login.ResLogin>(`/login`, params); // Standard post json request  ==>  application/json
  // return http.post<Login.ResLogin>(`/login`, params, { noLoading: true }); // Control the current request to not show loading
  // return http.post<Login.ResLogin>(`/login`, {}, { params }); // post request with query parameters  ==>  ?username=admin&password=123456
};

export const verifyKeyApi = (params: Login.ReqKeyValue) => {
  return http.get<Login.ResLogin>(`/verify`, params, { noLoading: false });
};

//Get Menu Permission
export const getAuthMenuListApi = () => {
  return http.get<Menu.MenuOptions[]>(`/permission/menu`, {}, { noLoading: true });

  //If you want to make the menu a local data,
  //comment out the previous line of code and introduce the local authMenuList.json data

  return authMenuList; //for test
};
