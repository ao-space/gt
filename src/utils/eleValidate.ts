export const validatorTimeFormat = (rule: any, value: any, callback: any) => {
  console.log("Calling validatorTimeFormat");
  const regex = /^(?:\d+(?:ns|us|µs|ms|s|m|h))+$/;
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
  const ipv4Pattern = "\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}";
  const ipv6Pattern = "\\[([0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}\\]";
  const domainPattern = "([a-zA-Z0-9-]+\\.)*[a-zA-Z0-9-]+";

  const regex = new RegExp(`^(?:(?:${ipv4Pattern}|${ipv6Pattern}|${domainPattern})?:?)${portPattern}$`);

  if (!value) {
    callback();
  } else if (regex.test(value)) {
    const parts = value.split(":");
    const port = parseInt(parts[parts.length - 1]);
    const MIN_PORT = 1;
    const MAX_PORT = 65535;
    if (MIN_PORT <= port && port <= MAX_PORT) {
      console.log("regex test passed");
      callback();
    } else {
      callback(new Error("Port number out of range"));
    }
  } else {
    callback(new Error("Please enter a valid address"));
  }
};
