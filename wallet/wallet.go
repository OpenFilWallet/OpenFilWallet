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

	rootPassword string

	db datastore.WalletDB
	lk sync.Mutex
}

func NewWallet(offline bool, rootPassword string, db datastore.WalletDB, close <-chan struct{}) (*Wallet, error) {
	login := newLogin(close)

	w := &Wallet{
		offline:      offline,
		login:        login,
		signer:       messagesigner.NewSigner(),
		rootPassword: rootPassword,
		db:           db,
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

	keys, err := account.LoadPrivateKeys(db, crypto.GenerateEncryptKey([]byte(rootPassword)))
	if err != nil {
		log.Warnw("NewWallet: LoadPrivateKeys", "err", err)
		return nil, err
	}

	err = w.signer.RegisterSigner(keys...)
	if err != nil {
		return nil, err
	}

	return w, nil
}
