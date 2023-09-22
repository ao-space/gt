import { ElMessage } from "element-plus";

/**
 * @description: Validate the network request status code
 * @param {Number} status
 * @return void
 */
export const checkStatus = (status: number) => {
  switch (status) {
    case 400:
      ElMessage.error("Request failed! Please try again later.");
      break;
    case 401:
      ElMessage.error("Login expired! Please log in again.");
      break;
    case 403:
      ElMessage.error("You do not have permission to access this resource.");
      break;
    case 404:
      ElMessage.error("The resource you are trying to access does not exist!");
      break;
    case 405:
      ElMessage.error("Invalid request method! Please try again later");
      break;
    case 408:
      ElMessage.error("Request timed out! Please try again later");
      break;
    case 500:
      ElMessage.error("Internal server error.");
      break;
    case 502:
      ElMessage.error("Bad gateway.");
      break;
    case 503:
      ElMessage.error("Service is currently unavailable. Please try again later.");
      break;
    case 504:
      ElMessage.error("Gateway timeout. The server took too long to respond.");
      break;
    default:
      ElMessage.error("An unexpected error occurred. Please try again.");
  }
};
