import { ElLoading } from "element-plus";

/* 全局请求 loading */
let loadingInstance: ReturnType<typeof ElLoading.service>;

/**
 * @description Start Loading
 * */
const startLoading = () => {
  loadingInstance = ElLoading.service({
    fullscreen: true,
    lock: true,
    text: "Loading",
    background: "rgba(0, 0, 0, 0.7)"
  });
};

/**
 * @description End the Loading
 * */
const endLoading = () => {
  loadingInstance.close();
};

/**
 * @description Display full-screen loading
 * */
let needLoadingRequestCount = 0;
export const showFullScreenLoading = () => {
  if (needLoadingRequestCount === 0) {
    startLoading();
  }
  needLoadingRequestCount++;
};

/**
 * @description Hide full-screen loading
 * */
export const tryHideFullScreenLoading = () => {
  if (needLoadingRequestCount <= 0) return;
  needLoadingRequestCount--;
  if (needLoadingRequestCount === 0) {
    endLoading();
  }
};
