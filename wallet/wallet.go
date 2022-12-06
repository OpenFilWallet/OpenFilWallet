package wallet

import (
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/OpenFilWallet/OpenFilWallet/modules/messagesigner"
	logging "github.com/ipfs/go-log/v2"
	"sync"
)

var log = logging.Logger("wallet-server")

type Wallet struct {
	offline bool

	*login
	*node

	signer messagesigner.Signer

	rootPassword string

	db datastore.WalletDB
	lk sync.Mutex
}

func NewWallet(offline bool, rootPassword string, db datastore.WalletDB, close <-chan struct{}) *Wallet {
	login := newLogin(close)

	w := &Wallet{
		offline:      offline,
		login:        login,
		signer:       messagesigner.NewSigner(),
		rootPassword: rootPassword,
		db:           db,
	}

	nodeInfo, err := w.getBestNode()
	if err == nil {
		node, err := newNode(nodeInfo.Name, nodeInfo.Endpoint, nodeInfo.Token)
		if err == nil {
			w.node = node
		} else {
			log.Warn("no lotus daemon node available")
		}
	}

	return w
}
