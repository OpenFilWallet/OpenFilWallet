package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
)

const (
	mnemonicPrefix      = "/keystore/mnemonic"
	mnemonicIndexPrefix = "/keystore/mnemonic/index"
	privatePrefix       = "/keystore/private"
	msigPrefix          = "/keystore/msig"
)

type KeyStore struct {
	mnemonicStore datastore.Batching
	mnemonicIndex *StoredIndex
	privateStore  *StateStore
	msigStore     *StateStore
}

func newKeyStore(ds datastore.Batching) *KeyStore {
	return &KeyStore{
		mnemonicStore: ds,
		mnemonicIndex: NewStoredIndex(ds, datastore.NewKey(mnemonicIndexPrefix)),
		privateStore:  NewStateStore(namespace.Wrap(ds, datastore.NewKey(privatePrefix))),
		msigStore:     NewStateStore(namespace.Wrap(ds, datastore.NewKey(msigPrefix))),
	}
}

func (db *KeyStore) putM(hdWallet *HdWallet, force bool) error {
	ctx := context.Background()
	if !force {
		isExist, err := db.mnemonicStore.Has(ctx, datastore.NewKey(mnemonicPrefix))
		if err != nil {
			return err
		}
		if isExist {
			return errors.New("mnemonic already exist")
		}
	}

	b, err := json.Marshal(hdWallet)
	if err != nil {
		return err
	}

	return db.mnemonicStore.Put(ctx, datastore.NewKey(mnemonicPrefix), b)
}

func (db *KeyStore) getM() (*HdWallet, error) {
	b, err := db.mnemonicStore.Get(context.Background(), datastore.NewKey(mnemonicPrefix))
	if err != nil {
		return nil, err
	}

	var hdWallet HdWallet
	err = json.Unmarshal(b, &hdWallet)
	if err != nil {
		return nil, err
	}

	return &hdWallet, nil
}

func (db *KeyStore) hasM() (bool, error) {
	return db.mnemonicStore.Has(context.Background(), datastore.NewKey(mnemonicPrefix))
}

func (db *KeyStore) deleteM() error {
	return db.mnemonicStore.Delete(context.Background(), datastore.NewKey(mnemonicPrefix))
}

func (db *KeyStore) index() (uint64, error) {
	return db.mnemonicIndex.Get()
}

func (db *KeyStore) nextIndex() (uint64, error) {
	return db.mnemonicIndex.Next()
}

func (db *KeyStore) putP(priWallet *PrivateWallet, force bool) error {
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

func (db *KeyStore) getP(addr string) (*PrivateWallet, error) {
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

func (db *KeyStore) hasP(addr string) (bool, error) {
	return db.privateStore.Has(addr)
}

func (db *KeyStore) deleteP(addr string) error {
	return db.privateStore.Get(addr).Delete()
}

func (db *KeyStore) listP() ([]PrivateWallet, error) {
	var priWallets []PrivateWallet
	err := db.privateStore.List(&priWallets)
	if err != nil {
		return nil, err
	}

	return priWallets, nil
}

func (db *KeyStore) putMsig(msigWallet *MsigWallet, force bool) error {
	if !force {
		isExist, err := db.msigStore.Has(msigWallet.MsigAddr)
		if err != nil {
			return err
		}
		if isExist {
			return errors.New("private already exist")
		}
	}

	return db.msigStore.Begin(msigWallet.MsigAddr, msigWallet, force)
}

func (db *KeyStore) getMsig(addr string) (*MsigWallet, error) {
	b, err := db.msigStore.Get(addr).Get()
	if err != nil {
		return nil, err
	}

	var msigWallet MsigWallet
	err = json.Unmarshal(b, &msigWallet)
	if err != nil {
		return nil, err
	}

	return &msigWallet, nil
}

func (db *KeyStore) hasMsig(addr string) (bool, error) {
	return db.msigStore.Has(addr)
}

func (db *KeyStore) deleteMsig(addr string) error {
	return db.msigStore.Get(addr).Delete()
}

func (db *KeyStore) listMsig() ([]MsigWallet, error) {
	var msigWallets []MsigWallet
	err := db.msigStore.List(&msigWallets)
	if err != nil {
		return nil, err
	}

	return msigWallets, nil
}
