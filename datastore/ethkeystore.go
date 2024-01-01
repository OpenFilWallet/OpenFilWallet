package datastore

import (
	"encoding/json"
	"errors"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
)

const (
	ethMnemonicIndexPrefix = "/eth/keystore/mnemonic/index"
	ethPrivatePrefix       = "/eth/keystore/private"
)

type EthKeyStore struct {
	mnemonicIndex *StoredIndex
	privateStore  *StateStore
}

func newEthKeyStore(ds datastore.Batching) *EthKeyStore {
	return &EthKeyStore{
		mnemonicIndex: NewStoredIndex(ds, datastore.NewKey(ethMnemonicIndexPrefix)), // Share fil mnemonic
		privateStore:  NewStateStore(namespace.Wrap(ds, datastore.NewKey(ethPrivatePrefix))),
	}
}

func (db *EthKeyStore) index() (uint64, error) {
	return db.mnemonicIndex.Get()
}

func (db *EthKeyStore) nextIndex() (uint64, error) {
	return db.mnemonicIndex.Next()
}

func (db *EthKeyStore) putP(priWallet *PrivateWallet, force bool) error {
	if !force {
		isExist, err := db.privateStore.Has(priWallet.Address)
		if err != nil {
			return err
		}
		if isExist {
			return errors.New("private already exist")
		}
	}

	return db.privateStore.Begin(priWallet.Address, priWallet, force)
}

func (db *EthKeyStore) getP(addr string) (*PrivateWallet, error) {
	b, err := db.privateStore.Get(addr).Get()
	if err != nil {
		return nil, err
	}

	var priWallet PrivateWallet
	err = json.Unmarshal(b, &priWallet)
	if err != nil {
		return nil, err
	}

	return &priWallet, nil
}

func (db *EthKeyStore) hasP(addr string) (bool, error) {
	return db.privateStore.Has(addr)
}

func (db *EthKeyStore) deleteP(addr string) error {
	return db.privateStore.Get(addr).Delete()
}

func (db *EthKeyStore) listP() ([]PrivateWallet, error) {
	var priWallets []PrivateWallet
	err := db.privateStore.List(&priWallets)
	if err != nil {
		return nil, err
	}

	return priWallets, nil
}
