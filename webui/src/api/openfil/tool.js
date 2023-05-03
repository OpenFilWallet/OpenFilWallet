import request from '@/utils/request'

export function decode(to, method, params) {
    const data = {
        "to_addr": to,
        "method": method,
        "params": params,
        "encoding": "hex",
    }

    return request({
        url: '/chain/decode',
        method: 'post',
        data: data
    })
}

export function encode(to, method, params) {
    const data = {
        "dest": to,
        "method": method,
        "params": params,
        "encoding": "hex",
    }

    return request({
        url: '/chain/encode',
        method: 'post',
        data: data
    })
}

export function balance(addr) {
    const data = {
        "address": addr,
    }

    return request({
        url: '/balance',
        method: 'get',
        params: data
    })
}

export function msigInspect(addr) {
    const data = {
        "msig_address": addr,
    }

    return request({
        url: '/msig/inspect',
        method: 'get',
        params: data
    })
}

export function minerControl(miner) {
    const data = {
        "miner_id": miner,
    }

    return request({
        url: '/miner/control_list',
        method: 'get',
        params: data
    })
}