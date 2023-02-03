package wallet

import (
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/gin-gonic/gin"
)

// WalletCreate Post
func (w *Wallet) WalletCreate(c *gin.Context) {
	param := client.CreateWalletRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("WalletCreate: BindJSON", "err", err.Error())
		ReturnError(c, ParamErr)
		return
	}

	var index = uint64(0)
	if param.Index <= 0 {
		index, err = w.db.NextMnemonicIndex()
		if err != nil {
			log.Warnw("WalletCreate: NextMnemonicIndex", "err", err.Error())
			ReturnError(c, NewError(500, err.Error()))
			return
		}
	} else {
		index = uint64(param.Index)
	}

	mnemonic, err := account.LoadMnemonic(w.db, crypto.GenerateEncryptKey([]byte(w.rootPassword)))
	if err != nil {
		log.Warnw("WalletCreate: LoadMnemonic", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	nks, err := account.GeneratePrivateKeyFromMnemonicIndex(w.db, mnemonic, int64(index), crypto.GenerateEncryptKey([]byte(w.rootPassword)))
	if err != nil {
		log.Warnw("WalletCreate: GeneratePrivateKeyFromMnemonicIndex", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	err = w.signer.RegisterSigner(nks...)
	if err != nil {
		log.Warnw("WalletCreate: RegisterSigner", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	var newWallet []string
	for _, nk := range nks {
		newWallet = append(newWallet, nk.Address.String())
	}

	ReturnOk(c, client.CreateWalletResponse{
		NewWalletAddrs: newWallet,
	})
}

// WalletList Get
func (w *Wallet) WalletList(c *gin.Context) {
	walletList, err := w.db.WalletList()
	if err != nil {
		log.Warnw("WalletList: WalletList", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	msigList, err := w.db.MsigWalletList()
	if err != nil {
		log.Warnw("WalletList: MsigWalletList", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	walletListMap := make(map[string][]datastore.PrivateWallet)
	for _, wallet := range walletList {
		if _, ok := walletListMap[walletType(wallet.Address)]; !ok {
			walletListMap[walletType(wallet.Address)] = []datastore.PrivateWallet{wallet}
			continue
		}

		walletListMap[walletType(wallet.Address)] = append(walletListMap[walletType(wallet.Address)], wallet)
	}

	data := make([]client.WalletListInfo, 0)
	for _, ms := range msigList {
		data = append(data, client.WalletListInfo{
			WalletType:    "msig",
			WalletAddress: ms.MsigAddr,
			WalletPath:    "",
		})
	}

	for _, key := range []string{"secp256k1", "bls"} {
		if wallets, ok := walletListMap[key]; ok {
			for _, wallet := range wallets {
				data = append(data, client.WalletListInfo{
					WalletType:    key,
					WalletAddress: wallet.Address,
					WalletPath:    wallet.Path,
				})
			}
		}
	}

	ReturnOk(c, data)
}

func walletType(address string) string {
	if address[:2] == "f1" || address == "t1" {
		return "secp256k1"
	}

	return "bls"
}
