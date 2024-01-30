import { ElNotification } from "element-plus";

/**
 * @description Global error handler
 * */
const errorHandler = (error: any) => {
  // Filter out HTTP request errors
  if (error.status || error.status === 0) return false;
  let errorMap: { [key: string]: string } = {
    InternalError: "Internal error in the JavaScript engine",
    ReferenceError: "Object not found",
    TypeError: "Incorrect type or object used",
    RangeError: "Parameter out of range when using a built-in object",
    SyntaxError: "Syntax error",
    EvalError: "Incorrect use of Eval",
    URIError: "URI error"
  };
  let errorName = errorMap[error.name] || "Unknown error";
  ElNotification({
    title: errorName,
    message: error,
    type: "error",
    duration: 3000
  });
};

export default errorHandler;
