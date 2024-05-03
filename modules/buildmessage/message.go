package buildmessage

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/builtin"
	miner13 "github.com/filecoin-project/go-state-types/builtin/v13/miner"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/types"
	specsminer8 "github.com/filecoin-project/specs-actors/v8/actors/builtin/miner"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"
	"strconv"
	"strings"
)

var log = logging.Logger("buildmessage")

type BaseParams struct {
	MaxFee     string `json:"max_fee"`
	GasFeeCap  string `json:"gas_feecap"`
	GasPremium string `json:"gas_premium"`
	GasLimit   int64  `json:"gas_limit"`
	Nonce      uint64 `json:"nonce"`
}

func (b BaseParams) String() string {
	return fmt.Sprintf("max_fee: %s, gas_feecap: %s, gas_premium: %s, gas_limit: %d, nonce: %d", b.MaxFee, b.GasFeeCap, b.GasPremium, b.GasLimit, b.Nonce)
}

func LotusMessageToString(msg *types.Message) string {
	return fmt.Sprintf("version: %d, to: %s, from: %s, nonce: %d, value: %s, gasLimit: %d, gasFeeCap: %s, gasPremium: %s, method: %d, params: %s", msg.Version, msg.To.String(), msg.From.String(), msg.Nonce, big.Int(msg.Value).String(), msg.GasLimit, big.Int(msg.GasFeeCap).String(), big.Int(msg.GasPremium).String(), msg.Method, hex.EncodeToString(msg.Params))
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

func NewWithdrawMessage(node api.FullNode, baseParams BaseParams, minerId string, amount string) (*types.Message, *specsminer8.WithdrawBalanceParams, error) {
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

	params := &specsminer8.WithdrawBalanceParams{
		AmountRequested: abi.TokenAmount(value),
	}

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, err
	}

	// todo Whether to use beneficiary withdrawal?
	from, err := node.StateAccountKey(ctx, mi.Beneficiary, types.EmptyTSK)
	if err != nil {
		if strings.Contains(err.Error(), "multisig") {
			err = xerrors.Errorf("minerId: %s owner: %s is multisig account, please use Msig Withdraw", minerId, mi.Owner.String())
		}

		return nil, nil, err
	}

	msg := &types.Message{
		To:     minerAddr,
		From:   from,
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

	mi, err := node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	var fromAddrId address.Address
	if sender == "" {
		fromAddrId = mi.Owner
	} else {
		fa, err := address.NewFromString(sender)
		if err != nil {
			return nil, nil, err
		}

		fromAddrId, err = node.StateLookupID(ctx, fa, types.EmptyTSK)
		if err != nil {
			return nil, nil, err
		}
	}

	if fromAddrId != mi.Owner && fromAddrId != newAddrId {
		return nil, nil, xerrors.New("from address must either be the old owner or the new owner")
	}

	params := &newAddrId

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, xerrors.Errorf("serializing params: %w", err)
	}

	from, err := node.StateAccountKey(ctx, fromAddrId, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	msg := &types.Message{
		From:   from,
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
func NewChangeWorkerMessage(node api.FullNode, baseParams BaseParams, minerId string, worker string, controlAddrs ...string) (*types.Message, *specsminer8.ChangeWorkerAddressParams, error) {
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

	params := &specsminer8.ChangeWorkerAddressParams{
		NewWorker:       newWorker,
		NewControlAddrs: newControlAddrs,
	}

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, fmt.Errorf("serializing params: %w", err)
	}

	from, err := node.StateAccountKey(ctx, mi.Owner, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	msg := &types.Message{
		From:   from,
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

	from, err := node.StateAccountKey(ctx, mi.Owner, types.EmptyTSK)
	if err != nil {
		return nil, err
	}

	msg := &types.Message{
		From:   from,
		To:     minerAddr,
		Method: builtin.MethodsMiner.ConfirmChangeWorkerAddress,
		Value:  big.Zero(),
	}

	msg, err = buildMessage(node, msg, baseParams)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func NewChangeBeneficiaryProposeMessage(node api.FullNode, baseParams BaseParams, minerId string, beneficiaryAddress, quota, expiration string, overwritePendingChange bool) (*types.Message, *miner13.ChangeBeneficiaryParams, error) {
	ctx := context.Background()

	na, err := address.NewFromString(beneficiaryAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing beneficiary address: %w", err)
	}

	newAddr, err := node.StateLookupID(ctx, na, types.EmptyTSK)
	if err != nil {
		return nil, nil, fmt.Errorf("looking up new beneficiary address: %w", err)
	}

	quotaParam, err := types.ParseFIL(quota)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing quota: %w", err)
	}

	expirationParam, err := strconv.ParseInt(expiration, 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing expiration: %w", err)
	}

	minerAddr, err := address.NewFromString(minerId)
	if err != nil {
		return nil, nil, err
	}

	mi, err := node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, xerrors.Errorf("getting miner info: %w", err)
	}

	if mi.Beneficiary == mi.Owner && newAddr == mi.Owner {
		return nil, nil, fmt.Errorf("beneficiary %s already set to owner address", mi.Beneficiary)
	}

	if mi.PendingBeneficiaryTerm != nil && !overwritePendingChange {
		return nil, nil, fmt.Errorf("WARNING: replacing Pending Beneficiary Term of: Beneficiary: %s, Quota: %s, Expiration Epoch:%d", mi.PendingBeneficiaryTerm.NewBeneficiary.String(), mi.PendingBeneficiaryTerm.NewQuota.String(), mi.PendingBeneficiaryTerm.NewExpiration)
	}

	params := &miner13.ChangeBeneficiaryParams{
		NewBeneficiary: newAddr,
		NewQuota:       abi.TokenAmount(quotaParam),
		NewExpiration:  abi.ChainEpoch(expirationParam),
	}

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, xerrors.Errorf("serializing params: %w", err)
	}

	from, err := node.StateAccountKey(ctx, mi.Owner, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}
	msg := &types.Message{
		From:   from,
		To:     minerAddr,
		Method: builtin.MethodsMiner.ChangeBeneficiary,
		Value:  big.Zero(),
		Params: sp,
	}

	msg, err = buildMessage(node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, params, nil
}

func NewConfirmChangeBeneficiary(node api.FullNode, baseParams BaseParams, minerId string) (*types.Message, *miner13.ChangeBeneficiaryParams, error) {
	ctx := context.Background()

	minerAddr, err := address.NewFromString(minerId)
	if err != nil {
		return nil, nil, err
	}

	mi, err := node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, xerrors.Errorf("getting miner info: %w", err)
	}

	if mi.PendingBeneficiaryTerm == nil {
		return nil, nil, fmt.Errorf("no pending beneficiary term found for miner %s", minerAddr)
	}

	var fromAddr address.Address
	if mi.Owner.String() == mi.Beneficiary.String() || mi.PendingBeneficiaryTerm != nil {
		fromAddr = mi.PendingBeneficiaryTerm.NewBeneficiary
	} else {
		fromAddr = mi.Beneficiary
	}

	params := &miner13.ChangeBeneficiaryParams{
		NewBeneficiary: mi.PendingBeneficiaryTerm.NewBeneficiary,
		NewQuota:       mi.PendingBeneficiaryTerm.NewQuota,
		NewExpiration:  mi.PendingBeneficiaryTerm.NewExpiration,
	}

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, xerrors.Errorf("serializing params: %w", err)
	}

	from, err := node.StateAccountKey(ctx, fromAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	msg := &types.Message{
		From:   from,
		To:     minerAddr,
		Method: builtin.MethodsMiner.ChangeBeneficiary,
		Value:  big.Zero(),
		Params: sp,
	}

	msg, err = buildMessage(node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, params, nil
}

func buildMessage(node api.FullNode, msg *types.Message, baseParams BaseParams) (*types.Message, error) {
	log.Debugw("buildMessage: start", "baseParams", baseParams.String())

	msg.GasFeeCap, _ = types.BigFromString(baseParams.GasFeeCap)
	msg.GasPremium, _ = types.BigFromString(baseParams.GasPremium)
	msg.GasLimit = baseParams.GasLimit
	msg.Nonce = baseParams.Nonce

	ctx := context.Background()
	if msg.GasLimit == 0 || msg.GasPremium == types.EmptyInt || types.BigCmp(msg.GasPremium, types.NewInt(0)) == 0 ||
		msg.GasFeeCap == types.EmptyInt || types.BigCmp(msg.GasFeeCap, types.NewInt(0)) == 0 {

		maxFee, err := types.ParseFIL(baseParams.MaxFee)
		if err != nil {
			log.Warnf("parsing max-fee: %s", err)
			maxFee, _ = types.ParseFIL("1 FIL")
		}

		msg, err = node.GasEstimateMessageGas(ctx, msg, &api.MessageSendSpec{MaxFee: abi.TokenAmount(maxFee)}, types.EmptyTSK)
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

	log.Debugw("buildMessage: end", "msg", LotusMessageToString(msg))
	return msg, nil
}
