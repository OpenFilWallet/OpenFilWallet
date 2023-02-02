package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/chain"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/OpenFilWallet/OpenFilWallet/modules/buildmessage"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/multisig"
	"github.com/filecoin-project/lotus/chain/consensus/filcns"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/gin-gonic/gin"
	cbor "github.com/ipfs/go-ipld-cbor"
	cbg "github.com/whyrusleeping/cbor-gen"
	"reflect"
	"sort"
)

// MsigWalletList Get
func (w *Wallet) MsigWalletList(c *gin.Context) {
	msigList, err := w.db.MsigWalletList()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	data := []client.MsigWalletListInfo{}
	for _, ms := range msigList {
		data = append(data, client.MsigWalletListInfo{
			MsigAddr:              ms.MsigAddr,
			Signers:               ms.Signers,
			NumApprovalsThreshold: ms.NumApprovalsThreshold,
		})
	}

	ReturnOk(c, data)
}

// MsigCreate Post
func (w *Wallet) MsigCreate(c *gin.Context) {
	param := client.MsigCreateRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigCreateMessage(param.BaseParams, param.Required, param.Duration, param.Value, param.From, param.Signers...)
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

// MsigAdd Post
func (w *Wallet) MsigAdd(c *gin.Context) {
	param := client.MsigAddRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	_, err = w.db.GetMsig(param.MsigAddress)
	if err == nil {
		ReturnError(c, NewError(500, "msig wallet already exists"))
		return
	}

	msigWallet, errResponse := w.inquireMsigInfo(param.MsigAddress)
	if errResponse != nil {
		ReturnError(c, errResponse)
		return
	}

	isExist := false
	for _, addr := range msigWallet.Signers {
		msigSignerAddr, err := address.NewFromString(addr)
		if err != nil {
			continue
		}
		signerActor, err := w.Api.StateAccountKey(context.Background(), msigSignerAddr, types.EmptyTSK)
		if err != nil {
			continue
		}
		if ok := w.signer.HasSigner(signerActor.String()); ok {
			isExist = true
			break
		}
	}

	if !isExist {
		ReturnError(c, NewError(500, "signers are not included in the wallet"))
		return
	}

	err = w.db.SetMsig(msigWallet)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, nil)
}

// MsigUpdate Post
func (w *Wallet) MsigUpdate(c *gin.Context) {
	param := client.MsigUpdateRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	_, err = w.db.GetMsig(param.MsigAddress)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	msigWallet, errResponse := w.inquireMsigInfo(param.MsigAddress)
	if errResponse != nil {
		ReturnError(c, errResponse)
		return
	}

	err = w.db.UpdateMsig(msigWallet)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, nil)
}

func (w *Wallet) inquireMsigInfo(msigAddress string) (*datastore.MsigWallet, *client.Response) {
	msigAddr, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, NewError(500, err.Error())
	}

	ctx := context.Background()
	head, err := w.Api.ChainHead(ctx)
	if err != nil {
		return nil, NewError(500, err.Error())
	}

	store := adt.WrapStore(ctx, cbor.NewCborStore(blockstore.NewAPIBlockstore(w.Api)))
	act, err := w.Api.StateGetActor(ctx, msigAddr, head.Key())
	if err != nil {
		return nil, NewError(500, err.Error())
	}

	mstate, err := multisig.Load(store, act)
	if err != nil {
		return nil, NewError(500, err.Error())
	}

	se, err := mstate.StartEpoch()
	if err != nil {
		return nil, NewError(500, err.Error())
	}

	ud, err := mstate.UnlockDuration()
	if err != nil {
		return nil, NewError(500, err.Error())
	}

	signers, err := mstate.Signers()
	if err != nil {
		return nil, NewError(500, err.Error())
	}

	threshold, err := mstate.Threshold()
	if err != nil {
		return nil, NewError(500, err.Error())
	}

	return &datastore.MsigWallet{
		MsigAddr:              msigAddress,
		Signers:               addr2Str(signers),
		NumApprovalsThreshold: threshold,
		UnlockDuration:        int64(ud),
		StartEpoch:            int64(se),
	}, nil
}

