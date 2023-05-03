import request from '@/utils/request'

export function msigApprove(from, msigAddress, txId) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "tx_id": txId
    }
    return request({
        url: '/msig/approve',
        method: 'post',
        data: data
    })
}

export function msigCancel(from, msigAddress, txId) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "tx_id": txId
    }

    return request({
        url: '/msig/cancel',
        method: 'post',
        data: data
    })
}

export function msigTransfer(from, msigAddress, to, amount) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "destination_address": to,
        "amount": amount
    }

    return request({
        url: '/msig/transfer_propose',
        method: 'post',
        data: data
    })
}

export function msigAdd(from, msigAddress, signerAddress, increaseThreshold) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "signer_address": signerAddress,
        "increase_threshold": increaseThreshold
    }

    return request({
        url: '/msig/add_signer_propose',
        method: 'post',
        data: data
    })
}

export function msigSwap(from, msigAddress, oldAddress, newAddress) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "old_address": oldAddress,
        "new_address": newAddress
    }

    return request({
        url: '/msig/swap_propose',
        method: 'post',
        data: data
    })
}

export function msigChangeThreshold(from, msigAddress, newThreshold) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "new_threshold": newThreshold
    }

    return request({
        url: '/msig/threshold_propose',
        method: 'post',
        data: data
    })
}

export function msigChangeOwner(from, msigAddress, minerId, newOwner) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "miner_id": minerId,
        "new_owner": newOwner
    }

    return request({
        url: '/msig/change_owner_propose',
        method: 'post',
        data: data
    })
}

export function msigWithdraw(from, msigAddress, minerId, amount) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "miner_id": minerId,
        "amount": amount
    }

    return request({
        url: '/msig/withdraw_propose',
        method: 'post',
        data: data
    })
}

export function msigChangeWorker(from, msigAddress, minerId, amount) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "miner_id": minerId,
        "new_worker": newWorker
    }

    return request({
        url: '/msig/change_worker_propose',
        method: 'post',
        data: data
    })
}

export function msigConfirmChangeWorker(from, msigAddress, minerId, amount) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "miner_id": minerId,
        "new_worker": newWorker
    }

    return request({
        url: '/msig/confirm_change_worker_propose',
        method: 'post',
        data: data
    })
}

export function msigChangeBeneficiary(from, msigAddress, minerId, beneficiaryAddress, quota, expiration, overwritePendingChange) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "miner_id": minerId,
        "beneficiary_address": beneficiaryAddress,
        "quota": quota,
        "expiration": expiration,
        "overwrite_pending_change": overwritePendingChange,
    }

    return request({
        url: '/msig/change_beneficiary_propose',
        method: 'post',
        data: data
    })
}

export function msigConfirmChangeBeneficiary(from, msigAddress, minerId) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "miner_id": minerId,
    }

    return request({
        url: '/msig/confirm_change_beneficiary_propose',
        method: 'post',
        data: data
    })
}

export function msigChangeControl(from, msigAddress, controlAddrs) {
    const data = {
        "from": from,
        "msig_address": msigAddress,
        "control_addrs": controlAddrs,
    }

    return request({
        url: '/msig/set_control_propose',
        method: 'post',
        data: data
    })
}