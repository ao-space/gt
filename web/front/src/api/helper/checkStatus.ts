import { ElMessage } from "element-plus";
import i18n from "@/languages";

/**
 * @description: Validate the network request status code
 * @param {Number} status
 * @return void
 */
export const checkStatus = (status: number) => {
  switch (status) {
    case 400:
      ElMessage.error(i18n.global.t("result.RequestFailed"));
      break;
    case 401:
      ElMessage.error(i18n.global.t("result.LoginExpired"));
      break;
    case 403:
      ElMessage.error(i18n.global.t("result.NoPermission"));
      break;
    case 404:
      ElMessage.error(i18n.global.t("result.ResourceNotFound"));
      break;
    case 405:
      ElMessage.error(i18n.global.t("result.InvalidRequestMethod"));
      break;
    case 408:
      ElMessage.error(i18n.global.t("result.RequestTimedOut"));
      break;
    case 500:
      ElMessage.error(i18n.global.t("result.InternalServerError"));
      break;
    case 502:
      ElMessage.error(i18n.global.t("result.BadGateway"));
      break;
    case 503:
      ElMessage.error(i18n.global.t("result.ServiceUnavailable"));
      break;
    case 504:
      ElMessage.error(i18n.global.t("result.GatewayTimeout"));
      break;
    default:
      ElMessage.error(i18n.global.t("result.UnexpectedError"));
  }
};
