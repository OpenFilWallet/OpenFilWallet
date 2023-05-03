import Vue from 'vue'

import Cookies from 'js-cookie'

import Element from 'element-ui'
import './assets/styles/element-variables.scss'

import '@/assets/styles/index.scss'
import "@/assets/styles/openfil.scss"
import App from './App'
import store from './store'
import router from './router'
import permission from './directive/permission'

import './assets/icons'
import './permission'
import Pagination from "@/components/Pagination";
import RightToolbar from "@/components/RightToolbar"
import Clipboard from "vue-clipboard2";
import i18n from "@/api/lang/i18n"

import * as echarts from 'echarts'
Vue.prototype.$echarts = echarts

Vue.prototype.msgSuccess = function (msg) {
  this.$message({ showClose: true, message: msg, type: "success" });
}

Vue.prototype.msgError = function (msg) {
  this.$message({ showClose: true, message: msg, type: "error" });
}

Vue.prototype.msgInfo = function (msg) {
  this.$message.info(msg);
}

Vue.component('Pagination', Pagination)
Vue.component('RightToolbar', RightToolbar)
Vue.use(permission)
Vue.use(Clipboard)
/**
 * If you don't want to use mock-server
 * you want to use MockJs for mock api
 * you can execute: mockXHR()
 *
 * Currently MockJs will be used in the production environment,
 * please remove it before going online! ! !
 */
Vue.use(Element, {
  i18n: (key, value) => i18n.t(key, value)
})

Vue.use(Element, {
  size: Cookies.get('size') || 'medium'
})

Vue.config.productionTip = false

new Vue({
  el: '#app',
  router,
  store,
  i18n,
  render: h => h(App)
})
