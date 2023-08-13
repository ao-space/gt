// ? Element 常用表单校验规则

/**
 *  @rule 手机号
 */
export function checkPhoneNumber(rule: any, value: any, callback: any) {
  const regexp = /^(((13[0-9]{1})|(15[0-9]{1})|(16[0-9]{1})|(17[3-8]{1})|(18[0-9]{1})|(19[0-9]{1})|(14[5-7]{1}))+\d{8})$/;
  if (value === "") callback("请输入手机号码");
  if (!regexp.test(value)) {
    callback(new Error("请输入正确的手机号码"));
  } else {
    return callback();
  }
}

export const validatorTimeFormat = (rule: any, value: any, callback: any) => {
  console.log("Calling validatorTimeFormat");
  const regex = /^(\d+(ns|us|µs|ms|s|m|h))+$/;
  if (!value) {
    callback(new Error("Please enter a value"));
  } else if (regex.test(value)) {
    console.log("regex test passed");
    console.log(value);
    callback();
  } else {
    console.log("regex test failed");
    console.log(value);
    callback(new Error("Please enter a valid time format"));
  }
};

export const validatorRange = (rule: any, value: any, callback: any) => {
  console.log("Calling validatorRange");
  const regex = /^\d+-\d+$/;
  if (!value) {
    callback(new Error("Please enter a value"));
  } else if (regex.test(value)) {
    console.log("regex test passed");
    callback();
  } else {
    console.log("regex test failed");
    callback(new Error("Please enter a valid range format"));
  }
};

export const validatorPositiveInteger = (rule: any, value: any, callback: any) => {
  console.log("Calling validatorPositiveInteger");
  const regex = /^\d+$/;
  if (!value) {
    callback(new Error("Please enter a value"));
  } else if (regex.test(value) && value > 0) {
    console.log("regex test passed");
    callback();
  } else {
    console.log("regex test failed");
    callback(new Error("Please enter a valid positive integer"));
  }
};

export const validatorAddr = (rule: any, value: any, callback: any) => {
  console.log("Calling validatorAddr");
  const portPattern = "\\d{1,5}";
  const ipPattern = "(?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}|0\\.0\\.0\\.0)";
  const regex = new RegExp(`^(?:${ipPattern}:)?${portPattern}$|^${portPattern}$`);
  if (!value) {
    callback(new Error("Please enter a value"));
  }
  if (regex.test(value)) {
    const parts = value.split(":");
    const port = parseInt(parts[parts.length - 1]);
    const MIN_PORT = 1;
    const MAX_PORT = 65535;
    if (MIN_PORT <= port && port <= MAX_PORT) {
      console.log("regex test passed");
      callback();
    }
  }
  return callback(new Error("Please enter a valid address"));
};
