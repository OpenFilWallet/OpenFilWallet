import request from '@/utils/request'

export function nodeList() {
    return request({
        url: '/node/list',
        method: 'get',
    })
}

export function nodeAdd(name, endpoint, token) {
    const data = {
        "name": name,
        "endpoint": endpoint,
        "token": token
    }

    return request({
        url: '/node/add',
        method: 'post',
        data: data
    })
}

export function nodeUpdate(name, endpoint, token) {
    const data = {
        "name": name,
        "endpoint": endpoint,
        "token": token
    }

    return request({
        url: '/node/update',
        method: 'post',
        data: data
    })
}

export function useNode(name) {
    const data = {
        "name": name,
    }

    return request({
        url: '/node/use_node',
        method: 'post',
        data: data
    })
}

export function nodeDelete(name) {
    const data = {
        "name": name,
    }

    return request({
        url: '/node/delete',
        method: 'post',
        data: data
    })
}

export function nodeBest() {
    return request({
        url: '/node/best',
        method: 'get',
    })
}