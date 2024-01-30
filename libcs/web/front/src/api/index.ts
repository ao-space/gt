import axios, { AxiosInstance, AxiosError, AxiosRequestConfig, InternalAxiosRequestConfig, AxiosResponse } from "axios";
import { showFullScreenLoading, tryHideFullScreenLoading } from "@/config/serviceLoading";
import { LOGIN_URL } from "@/config";
import { ElMessage } from "element-plus";
import { ResultData } from "@/api/interface";
import { ResultEnum } from "@/enums/httpEnum";
import { checkStatus } from "./helper/checkStatus";
import { useUserStore } from "@/stores/modules/user";
import router from "@/routers";

export interface CustomAxiosRequestConfig extends InternalAxiosRequestConfig {
  noLoading?: boolean;
}

const config = {
  // Default API URL, can be modified in .env.** files
  baseURL: import.meta.env.VITE_API_URL as string,
  // Setting timeout duration
  timeout: ResultEnum.TIMEOUT as number,
  // Allow credentials for cross-origin requests
  withCredentials: true
};

class RequestHttp {
  service: AxiosInstance;
  public constructor(config: AxiosRequestConfig) {
    // Creating an Axios instance
    this.service = axios.create(config);

    /**
     * @description Request interceptor
     * Client sends request -> [Request Interceptor] -> Server
     * Token validation (JWT): Receive token from the server, store it in vuex/pinia/local storage
     */
    this.service.interceptors.request.use(
      (config: CustomAxiosRequestConfig) => {
        const userStore = useUserStore();
        // if current request is no need for loading,
        // it will be controlled by the third parameter: { noLoading: true } in the api service
        config.noLoading || showFullScreenLoading();
        config.headers["x-token"] = userStore.token;
        return config;
      },
      (error: AxiosError) => {
        return Promise.reject(error);
      }
    );

    /**
     * @description Response interceptor
     * Server returns information -> [Interceptor processes it] -> Client JS receives the information
     */
    this.service.interceptors.response.use(
      (response: AxiosResponse) => {
        const { data } = response;
        const userStore = useUserStore();
        tryHideFullScreenLoading();
        // If login has expired
        if (data.code == ResultEnum.OVERDUE) {
          userStore.setToken("");
          router.replace(LOGIN_URL);
          ElMessage.error(data.msg);
          return Promise.reject(data);
        }
        // Global error message interception (to prevent errors when downloading files that return data streams without a code)
        if (data.code && data.code !== ResultEnum.SUCCESS) {
          ElMessage.error(data.msg);
          return Promise.reject(data);
        }
        // Successful request (on the page, unless there's a special case, no need to handle failure logic)
        return data;
      },
      async (error: AxiosError) => {
        const { response } = error;
        tryHideFullScreenLoading();
        // Separate checks for request timeout and network errors, as they don't have a response
        if (error.message.indexOf("timeout") !== -1) ElMessage.error("Request timed out! Please try again later");
        if (error.message.indexOf("Network Error") !== -1) ElMessage.error("Network error! Please try again later");
        // Handle different server error status codes
        if (response) checkStatus(response.status);
        // If there's no server response (could be a server error or client is offline), handle offline scenario: can redirect to an error page
        if (!window.navigator.onLine) router.replace("/500");
        return Promise.reject(error);
      }
    );
  }

  /**
   * @description Encapsulation of common request methods
   */
  get<T>(url: string, params?: object, _object = {}): Promise<ResultData<T>> {
    return this.service.get(url, { params, ..._object });
  }
  post<T>(url: string, params?: object | string, _object = {}): Promise<ResultData<T>> {
    return this.service.post(url, params, _object);
  }
  put<T>(url: string, params?: object, _object = {}): Promise<ResultData<T>> {
    return this.service.put(url, params, _object);
  }
  delete<T>(url: string, params?: any, _object = {}): Promise<ResultData<T>> {
    return this.service.delete(url, { params, ..._object });
  }
  download(url: string, params?: object, _object = {}): Promise<BlobPart> {
    return this.service.post(url, params, { ..._object, responseType: "blob" });
  }
}

export default new RequestHttp(config);
