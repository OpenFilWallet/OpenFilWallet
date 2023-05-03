import axios from 'axios'
import { Notification, MessageBox, Message } from 'element-ui'
import store from '@/store'
import { getToken } from '@/utils/auth'

axios.defaults.headers['Content-Type'] = 'application/json;charset=utf-8'
const service = axios.create({
  baseURL: process.env.VUE_APP_BASE_API,
  timeout: 10000000
})

service.interceptors.request.use(config => {

  if (getToken()) {
    config.headers['Authorization'] = 'Bearer ' + getToken()
  }

  if (config.method === 'get' && config.params) {
    let url = config.url + '?';
    for (const propName of Object.keys(config.params)) {
      const value = config.params[propName];
      var part = encodeURIComponent(propName) + "=";
      if (value !== null && typeof (value) !== "undefined") {
        if (typeof value === 'object') {
          for (const key of Object.keys(value)) {
            let params = propName + '[' + key + ']';
            var subPart = encodeURIComponent(params) + "=";
            url += subPart + encodeURIComponent(value[key]) + "&";
          }
        } else {
          url += part + encodeURIComponent(value) + "&";
        }
      }
    }
    url = url.slice(0, -1);
    config.params = {};
    config.url = url;
  }
  return config
}, error => {
  console.log(error)
  Promise.reject(error)
})

service.interceptors.response.use(res => {
  const code = res.data.code || 200;
  const msg = res.data.message

  if (msg == "wallet is locked, please login") {
    MessageBox.confirm(msg, '', {
      confirmButtonText: 'OK',
      cancelButtonText: 'Cancel',
      type: 'warning'
    }
    ).then(() => {
      store.dispatch('LogOut').then(() => {
        location.href = '/index';
      })
    })
  } else if (code === 500) {
    Message({
      message: msg,
      type: 'error'
    })
    return Promise.reject(new Error(msg))
  } else if (code !== 200) {
    Notification.error({
      title: msg
    })
    return Promise.reject('error')
  } else {
    return res.data
  }
},
  error => {
    console.log('err' + error)
    let { message } = error;

    Message({
      message: message,
      type: 'error',
      duration: 1 * 1000
    })
    return Promise.reject(error)
  }
)

export default service
