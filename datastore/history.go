package datastore

import (
	"encoding/json"
	"errors"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"path/filepath"
	"sync"
)

const historyBasePrefix = "/transaction/history"

type HistoryStore struct {
	ds       datastore.Batching
	recorder map[string]*StateStore
	lk       sync.Mutex
}

func newHistoryStore(ds datastore.Batching) *HistoryStore {
	return &HistoryStore{
		ds:       ds,
		recorder: make(map[string]*StateStore),
	}
}

func (db *HistoryStore) put(msg *History, force bool) error {
	var store *StateStore
	db.lk.Lock()
	if _, ok := db.recorder[msg.From]; !ok {
		db.recorder[msg.From] = NewStateStore(namespace.Wrap(db.ds, txHistoryKey(msg.From)))
	}
	store = db.recorder[msg.From]
	db.lk.Unlock()

	return store.Begin(msg.Nonce, msg, force)
}

func (db *HistoryStore) get(addr string, nonce uint64) (*History, error) {
	store, err := db.getStore(addr)
	if err != nil {
		return nil, err
	}

	b, err := store.Get(nonce).Get()
	if err != nil {
		return nil, err
	}

	var msg History
	err = json.Unmarshal(b, &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (db *HistoryStore) has(addr string, nonce uint64) (bool, error) {
	store, err := db.getStore(addr)
	if err != nil {
		return false, err
	}

	return store.Has(nonce)
}

func (db *HistoryStore) list(addr string) ([]History, error) {
	store, err := db.getStore(addr)
	if err != nil {
		return nil, err
	}

	var msgs []History
	err = store.List(&msgs)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (db *HistoryStore) getStore(addr string) (*StateStore, error) {
	db.lk.Lock()
	defer db.lk.Unlock()
	var store *StateStore

	if st, ok := db.recorder[addr]; !ok {
		return nil, errors.New("history does not exist")
	} else {
		store = st
	}

	return store, nil
}

func txHistoryKey(addr string) datastore.Key {
	return datastore.NewKey(filepath.Join(historyBasePrefix, addr))
}
