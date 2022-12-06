package wallet

import (
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/gin-gonic/gin"
)

// WalletCreate Post
func (w *Wallet) WalletCreate(c *gin.Context) {
	param := client.CreateWalletRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	var index = uint64(param.Index)
	if param.Index == -1 {
		index, err = w.db.MnemonicIndex()
		if err != nil {
			ReturnError(c, NewError(500, err.Error()))
			return
		}
	}

	mnemonic, err := account.LoadMnemonic(w.db, crypto.GenerateEncryptKey([]byte(w.rootPassword)))
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	nks, err := account.GeneratePrivateKeyFromMnemonicIndex(w.db, mnemonic, int64(index), crypto.GenerateEncryptKey([]byte(w.rootPassword)))
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	err = w.signer.RegisterSigner(nks...)
	if err != nil {
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
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	msigList, err := w.db.MsigWalletList()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	data := make(map[string]interface{})

	walletListMap := make(map[string][]string)
	for _, wallet := range walletList {
		if _, ok := walletListMap[walletType(wallet.Address)]; !ok {
			walletListMap[walletType(wallet.Address)] = []string{wallet.Address}
			continue
		}

		walletListMap[walletType(wallet.Address)] = append(walletListMap[walletType(wallet.Address)], wallet.Address)
	}

	data["msig"] = msigList
	for key, value := range walletListMap {
		data[key] = value
	}

	ReturnOk(c, data)
}

func walletType(address string) string {
	if address[:2] == "f1" || address == "t1" {
		return "secp256k1"
	}

	return "bls"
}
