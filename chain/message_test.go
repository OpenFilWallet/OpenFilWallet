package chain

import (
	"encoding/json"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	actorstypes "github.com/filecoin-project/go-state-types/actors"
	init13 "github.com/filecoin-project/go-state-types/builtin/v13/init"
	multisig13 "github.com/filecoin-project/go-state-types/builtin/v13/multisig"
	"github.com/filecoin-project/go-state-types/manifest"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/actors/builtin/power"
	"github.com/filecoin-project/lotus/chain/types"
	specsminer8 "github.com/filecoin-project/specs-actors/v8/actors/builtin/miner"
	specspower8 "github.com/filecoin-project/specs-actors/v8/actors/builtin/power"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestEncodeParamsAndDecodeParams(t *testing.T) {
	testAddr, _ := address.NewFromString("f13p72btfd5ielrdibduudppjhrvg2ahuecd6xapy")

	var paramsSlice = make([]interface{}, 0)
	var serializedSlice = make([][]byte, 0)

	createMinerParams := specspower8.CreateMinerParams{
		Owner:               testAddr,
		Worker:              testAddr,
		WindowPoStProofType: abi.RegisteredPoStProof_StackedDrgWindow64GiBV1,
		Peer:                abi.PeerID("not really a peer id"),
		Multiaddrs:          []abi.Multiaddrs{{1}},
	}

	paramsSlice = append(paramsSlice, &createMinerParams)

	sp, err := actors.SerializeParams(&createMinerParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	withdrawBalanceParams := specsminer8.WithdrawBalanceParams{
		AmountRequested: abi.NewTokenAmount(10000),
	}

	paramsSlice = append(paramsSlice, &withdrawBalanceParams)

	sp, err = actors.SerializeParams(&withdrawBalanceParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	paramsSlice = append(paramsSlice, &testAddr)

	sp, err = actors.SerializeParams(&testAddr)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	changeWorkerAddressParams := specsminer8.ChangeWorkerAddressParams{
		NewWorker:       testAddr,
		NewControlAddrs: []address.Address{testAddr},
	}
	paramsSlice = append(paramsSlice, &changeWorkerAddressParams)

	sp, err = actors.SerializeParams(&changeWorkerAddressParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	constructorParams := multisig13.ConstructorParams{
		Signers:               []address.Address{testAddr},
		NumApprovalsThreshold: 3,
		UnlockDuration:        abi.ChainEpoch(10000),
		StartEpoch:            abi.ChainEpoch(10001),
	}
	paramsSlice = append(paramsSlice, &constructorParams)

	require.NoError(t, err)
	enc, actErr := actors.SerializeParams(&constructorParams)
	require.NoError(t, actErr)

	code, ok := actors.GetActorCodeID(actorstypes.Version11, manifest.MultisigKey)
	require.True(t, ok)

	// new actors are created by invoking 'exec' on the init actor with the constructor params
	execParams := &init13.ExecParams{
		CodeCID:           code,
		ConstructorParams: enc,
	}

	sp, actErr = actors.SerializeParams(execParams)
	require.NoError(t, actErr)

	serializedSlice = append(serializedSlice, sp)

	proposeParams := multisig13.ProposeParams{
		To:     testAddr,
		Value:  abi.NewTokenAmount(1000),
		Method: abi.MethodNum(1),
		Params: []byte("test params"),
	}
	paramsSlice = append(paramsSlice, &proposeParams)

	sp, err = actors.SerializeParams(&proposeParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	txnIDParams := multisig13.TxnIDParams{
		ID:           multisig13.TxnID(1),
		ProposalHash: []byte("test hash"),
	}
	paramsSlice = append(paramsSlice, &txnIDParams)

	sp, err = actors.SerializeParams(&txnIDParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	addSignerParams := multisig13.AddSignerParams{
		Signer:   testAddr,
		Increase: true,
	}
	paramsSlice = append(paramsSlice, &addSignerParams)

	sp, err = actors.SerializeParams(&addSignerParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	removeSignerParams := multisig13.RemoveSignerParams{
		Signer:   testAddr,
		Decrease: true,
	}
	paramsSlice = append(paramsSlice, &removeSignerParams)

	sp, err = actors.SerializeParams(&removeSignerParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	swapSignerParams := multisig13.SwapSignerParams{
		From: testAddr,
		To:   testAddr,
	}
	paramsSlice = append(paramsSlice, &swapSignerParams)

	sp, err = actors.SerializeParams(&swapSignerParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	changeNumApprovalsThresholdParams := multisig13.ChangeNumApprovalsThresholdParams{
		NewThreshold: 5,
	}
	paramsSlice = append(paramsSlice, &changeNumApprovalsThresholdParams)

	sp, err = actors.SerializeParams(&changeNumApprovalsThresholdParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	lockBalanceParams := multisig13.LockBalanceParams{
		StartEpoch:     abi.ChainEpoch(10000),
		UnlockDuration: abi.ChainEpoch(10000),
		Amount:         abi.NewTokenAmount(10000),
	}
	paramsSlice = append(paramsSlice, &lockBalanceParams)

	sp, err = actors.SerializeParams(&lockBalanceParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	for i, params := range paramsSlice {
		p, err := EncodeParams(params)
		if err != nil {
			t.Fatalf("%d, err: %s", i, err.Error())
		}
		b, err := json.Marshal(p)
		require.NoError(t, err)
		t.Log(i, string(b))

		serializedParams, err := DecodeParams(*p)
		require.NoError(t, err)
		require.True(t, reflect.DeepEqual(serializedSlice[i], serializedParams))
	}
}

func TestEncodeMessage(t *testing.T) {
	testAddr, _ := address.NewFromString("f13p72btfd5ielrdibduudppjhrvg2ahuecd6xapy")

	var paramsSlice = make([]interface{}, 0)
	var serializedSlice = make([][]byte, 0)

	createMinerParams := specspower8.CreateMinerParams{
		Owner:               testAddr,
		Worker:              testAddr,
		WindowPoStProofType: abi.RegisteredPoStProof_StackedDrgWindow64GiBV1,
		Peer:                abi.PeerID("not really a peer id"),
		Multiaddrs:          []abi.Multiaddrs{{1}},
	}

	paramsSlice = append(paramsSlice, &createMinerParams)

	sp, err := actors.SerializeParams(&createMinerParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	withdrawBalanceParams := specsminer8.WithdrawBalanceParams{
		AmountRequested: abi.NewTokenAmount(10000),
	}

	paramsSlice = append(paramsSlice, &withdrawBalanceParams)

	sp, err = actors.SerializeParams(&withdrawBalanceParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	paramsSlice = append(paramsSlice, &testAddr)

	sp, err = actors.SerializeParams(&testAddr)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	changeWorkerAddressParams := specsminer8.ChangeWorkerAddressParams{
		NewWorker:       testAddr,
		NewControlAddrs: []address.Address{testAddr},
	}
	paramsSlice = append(paramsSlice, &changeWorkerAddressParams)

	sp, err = actors.SerializeParams(&changeWorkerAddressParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	constructorParams := multisig13.ConstructorParams{
		Signers:               []address.Address{testAddr},
		NumApprovalsThreshold: 3,
		UnlockDuration:        abi.ChainEpoch(10000),
		StartEpoch:            abi.ChainEpoch(10001),
	}
	paramsSlice = append(paramsSlice, &constructorParams)

	enc, actErr := actors.SerializeParams(&constructorParams)
	require.NoError(t, actErr)

	code, ok := actors.GetActorCodeID(actorstypes.Version11, manifest.MultisigKey)
	if !ok {
		t.Fatalf("failed to get multisig code ID")
	}

	// new actors are created by invoking 'exec' on the init actor with the constructor params
	execParams := &init13.ExecParams{
		CodeCID:           code,
		ConstructorParams: enc,
	}

	sp, err = actors.SerializeParams(execParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	proposeParams := multisig13.ProposeParams{
		To:     testAddr,
		Value:  abi.NewTokenAmount(1000),
		Method: abi.MethodNum(1),
		Params: []byte("test params"),
	}
	paramsSlice = append(paramsSlice, &proposeParams)

	sp, err = actors.SerializeParams(&proposeParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	txnIDParams := multisig13.TxnIDParams{
		ID:           multisig13.TxnID(1),
		ProposalHash: []byte("test hash"),
	}
	paramsSlice = append(paramsSlice, &txnIDParams)

	sp, err = actors.SerializeParams(&txnIDParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	addSignerParams := multisig13.AddSignerParams{
		Signer:   testAddr,
		Increase: true,
	}
	paramsSlice = append(paramsSlice, &addSignerParams)

	sp, err = actors.SerializeParams(&addSignerParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	removeSignerParams := multisig13.RemoveSignerParams{
		Signer:   testAddr,
		Decrease: true,
	}
	paramsSlice = append(paramsSlice, &removeSignerParams)

	sp, err = actors.SerializeParams(&removeSignerParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	swapSignerParams := multisig13.SwapSignerParams{
		From: testAddr,
		To:   testAddr,
	}
	paramsSlice = append(paramsSlice, &swapSignerParams)

	sp, err = actors.SerializeParams(&swapSignerParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	changeNumApprovalsThresholdParams := multisig13.ChangeNumApprovalsThresholdParams{
		NewThreshold: 5,
	}
	paramsSlice = append(paramsSlice, &changeNumApprovalsThresholdParams)

	sp, err = actors.SerializeParams(&changeNumApprovalsThresholdParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	lockBalanceParams := multisig13.LockBalanceParams{
		StartEpoch:     abi.ChainEpoch(10000),
		UnlockDuration: abi.ChainEpoch(10000),
		Amount:         abi.NewTokenAmount(10000),
	}
	paramsSlice = append(paramsSlice, &lockBalanceParams)

	sp, err = actors.SerializeParams(&lockBalanceParams)
	require.NoError(t, err)
	serializedSlice = append(serializedSlice, sp)

	for i, sp := range serializedSlice {
		msg := &types.Message{
			Version:    0,
			To:         power.Address,
			From:       testAddr,
			Nonce:      0,
			Value:      abi.NewTokenAmount(0),
			GasLimit:   56518036,
			GasFeeCap:  abi.NewTokenAmount(1238542683),
			GasPremium: abi.NewTokenAmount(99967),
			Method:     power.Methods.CreateMiner,
			Params:     sp,
		}

		myMsg, err := EncodeMessage(msg, paramsSlice[i])
		require.NoError(t, err)
		dMsg, err := DecodeMessage(myMsg)
		require.NoError(t, err)
		require.Equal(t, msg, dMsg)
	}
}
