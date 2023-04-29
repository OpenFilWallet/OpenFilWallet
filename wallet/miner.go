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
		log.Warnw("Miner: Withdraw: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}

	msg, msgParams, err := buildmessage.NewWithdrawMessage(w.Api, param.BaseParams, param.MinerId, param.Amount)
	if err != nil {
		log.Warnw("Miner: Withdraw: NewWithdrawMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		log.Warnw("Miner: Withdraw: EncodeMessage", "err", err)
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
		log.Warnw("Miner: ChangeOwner: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}

	msg, msgParams, err := buildmessage.NewChangeOwnerMessage(w.Api, param.BaseParams, param.MinerId, param.NewOwner, param.From)
	if err != nil {
		log.Warnw("Miner: ChangeOwner: NewChangeOwnerMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		log.Warnw("Miner: ChangeOwner: EncodeMessage", "err", err)
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
		log.Warnw("Miner: ChangeWorker: EncodeMessage", "err", err)
		ReturnError(c, ParamErr)
		return
	}

	msg, msgParams, err := buildmessage.NewChangeWorkerMessage(w.Api, param.BaseParams, param.MinerId, param.NewWorker, param.NewControlAddrs...)
	if err != nil {
		log.Warnw("Miner: ChangeWorker: NewChangeWorkerMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		log.Warnw("Miner: ChangeWorker: EncodeMessage", "err", err)
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
		log.Warnw("Miner: ConfirmChangeWorker: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}
	msg, err := buildmessage.NewConfirmUpdateWorkerMessage(w.Api, param.BaseParams, param.MinerId, param.NewWorker)
	if err != nil {
		log.Warnw("Miner: ConfirmChangeWorker: NewConfirmUpdateWorkerMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, nil)
	if err != nil {
		log.Warnw("Miner: ConfirmChangeWorker: EncodeMessage", "err", err)
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
		log.Warnw("Miner: ChangeControl: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}

	msg, msgParams, err := buildmessage.NewChangeWorkerMessage(w.Api, param.BaseParams, param.MinerId, "", param.NewControlAddrs...)
	if err != nil {
		log.Warnw("Miner: ChangeControl: NewChangeWorkerMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		log.Warnw("Miner: ChangeControl: EncodeMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, myMsg)
}

// ControlList Get
func (w *Wallet) ControlList(c *gin.Context) {
	minerId, ok := c.GetQuery("miner_id")
	if !ok {
		log.Warnw("Miner: ControlList: GetQuery", "err", "key: miner_id does not exist")
		ReturnError(c, ParamErr)
		return
	}

	minerAddr, err := address.NewFromString(minerId)
	if err != nil {
		log.Warnw("Miner: ControlList: NewFromString", "minerId", minerId, "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mi, err := w.Api.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		log.Warnw("Miner: ControlList: StateMinerInfo", "minerId", minerId, "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	printMeta := func(addr address.Address) client.Meta {
		meta := client.Meta{
			ID:      addr.String(),
			Balance: "",
		}

		k, err := w.Api.StateAccountKey(ctx, addr, types.EmptyTSK)
		if err == nil {
			meta.Address = k.String()
		} else {
			meta.ID = addr.String() + " (multisig)"
		}
		amount, err := w.node.Api.WalletBalance(ctx, addr)
		if err != nil {
			log.Warnw("Balance: WalletBalance", "err", err.Error())
		} else {
			meta.Balance = types.FIL(amount).String()
		}
		return meta
	}

	controlAddrs := make([]client.Meta, 0)
	for _, addr := range mi.ControlAddresses {
		controlAddrs = append(controlAddrs, printMeta(addr))
	}

	minerControl := client.MinerControl{
		Owner:       printMeta(mi.Owner),
		Beneficiary: printMeta(mi.Beneficiary),
		Worker:      printMeta(mi.Worker),
	}

	if mi.NewWorker.String() != address.UndefAddressString {
		m := printMeta(mi.NewWorker)
		minerControl.NewWorker = &m
	}
	if len(controlAddrs) != 0 {
		minerControl.ControlAddresses = controlAddrs
	}

	ReturnOk(c, minerControl)
}

// ChangeBeneficiary Post
func (w *Wallet) ChangeBeneficiary(c *gin.Context) {
	param := client.ChangeBeneficiaryRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("Miner: ChangeBeneficiary: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}

	msg, msgParams, err := buildmessage.NewChangeBeneficiaryProposeMessage(w.Api, param.BaseParams, param.MinerId, param.BeneficiaryAddress, param.Quota, param.Expiration, param.OverwritePendingChange)
	if err != nil {
		log.Warnw("Miner: ChangeBeneficiary: NewChangeBeneficiaryProposeMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		log.Warnw("Miner: ChangeBeneficiary: EncodeMessage", "err", err)
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
		log.Warnw("Miner: ConfirmChangeBeneficiary: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}
	msg, msgParams, err := buildmessage.NewConfirmChangeBeneficiary(w.Api, param.BaseParams, param.MinerId)
	if err != nil {
		log.Warnw("Miner: ConfirmChangeBeneficiary: NewConfirmChangeBeneficiary", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, msgParams)
	if err != nil {
		log.Warnw("Miner: ConfirmChangeBeneficiary: EncodeMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, myMsg)
}
