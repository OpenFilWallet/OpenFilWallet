package buildmessage

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	actorstypes "github.com/filecoin-project/go-state-types/actors"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/go-state-types/builtin/v9/miner"
	multisig9 "github.com/filecoin-project/go-state-types/builtin/v9/multisig"
	"github.com/filecoin-project/go-state-types/network"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors"
	lotusbuiltin "github.com/filecoin-project/lotus/chain/actors/builtin"
	"github.com/filecoin-project/lotus/chain/actors/builtin/multisig"
	"github.com/filecoin-project/lotus/chain/types"
	miner8 "github.com/filecoin-project/specs-actors/v8/actors/builtin/miner"
	"github.com/minio/blake2b-simd"
	"golang.org/x/xerrors"
	"strconv"
)

type Msiger struct {
	node api.FullNode
}

func NewMsiger(node api.FullNode) *Msiger {
	return &Msiger{
		node: node,
	}
}

func (m *Msiger) NewMsigCreateMessage(baseParams BaseParams, required, duration uint64, value, from string, signers ...string) (*types.Message, *multisig9.ConstructorParams, error) {
	var signerAddrs []address.Address
	for _, signer := range signers {
		signerAddr, err := address.NewFromString(signer)
		if err != nil {
			return nil, nil, err
		}
		signerAddrs = append(signerAddrs, signerAddr)
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	filval, err := types.ParseFIL(value)
	if err != nil {
		return nil, nil, err
	}

	intVal := types.BigInt(filval)

	if required == 0 {
		required = uint64(len(signers))
	}

	msg, params, err := m.MsigCreate(required, signerAddrs, abi.ChainEpoch(duration), intVal, sendAddr)
	if err != nil {
		return nil, nil, err
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, params, nil
}

func (m *Msiger) NewMsigApproveMessage(baseParams BaseParams, msigAddress, txId string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	act, err := m.node.StateGetActor(ctx, msig, types.EmptyTSK)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to look up multisig %s: %w", msig, err)
	}

	if !lotusbuiltin.IsMultisigActor(act.Code) {
		return nil, nil, fmt.Errorf("actor %s is not a multisig actor", msig)
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, approveParams, err := m.MsigApprove(msig, txid, sendAddr)
	if err != nil {
		return nil, nil, err
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigCancelMessage(baseParams BaseParams, msigAddress, txId string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	act, err := m.node.StateGetActor(ctx, msig, types.EmptyTSK)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to look up multisig %s: %w", msig, err)
	}

	if !lotusbuiltin.IsMultisigActor(act.Code) {
		return nil, nil, fmt.Errorf("actor %s is not a multisig actor", msig)
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, cancelParams, err := m.MsigCancel(msig, txid, sendAddr)
	if err != nil {
		return nil, nil, err
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, cancelParams, nil
}

func (m *Msiger) NewMsigTransferProposeMessage(baseParams BaseParams, msigAddress, destinationAddress, amount string, from string) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	dest, err := address.NewFromString(destinationAddress)
	if err != nil {
		return nil, nil, err
	}

	value, err := types.ParseFIL(amount)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, proposeParams, err := m.MsigPropose(msig, dest, types.BigInt(value), sendAddr, uint64(builtin.MethodSend), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigTransferApproveMessage(baseParams BaseParams, msigAddress, txId string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, approveParams, err := m.MsigApprove(msig, txid, sendAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigApprove: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigTransferCancelMessage(baseParams BaseParams, msigAddress, txId string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, cancelParams, err := m.MsigCancel(msig, txid, sendAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigCancel: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, cancelParams, nil
}

func (m *Msiger) NewMsigAddSignerProposeMessage(baseParams BaseParams, msigAddress, signerAddress string, increaseThreshold bool, from string) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	signer, err := address.NewFromString(signerAddress)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, proposeParams, err := m.MsigAddPropose(msig, sendAddr, signer, increaseThreshold)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigAddPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigAddSignerApproveMessage(baseParams BaseParams, msigAddress, proposerAddress, txId, signerAddress string, increaseThreshold bool, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	prop, err := address.NewFromString(proposerAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	signer, err := address.NewFromString(signerAddress)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, approveParams, err := m.MsigAddApprove(msig, sendAddr, txid, prop, signer, increaseThreshold)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigAddApprove: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigAddSignerCancelMessage(baseParams BaseParams, msigAddress, txId, signerAddress string, increaseThreshold bool, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	signer, err := address.NewFromString(signerAddress)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, cancelParams, err := m.MsigAddCancel(msig, sendAddr, txid, signer, increaseThreshold)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigAddCancel: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, cancelParams, nil
}

func (m *Msiger) NewMsigSwapProposeMessage(baseParams BaseParams, msigAddress, oldAddress, newAddress string, from string) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	oldAddr, err := address.NewFromString(oldAddress)
	if err != nil {
		return nil, nil, err
	}

	newAddr, err := address.NewFromString(newAddress)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, proposeParams, err := m.MsigSwapPropose(msig, sendAddr, oldAddr, newAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigSwapPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigSwapApproveMessage(baseParams BaseParams, msigAddress, proposerAddress, txId, oldAddress, newAddress string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	prop, err := address.NewFromString(proposerAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	oldAddr, err := address.NewFromString(oldAddress)
	if err != nil {
		return nil, nil, err
	}

	newAddr, err := address.NewFromString(newAddress)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, approveParams, err := m.MsigSwapApprove(msig, sendAddr, txid, prop, oldAddr, newAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigSwapApprove: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigSwapCancelMessage(baseParams BaseParams, msigAddress, txId, oldAddress, newAddress string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	oldAddr, err := address.NewFromString(oldAddress)
	if err != nil {
		return nil, nil, err
	}

	newAddr, err := address.NewFromString(newAddress)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, cancelParams, err := m.MsigSwapCancel(msig, sendAddr, txid, oldAddr, newAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigSwapCancel: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, cancelParams, nil
}

func (m *Msiger) NewMsigLockProposeMessage(baseParams BaseParams, msigAddress, startEpoch, unlockDuration, amount string, from string) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	start, err := strconv.ParseUint(startEpoch, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	duration, err := strconv.ParseUint(unlockDuration, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	value, err := types.ParseFIL(amount)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	params, actErr := actors.SerializeParams(&multisig9.LockBalanceParams{
		StartEpoch:     abi.ChainEpoch(start),
		UnlockDuration: abi.ChainEpoch(duration),
		Amount:         big.Int(value),
	})

	if actErr != nil {
		return nil, nil, actErr
	}

	msg, proposeParams, err := m.MsigPropose(msig, msig, big.Zero(), sendAddr, uint64(multisig.Methods.LockBalance), params)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigLockApproveMessage(baseParams BaseParams, msigAddress, proposerAddress, txId, startEpoch, unlockDuration, amount string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	prop, err := address.NewFromString(proposerAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	start, err := strconv.ParseUint(startEpoch, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	duration, err := strconv.ParseUint(unlockDuration, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	value, err := types.ParseFIL(amount)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	params, actErr := actors.SerializeParams(&multisig9.LockBalanceParams{
		StartEpoch:     abi.ChainEpoch(start),
		UnlockDuration: abi.ChainEpoch(duration),
		Amount:         big.Int(value),
	})

	if actErr != nil {
		return nil, nil, actErr
	}

	msg, approveParams, err := m.MsigApproveTxnHash(msig, txid, prop, msig, big.Zero(), sendAddr, uint64(multisig.Methods.LockBalance), params)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigApproveTxnHash: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigLockCancelMessage(baseParams BaseParams, msigAddress, txId, startEpoch, unlockDuration, amount string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	start, err := strconv.ParseUint(startEpoch, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	duration, err := strconv.ParseUint(unlockDuration, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	value, err := types.ParseFIL(amount)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	params, actErr := actors.SerializeParams(&multisig9.LockBalanceParams{
		StartEpoch:     abi.ChainEpoch(start),
		UnlockDuration: abi.ChainEpoch(duration),
		Amount:         big.Int(value),
	})

	if actErr != nil {
		return nil, nil, actErr
	}

	msg, cancelParams, err := m.MsigCancelTxnHash(msig, txid, msig, big.Zero(), sendAddr, uint64(multisig.Methods.LockBalance), params)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigCancelTxnHash: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, cancelParams, nil
}

func (m *Msiger) NewMsigThresholdProposeMessage(baseParams BaseParams, msigAddress, newThreshold string, from string) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	newT, err := strconv.ParseUint(newThreshold, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	params, actErr := actors.SerializeParams(&multisig9.ChangeNumApprovalsThresholdParams{
		NewThreshold: newT,
	})

	if actErr != nil {
		return nil, nil, actErr
	}

	msg, proposeParams, err := m.MsigPropose(msig, msig, big.Zero(), sendAddr, uint64(multisig.Methods.ChangeNumApprovalsThreshold), params)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigThresholdApproveMessage(baseParams BaseParams, msigAddress, proposerAddress, txId, newThreshold string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	prop, err := address.NewFromString(proposerAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	newT, err := strconv.ParseUint(newThreshold, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	params, actErr := actors.SerializeParams(&multisig9.ChangeNumApprovalsThresholdParams{
		NewThreshold: newT,
	})

	if actErr != nil {
		return nil, nil, actErr
	}

	msg, approveParams, err := m.MsigApproveTxnHash(msig, txid, prop, msig, big.Zero(), sendAddr, uint64(multisig.Methods.ChangeNumApprovalsThreshold), params)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigApproveTxnHash: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigThresholdCancelMessage(baseParams BaseParams, msigAddress, txId, newThreshold string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	newT, err := strconv.ParseUint(newThreshold, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	params, actErr := actors.SerializeParams(&multisig9.ChangeNumApprovalsThresholdParams{
		NewThreshold: newT,
	})

	if actErr != nil {
		return nil, nil, actErr
	}

	msg, cancelParams, err := m.MsigCancelTxnHash(msig, txid, msig, big.Zero(), sendAddr, uint64(multisig.Methods.ChangeNumApprovalsThreshold), params)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigCancelTxnHash: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, cancelParams, nil
}

func (m *Msiger) NewMsigChangeOwnerProposeMessage(baseParams BaseParams, msigAddress, miner, newOwner string, from string) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		return nil, nil, err
	}

	na, err := address.NewFromString(newOwner)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	newAddr, err := m.node.StateLookupID(ctx, na, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	mi, err := m.node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	if mi.Owner == newAddr {
		return nil, nil, fmt.Errorf("owner address already set to %s", na)
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	sp, err := actors.SerializeParams(&newAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("serializing params: %w", err)
	}

	msg, proposeParams, err := m.MsigPropose(msig, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ChangeOwnerAddress), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigChangeOwnerApproveMessage(baseParams BaseParams, msigAddress, proposerAddress, txId, miner, newOwner string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	prop, err := address.NewFromString(proposerAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	na, err := address.NewFromString(newOwner)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	newAddr, err := m.node.StateLookupID(ctx, na, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		return nil, nil, err
	}

	mi, err := m.node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	if mi.Owner == newAddr {
		return nil, nil, fmt.Errorf("owner address already set to %s", na)
	}

	sp, err := actors.SerializeParams(&newAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("serializing params: %w", err)
	}

	msg, approveParams, err := m.MsigApproveTxnHash(msig, txid, prop, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ChangeOwnerAddress), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigApproveTxnHash: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigWithdrawProposeMessage(baseParams BaseParams, msigAddress, miner, amount string, from string) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		return nil, nil, err
	}

	val, err := types.ParseFIL(amount)
	if err != nil {
		return nil, nil, err
	}

	sp, err := actors.SerializeParams(&miner8.WithdrawBalanceParams{
		AmountRequested: abi.TokenAmount(val),
	})

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, proposeParams, err := m.MsigPropose(msig, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.WithdrawBalance), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigWithdrawApproveMessage(baseParams BaseParams, msigAddress, proposerAddress, txId, miner, amount string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	prop, err := address.NewFromString(proposerAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	val, err := types.ParseFIL(amount)
	if err != nil {
		return nil, nil, err
	}

	sp, err := actors.SerializeParams(&miner8.WithdrawBalanceParams{
		AmountRequested: abi.TokenAmount(val),
	})

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		return nil, nil, err
	}

	msg, approveParams, err := m.MsigApproveTxnHash(msig, txid, prop, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.WithdrawBalance), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigApproveTxnHash: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigChangeWorkerProposeMessage(baseParams BaseParams, msigAddress, miner, newWorker string, from string) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		return nil, nil, err
	}

	na, err := address.NewFromString(newWorker)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	newAddr, err := m.node.StateLookupID(ctx, na, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	mi, err := m.node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	if mi.NewWorker.Empty() {
		if mi.Worker == newAddr {
			return nil, nil, fmt.Errorf("worker address already set to %s", na)
		}
	} else {
		if mi.NewWorker == newAddr {
			return nil, nil, fmt.Errorf("change to worker address %s already pending", na)
		}
	}

	cwp := &miner8.ChangeWorkerAddressParams{
		NewWorker:       newAddr,
		NewControlAddrs: mi.ControlAddresses,
	}

	sp, err := actors.SerializeParams(cwp)
	if err != nil {
		return nil, nil, fmt.Errorf("serializing params: %w", err)
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, proposeParams, err := m.MsigPropose(msig, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ChangeWorkerAddress), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigChangeWorkerApproveMessage(baseParams BaseParams, msigAddress, proposerAddress, txId, miner, newWorker string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	prop, err := address.NewFromString(proposerAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	na, err := address.NewFromString(newWorker)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	newAddr, err := m.node.StateLookupID(ctx, na, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		return nil, nil, err
	}

	mi, err := m.node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	if mi.NewWorker.Empty() {
		if mi.Worker == newAddr {
			return nil, nil, fmt.Errorf("worker address already set to %s", na)
		}
	} else {
		if mi.NewWorker == newAddr {
			return nil, nil, fmt.Errorf("change to worker address %s already pending", na)
		}
	}

	cwp := &miner8.ChangeWorkerAddressParams{
		NewWorker:       newAddr,
		NewControlAddrs: mi.ControlAddresses,
	}

	sp, err := actors.SerializeParams(cwp)
	if err != nil {
		return nil, nil, fmt.Errorf("serializing params: %w", err)
	}

	msg, approveParams, err := m.MsigApproveTxnHash(msig, txid, prop, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ChangeWorkerAddress), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigApproveTxnHash: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigConfirmChangeWorkerProposeMessage(baseParams BaseParams, msigAddress, miner, newWorker string, from string) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		return nil, nil, err
	}

	na, err := address.NewFromString(newWorker)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	newAddr, err := m.node.StateLookupID(ctx, na, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	mi, err := m.node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	if mi.NewWorker.Empty() {
		return nil, nil, xerrors.Errorf("no worker key change proposed")
	} else if mi.NewWorker != newAddr {
		return nil, nil, xerrors.Errorf("worker key %s does not match current worker key proposal %s", newAddr, mi.NewWorker)
	}

	if head, err := m.node.ChainHead(ctx); err != nil {
		return nil, nil, xerrors.Errorf("failed to get the chain head: %w", err)
	} else if head.Height() < mi.WorkerChangeEpoch {
		return nil, nil, xerrors.Errorf("worker key change cannot be confirmed until %d, current height is %d", mi.WorkerChangeEpoch, head.Height())
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, proposeParams, err := m.MsigPropose(msig, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ConfirmUpdateWorkerKey), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigConfirmChangeWorkerApproveMessage(baseParams BaseParams, msigAddress, proposerAddress, txId, miner, newWorker string, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	prop, err := address.NewFromString(proposerAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	na, err := address.NewFromString(newWorker)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	newAddr, err := m.node.StateLookupID(ctx, na, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		return nil, nil, err
	}

	mi, err := m.node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	if mi.NewWorker.Empty() {
		return nil, nil, xerrors.Errorf("no worker key change proposed")
	} else if mi.NewWorker != newAddr {
		return nil, nil, xerrors.Errorf("worker key %s does not match current worker key proposal %s", newAddr, mi.NewWorker)
	}

	if head, err := m.node.ChainHead(ctx); err != nil {
		return nil, nil, xerrors.Errorf("failed to get the chain head: %w", err)
	} else if head.Height() < mi.WorkerChangeEpoch {
		return nil, nil, xerrors.Errorf("worker key change cannot be confirmed until %d, current height is %d", mi.WorkerChangeEpoch, head.Height())
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, approveParams, err := m.MsigApproveTxnHash(msig, txid, prop, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ConfirmUpdateWorkerKey), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigApproveTxnHash: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigSetControlProposeMessage(baseParams BaseParams, msigAddress, miner string, from string, controlAddrs ...string) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	mi, err := m.node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	var toSet []address.Address
	for i, as := range controlAddrs {
		a, err := address.NewFromString(as)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing address %d: %w", i, err)
		}

		ka, err := m.node.StateAccountKey(ctx, a, types.EmptyTSK)
		if err != nil {
			return nil, nil, err
		}

		_, err = m.node.StateLookupID(ctx, ka, types.EmptyTSK)
		if err != nil {
			return nil, nil, fmt.Errorf("looking up %s: %w", ka, err)
		}

		toSet = append(toSet, ka)
	}

	cwp := &miner8.ChangeWorkerAddressParams{
		NewWorker:       mi.Worker,
		NewControlAddrs: toSet,
	}

	sp, err := actors.SerializeParams(cwp)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, proposeParams, err := m.MsigPropose(msig, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ChangeWorkerAddress), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigSetControlApproveMessage(baseParams BaseParams, msigAddress, proposerAddress, txId, miner string, from string, controlAddrs ...string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	prop, err := address.NewFromString(proposerAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		return nil, nil, err
	}

	mi, err := m.node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	var toSet []address.Address
	for i, as := range controlAddrs {
		a, err := address.NewFromString(as)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing address %d: %w", i, err)
		}

		ka, err := m.node.StateAccountKey(ctx, a, types.EmptyTSK)
		if err != nil {
			return nil, nil, err
		}

		_, err = m.node.StateLookupID(ctx, ka, types.EmptyTSK)
		if err != nil {
			return nil, nil, fmt.Errorf("looking up %s: %w", ka, err)
		}

		toSet = append(toSet, ka)
	}

	cwp := &miner8.ChangeWorkerAddressParams{
		NewWorker:       mi.Worker,
		NewControlAddrs: toSet,
	}

	sp, err := actors.SerializeParams(cwp)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	msg, approveParams, err := m.MsigApproveTxnHash(msig, txid, prop, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ChangeWorkerAddress), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigApproveTxnHash: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigChangeBeneficiaryProposeMessage(baseParams BaseParams, msigAddress, minerId string, from string, beneficiaryAddress, quota, expiration string, overwritePendingChange bool) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	na, err := address.NewFromString(beneficiaryAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing beneficiary address: %w", err)
	}

	newAddr, err := m.node.StateLookupID(ctx, na, types.EmptyTSK)
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

	mi, err := m.node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	if mi.PendingBeneficiaryTerm != nil && !overwritePendingChange {
		return nil, nil, fmt.Errorf("WARNING: replacing Pending Beneficiary Term of: Beneficiary: %s, Quota: %s, Expiration Epoch:%d", mi.PendingBeneficiaryTerm.NewBeneficiary.String(), mi.PendingBeneficiaryTerm.NewQuota.String(), mi.PendingBeneficiaryTerm.NewExpiration)
	}

	params := &miner.ChangeBeneficiaryParams{
		NewBeneficiary: newAddr,
		NewQuota:       abi.TokenAmount(quotaParam),
		NewExpiration:  abi.ChainEpoch(expirationParam),
	}

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, xerrors.Errorf("serializing params: %w", err)
	}

	msg, proposeParams, err := m.MsigPropose(msig, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ChangeBeneficiary), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigChangeBeneficiaryApproveMessage(baseParams BaseParams, msigAddress, proposerAddress, txId, minerId string, from string, beneficiaryAddress, quota, expiration string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	prop, err := address.NewFromString(proposerAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	na, err := address.NewFromString(beneficiaryAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing beneficiary address: %w", err)
	}

	newAddr, err := m.node.StateLookupID(ctx, na, types.EmptyTSK)
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

	params := &miner.ChangeBeneficiaryParams{
		NewBeneficiary: newAddr,
		NewQuota:       abi.TokenAmount(quotaParam),
		NewExpiration:  abi.ChainEpoch(expirationParam),
	}

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, xerrors.Errorf("serializing params: %w", err)
	}

	msg, approveParams, err := m.MsigApproveTxnHash(msig, txid, prop, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ChangeBeneficiary), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigApproveTxnHash: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

func (m *Msiger) NewMsigConfirmChangeBeneficiaryProposeMessage(baseParams BaseParams, msigAddress, minerId string, from string) (*types.Message, *multisig9.ProposeParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	minerAddr, err := address.NewFromString(minerId)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	mi, err := m.node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	if mi.PendingBeneficiaryTerm == nil {
		return nil, nil, fmt.Errorf("no pending beneficiary term found for miner %s", minerAddr)
	}

	params := &miner.ChangeBeneficiaryParams{
		NewBeneficiary: mi.PendingBeneficiaryTerm.NewBeneficiary,
		NewQuota:       mi.PendingBeneficiaryTerm.NewQuota,
		NewExpiration:  mi.PendingBeneficiaryTerm.NewExpiration,
	}

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, xerrors.Errorf("serializing params: %w", err)
	}

	msg, proposeParams, err := m.MsigPropose(msig, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ChangeBeneficiary), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigPropose: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, proposeParams, nil
}

func (m *Msiger) NewMsigConfirmChangeBeneficiaryApproveMessage(baseParams BaseParams, msigAddress, minerId, proposerAddress, txId, from string) (*types.Message, *multisig9.TxnIDParams, error) {
	msig, err := address.NewFromString(msigAddress)
	if err != nil {
		return nil, nil, err
	}

	sendAddr, err := address.NewFromString(from)
	if err != nil {
		return nil, nil, err
	}

	prop, err := address.NewFromString(proposerAddress)
	if err != nil {
		return nil, nil, err
	}

	txid, err := strconv.ParseUint(txId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	minerAddr, err := address.NewFromString(minerId)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	mi, err := m.node.StateMinerInfo(ctx, minerAddr, types.EmptyTSK)
	if err != nil {
		return nil, nil, err
	}

	if mi.PendingBeneficiaryTerm == nil {
		return nil, nil, fmt.Errorf("no pending beneficiary term found for miner %s", minerAddr)
	}

	params := &miner.ChangeBeneficiaryParams{
		NewBeneficiary: mi.PendingBeneficiaryTerm.NewBeneficiary,
		NewQuota:       mi.PendingBeneficiaryTerm.NewQuota,
		NewExpiration:  mi.PendingBeneficiaryTerm.NewExpiration,
	}

	sp, err := actors.SerializeParams(params)
	if err != nil {
		return nil, nil, xerrors.Errorf("serializing params: %w", err)
	}

	msg, approveParams, err := m.MsigApproveTxnHash(msig, txid, prop, minerAddr, big.Zero(), sendAddr, uint64(builtin.MethodsMiner.ChangeBeneficiary), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("MsigApproveTxnHash: %w", err)
	}

	msg, err = buildMessage(m.node, msg, baseParams)
	if err != nil {
		return nil, nil, err
	}

	return msg, approveParams, nil
}

// --------------------

func (m *Msiger) messageBuilder(from address.Address) (multisig.MessageBuilder, error) {
	av, err := actorstypes.VersionForNetwork(network.Version17)
	if err != nil {
		return nil, err
	}

	return multisig.Message(av, from), nil
}

func (m *Msiger) MsigCreate(req uint64, addrs []address.Address, duration abi.ChainEpoch, val types.BigInt, src address.Address) (*types.Message, *multisig9.ConstructorParams, error) {
	mb, err := m.messageBuilder(src)
	if err != nil {
		return nil, nil, err
	}

	msg, err := mb.Create(addrs, req, 0, duration, val)
	if err != nil {
		return nil, nil, err
	}

	msigParams := &multisig9.ConstructorParams{
		Signers:               addrs,
		NumApprovalsThreshold: req,
		UnlockDuration:        duration,
		StartEpoch:            0,
	}

	// msg.Params
	//  enc, _ := actors.SerializeParams(msigParams)
	//
	//	code, _ := builtin.GetMultisigActorCodeID(actors.Version8)
	//
	//	execParams := &init8.ExecParams{
	//		CodeCID:           code,
	//		ConstructorParams: enc,
	//	}
	//
	//	enc, _ = actors.SerializeParams(execParams)

	return msg, msigParams, nil
}

func (m *Msiger) MsigPropose(msig address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.Message, *multisig9.ProposeParams, error) {
	mb, err := m.messageBuilder(src)
	if err != nil {
		return nil, nil, err
	}

	msg, err := mb.Propose(msig, to, amt, abi.MethodNum(method), params)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to create proposal: %w", err)
	}

	proposeParams := &multisig9.ProposeParams{
		To:     to,
		Value:  amt,
		Method: abi.MethodNum(method),
		Params: params,
	}

	return msg, proposeParams, nil
}

func (m *Msiger) MsigAddPropose(msig address.Address, src address.Address, newAdd address.Address, inc bool) (*types.Message, *multisig9.ProposeParams, error) {
	enc, actErr := serializeAddParams(newAdd, inc)
	if actErr != nil {
		return nil, nil, actErr
	}

	return m.MsigPropose(msig, msig, big.Zero(), src, uint64(multisig.Methods.AddSigner), enc)
}

func (m *Msiger) MsigAddApprove(msig address.Address, src address.Address, txID uint64, proposer address.Address, newAdd address.Address, inc bool) (*types.Message, *multisig9.TxnIDParams, error) {
	enc, actErr := serializeAddParams(newAdd, inc)
	if actErr != nil {
		return nil, nil, actErr
	}

	return m.MsigApproveTxnHash(msig, txID, proposer, msig, big.Zero(), src, uint64(multisig.Methods.AddSigner), enc)
}

func (m *Msiger) MsigAddCancel(msig address.Address, src address.Address, txID uint64, newAdd address.Address, inc bool) (*types.Message, *multisig9.TxnIDParams, error) {
	enc, actErr := serializeAddParams(newAdd, inc)
	if actErr != nil {
		return nil, nil, actErr
	}

	return m.MsigCancelTxnHash(msig, txID, msig, big.Zero(), src, uint64(multisig.Methods.AddSigner), enc)
}

func (m *Msiger) MsigSwapPropose(msig address.Address, src address.Address, oldAdd address.Address, newAdd address.Address) (*types.Message, *multisig9.ProposeParams, error) {
	enc, actErr := serializeSwapParams(oldAdd, newAdd)
	if actErr != nil {
		return nil, nil, actErr
	}

	return m.MsigPropose(msig, msig, big.Zero(), src, uint64(multisig.Methods.SwapSigner), enc)
}

func (m *Msiger) MsigSwapApprove(msig address.Address, src address.Address, txID uint64, proposer address.Address, oldAdd address.Address, newAdd address.Address) (*types.Message, *multisig9.TxnIDParams, error) {
	enc, actErr := serializeSwapParams(oldAdd, newAdd)
	if actErr != nil {
		return nil, nil, actErr
	}

	return m.MsigApproveTxnHash(msig, txID, proposer, msig, big.Zero(), src, uint64(multisig.Methods.SwapSigner), enc)
}

func (m *Msiger) MsigSwapCancel(msig address.Address, src address.Address, txID uint64, oldAdd address.Address, newAdd address.Address) (*types.Message, *multisig9.TxnIDParams, error) {
	enc, actErr := serializeSwapParams(oldAdd, newAdd)
	if actErr != nil {
		return nil, nil, actErr
	}

	return m.MsigCancelTxnHash(msig, txID, msig, big.Zero(), src, uint64(multisig.Methods.SwapSigner), enc)
}

func (m *Msiger) MsigApprove(msig address.Address, txID uint64, src address.Address) (*types.Message, *multisig9.TxnIDParams, error) {

	return m.MsigApproveOrCancelSimple(api.MsigApprove, msig, txID, src)
}

func (m *Msiger) MsigApproveTxnHash(msig address.Address, txID uint64, proposer address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.Message, *multisig9.TxnIDParams, error) {
	return m.MsigApproveOrCancelTxnHash(api.MsigApprove, msig, txID, proposer, to, amt, src, method, params)
}

func (m *Msiger) MsigCancel(msig address.Address, txID uint64, src address.Address) (*types.Message, *multisig9.TxnIDParams, error) {
	return m.MsigApproveOrCancelSimple(api.MsigCancel, msig, txID, src)
}

func (m *Msiger) MsigCancelTxnHash(msig address.Address, txID uint64, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.Message, *multisig9.TxnIDParams, error) {
	return m.MsigApproveOrCancelTxnHash(api.MsigCancel, msig, txID, src, to, amt, src, method, params)
}

func (m *Msiger) MsigRemoveSigner(msig address.Address, proposer address.Address, toRemove address.Address, decrease bool) (*types.Message, *multisig9.ProposeParams, error) {
	enc, actErr := serializeRemoveParams(toRemove, decrease)
	if actErr != nil {
		return nil, nil, actErr
	}

	return m.MsigPropose(msig, msig, types.NewInt(0), proposer, uint64(multisig.Methods.RemoveSigner), enc)
}

func (m *Msiger) MsigApproveOrCancelSimple(operation api.MsigProposeResponse, msig address.Address, txID uint64, src address.Address) (*types.Message, *multisig9.TxnIDParams, error) {
	if msig == address.Undef {
		return nil, nil, xerrors.Errorf("must provide multisig address")
	}

	if src == address.Undef {
		return nil, nil, xerrors.Errorf("must provide source address")
	}

	mb, err := m.messageBuilder(src)
	if err != nil {
		return nil, nil, err
	}

	var msg *types.Message
	switch operation {
	case api.MsigApprove:
		msg, err = mb.Approve(msig, txID, nil)
	case api.MsigCancel:
		msg, err = mb.Cancel(msig, txID, nil)
	default:
		return nil, nil, xerrors.Errorf("Invalid operation for msigApproveOrCancel")
	}
	if err != nil {
		return nil, nil, err
	}

	params, err := txnParams(txID, nil)
	if err != nil {
		return nil, nil, err
	}

	return msg, params, nil
}

func (m *Msiger) MsigApproveOrCancelTxnHash(operation api.MsigProposeResponse, msig address.Address, txID uint64, proposer address.Address, to address.Address, amt types.BigInt, src address.Address, method uint64, params []byte) (*types.Message, *multisig9.TxnIDParams, error) {
	if msig == address.Undef {
		return nil, nil, xerrors.Errorf("must provide multisig address")
	}

	if src == address.Undef {
		return nil, nil, xerrors.Errorf("must provide source address")
	}

	if proposer.Protocol() != address.ID {
		proposerID, err := m.node.StateLookupID(context.Background(), proposer, types.EmptyTSK)
		if err != nil {
			return nil, nil, err
		}
		proposer = proposerID
	}

	p := multisig.ProposalHashData{
		Requester: proposer,
		To:        to,
		Value:     amt,
		Method:    abi.MethodNum(method),
		Params:    params,
	}

	mb, err := m.messageBuilder(src)
	if err != nil {
		return nil, nil, err
	}

	var msg *types.Message
	switch operation {
	case api.MsigApprove:
		msg, err = mb.Approve(msig, txID, &p)
	case api.MsigCancel:
		msg, err = mb.Cancel(msig, txID, &p)
	default:
		return nil, nil, xerrors.Errorf("Invalid operation for msigApproveOrCancel")
	}
	if err != nil {
		return nil, nil, err
	}

	methodParams, err := txnParams(txID, &p)
	if err != nil {
		return nil, nil, err
	}

	return msg, methodParams, nil
}

func serializeAddParams(new address.Address, inc bool) ([]byte, error) {
	enc, actErr := actors.SerializeParams(&multisig9.AddSignerParams{
		Signer:   new,
		Increase: inc,
	})
	if actErr != nil {
		return nil, actErr
	}

	return enc, nil
}

func serializeSwapParams(old address.Address, new address.Address) ([]byte, error) {
	enc, actErr := actors.SerializeParams(&multisig9.SwapSignerParams{
		From: old,
		To:   new,
	})
	if actErr != nil {
		return nil, actErr
	}

	return enc, nil
}

func serializeRemoveParams(rem address.Address, dec bool) ([]byte, error) {
	enc, actErr := actors.SerializeParams(&multisig9.RemoveSignerParams{
		Signer:   rem,
		Decrease: dec,
	})
	if actErr != nil {
		return nil, actErr
	}

	return enc, nil
}

func txnParams(id uint64, data *multisig9.ProposalHashData) (*multisig9.TxnIDParams, error) {
	params := multisig9.TxnIDParams{ID: multisig9.TxnID(id)}
	if data != nil {
		if data.Requester.Protocol() != address.ID {
			return nil, xerrors.Errorf("proposer address must be an ID address, was %s", data.Requester)
		}
		if data.Value.Sign() == -1 {
			return nil, xerrors.Errorf("proposal value must be non-negative, was %s", data.Value)
		}
		if data.To == address.Undef {
			return nil, xerrors.Errorf("proposed destination address must be set")
		}
		pser, err := data.Serialize()
		if err != nil {
			return nil, err
		}
		hash := blake2b.Sum256(pser)
		params.ProposalHash = hash[:]
	}

	return &params, nil
}
