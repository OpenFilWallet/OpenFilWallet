import router from './router'
import store from './store'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import { getToken } from '@/utils/auth'

NProgress.configure({ showSpinner: false })

const loginPath = '/login'

router.beforeEach((to, from, next) => {
  NProgress.start()
  if (getToken()) {
    if (to.path === loginPath) {
      next({ path: '/' })
      NProgress.done()
    } else {
      store.dispatch('GenerateRoutes').then(accessRoutes => {
        router.addRoutes(accessRoutes)
        next({ ...to, replace: true })
      }).catch(err => {
        store.dispatch('LogOut').then(() => {
          next({ path: '/' })
        })
      })
      next()
    }
  } else {
    if (to.path === loginPath) {
      next()
    } else {
      next(`/login`)
      NProgress.done()
    }
  }
})

router.afterEach(() => {
  NProgress.done()
})
