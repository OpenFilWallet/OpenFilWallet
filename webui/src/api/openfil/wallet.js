import request from '@/utils/request'

export function create(index) {
    const data = {
        "index": index,
    }

    return request({
        url: '/wallet/create',
        method: 'post',
        data: data
    })
}

export function txHistory(addr) {
    const data = {
        "address": addr,
    }

    return request({
        url: '/tx_history',
        method: 'get',
        params: data
    })
}

export function msigCreate(from, required, duration, value, signers) {
    const data = {
        "from": from,
        "required": required,
        "duration": duration,
        "value": value,
        "signers": signers,
    }
    console.log("-----------------------", data);

    return request({
        url: '/msig/create',
        method: 'post',
        data: data
    })
}

export function msigAdd(msigAddress) {
    const data = {
        "msig_address": msigAddress,
    }

    return request({
        url: '/msig/add',
        method: 'post',
        data: data
    })
}


export function walletList(query) {
    return request({
        url: '/wallet/list',
        method: 'get',
        params: query
    })
}

export function msigWalletList(query) {
    return request({
        url: '/msig/list',
        method: 'get',
        params: query
    })
}