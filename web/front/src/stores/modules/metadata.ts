import { defineStore } from "pinia";
import { ref, watch } from "vue";
import i18n from "@/languages";
import { useAuthStore } from "@/stores/modules/auth";

export const useMetadataStore = defineStore("GT-metadata", () => {
  let auth_store = useAuthStore();
  let data = localStorage.getItem("lang");
  let language = ref("en");
  if (data != null) {
    language.value = data;
  }
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
  }
  return { language, changeLangStatus };
});
