package datastore

import (
	"encoding/json"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
)

const nodePrefix = "/node/info"

type NodeStore struct {
	nodeStore *StateStore
}

func newNodeStore(ds datastore.Batching) *NodeStore {
	return &NodeStore{
		nodeStore: NewStateStore(namespace.Wrap(ds, datastore.NewKey(nodePrefix))),
	}
}

func (db *NodeStore) put(info *NodeInfo, force bool) error {
	return db.nodeStore.Begin(info.Name, info, force)
}

func (db *NodeStore) get(name string) (*NodeInfo, error) {
	var nodeInfo NodeInfo
	val, err := db.nodeStore.Get(name).Get()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(val, &nodeInfo)
	if err != nil {
		return nil, err
	}

	return &nodeInfo, nil
}

func (db *NodeStore) has(name string) (bool, error) {
	return db.nodeStore.Has(name)
}

func (db *NodeStore) delete(name string) error {
	return db.nodeStore.Get(name).Delete()
}

func (db *NodeStore) list() ([]NodeInfo, error) {
	var nodeInfos []NodeInfo
	err := db.nodeStore.List(&nodeInfos)
	if err != nil {
		return nil, err
	}

	return nodeInfos, nil
}
