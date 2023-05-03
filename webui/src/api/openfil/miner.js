import request from '@/utils/request'

export function withdraw(minerId, amount) {
    const data = {
        "miner_id": minerId,
        "amount": amount
    }

    return request({
        url: '/miner/withdraw',
        method: 'post',
        data: data
    })
}

export function changeOwner(minerId, newOwnerAddr, fromAddr) {
    const data = {
        "miner_id": minerId,
        "new_owner": newOwnerAddr,
        "from": fromAddr
    }

    return request({
        url: '/miner/change_owner',
        method: 'post',
        data: data
    })
}

export function changeWorker(minerId, newWorker) {
    const data = {
        "miner_id": minerId,
        "new_worker": newWorker
    }

    return request({
        url: '/miner/change_worker',
        method: 'post',
        data: data
    })
}

export function confirmWorker(minerId, newWorker) {
    const data = {
        "miner_id": minerId,
        "new_worker": newWorker
    }

    return request({
        url: '/miner/confirm_change_worker',
        method: 'post',
        data: data
    })
}

export function changeBeneficiary(minerId, beneficiaryAddress, quota, expiration, overwritePendingChange) {
    const data = {
        "miner_id": minerId,
        "beneficiary_address": beneficiaryAddress,
        "quota": quota,
        "expiration": expiration,
        "overwrite_pending_change": overwritePendingChange
    }

    return request({
        url: '/miner/change_beneficiary',
        method: 'post',
        data: data
    })
}

export function confirmBeneficiary(minerId) {
    const data = {
        "miner_id": minerId,
    }

    return request({
        url: '/miner/confirm_change_beneficiary',
        method: 'post',
        data: data
    })
}

export function changeControl(minerId, newControlAddrs) {
    const data = {
        "miner_id": minerId,
        "new_controlAddrs": newControlAddrs,
    }

    return request({
        url: '/miner/change_control',
        method: 'post',
        data: data
    })
}