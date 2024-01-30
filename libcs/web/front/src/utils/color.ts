import { ElMessage } from "element-plus";

/**
 * @description Convert hex color to RGB color
 * @param {String} str Color value string
 * @returns {String} Returns the processed color value
 */
export function hexToRgb(str: any) {
  let hexs: any = "";
  let reg = /^\#?[0-9A-Fa-f]{6}$/;
  if (!reg.test(str)) return ElMessage.warning("输入错误的hex");
  str = str.replace("#", "");
  hexs = str.match(/../g);
  for (let i = 0; i < 3; i++) hexs[i] = parseInt(hexs[i], 16);
  return hexs;
}

/**
 * @description Convert RGB color to Hex color
 * @param {*} r Represents red
 * @param {*} g Represents green
 * @param {*} b Represents blue
 * @returns {String} Returns the processed color value
 */
export function rgbToHex(r: any, g: any, b: any) {
  let reg = /^\d{1,3}$/;
  if (!reg.test(r) || !reg.test(g) || !reg.test(b)) return ElMessage.warning("输入错误的rgb颜色值");
  let hexs = [r.toString(16), g.toString(16), b.toString(16)];
  for (let i = 0; i < 3; i++) if (hexs[i].length == 1) hexs[i] = `0${hexs[i]}`;
  return `#${hexs.join("")}`;
}

/**
 * @description Darken the color value
 * @param {String} color Color value string
 * @param {Number} level Darkening degree, between 0-1
 * @returns {String} Returns the processed color value
 */
export function getDarkColor(color: string, level: number) {
  let reg = /^\#?[0-9A-Fa-f]{6}$/;
  if (!reg.test(color)) return ElMessage.warning("输入错误的hex颜色值");
  let rgb = hexToRgb(color);
  for (let i = 0; i < 3; i++) rgb[i] = Math.round(20.5 * level + rgb[i] * (1 - level));
  return rgbToHex(rgb[0], rgb[1], rgb[2]);
}

/**
 * @description Lighten the color value
 * @param {String} color Color value string
 * @param {Number} level Lightening degree, between 0-1
 * @returns {String} Returns the processed color value
 */
export function getLightColor(color: string, level: number) {
  let reg = /^\#?[0-9A-Fa-f]{6}$/;
  if (!reg.test(color)) return ElMessage.warning("输入错误的hex颜色值");
  let rgb = hexToRgb(color);
  for (let i = 0; i < 3; i++) rgb[i] = Math.round(255 * level + rgb[i] * (1 - level));
  return rgbToHex(rgb[0], rgb[1], rgb[2]);
}
