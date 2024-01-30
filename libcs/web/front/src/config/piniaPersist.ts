import { PersistedStateOptions } from "pinia-plugin-persistedstate";

/**
 * @description Configuration parameters for Pinia persistence
 * @param {String} key The name to be stored in persistence
 * @param {Storage} storage The storage method used, such as: localStorage, sessionStorage, etc.
 * @param {Array} paths The state names that need to be persisted
 * @return persist
 * */
const piniaPersistConfig = (key: string, storage: Storage, paths?: string[]) => {
  const persist: PersistedStateOptions = {
    key,
    storage,
    paths
  };
  return persist;
};

export default piniaPersistConfig;
