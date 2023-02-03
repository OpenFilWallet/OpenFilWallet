package wallet

import (
	"github.com/OpenFilWallet/OpenFilWallet/chain"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/modules/buildmessage"
	"github.com/gin-gonic/gin"
)

// Transfer Post
func (w *Wallet) Transfer(c *gin.Context) {
	param := client.TransferRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("Transfer: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}

	msg, err := buildmessage.NewTransferMessage(w.node.Api, param.BaseParams, param.From, param.To, param.Amount)
	if err != nil {
		log.Warnw("Transfer: NewTransferMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, nil)
	if err != nil {
		log.Warnw("Transfer: EncodeMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, myMsg)
	return
}
