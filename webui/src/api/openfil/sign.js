import request from '@/utils/request'

export function sign(data) {
    return request({
        url: '/sign',
        method: 'post',
        data: data
    })
}

export function signMsg(from, msg) {
    const data = {
        "from": from,
        "hex_message": msg,
    }

    return request({
        url: '/sign_msg',
        method: 'post',
        data: data
    })
}

export function signAndSend(data) {
    return request({
        url: '/sign_send',
        method: 'post',
        data: data
    })
}
