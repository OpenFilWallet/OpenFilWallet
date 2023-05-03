package chain

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	actorstypes "github.com/filecoin-project/go-state-types/actors"
	multisig11 "github.com/filecoin-project/go-state-types/builtin/v11/multisig"
	init8 "github.com/filecoin-project/go-state-types/builtin/v9/init"
	"github.com/filecoin-project/go-state-types/builtin/v9/miner"
	"github.com/filecoin-project/go-state-types/manifest"
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

func (m *Message) String() string {
	msg, _ := json.Marshal(m)
	return string(msg)
}

func EncodeMessage(msg *types.Message, params interface{}) (*Message, error) {
	paramsInfo, err := EncodeParams(params)
	if err != nil {
		return nil, err
	}

	dp, err := DecodeParams(*paramsInfo)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(dp, msg.Params) {
		return nil, errors.New("EncodeParams and DecodeParams of params do not match")
	}

	return &Message{
		Version:    msg.Version,
		To:         msg.To.String(),
		From:       msg.From.String(),
		Nonce:      msg.Nonce,
		Value:      msg.Value.Int64(),
		GasLimit:   msg.GasLimit,
		GasFeeCap:  msg.GasFeeCap.Int64(),
		GasPremium: msg.GasPremium.Int64(),
		Method:     uint64(msg.Method),
		Params:     *paramsInfo,
	}, nil
}

func DecodeMessage(msg *Message) (*types.Message, error) {
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
		var p multisig11.ConstructorParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}

		enc, actErr := actors.SerializeParams(&p)
		if actErr != nil {
			return nil, actErr
		}

		code, ok := actors.GetActorCodeID(actorstypes.Version11, manifest.MultisigKey)
		if !ok {
			return nil, xerrors.Errorf("failed to get multisig code ID")
		}

		ep := &init8.ExecParams{
			CodeCID:           code,
			ConstructorParams: enc,
		}

		cbor = ep
	case "ProposeParams":
		var p multisig11.ProposeParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "TxnIDParams":
		var p multisig11.TxnIDParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "AddSignerParams":
		var p multisig11.AddSignerParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "RemoveSignerParams":
		var p multisig11.RemoveSignerParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "SwapSignerParams":
		var p multisig11.SwapSignerParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "ChangeNumApprovalsThresholdParams":
		var p multisig11.ChangeNumApprovalsThresholdParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "LockBalanceParams":
		var p multisig11.LockBalanceParams
		err = json.Unmarshal([]byte(params.Params), &p)
		if err != nil {
			return nil, err
		}
		cbor = &p
	case "ChangeBeneficiaryParams":
		var p miner.ChangeBeneficiaryParams
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
	case nil:
		return &ParamsInfo{
			Name:   "",
			Params: "",
		}, nil
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
	case *multisig11.ConstructorParams: // Msig Constructor
		return &ParamsInfo{
			Name:   "ConstructorParams",
			Params: string(b),
		}, nil
	case *multisig11.ProposeParams: // Propose
		return &ParamsInfo{
			Name:   "ProposeParams",
			Params: string(b),
		}, nil
	case *multisig11.TxnIDParams: // Cancel & Approve
		return &ParamsInfo{
			Name:   "TxnIDParams",
			Params: string(b),
		}, nil
	case *multisig11.AddSignerParams: // AddSigner
		return &ParamsInfo{
			Name:   "AddSignerParams",
			Params: string(b),
		}, nil
	case *multisig11.RemoveSignerParams: // RemoveSigner
		return &ParamsInfo{
			Name:   "RemoveSignerParams",
			Params: string(b),
		}, nil
	case *multisig11.SwapSignerParams: // SwapSigner
		return &ParamsInfo{
			Name:   "SwapSignerParams",
			Params: string(b),
		}, nil
	case *multisig11.ChangeNumApprovalsThresholdParams: // ChangeNumApprovalsThreshold
		return &ParamsInfo{
			Name:   "ChangeNumApprovalsThresholdParams",
			Params: string(b),
		}, nil
	case *multisig11.LockBalanceParams: // LockBalance
		return &ParamsInfo{
			Name:   "LockBalanceParams",
			Params: string(b),
		}, nil
	case *miner.ChangeBeneficiaryParams:
		return &ParamsInfo{
			Name:   "ChangeBeneficiaryParams",
			Params: string(b),
		}, nil
	default:
		return nil, ErrNotSupported
	}
}
