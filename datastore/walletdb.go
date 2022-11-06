package datastore

import (
	"errors"
	"github.com/ipfs/go-datastore"
)

type WalletDB struct {
	hStore *HistoryStore
	kStore *KeyStore
	nStore *NodeStore
	sStore *ScryptStore
}

func NewWalletDB(ds datastore.Batching) WalletDB {
	return WalletDB{
		hStore: newHistoryStore(ds),
		kStore: newKeyStore(ds),
		nStore: newNodeStore(ds),
		sStore: newScryptStore(ds),
	}
}

// ------ scrypt ------

func (db *WalletDB) GetRootPassword() ([]byte, error) {
	return db.sStore.get(root)
}

func (db *WalletDB) GetLoginPassword() ([]byte, error) {
	return db.sStore.get(login)
}

func (db *WalletDB) SetRootPassword(password []byte) error {
	if len(password) == 0 {
		return errors.New("password cannot be empty")
	}

	return db.sStore.put(root, password, false)
}

func (db *WalletDB) SetLoginPassword(password []byte) error {
	if len(password) == 0 {
		return errors.New("password cannot be empty")
	}

	return db.sStore.put(login, password, false)
}

func (db *WalletDB) UpdateRootPassword(password []byte) error {
	if len(password) == 0 {
		return errors.New("password cannot be empty")
	}

	return db.sStore.put(root, password, true)
}

func (db *WalletDB) UpdateLoginPassword(password []byte) error {
	if len(password) == 0 {
		return errors.New("password cannot be empty")
	}

	return db.sStore.put(login, password, true)
}

func (db *WalletDB) DeleteRootPassword() error {
	return db.sStore.delete(root)
}

func (db *WalletDB) DeleteLoginPassword() error {
	return db.sStore.delete(login)
}

// ------ node ------

func (db *WalletDB) GetNode(name string) (*NodeInfo, error) {
	if name == "" {
		return nil, errors.New("node name cannot be empty")
	}

	return db.nStore.get(name)
}

func (db *WalletDB) SetNode(nodeInfo *NodeInfo) error {
	if nodeInfo.Name == "" {
		return errors.New("node name cannot be empty")
	}

	return db.nStore.put(nodeInfo, false)
}

func (db *WalletDB) UpdateNode(nodeInfo *NodeInfo) error {
	if nodeInfo.Name == "" {
		return errors.New("node name cannot be empty")
	}

	return db.nStore.put(nodeInfo, true)
}

func (db *WalletDB) DeleteNode(name string) error {
	if name == "" {
		return errors.New("node name cannot be empty")
	}

	return db.nStore.delete(name)
}

func (db *WalletDB) NodeList() ([]NodeInfo, error) {
	return db.nStore.list()
}

// ------ history -------

func (db *WalletDB) GetHistory(addr string, nonce uint64) (*History, error) {
	if addr == "" {
		return nil, errors.New("addr cannot be empty")
	}

	return db.hStore.get(addr, nonce)
}

func (db *WalletDB) SetHistory(msg *History) error {
	if msg.From == "" {
		return errors.New("addr cannot be empty")
	}

	return db.hStore.put(msg, false)
}

func (db *WalletDB) UpdateHistory(msg *History) error {
	if msg.From == "" {
		return errors.New("addr cannot be empty")
	}

	return db.hStore.put(msg, true)
}

func (db *WalletDB) HistoryList(addr string) ([]History, error) {
	if addr == "" {
		return nil, errors.New("addr cannot be empty")
	}

	return db.hStore.list(addr)
}

// ------ keystore ------

func (db *WalletDB) GetMnemonic() (*HdWallet, error) {
	return db.kStore.getM()
}

func (db *WalletDB) SetMnemonic(hdWallet *HdWallet) error {
	return db.kStore.putM(hdWallet, false)
}

func (db *WalletDB) UpdateMnemonic(hdWallet *HdWallet) error {
	return db.kStore.putM(hdWallet, true)
}

func (db *WalletDB) MnemonicIndex() (uint64, error) {
	return db.kStore.index()
}

func (db *WalletDB) NextMnemonicIndex() (uint64, error) {
	return db.kStore.nextIndex()
}

func (db *WalletDB) GetPrivate(addr string) (*PrivateWallet, error) {
	return db.kStore.getP(addr)
}

func (db *WalletDB) SetPrivate(priWallet *PrivateWallet) error {
	if priWallet == nil {
		return errors.New("priWallet is not allowed to be nil")
	}

	if priWallet.Address == "" {
		return errors.New("address is not allowed to be empty")
	}

	return db.kStore.putP(priWallet, false)
}

func (db *WalletDB) UpdatePrivate(priWallet *PrivateWallet) error {
	if priWallet == nil {
		return errors.New("priWallet is not allowed to be nil")
	}

	if priWallet.Address == "" {
		return errors.New("address is not allowed to be empty")
	}

	return db.kStore.putP(priWallet, true)
}

func (db *WalletDB) WalletList() ([]PrivateWallet, error) {
	return db.kStore.listP()
}

func (db *WalletDB) GetMsig(addr string) (*MsigWallet, error) {
	return db.kStore.getMsig(addr)
}

func (db *WalletDB) SetMsig(msigWallet *MsigWallet) error {
	if msigWallet == nil {
		return errors.New("msigWallet is not allowed to be nil")
	}

	if msigWallet.MsigAddr == "" {
		return errors.New("MsigAddr is not allowed to be empty")
	}

	return db.kStore.putMsig(msigWallet, false)
}

func (db *WalletDB) UpdateMsig(msigWallet *MsigWallet) error {
	if msigWallet == nil {
		return errors.New("msigWallet is not allowed to be nil")
	}

	if msigWallet.MsigAddr == "" {
		return errors.New("MsigAddr is not allowed to be empty")
	}

	return db.kStore.putMsig(msigWallet, true)
}

func (db *WalletDB) MsigWalletList() ([]MsigWallet, error) {
	return db.kStore.listMsig()
}
