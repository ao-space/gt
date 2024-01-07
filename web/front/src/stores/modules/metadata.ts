import { defineStore } from "pinia";
import { reactive, ref, watch } from "vue";
import i18n from "@/languages";
import { useAuthStore } from "@/stores/modules/auth";

export const useMetadataStore = defineStore("GT-metadata", () => {
  let auth_store = useAuthStore();
  const language = ref(navigator.language.split("-")[0].toLocaleLowerCase());
  let local_language = localStorage.getItem("lang");
  if (local_language != null) {
    language.value = local_language;
  }
  console.log("lang:", language);
  watch(language, changeLangStatus);
  async function changeLangStatus(new_data: string) {
    await auth_store.getAuthMenuList();
    language.value = new_data;
    if (new_data == "zh") {
      i18n.global.locale = "zh";
      localStorage.setItem("lang", "zh");
    } else {
      i18n.global.locale = "en";
      localStorage.setItem("lang", "en");
    }
    update();
  }
  let serverComponentList = reactive([
    { name: "GeneralSetting", title: i18n.global.t("sconfig.GeneralSetting") },
    { name: "NetworkSetting", title: i18n.global.t("sconfig.NetworkSetting") },
    { name: "SecuritySetting", title: i18n.global.t("sconfig.SecuritySetting") },
    { name: "ConnectionSetting", title: i18n.global.t("sconfig.ConnectionSetting") },
    { name: "APISetting", title: i18n.global.t("sconfig.APISetting") },
    { name: "SentrySetting", title: i18n.global.t("sconfig.SentrySetting") },
    { name: "LogSetting", title: i18n.global.t("sconfig.LogSetting") },
    { name: "User1Setting", title: i18n.global.t("sconfig.User") + 1 + i18n.global.t("sconfig.Setting") }
  ]);

  let clientComponentList = reactive([
    { name: "GeneralSetting", title: i18n.global.t("cconfig.GeneralSetting") },
    { name: "SentrySetting", title: i18n.global.t("cconfig.SentrySetting") },
    { name: "WebRTCSetting", title: i18n.global.t("cconfig.WebRTCSetting") },
    { name: "TCPForwardSetting", title: i18n.global.t("cconfig.TCPForwardSetting") },
    { name: "LogSetting", title: i18n.global.t("cconfig.LogSetting") },
    { name: "Service1Setting", title: i18n.global.t("cconfig.Service") + 1 + i18n.global.t("cconfig.Setting") }
  ]);
  let update = () => {
    Object.assign(serverComponentList, [
      { name: "GeneralSetting", title: i18n.global.t("sconfig.GeneralSetting") },
      { name: "NetworkSetting", title: i18n.global.t("sconfig.NetworkSetting") },
      { name: "SecuritySetting", title: i18n.global.t("sconfig.SecuritySetting") },
      { name: "ConnectionSetting", title: i18n.global.t("sconfig.ConnectionSetting") },
      { name: "APISetting", title: i18n.global.t("sconfig.APISetting") },
      { name: "SentrySetting", title: i18n.global.t("sconfig.SentrySetting") },
      { name: "LogSetting", title: i18n.global.t("sconfig.LogSetting") },
      { name: "User1Setting", title: i18n.global.t("sconfig.User") + 1 + i18n.global.t("sconfig.Setting") }
    ]);
    Object.assign(clientComponentList, [
      { name: "GeneralSetting", title: i18n.global.t("cconfig.GeneralSetting") },
      { name: "SentrySetting", title: i18n.global.t("cconfig.SentrySetting") },
      { name: "WebRTCSetting", title: i18n.global.t("cconfig.WebRTCSetting") },
      { name: "TCPForwardSetting", title: i18n.global.t("cconfig.TCPForwardSetting") },
      { name: "LogSetting", title: i18n.global.t("cconfig.LogSetting") },
      { name: "Service1Setting", title: i18n.global.t("cconfig.Service") + 1 + i18n.global.t("cconfig.Setting") }
    ]);
  };
  return { language, changeLangStatus, serverComponentList, clientComponentList, update };
});
