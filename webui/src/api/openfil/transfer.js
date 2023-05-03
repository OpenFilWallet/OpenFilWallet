import request from '@/utils/request'

export function transfer(from, to, amount) {
    const data = {
        "from": from,
        "to": to,
        "amount": amount
    }

    return request({
        url: '/transfer',
        method: 'post',
        data: data
    })
}