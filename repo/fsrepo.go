package repo

import (
	"errors"
	fslock "github.com/ipfs/go-fs-lock"
	logging "github.com/ipfs/go-log/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/xerrors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Repo interface {
	// APIEndpoint returns multiaddress for communication with Lotus API
	APIEndpoint() (multiaddr.Multiaddr, error)

	// Lock locks the repo for exclusive use.
	Lock() (LockedRepo, error)
}

const (
	fsAPI       = "api"
	fsDatastore = "datastore"
	fsLock      = "repo.lock"
)

var (
	ErrNoAPIEndpoint     = errors.New("API not running (no endpoint)")
	ErrRepoAlreadyLocked = errors.New("repo is already locked (openfild already running)")
	ErrClosedRepo        = errors.New("repo is no longer open")
)

var log = logging.Logger("repo")

type FsRepo struct {
	path string
}

func NewFS(path string) (*FsRepo, error) {
	path, err := homedir.Expand(path)
	if err != nil {
		return nil, err
	}

	return &FsRepo{
		path: path,
	}, nil
}

func (fsr *FsRepo) Exists() (bool, error) {
	_, err := os.Stat(filepath.Join(fsr.path, fsDatastore))

	notexist := os.IsNotExist(err)
	if notexist {
		return false, nil
	}

	return true, err
}

func (fsr *FsRepo) Init() error {
	exist, err := fsr.Exists()
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	log.Infof("Initializing repo at '%s'", fsr.path)
	err = os.MkdirAll(fsr.path, 0755) //nolint: gosec
	if err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}

func (fsr *FsRepo) APIEndpoint() (multiaddr.Multiaddr, error) {
	p := filepath.Join(fsr.path, fsAPI)

	f, err := os.Open(p)
	if os.IsNotExist(err) {
		return nil, ErrNoAPIEndpoint
	} else if err != nil {
		return nil, err
	}
	defer f.Close() //nolint: errcheck // Read only op

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, xerrors.Errorf("failed to read %q: %w", p, err)
	}
	strma := string(data)
	strma = strings.TrimSpace(strma)

	apima, err := multiaddr.NewMultiaddr(strma)
	if err != nil {
		return nil, err
	}

	return apima, nil
}

func (fsr *FsRepo) Lock() (LockedRepo, error) {
	locked, err := fslock.Locked(fsr.path, fsLock)
	if err != nil {
		return nil, xerrors.Errorf("could not check lock status: %w", err)
	}
	if locked {
		return nil, ErrRepoAlreadyLocked
	}

	closer, err := fslock.Lock(fsr.path, fsLock)
	if err != nil {
		return nil, xerrors.Errorf("could not lock the repo: %w", err)
	}

	return &fsLockedRepo{
		path:   fsr.path,
		closer: closer,
	}, nil
}

func (fsr *FsRepo) LockRO() (LockedRepo, error) {
	lr, err := fsr.Lock()
	if err != nil {
		return nil, err
	}

	lr.(*fsLockedRepo).readonly = true
	return lr, nil
}
