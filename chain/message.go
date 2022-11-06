package chain

import (
	"encoding/json"
	"errors"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	multisig8 "github.com/filecoin-project/go-state-types/builtin/v8/multisig"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/types"
	miner8 "github.com/filecoin-project/specs-actors/v8/actors/builtin/miner"
	power8 "github.com/filecoin-project/specs-actors/v8/actors/builtin/power"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"
)

var ErrNotSupported = errors.New("not supported")

type Message struct {
	Version    uint64     `json:"version"`
	To         string     `json:"to"`
	From       string     `json:"from"`
	Nonce      uint64     `json:"nonce"`
	Value      int64      `json:"value"`
	GasLimit   int64      `json:"gas_limit"`
	GasFeeCap  int64      `json:"gas_feecap"`
	GasPremium int64      `json:"gas_premium"`
	Method     uint64     `json:"method"`
	Params     ParamsInfo `json:"params"`
}

type ParamsInfo struct {
	Name   string `json:"name"`
	Params string `json:"params"`
}

func BuildMessage(msg *Message) (*types.Message, error) {
	from, err := address.NewFromString(msg.From)
	if err != nil {
		return nil, err
	}

	to, err := address.NewFromString(msg.To)
	if err != nil {
		return nil, err
	}

	params, err := DecodeParams(msg.Params)
	if err != nil {
		return nil, err
	}

	return &types.Message{
		Version:    msg.Version,
		To:         to,
		From:       from,
		Nonce:      msg.Nonce,
		Value:      abi.NewTokenAmount(msg.Value),
		GasLimit:   msg.GasLimit,
		GasFeeCap:  abi.NewTokenAmount(msg.GasFeeCap),
		GasPremium: abi.NewTokenAmount(msg.GasPremium),
		Method:     abi.MethodNum(msg.Method),
		Params:     params,
	}, nil
}

func DecodeParams(params ParamsInfo) ([]byte, error) {
	var cbor cbg.CBORMarshaler
	var err error

	switch params.Name {
	case "": // Send & ConfirmUpdateWorkerKey
		return []byte{}, nil
	case "CreateMinerParams":
		var p power8.CreateMinerParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "WithdrawBalanceParams":
		var p miner8.WithdrawBalanceParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "Address":
		addr, err := address.NewFromString(params.Params)
		if err != nil {
			return nil, err
		}
		cbor = &addr
	case "ChangeWorkerAddressParams":
		var p miner8.ChangeWorkerAddressParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "ConstructorParams":
		var p multisig8.ConstructorParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "ProposeParams":
		var p multisig8.ProposeParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "TxnIDParams":
		var p multisig8.TxnIDParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "AddSignerParams":
		var p multisig8.AddSignerParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "RemoveSignerParams":
		var p multisig8.RemoveSignerParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "SwapSignerParams":
		var p multisig8.SwapSignerParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "ChangeNumApprovalsThresholdParams":
		var p multisig8.ChangeNumApprovalsThresholdParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "LockBalanceParams":
		var p multisig8.LockBalanceParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p

	default:
		return nil, ErrNotSupported
	}

	sp, err := actors.SerializeParams(cbor)
	if err != nil {
		return nil, xerrors.Errorf("serializing params: %w", err)
	}

	return sp, nil
}

func EncodeParams(params interface{}) (*ParamsInfo, error) {
	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	switch params.(type) {
	case *power8.CreateMinerParams: // CreateMiner
		return &ParamsInfo{
			Name:   "CreateMinerParams",
			Params: string(b),
		}, nil
	case *miner8.WithdrawBalanceParams: // WithdrawBalance
		return &ParamsInfo{
			Name:   "WithdrawBalanceParams",
			Params: string(b),
		}, nil
	case *address.Address: // ChangeOwnerAddress
		return &ParamsInfo{
			Name:   "Address",
			Params: params.(*address.Address).String(),
		}, nil
	case *miner8.ChangeWorkerAddressParams: // ChangeWorkerAddress, and ConfirmUpdateWorkerKey has no params
		return &ParamsInfo{
			Name:   "ChangeWorkerAddressParams",
			Params: string(b),
		}, nil
	case *multisig8.ConstructorParams: // Constructor
		return &ParamsInfo{
			Name:   "ConstructorParams",
			Params: string(b),
		}, nil
	case *multisig8.ProposeParams: // Propose
		return &ParamsInfo{
			Name:   "ProposeParams",
			Params: string(b),
		}, nil
	case *multisig8.TxnIDParams: // Cancel & Approve
		return &ParamsInfo{
			Name:   "TxnIDParams",
			Params: string(b),
		}, nil
	case *multisig8.AddSignerParams: // AddSigner
		return &ParamsInfo{
			Name:   "AddSignerParams",
			Params: string(b),
		}, nil
	case *multisig8.RemoveSignerParams: // RemoveSigner
		return &ParamsInfo{
			Name:   "RemoveSignerParams",
			Params: string(b),
		}, nil
	case *multisig8.SwapSignerParams: // SwapSigner
		return &ParamsInfo{
			Name:   "SwapSignerParams",
			Params: string(b),
		}, nil
	case *multisig8.ChangeNumApprovalsThresholdParams: // ChangeNumApprovalsThreshold
		return &ParamsInfo{
			Name:   "ChangeNumApprovalsThresholdParams",
			Params: string(b),
		}, nil
	case *multisig8.LockBalanceParams: // LockBalance
		return &ParamsInfo{
			Name:   "LockBalanceParams",
			Params: string(b),
		}, nil
	default:
		return nil, ErrNotSupported
	}
}
