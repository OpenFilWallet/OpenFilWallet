import request from '@/utils/request'

export function send(data) {
    return request({
        url: '/send',
        method: 'post',
        data: data
    })
}