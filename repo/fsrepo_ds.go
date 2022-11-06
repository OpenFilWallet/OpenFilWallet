package repo

import (
	"context"
	"errors"
	"github.com/ipfs/go-datastore"
	levelds "github.com/ipfs/go-ds-leveldb"
	"github.com/multiformats/go-multiaddr"
	ldbopts "github.com/syndtr/goleveldb/leveldb/opt"
	"golang.org/x/xerrors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type LockedRepo interface {
	// Close closes repo and removes lock.
	Close() error

	// Datastore Returns datastore defined in this repo.
	// The supplied context must only be used to initialize the datastore.
	// The implementation should not retain the context for usage throughout
	// the lifecycle.
	Datastore(ctx context.Context) (datastore.Batching, error)

	// Path returns absolute path of the repo
	Path() string

	// Readonly returns true if the repo is readonly
	Readonly() bool

	// SetAPIEndpoint sets the endpoint of the current API
	SetAPIEndpoint(multiaddr.Multiaddr) error
}

type fsLockedRepo struct {
	path     string
	closer   io.Closer
	readonly bool

	ds     datastore.Batching
	dsErr  error
	dsOnce sync.Once
}

func (fsr *fsLockedRepo) Close() error {
	err := os.Remove(fsr.join(fsAPI))

	if err != nil && !os.IsNotExist(err) {
		return xerrors.Errorf("could not remove API file: %w", err)
	}
	if fsr.ds != nil {
		if err := fsr.ds.Close(); err != nil {
			return xerrors.Errorf("could not close datastore: %w", err)
		}
	}

	err = fsr.closer.Close()
	fsr.closer = nil
	return err
}

func (fsr *fsLockedRepo) Datastore(_ context.Context) (datastore.Batching, error) {
	fsr.dsOnce.Do(func() {
		fsr.ds, fsr.dsErr = fsr.openDatastore(fsr.readonly)
	})

	if fsr.dsErr != nil {
		return nil, fsr.dsErr
	}
	if fsr.ds != nil {
		return fsr.ds, nil
	}

	return nil, errors.New("no such datastore")
}

func (fsr *fsLockedRepo) Readonly() bool {
	return fsr.readonly
}

func (fsr *fsLockedRepo) Path() string {
	return fsr.path
}

func (fsr *fsLockedRepo) join(paths ...string) string {
	return filepath.Join(append([]string{fsr.path}, paths...)...)
}

func (fsr *fsLockedRepo) SetAPIEndpoint(ma multiaddr.Multiaddr) error {
	if err := fsr.stillValid(); err != nil {
		return err
	}
	return ioutil.WriteFile(fsr.join(fsAPI), []byte(ma.String()), 0644)
}

func (fsr *fsLockedRepo) stillValid() error {
	if fsr.closer == nil {
		return ErrClosedRepo
	}
	return nil
}

func (fsr *fsLockedRepo) openDatastore(readonly bool) (datastore.Batching, error) {
	if err := os.MkdirAll(fsr.join(fsDatastore), 0755); err != nil {
		return nil, xerrors.Errorf("mkdir %s: %w", fsr.join(fsDatastore), err)
	}

	ds, err := levelDs(fsr.join(fsDatastore), readonly)
	if err != nil {
		return nil, xerrors.Errorf("opening datastore: %w", err)
	}

	return ds, nil
}

func levelDs(path string, readonly bool) (datastore.Batching, error) {
	return levelds.NewDatastore(path, &levelds.Options{
		Compression: ldbopts.NoCompression,
		NoSync:      false,
		Strict:      ldbopts.StrictAll,
		ReadOnly:    readonly,
	})
}
