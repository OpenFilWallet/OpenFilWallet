package wallet

import (
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/filecoin-project/go-address"
	"github.com/gin-gonic/gin"
)

// TxHistory Get
func (w *Wallet) TxHistory(c *gin.Context) {
	addr := c.Query("address")
	_, err := address.NewFromString(addr)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	historys, err := w.db.HistoryList(addr)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	var hs []client.HistoryResponse
	for _, h := range historys {
		hs = append(hs, client.HistoryResponse{
			Version:    h.Version,
			To:         h.To,
			From:       h.From,
			Nonce:      h.Nonce,
			Value:      h.Value,
			GasLimit:   h.GasLimit,
			GasFeeCap:  h.GasFeeCap,
			GasPremium: h.GasPremium,
			Method:     h.Method,
			Params:     h.Params,
			TxCid:      h.TxCid,
			TxState:    string(h.TxState),
		})
	}

	ReturnOk(c, hs)
}
