package datastore

import (
	"context"
	"errors"
	"github.com/ipfs/go-datastore"
)

type scryptType int

const (
	master scryptType = iota
	login
)

const (
	scryptMasterPrefix = "/scrypt/root"
	scryptLoginPrefix  = "/scrypt/login"
)

type ScryptStore struct {
	ds datastore.Batching
}

func newScryptStore(ds datastore.Batching) *ScryptStore {
	return &ScryptStore{
		ds: ds,
	}
}

func (db *ScryptStore) put(st scryptType, scryptKey []byte, force bool) error {
	if !force {
		isExist, err := db.has(st)
		if err != nil {
			return err
		}
		if isExist {
			return errors.New("scrypt already exist")
		}
	}

	return db.ds.Put(context.Background(), prefix(st), scryptKey)
}

func (db *ScryptStore) get(st scryptType) ([]byte, error) {
	return db.ds.Get(context.Background(), prefix(st))
}

func (db *ScryptStore) has(st scryptType) (bool, error) {
	return db.ds.Has(context.Background(), prefix(st))
}

func (db *ScryptStore) delete(st scryptType) error {
	return db.ds.Delete(context.Background(), prefix(st))
}

func prefix(st scryptType) datastore.Key {
	switch st {
	case master:
		return datastore.NewKey(scryptMasterPrefix)
	case login:
		return datastore.NewKey(scryptLoginPrefix)
	default:
		panic("Unexpected scrypt type")
	}
}
