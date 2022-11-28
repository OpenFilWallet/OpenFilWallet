package buildmessage

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/v8/actors/builtin"
	miner8 "github.com/filecoin-project/specs-actors/v8/actors/builtin/miner"
	"golang.org/x/xerrors"
)

type BaseParams struct {
	MaxFee     int64
	GasFeeCap  string
	GasPremium string
	GasLimit   int64
	Nonce      uint64
}

func NewTransferMessage(node api.FullNode, baseParams BaseParams, from, to string, amount string) (*types.Message, error) {
	fromAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, err
	}

	toAddr, err := address.NewFromString(to)
	if err != nil {
		return nil, err
	}

	value, err := types.ParseFIL(amount)
	if err != nil {
		return nil, err
	}

	msg := &types.Message{
		To:     toAddr,
		From:   fromAddr,
		Value:  abi.TokenAmount(value),
		Method: builtin.MethodSend,
	}

	msg, err = buildMessage(node, msg, baseParams)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func NewWithdrawMessage(node api.FullNode, baseParams BaseParams, minerId string, amount string) (*types.Message, *miner8.WithdrawBalanceParams, error) {
	minerAddr, err := address.NewFromString(minerId)
	if err != nil {
		return nil, nil, err
	}

	value, err := types.ParseFIL(amount)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	mi, err := node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	available, err := node.StateMinerAvailableBalance(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	if value.Int64() == 0 {
		value = types.FIL(available)
	}

	params := &miner8.WithdrawBalanceParams{
		AmountRequested: abi.TokenAmount(value),
	}

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, err
	}

	msg := &types.Message{
		To:     minerAddr,
		From:   mi.Owner,
		Value:  types.NewInt(0),
		Method: builtin.MethodsMiner.WithdrawBalance,
		Params: sp,
	}

	msg, err = buildMessage(node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, params, nil
}

func NewChangeOwnerMessage(node api.FullNode, baseParams BaseParams, minerId, newOwner, sender string) (*types.Message, *address.Address, error) {
	minerAddr, err := address.NewFromString(minerId)
	if err != nil {
		return nil, nil, err
	}

	na, err := address.NewFromString(newOwner)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	newAddrId, err := node.StateLookupID(ctx, na, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	fa, err := address.NewFromString(sender)
	if err != nil {
		return nil, nil, err
	}

	fromAddrId, err := node.StateLookupID(ctx, fa, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	mi, err := node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	if fromAddrId != mi.Owner && fromAddrId != newAddrId {
		return nil, nil, xerrors.New("from address must either be the old owner or the new owner")
	}

	params := &newAddrId

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, xerrors.Errorf("serializing params: %w", err)
	}

	msg := &types.Message{
		From:   fromAddrId,
		To:     minerAddr,
		Method: builtin.MethodsMiner.ChangeOwnerAddress,
		Value:  big.Zero(),
		Params: sp,
	}

	msg, err = buildMessage(node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, params, nil
}

// NewChangeWorkerMessage ChangeWorker and Change Control
func NewChangeWorkerMessage(node api.FullNode, baseParams BaseParams, minerId string, worker string, controlAddrs ...string) (*types.Message, *miner8.ChangeWorkerAddressParams, error) {
	minerAddr, err := address.NewFromString(minerId)
	if err != nil {
		return nil, nil, err
	}

	workerAddr, err := address.NewFromString(worker)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	mi, err := node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	var newControlAddrs []address.Address
	for i, as := range controlAddrs {
		a, err := address.NewFromString(as)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing address %d: %w", i, err)
		}

		ka, err := node.StateAccountKey(ctx, a, types.EmptyTSK)
		if err != nil {
			return nil, nil, err
		}

		// make sure the address exists on chain
		_, err = node.StateLookupID(ctx, ka, types.EmptyTSK)
		if err != nil {
			return nil, nil, fmt.Errorf("looking up %s: %w", ka, err)
		}

		newControlAddrs = append(newControlAddrs, ka)
	}

	if len(newControlAddrs) == 0 {
		newControlAddrs = mi.ControlAddresses
	}

	newWorker := mi.Worker
	if workerAddr != address.Undef {
		newAddr, err := node.StateLookupID(ctx, workerAddr, types.EmptyTSK)
		if err != nil {
			return nil, nil, err
		}

		if mi.NewWorker.Empty() {
			if mi.Worker == newAddr {
				return nil, nil, fmt.Errorf("worker address already set to %s", workerAddr)
			}
		} else {
			if mi.NewWorker == newAddr {
				return nil, nil, fmt.Errorf("change to worker address %s already pending", workerAddr)
			}
		}

		newWorker = newAddr
	}

	params := &miner8.ChangeWorkerAddressParams{
		NewWorker:       newWorker,
		NewControlAddrs: newControlAddrs,
	}

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, fmt.Errorf("serializing params: %w", err)
	}

	msg := &types.Message{
		From:   mi.Owner,
		To:     minerAddr,
		Method: builtin.MethodsMiner.ChangeWorkerAddress,
		Value:  big.Zero(),
		Params: sp,
	}

	msg, err = buildMessage(node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, params, nil
}

func NewConfirmUpdateWorkerMessage(node api.FullNode, baseParams BaseParams, minerId string, worker string) (*types.Message, error) {
	minerAddr, err := address.NewFromString(minerId)
	if err != nil {
		return nil, err
	}

	workerAddr, err := address.NewFromString(worker)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	mi, err := node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}

	newAddr, err := node.StateLookupID(ctx, workerAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}

	if mi.NewWorker.Empty() {
		return nil, fmt.Errorf("no worker key change proposed")
	} else if mi.NewWorker != newAddr {
		return nil, fmt.Errorf("worker key %s does not match current worker key proposal %s", newAddr, mi.NewWorker)
	}

	if head, err := node.ChainHead(ctx); err != nil {
		return nil, fmt.Errorf("failed to get the chain head: %w", err)
	} else if head.Height() < mi.WorkerChangeEpoch {
		return nil, fmt.Errorf("worker key change cannot be confirmed until %d, current height is %d", mi.WorkerChangeEpoch, head.Height())
	}

	msg := &types.Message{
		From:   mi.Owner,
		To:     minerAddr,
		Method: builtin.MethodsMiner.ConfirmUpdateWorkerKey,
		Value:  big.Zero(),
	}

	msg, err = buildMessage(node, msg, baseParams)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func buildMessage(node api.FullNode, msg *types.Message, baseParams BaseParams) (*types.Message, error) {
	msg.GasFeeCap, _ = types.BigFromString(baseParams.GasFeeCap)
	msg.GasPremium, _ = types.BigFromString(baseParams.GasPremium)
	msg.GasLimit = baseParams.GasLimit
	msg.Nonce = baseParams.Nonce

	ctx := context.Background()
	if msg.GasLimit == 0 || msg.GasPremium == types.EmptyInt || types.BigCmp(msg.GasPremium, types.NewInt(0)) == 0 ||
		msg.GasFeeCap == types.EmptyInt || types.BigCmp(msg.GasFeeCap, types.NewInt(0)) == 0 {
		var err error
		msg, err = node.GasEstimateMessageGas(ctx, msg, &api.MessageSendSpec{MaxFee: abi.NewTokenAmount(baseParams.MaxFee)}, types.EmptyTSK)
		if err != nil {
			return nil, err
		}
	}

	if msg.Nonce == 0 {
		mpoolNonce, err := node.MpoolGetNonce(ctx, msg.From)
		if err != nil {
			return nil, err
		}
		msg.Nonce = mpoolNonce
	}

	return msg, nil
}
