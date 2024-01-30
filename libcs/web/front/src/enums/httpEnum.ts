/**
 * @description：Request configuration
 */
export enum ResultEnum {
  SUCCESS = 200,
  ERROR = 500,
  OVERDUE = 401,
  TIMEOUT = 30000,
  TYPE = "success"
}

/**
 * @description：HTTP request methods
 */
export enum RequestEnum {
  GET = "GET",
  POST = "POST",
  PATCH = "PATCH",
  PUT = "PUT",
  DELETE = "DELETE"
}

/**
 * @description：Commonly used content type values
 */
export enum ContentTypeEnum {
  // json
  JSON = "application/json;charset=UTF-8",
  // text
  TEXT = "text/plain;charset=UTF-8",
  // form-data
  FORM_URLENCODED = "application/x-www-form-urlencoded;charset=UTF-8",
  // form-data upload
  FORM_DATA = "multipart/form-data;charset=UTF-8"
}
