package wallet

import (
	"context"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/gin-gonic/gin"
	"time"
)

func (w *Wallet) Balance(c *gin.Context) {
	addrStr, ok := c.GetQuery("address")
	if !ok {
		ReturnError(c, ParamErr)
		return
	}

	ctx := context.Background()

	addr, err := address.NewFromString(addrStr)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	amount, err := w.node.Api.WalletBalance(timeoutCtx, addr)
	cancel()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, client.BalanceInfo{
		Address: addrStr,
		Amount:  types.FIL(amount).String(),
	})
}
