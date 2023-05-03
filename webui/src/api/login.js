import request from '@/utils/request'

export function login(password) {
  const data = {
    "login_password": password,
  }
  return request({
    url: '/login',
    method: 'post',
    data: data
  })
}

export function logout() {
  return request({
    url: '/logout',
    method: 'post'
  })
}
