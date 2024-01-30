/**
 * @description gain the prompt language of current time
 * @returns {String}
 */
export function getTimeState() {
  let timeNow = new Date();
  let hours = timeNow.getHours();
  if (hours >= 0 && hours < 5) return `Good midnight 🌛`;
  if (hours >= 5 && hours < 12) return `Good morning ⛅`;
  if (hours >= 12 && hours < 18) return `Good afternoon 🌞`;
  if (hours >= 18 && hours <= 24) return `Good evening 🌆`;
}

/**
 * @description use recursion to flatten the menu, which is convenient for adding dynamic routes
 * @param {Array} menuList
 * @returns {Array}
 */
export function getFlatMenuList(menuList: Menu.MenuOptions[]): Menu.MenuOptions[] {
  let newMenuList: Menu.MenuOptions[] = JSON.parse(JSON.stringify(menuList));
  return newMenuList.flatMap(item => [item, ...(item.children ? getFlatMenuList(item.children) : [])]);
}

/**
 * @description use recursion to filter out the list that needs to be rendered in the left menu (the menu with isHide == true needs to be removed)
 * @param {Array} menuList
 * @returns {Array}
 * */
export function getShowMenuList(menuList: Menu.MenuOptions[]) {
  let newMenuList: Menu.MenuOptions[] = JSON.parse(JSON.stringify(menuList));
  return newMenuList.filter(item => {
    item.children?.length && (item.children = getShowMenuList(item.children));
    return !item.meta?.isHide;
  });
}

/**
 * @description use recursion to find out all breadcrumbs and store them in pinia/vuex
 * @param {Array} menuList
 * @param {Array} parent parent menu
 * @param {Object} result
 * @returns {Object}
 */
export const getAllBreadcrumbList = (menuList: Menu.MenuOptions[], parent = [], result: { [key: string]: any } = {}) => {
  for (const item of menuList) {
    result[item.path] = [...parent, item];
    if (item.children) getAllBreadcrumbList(item.children, result[item.path], result);
  }
  return result;
};