// MsigInspect Get
func (w *Wallet) MsigInspect(c *gin.Context) {
	msigAddress, ok := c.GetQuery("msig_address")
	if !ok {
		ReturnError(c, ParamErr)
		return
	}

	msigAddr, err := address.NewFromString(msigAddress)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ctx := context.Background()
	head, err := w.Api.ChainHead(ctx)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	store := adt.WrapStore(ctx, cbor.NewCborStore(blockstore.NewAPIBlockstore(w.Api)))
	act, err := w.Api.StateGetActor(ctx, msigAddr, head.Key())
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	mstate, err := multisig.Load(store, act)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	var inspect = client.MsigInspect{
		MsigAddr: msigAddr.String(),
	}
	inspect.Balance = act.Balance.String()

	ib, err := mstate.InitialBalance()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}
	se, err := mstate.StartEpoch()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}
	ud, err := mstate.UnlockDuration()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}
	locked, err := mstate.LockedBalance(head.Height())
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	inspect.Spendable = types.BigSub(act.Balance, locked).String()
	inspect.Lock.InitialBalance = ib.String()
	inspect.Lock.LockAmount = locked.String()
	inspect.Lock.StartEpoch = uint64(se)
	inspect.Lock.UnlockDuration = uint64(ud)

	signers, err := mstate.Signers()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}
	threshold, err := mstate.Threshold()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}
	inspect.Signers = addr2Str(signers)
	inspect.Threshold = threshold

	pending := make(map[int64]multisig.Transaction)
	if err := mstate.ForEachPendingTxn(func(id int64, txn multisig.Transaction) error {
		pending[id] = txn
		return nil
	}); err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	var transactions []client.MsigTransaction
	if len(pending) > 0 {
		var txids []int64
		for txid := range pending {
			txids = append(txids, txid)
		}
		sort.Slice(txids, func(i, j int) bool {
			return txids[i] < txids[j]
		})

		for _, txid := range txids {
			tx := pending[txid]

			targAct, err := w.Api.StateGetActor(ctx, tx.To, types.EmptyTSK)
			paramStr := fmt.Sprintf("%x", tx.Params)

			if err != nil {
				if tx.Method == 0 {
					transactions = append(transactions, client.MsigTransaction{
						Txid:     txid,
						To:       tx.To.String(),
						Value:    tx.Value.String(),
						Method:   fmt.Sprintf("Send(%d)", tx.Method),
						Params:   "",
						Approved: addr2Str(tx.Approved),
					})
				} else {
					transactions = append(transactions, client.MsigTransaction{
						Txid:     txid,
						To:       tx.To.String(),
						Value:    tx.Value.String(),
						Method:   fmt.Sprintf("%d", tx.Method),
						Params:   paramStr,
						Approved: addr2Str(tx.Approved),
					})
				}
			} else {
				method := filcns.NewActorRegistry().Methods[targAct.Code][tx.Method]

				if tx.Method != 0 {
					ptyp := reflect.New(method.Params.Elem()).Interface().(cbg.CBORUnmarshaler)
					if err := ptyp.UnmarshalCBOR(bytes.NewReader(tx.Params)); err != nil {
						ReturnError(c, NewError(500, fmt.Errorf("failed to decode parameters of transaction %d: %w", txid, err).Error()))
						return
					}

					b, err := json.Marshal(ptyp)
					if err != nil {
						ReturnError(c, NewError(500, fmt.Errorf("could not json marshal parameter type: %w", err).Error()))
						return
					}

					paramStr = string(b)
				}
				transactions = append(transactions, client.MsigTransaction{
					Txid:     txid,
					To:       tx.To.String(),
					Value:    tx.Value.String(),
					Method:   fmt.Sprintf("%d", tx.Method),
					Params:   paramStr,
					Approved: addr2Str(tx.Approved),
				})
			}
		}
	}

	inspect.Transactions = transactions
	ReturnOk(c, inspect)
}

