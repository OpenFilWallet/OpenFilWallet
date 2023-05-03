import Vue from 'vue';
import VueI18n from 'vue-i18n';
import enLocale from './en-us'
import cnLocal from './zh-cn'
import uiEnLocal from "element-ui/lib/locale/lang/en"
import uiCnLocal from "element-ui/lib/locale/lang/zh-CN"
Vue.use(VueI18n);
const messages = {
  en: {
    ...enLocale,
    ...uiEnLocal
  },
  cn: {
    ...cnLocal,
    ...uiCnLocal
  }
}
const i18n = new VueI18n({
  locale: localStorage.getItem('lang') || 'en',
  messages,
});


export default i18n;
