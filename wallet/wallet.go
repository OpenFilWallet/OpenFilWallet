package wallet

import (
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/OpenFilWallet/OpenFilWallet/modules/messagesigner"
	logging "github.com/ipfs/go-log/v2"
	"sync"
)

var log = logging.Logger("wallet-server")

type Wallet struct {
	*login
	*node
	*txTracker

	offline bool

	signer messagesigner.Signer

	masterPassword string

	db datastore.WalletDB
	lk sync.Mutex
}

func NewWallet(offline bool, masterPassword string, db datastore.WalletDB, close <-chan struct{}) (*Wallet, error) {
	login := newLogin(close)

	w := &Wallet{
		offline:        offline,
		login:          login,
		signer:         messagesigner.NewSigner(),
		masterPassword: masterPassword,
		db:             db,
	}

	nodeInfo, err := w.getBestNode()
	if err != nil {
		return nil, err
	}

	n, err := newNode(nodeInfo.Name, nodeInfo.Endpoint, nodeInfo.Token)
	if err == nil {
		w.node = n
	} else {
		log.Warn("no nodes available")
	}

	txTracker := newTxTracker(n, db, close)
	w.txTracker = txTracker

	keys, err := account.LoadPrivateKeys(db, crypto.GenerateEncryptKey([]byte(masterPassword)))
	if err != nil {
		log.Warnw("NewWallet: LoadPrivateKeys", "err", err)
		return nil, err
	}

	err = w.signer.RegisterSigner(keys...)
	if err != nil {
		return nil, err
	}

	ethKeys, err := account.LoadEthPrivateKeys(db, crypto.GenerateEncryptKey([]byte(masterPassword)))
	if err != nil {
		log.Warnw("NewWallet: LoadEthPrivateKeys", "err", err)
		return nil, err
	}

	err = w.signer.RegisterEthSigner(ethKeys...)
	if err != nil {
		return nil, err
	}

	return w, nil
}
