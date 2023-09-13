import { PersistedStateOptions } from "pinia-plugin-persistedstate";

/**
 * @description Configuration parameters for Pinia persistence
 * @param {String} key The name to be stored in persistence
 * @param {Array} paths The state names that need to be persisted
 * @return persist
 * */
const piniaPersistConfig = (key: string, paths?: string[]) => {
  const persist: PersistedStateOptions = {
    key,
    storage: localStorage,
    paths
  };
  return persist;
};

export default piniaPersistConfig;
