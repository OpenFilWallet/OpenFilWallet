package wallet

import (
	"context"
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	"github.com/gin-gonic/gin"
	"time"
)

// EthWalletCreate Post
func (w *Wallet) EthWalletCreate(c *gin.Context) {
	param := client.CreateWalletRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("WalletCreate: BindJSON", "err", err.Error())
		ReturnError(c, ParamErr)
		return
	}

	log.Infow("WalletCreate", "index", param.Index)
	var index = uint64(0)
	if param.Index <= 0 {
		index, err = w.db.NextMnemonicEthIndex()
		if err != nil {
			log.Warnw("WalletCreate: NextMnemonicEthIndex", "err", err.Error())
			ReturnError(c, NewError(500, err.Error()))
			return
		}
	} else {
		index = uint64(param.Index)
	}

	mnemonic, err := account.LoadMnemonic(w.db, crypto.GenerateEncryptKey([]byte(w.masterPassword)))
	if err != nil {
		log.Warnw("WalletCreate: LoadMnemonic", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ethKey, err := account.GenerateEthPrivateKeyFromMnemonicIndex(w.db, mnemonic, int64(index), crypto.GenerateEncryptKey([]byte(w.masterPassword)))
	if err != nil {
		log.Warnw("WalletCreate: GenerateEthPrivateKeyFromMnemonicIndex", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	err = w.signer.RegisterEthSigner(*ethKey)
	if err != nil {
		log.Warnw("WalletCreate: RegisterSigner", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	var newWallet = []string{ethKey.Address.String()}

	ReturnOk(c, client.CreateWalletResponse{
		NewWalletAddrs: newWallet,
	})
}

// EthWalletList Get
func (w *Wallet) EthWalletList(c *gin.Context) {
	_, isBalance := c.GetQuery("balance")
	walletList, err := w.db.EthWalletList()
	if err != nil {
		log.Warnw("WalletList: WalletList", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	data := make([]client.WalletListInfo, 0)
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	for _, wallet := range walletList {
		var amount = types.NewInt(0)

		addr, err := ethtypes.ParseEthAddress(wallet.Address)
		if err != nil {
			log.Warnw("WalletList: ParseEthAddress", "err", err.Error())
			ReturnError(c, NewError(500, err.Error()))
			return
		}
		f4Addr, err := addr.ToFilecoinAddress()
		if err != nil {
			log.Warnw("WalletList: ToFilecoinAddress", "err", err.Error())
			ReturnError(c, NewError(500, err.Error()))
			return
		}

		if isBalance {
			amount, err = w.node.Api.WalletBalance(timeoutCtx, f4Addr)
			if err != nil {
				log.Warnw("Balance: WalletBalance", "err", err.Error())
				ReturnError(c, NewError(500, err.Error()))
				return
			}
		}

		walletId := ""
		id, err := w.node.Api.StateLookupID(timeoutCtx, f4Addr, types.EmptyTSK)
		if err != nil {
			log.Infow("StateLookupID", "err", err.Error())
			walletId = "NotFound"
		} else {
			walletId = id.String()
		}

		data = append(data, client.WalletListInfo{
			WalletType:    "fevm",
			WalletAddress: wallet.Address,
			FilAddress:    f4Addr.String(),
			WalletId:      walletId,
			WalletPath:    wallet.Path,
			Balance:       types.FIL(amount).String(),
		})
	}

	ReturnOk(c, data)
}

func (w *Wallet) EthBalance(c *gin.Context) {
	addrStr, ok := c.GetQuery("address")
	if !ok {
		log.Warnw("Balance: GetQuery", "err", "key: address does not exist")
		ReturnError(c, ParamErr)
		return
	}

	ctx := context.Background()

	addr, err := ethtypes.ParseEthAddress(addrStr)
	if err != nil {
		log.Warnw("WalletList: ParseEthAddress", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	f4Addr, err := addr.ToFilecoinAddress()
	if err != nil {
		log.Warnw("WalletList: ToFilecoinAddress", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	amount, err := w.node.Api.WalletBalance(timeoutCtx, f4Addr)
	cancel()
	if err != nil {
		log.Warnw("Balance: WalletBalance", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, client.BalanceInfo{
		Address:    addrStr,
		FilAddress: f4Addr.String(),
		Amount:     types.FIL(amount).String(),
	})
}

// todo send fil f1/f2/f3 to 0x
// todo send fil 0x to f1/f2/f3
// todo send fil 0x to 0x
// todo fevm tx history
// todo fevm send: check 0x is contract
// todo fevm ui
