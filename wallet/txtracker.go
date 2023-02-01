package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	multisig9 "github.com/filecoin-project/go-state-types/builtin/v9/multisig"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/stmgr"
	"github.com/filecoin-project/lotus/chain/types"
	init2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/init"
	"github.com/ipfs/go-cid"
	"time"
)

type txTracker struct {
	node       *node
	db         datastore.WalletDB
	txReceiver chan *datastore.History
	close      <-chan struct{}
}

func newTxTracker(node *node, db datastore.WalletDB, close <-chan struct{}) *txTracker {
	txTracker := &txTracker{
		node:       node,
		db:         db,
		txReceiver: make(chan *datastore.History, 50),
		close:      close,
	}

	go txTracker.txMonitor()

	return txTracker
}

func (tt *txTracker) trackTx(msg *datastore.History) {
	tt.txReceiver <- msg
}

func (tt *txTracker) txMonitor() {
	for {
		select {
		case tx := <-tt.txReceiver:
			go tt.monitor(tx)
		case <-tt.close:
			return
		}
	}
}

func (tt *txTracker) monitor(msg *datastore.History) {
	for {
		time.Sleep(1 * time.Minute)

		if tt.node == nil {
			continue
		}

		recordFailedTx := func(err error) {
			msg.TxState = datastore.Failed
			msg.Detail = err.Error()
			tt.recordTx(msg)
		}

		recordSuccessTx := func() {
			msg.TxState = datastore.Success
			tt.recordTx(msg)
		}

		msgCid, _ := cid.Parse(msg.TxCid)
		searchRes, err := tt.node.Api.StateSearchMsg(context.Background(), types.EmptyTSK, msgCid, stmgr.LookbackNoLimit, true)
		if err != nil {
			// For some public node services, StateSearchMsg request parameters are optimized: only one msg cid parameter is required
			r, err := client.LotusStateSearchMsg(tt.node.nodeEndpoint, tt.node.token, msg.TxCid)
			if err != nil {
				recordFailedTx(err)
				return
			}

			if r == nil {
				continue
			}

			searchRes = &api.MsgLookup{
				Message:   r.Message,
				Receipt:   r.Receipt,
				ReturnDec: r.ReturnDec,
				TipSet:    r.TipSet,
				Height:    r.Height,
			}
		}

		if searchRes != nil {
			if searchRes.Receipt.ExitCode.IsError() {
				recordFailedTx(err)
				return
			}

			if msg.ParamName == "ConstructorParams" { // create msig tx
				var execreturn init2.ExecReturn
				if err := execreturn.UnmarshalCBOR(bytes.NewReader(searchRes.Receipt.Return)); err != nil {
					recordFailedTx(err)
					return
				}

				msig := execreturn.RobustAddress.String()
				var p multisig9.ConstructorParams
				err = json.Unmarshal([]byte(msg.Params), &p)
				if err != nil {
					log.Warnw("txTracker: Unmarshal Msig ConstructorParams fail", "err", err.Error())
					recordFailedTx(err)
					return
				}

				signers := make([]string, 0)
				for _, signer := range p.Signers {
					signers = append(signers, signer.String())
				}
				if err = tt.addMsig(&datastore.MsigWallet{
					MsigAddr:              msig,
					Signers:               signers,
					NumApprovalsThreshold: p.NumApprovalsThreshold,
					UnlockDuration:        int64(p.UnlockDuration),
					StartEpoch:            int64(p.StartEpoch),
				}); err != nil {
					log.Warnw("txTracker: addMsig to db fail", "err", err.Error())
					recordFailedTx(err)
					return
				}
			}

			recordSuccessTx()
			return
		}
	}
}

func (tt *txTracker) recordTx(msg *datastore.History) {
	err := tt.db.SetHistory(msg)
	if err != nil {
		log.Warnw("RecordTx fail", "msg", fmt.Sprintf("From: %s To: %s Method: %d", msg.From, msg.To, msg.Method), "err", err)
	}
}

func (tt *txTracker) addMsig(msig *datastore.MsigWallet) error {
	return tt.db.SetMsig(msig)
}
