package datastore

import (
	"errors"
	"github.com/ipfs/go-datastore"
)

type WalletDB struct {
	hStore  *HistoryStore
	ekStore *EthKeyStore
	kStore  *KeyStore
	nStore  *NodeStore
	sStore  *ScryptStore
}

func NewWalletDB(ds datastore.Batching) WalletDB {
	walletDB := WalletDB{
		hStore:  newHistoryStore(ds),
		ekStore: newEthKeyStore(ds),
		kStore:  newKeyStore(ds),
		nStore:  newNodeStore(ds),
		sStore:  newScryptStore(ds),
	}

	walletLists, _ := walletDB.WalletList()
	ethWalletLists, _ := walletDB.EthWalletList()
	lists := append(walletLists, ethWalletLists...)
	if len(lists) != 0 {
		for _, wallet := range lists {
			walletDB.hStore.setupRecorder(wallet.Address)
		}
	}

	return walletDB
}

// ------ scrypt ------

func (db *WalletDB) HasMasterPassword() (bool, error) {
	return db.sStore.has(master)
}

func (db *WalletDB) HasLoginPassword() (bool, error) {
	return db.sStore.has(login)
}

func (db *WalletDB) GetMasterPassword() ([]byte, error) {
	return db.sStore.get(master)
}

func (db *WalletDB) GetLoginPassword() ([]byte, error) {
	return db.sStore.get(login)
}

func (db *WalletDB) SetMasterPassword(password []byte) error {
	if len(password) == 0 {
		return errors.New("password cannot be empty")
	}

	return db.sStore.put(master, password, false)
}

func (db *WalletDB) SetLoginPassword(password []byte) error {
	if len(password) == 0 {
		return errors.New("password cannot be empty")
	}

	return db.sStore.put(login, password, false)
}

func (db *WalletDB) UpdateMasterPassword(password []byte) error {
	if len(password) == 0 {
		return errors.New("password cannot be empty")
	}

	return db.sStore.put(master, password, true)
}

func (db *WalletDB) UpdateLoginPassword(password []byte) error {
	if len(password) == 0 {
		return errors.New("password cannot be empty")
	}

	return db.sStore.put(login, password, true)
}

func (db *WalletDB) DeleteMasterPassword() error {
	return db.sStore.delete(master)
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

func (db *WalletDB) HasMnemonic() (bool, error) {
	return db.kStore.hasM()
}

func (db *WalletDB) GetMnemonic() (*HdWallet, error) {
	return db.kStore.getM()
}

func (db *WalletDB) SetMnemonic(hdWallet *HdWallet) error {
	return db.kStore.putM(hdWallet, false)
}

func (db *WalletDB) UpdateMnemonic(hdWallet *HdWallet) error {
	return db.kStore.putM(hdWallet, true)
}

func (db *WalletDB) DeleteMnemonic() error {
	return db.kStore.deleteM()
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

func (db *WalletDB) DeletePrivate(addr string) error {
	return db.kStore.deleteP(addr)
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

func (db *WalletDB) DeleteMsig(addr string) error {
	return db.kStore.deleteMsig(addr)
}

func (db *WalletDB) MnemonicEthIndex() (uint64, error) {
	return db.ekStore.index()
}

func (db *WalletDB) NextMnemonicEthIndex() (uint64, error) {
	return db.ekStore.nextIndex()
}

func (db *WalletDB) GetEthPrivate(addr string) (*PrivateWallet, error) {
	return db.ekStore.getP(addr)
}

func (db *WalletDB) SetEthPrivate(priWallet *PrivateWallet) error {
	if priWallet == nil {
		return errors.New("priWallet is not allowed to be nil")
	}

	if priWallet.Address == "" {
		return errors.New("address is not allowed to be empty")
	}

	if priWallet.Address[:2] != "0x" {
		return errors.New("address format error")
	}

	return db.ekStore.putP(priWallet, false)
}

func (db *WalletDB) UpdateEthPrivate(priWallet *PrivateWallet) error {
	if priWallet == nil {
		return errors.New("priWallet is not allowed to be nil")
	}

	if priWallet.Address == "" {
		return errors.New("address is not allowed to be empty")
	}

	if priWallet.Address[:2] != "0x" {
		return errors.New("address format error")
	}

	return db.ekStore.putP(priWallet, true)
}

func (db *WalletDB) EthWalletList() ([]PrivateWallet, error) {
	return db.ekStore.listP()
}

func (db *WalletDB) DeleteEthPrivate(addr string) error {
	return db.ekStore.deleteP(addr)
}