// MsigApprove Post
func (w *Wallet) MsigApprove(c *gin.Context) {
	param := client.MsigBaseRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigApproveMessage(param.BaseParams, param.MsigAddress, param.TxId, param.From)
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

// MsigCancel Post
func (w *Wallet) MsigCancel(c *gin.Context) {
	param := client.MsigBaseRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigCancelMessage(param.BaseParams, param.MsigAddress, param.TxId, param.From)
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

// MsigTransferPropose Post
func (w *Wallet) MsigTransferPropose(c *gin.Context) {
	param := client.MsigTransferProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigTransferProposeMessage(param.BaseParams, param.MsigAddress, param.DestinationAddress, param.Amount, param.From)

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

// MsigTransferApprove Post
func (w *Wallet) MsigTransferApprove(c *gin.Context) {
	param := client.MsigBaseRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigTransferApproveMessage(param.BaseParams, param.MsigAddress, param.TxId, param.From)
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

// MsigTransferCancel Post
func (w *Wallet) MsigTransferCancel(c *gin.Context) {
	param := client.MsigBaseRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigTransferCancelMessage(param.BaseParams, param.MsigAddress, param.TxId, param.From)
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

// MsigAddPropose Post
func (w *Wallet) MsigAddPropose(c *gin.Context) {
	param := client.MsigAddSignerProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigAddSignerProposeMessage(param.BaseParams, param.MsigAddress, param.SignerAddress, param.IncreaseThreshold, param.From)
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

// MsigAddApprove Post
func (w *Wallet) MsigAddApprove(c *gin.Context) {
	param := client.MsigAddSignerApprovRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigAddSignerApproveMessage(param.BaseParams, param.MsigAddress, param.ProposerAddress, param.TxId, param.SignerAddress, param.IncreaseThreshold, param.From)
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

// MsigAddCancel Post
func (w *Wallet) MsigAddCancel(c *gin.Context) {
	param := client.MsigAddSignerCancelRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigAddSignerCancelMessage(param.BaseParams, param.MsigAddress, param.TxId, param.SignerAddress, param.IncreaseThreshold, param.From)
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

// todo remove signer

// MsigSwapPropose Post
func (w *Wallet) MsigSwapPropose(c *gin.Context) {
	param := client.MsigSwapProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigSwapProposeMessage(param.BaseParams, param.MsigAddress, param.OldAddress, param.NewAddress, param.From)
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

// MsigSwapApprove Post
func (w *Wallet) MsigSwapApprove(c *gin.Context) {
	param := client.MsigSwapApproveRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigSwapApproveMessage(param.BaseParams, param.MsigAddress, param.ProposerAddress, param.TxId, param.OldAddress, param.NewAddress, param.From)
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

// MsigSwapCancel Post
func (w *Wallet) MsigSwapCancel(c *gin.Context) {
	param := client.MsigSwapCancelRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigSwapCancelMessage(param.BaseParams, param.MsigAddress, param.TxId, param.OldAddress, param.NewAddress, param.From)
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

// MsigLockPropose Post
func (w *Wallet) MsigLockPropose(c *gin.Context) {
	param := client.MsigLockProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigLockProposeMessage(param.BaseParams, param.MsigAddress, param.StartEpoch, param.UnlockDuration, param.Amount, param.From)
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

// MsigLockApprove Post
func (w *Wallet) MsigLockApprove(c *gin.Context) {
	param := client.MsigLockApproveRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigLockApproveMessage(param.BaseParams, param.MsigAddress, param.ProposerAddress, param.TxId, param.StartEpoch, param.UnlockDuration, param.Amount, param.From)
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

// MsigLockCancel Post
func (w *Wallet) MsigLockCancel(c *gin.Context) {
	param := client.MsigLockCancelRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigLockCancelMessage(param.BaseParams, param.MsigAddress, param.TxId, param.StartEpoch, param.UnlockDuration, param.Amount, param.From)
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

// MsigThresholdPropose Post
func (w *Wallet) MsigThresholdPropose(c *gin.Context) {
	param := client.MsigThresholdProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigThresholdProposeMessage(param.BaseParams, param.MsigAddress, param.NewThreshold, param.From)
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

// MsigThresholdApprove Post
func (w *Wallet) MsigThresholdApprove(c *gin.Context) {
	param := client.MsigThresholdApproveRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigThresholdApproveMessage(param.BaseParams, param.MsigAddress, param.ProposerAddress, param.TxId, param.NewThreshold, param.From)
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

// MsigThresholdCancel Post
func (w *Wallet) MsigThresholdCancel(c *gin.Context) {
	param := client.MsigThresholdCancelRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigThresholdCancelMessage(param.BaseParams, param.MsigAddress, param.TxId, param.NewThreshold, param.From)
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

// MsigChangeOwnerPropose Post
func (w *Wallet) MsigChangeOwnerPropose(c *gin.Context) {
	param := client.MsigChangeOwnerProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigChangeOwnerProposeMessage(param.BaseParams, param.MsigAddress, param.MinerId, param.NewOwner, param.From)
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

// MsigChangeOwnerApprove Post
func (w *Wallet) MsigChangeOwnerApprove(c *gin.Context) {
	param := client.MsigChangeOwnerApproveRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigChangeOwnerApproveMessage(param.BaseParams, param.MsigAddress, param.ProposerAddress, param.TxId, param.MinerId, param.NewOwner, param.From)
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

// MsigWithdrawPropose Post
func (w *Wallet) MsigWithdrawPropose(c *gin.Context) {
	param := client.MsigWithdrawProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigWithdrawProposeMessage(param.BaseParams, param.MsigAddress, param.MinerId, param.Amount, param.From)
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

// MsigWithdrawApprove Post
func (w *Wallet) MsigWithdrawApprove(c *gin.Context) {
	param := client.MsigWithdrawApproveRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigWithdrawApproveMessage(param.BaseParams, param.MsigAddress, param.ProposerAddress, param.TxId, param.MinerId, param.Amount, param.From)
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

// MsigChangeWorkerPropose Post
func (w *Wallet) MsigChangeWorkerPropose(c *gin.Context) {
	param := client.MsigChangeWorkerProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigChangeWorkerProposeMessage(param.BaseParams, param.MsigAddress, param.MinerId, param.NewWorker, param.From)
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

// MsigChangeWorkerApprove Post
func (w *Wallet) MsigChangeWorkerApprove(c *gin.Context) {
	param := client.MsigChangeWorkerApproveRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigChangeWorkerApproveMessage(param.BaseParams, param.MsigAddress, param.ProposerAddress, param.TxId, param.MinerId, param.NewWorker, param.From)
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

// MsigConfirmChangeWorkerPropose Post
func (w *Wallet) MsigConfirmChangeWorkerPropose(c *gin.Context) {
	param := client.MsigChangeWorkerProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigConfirmChangeWorkerProposeMessage(param.BaseParams, param.MsigAddress, param.MinerId, param.NewWorker, param.From)
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

// MsigConfirmChangeWorkerApprove Post
func (w *Wallet) MsigConfirmChangeWorkerApprove(c *gin.Context) {
	param := client.MsigChangeWorkerApproveRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigConfirmChangeWorkerApproveMessage(param.BaseParams, param.MsigAddress, param.ProposerAddress, param.TxId, param.MinerId, param.NewWorker, param.From)
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

// MsigSetControlPropose Post
func (w *Wallet) MsigSetControlPropose(c *gin.Context) {
	param := client.MsigSetControlProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigSetControlProposeMessage(param.BaseParams, param.MsigAddress, param.MinerId, param.From, param.ControlAddrs...)
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

// MsigSetControlApprove Post
func (w *Wallet) MsigSetControlApprove(c *gin.Context) {
	param := client.MsigSetControlApproveRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigSetControlApproveMessage(param.BaseParams, param.MsigAddress, param.ProposerAddress, param.TxId, param.MinerId, param.From, param.ControlAddrs...)
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

// MsigChangeBeneficiaryPropose Post
func (w *Wallet) MsigChangeBeneficiaryPropose(c *gin.Context) {
	param := client.MsigChangeBeneficiaryProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigChangeBeneficiaryProposeMessage(param.BaseParams, param.MsigAddress, param.MinerId, param.From, param.BeneficiaryAddress, param.Quota, param.Expiration, param.OverwritePendingChange)
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

// MsigChangeBeneficiaryApprove Post
func (w *Wallet) MsigChangeBeneficiaryApprove(c *gin.Context) {
	param := client.MsigChangeBeneficiaryApproveRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigChangeBeneficiaryApproveMessage(param.BaseParams, param.MsigAddress, param.ProposerAddress, param.TxId, param.MinerId, param.From, param.BeneficiaryAddress, param.Quota, param.Expiration)
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

// MsigConfirmChangeBeneficiaryPropose Post
func (w *Wallet) MsigConfirmChangeBeneficiaryPropose(c *gin.Context) {
	param := client.MsigConfirmChangeBeneficiaryProposeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigConfirmChangeBeneficiaryProposeMessage(param.BaseParams, param.MsigAddress, param.MinerId, param.From)
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

// MsigConfirmChangeBeneficiaryApprove Post
func (w *Wallet) MsigConfirmChangeBeneficiaryApprove(c *gin.Context) {
	param := client.MsigConfirmChangeBeneficiaryApproveRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	msig := buildmessage.NewMsiger(w.Api)
	msg, msgParams, err := msig.NewMsigConfirmChangeBeneficiaryApproveMessage(param.BaseParams, param.MsigAddress, param.ProposerAddress, param.TxId, param.MinerId, param.From)
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
