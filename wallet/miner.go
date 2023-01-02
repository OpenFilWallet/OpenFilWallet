package wallet

import (
	"context"
	"github.com/OpenFilWallet/OpenFilWallet/chain"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/modules/buildmessage"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/gin-gonic/gin"
	"time"
)

// Withdraw Post
func (w *Wallet) Withdraw(c *gin.Context) {
	param := client.WithdrawRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msg, msgParams, err := buildmessage.NewWithdrawMessage(w.Api, param.BaseParams, param.MinerId, param.Amount)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, myMsg)
	return
}

// ChangeOwner Post
func (w *Wallet) ChangeOwner(c *gin.Context) {
	param := client.ChangeOwnerRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msg, msgParams, err := buildmessage.NewChangeOwnerMessage(w.Api, param.BaseParams, param.MinerId, param.NewOwner, param.From)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, myMsg)
}

// ChangeWorker Post
func (w *Wallet) ChangeWorker(c *gin.Context) {
	param := client.ChangeWorkerRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msg, msgParams, err := buildmessage.NewChangeWorkerMessage(w.Api, param.BaseParams, param.MinerId, param.NewWorker, param.NewControlAddrs...)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, myMsg)
}

// ConfirmChangeWorker Post
func (w *Wallet) ConfirmChangeWorker(c *gin.Context) {
	param := client.ConfirmChangeWorkerRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}
	msg, err := buildmessage.NewConfirmUpdateWorkerMessage(w.Api, param.BaseParams, param.MinerId, param.NewWorker)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, nil)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, myMsg)
}

// ChangeControl Post
func (w *Wallet) ChangeControl(c *gin.Context) {
	param := client.ChangeWorkerRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msg, msgParams, err := buildmessage.NewChangeWorkerMessage(w.Api, param.BaseParams, param.MinerId, "", param.NewControlAddrs...)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, myMsg)
}

// ControlList Get
func (w *Wallet) ControlList(c *gin.Context) {
	minerId, ok := c.GetQuery("miner_id")
	if !ok {
		ReturnError(c, ParamErr)
		return
	}

	minerAddr, err := address.NewFromString(minerId)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	mi, err := w.Api.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	cancel()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, client.MinerControl{
		Owner:            mi.Owner.String(),
		Worker:           mi.Worker.String(),
		NewWorker:        mi.NewWorker.String(),
		ControlAddresses: addr2Str(mi.ControlAddresses),
	})
}

// ChangeBeneficiary Post
func (w *Wallet) ChangeBeneficiary(c *gin.Context) {
	param := client.ChangeBeneficiaryRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msg, msgParams, err := buildmessage.NewChangeBeneficiaryProposeMessage(w.Api, param.BaseParams, param.MinerId, param.BeneficiaryAddress, param.Quota, param.Expiration, param.OverwritePendingChange)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, myMsg)
}

// ConfirmChangeBeneficiary Post
func (w *Wallet) ConfirmChangeBeneficiary(c *gin.Context) {
	param := client.ConfirmChangeBeneficiaryRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}
	msg, msgParams, err := buildmessage.NewConfirmChangeBeneficiary(w.Api, param.BaseParams, param.MinerId, param.ExistingBeneficiary, param.NewBeneficiary)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, myMsg)
}
