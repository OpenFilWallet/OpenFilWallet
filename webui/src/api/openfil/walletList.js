import request from '@/utils/request'

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