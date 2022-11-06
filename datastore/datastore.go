package datastore

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"go.uber.org/multierr"
	"golang.org/x/xerrors"
	"reflect"
	"sync"
)

type StateStore struct {
	ds datastore.Datastore
}

func NewStateStore(ds datastore.Datastore) *StateStore {
	return &StateStore{ds: ds}
}

func ToKey(k interface{}) datastore.Key {
	switch t := k.(type) {
	case uint64:
		return datastore.NewKey(fmt.Sprint(t))
	case fmt.Stringer:
		return datastore.NewKey(t.String())
	default:
		panic("unexpected key type")
	}
}

func (st *StateStore) Begin(i interface{}, state interface{}, force bool) error {
	k := ToKey(i)
	has, err := st.ds.Has(context.TODO(), k)
	if err != nil {
		return err
	}
	if has && !force {
		return xerrors.Errorf("already tracking state for %v", i)
	}

	b, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return st.ds.Put(context.TODO(), k, b)
}

func (st *StateStore) Get(i interface{}) *StoredState {
	return &StoredState{
		ds:   st.ds,
		name: ToKey(i),
	}
}

func (st *StateStore) Has(i interface{}) (bool, error) {
	return st.ds.Has(context.TODO(), ToKey(i))
}

func (st *StateStore) List(out interface{}) error {
	res, err := st.ds.Query(context.TODO(), query.Query{})
	if err != nil {
		return err
	}
	defer res.Close()

	outT := reflect.TypeOf(out).Elem().Elem()
	rout := reflect.ValueOf(out)

	var errs error

	for {
		res, ok := res.NextSync()
		if !ok {
			break
		}
		if res.Error != nil {
			return res.Error
		}

		elem := reflect.New(outT)
		err := json.Unmarshal(res.Value, elem.Interface())
		if err != nil {
			errs = multierr.Append(errs, xerrors.Errorf("Unmarshal state for key '%s': %w", res.Key, err))
			continue
		}

		rout.Elem().Set(reflect.Append(rout.Elem(), elem.Elem()))
	}

	return errs
}

type StoredState struct {
	ds   datastore.Datastore
	name datastore.Key
}

func (st *StoredState) Delete() error {
	has, err := st.ds.Has(context.TODO(), st.name)
	if err != nil {
		return err
	}
	if !has {
		return xerrors.Errorf("No state for %s", st.name)
	}
	if err := st.ds.Delete(context.TODO(), st.name); err != nil {
		return xerrors.Errorf("removing state from datastore: %w", err)
	}
	st.name = datastore.Key{}
	st.ds = nil

	return nil
}

func (st *StoredState) Get() ([]byte, error) {
	val, err := st.ds.Get(context.TODO(), st.name)
	if err != nil {
		if xerrors.Is(err, datastore.ErrNotFound) {
			return nil, xerrors.Errorf("No state for %s: %w", st.name, err)
		}
		return nil, err
	}

	return val, nil
}

func (st *StoredState) Mutate(mutator interface{}) error {
	return st.mutate(jsonMutator(mutator))
}

func (st *StoredState) mutate(mutator func([]byte) ([]byte, error)) error {
	has, err := st.ds.Has(context.TODO(), st.name)
	if err != nil {
		return err
	}
	if !has {
		return xerrors.Errorf("No state for %s", st.name)
	}

	cur, err := st.ds.Get(context.TODO(), st.name)
	if err != nil {
		return err
	}

	mutated, err := mutator(cur)
	if err != nil {
		return err
	}

	if bytes.Equal(mutated, cur) {
		return nil
	}

	return st.ds.Put(context.TODO(), st.name, mutated)
}

func jsonMutator(mutator interface{}) func([]byte) ([]byte, error) {
	rmut := reflect.ValueOf(mutator)

	return func(in []byte) ([]byte, error) {
		state := reflect.New(rmut.Type().In(0).Elem())

		err := json.Unmarshal(in, state.Interface())
		if err != nil {
			return nil, err
		}

		out := rmut.Call([]reflect.Value{state})

		if err := out[0].Interface(); err != nil {
			return nil, err.(error)
		}

		return json.Marshal(state.Interface())
	}
}

// StoredIndex is a counter that persists to a datastore as it increments
type StoredIndex struct {
	lock sync.Mutex
	ds   datastore.Datastore
	name datastore.Key
}

// NewStoredIndex returns a new StoredCounter for the given datastore and key
func NewStoredIndex(ds datastore.Datastore, name datastore.Key) *StoredIndex {
	return &StoredIndex{ds: ds, name: name}
}

// Next returns the next counter value, updating it on disk in the process
// if no counter is present, it creates one and returns a 0 value
func (si *StoredIndex) Next() (uint64, error) {
	ctx := context.TODO()
	si.lock.Lock()
	defer si.lock.Unlock()

	has, err := si.ds.Has(ctx, si.name)
	if err != nil {
		return 0, err
	}

	var next uint64 = 0
	if has {
		curBytes, err := si.ds.Get(ctx, si.name)
		if err != nil {
			return 0, err
		}
		cur, _ := binary.Uvarint(curBytes)
		next = cur + 1
	}
	buf := make([]byte, binary.MaxVarintLen64)
	size := binary.PutUvarint(buf, next)

	return next, si.ds.Put(ctx, si.name, buf[:size])
}

// Get returns current counter value
func (si *StoredIndex) Get() (uint64, error) {
	ctx := context.TODO()
	si.lock.Lock()
	defer si.lock.Unlock()

	has, err := si.ds.Has(ctx, si.name)
	if err != nil {
		return 0, err
	}

	var cur uint64 = 0
	if has {
		curBytes, err := si.ds.Get(ctx, si.name)
		if err != nil {
			return 0, err
		}
		cur, _ = binary.Uvarint(curBytes)
	}

	return cur, nil
}
