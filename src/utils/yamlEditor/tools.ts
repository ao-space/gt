// import YAML from "yaml";
// import yaml from "js-yaml";

// interface ConversionResult {
//   data: string | object;
//   error: boolean;
// }

// // Convert JSON to YAML format
// export const json2yaml = (jsonData: string | object): ConversionResult => {
//   try {
//     const data = typeof jsonData === "string" ? yaml.dump(JSON.parse(jsonData)) : yaml.dump(jsonData);
//     return {
//       data,
//       error: false
//     };
//   } catch (err) {
//     return {
//       data: "",
//       error: true
//     };
//   }
// };

// // Convert YAML to JSON format
// export const yaml2json = (yamlStr: string, returnString: boolean): ConversionResult => {
//   try {
//     const data = returnString ? JSON.stringify(YAML.parse(yamlStr), null, 2) : YAML.parse(yamlStr);
//     return {
//       data,
//       error: false
//     };
//   } catch (err) {
//     return {
//       data: "",
//       error: true
//     };
//   }
// };
