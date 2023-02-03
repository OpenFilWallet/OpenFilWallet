package wallet

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/cbor"
	"github.com/filecoin-project/lotus/chain/consensus/filcns"
	"github.com/filecoin-project/lotus/chain/vm"
	exported7 "github.com/filecoin-project/specs-actors/v7/actors/builtin/exported"
	"github.com/gin-gonic/gin"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"
	"reflect"
)

func (w *Wallet) Decode(c *gin.Context) {
	param := client.DecodeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("Decode: BindJSON", "err", err.Error())
		ReturnError(c, ParamErr)
		return
	}

	var params []byte
	switch param.Encoding {
	case "base64":
		params, err = base64.StdEncoding.DecodeString(param.Params)
		if err != nil {
			log.Warnw("Decode: DecodeString", "err", err.Error())
			ReturnError(c, NewError(500, fmt.Sprintf("decoding base64 value: %s", err)))
			return
		}
	case "hex":
		params, err = hex.DecodeString(param.Params)
		if err != nil {
			log.Warnw("Decode: DecodeString", "err", err.Error())
			ReturnError(c, NewError(500, fmt.Sprintf("decoding hex value: %s", err)))
			return
		}
	default:
		log.Warnw("Decode: DecodeString", "err", fmt.Sprintf("unrecognized encoding: %s", param.Encoding))
		ReturnError(c, NewError(500, fmt.Sprintf("unrecognized encoding: %s", param.Encoding)))
		return
	}

	decParams, err := DecodeParams(abi.MethodNum(param.Method), params)
	if err != nil {
		log.Warnw("Decode: DecodeParams", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
	}

	ReturnOk(c, string(decParams))
}

func (w *Wallet) Encode(c *gin.Context) {
	param := client.EncodeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("Encode: BindJSON", "err", err.Error())
		ReturnError(c, ParamErr)
		return
	}

	encParams, err := EncodeParams(abi.MethodNum(param.Method), param.Params)
	if err != nil {
		log.Warnw("Encode: EncodeParams", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
	}

	encodeMsg := ""
	switch param.Encoding {
	case "base64", "b64":
		encodeMsg = base64.StdEncoding.EncodeToString(encParams)
	case "hex":
		encodeMsg = hex.EncodeToString(encParams)
	default:
		log.Warnw("Encode: EncodeToString", "err", fmt.Sprintf("not support encoding: %s", param.Encoding))
		ReturnError(c, NewError(500, "not support encoding"))
	}

	ReturnOk(c, encodeMsg)
}

func EncodeParams(method abi.MethodNum, params string) ([]byte, error) {
	var paramType cbg.CBORUnmarshaler
	for _, actor := range exported7.BuiltinActors() {
		if MethodMetaMap, ok := filcns.NewActorRegistry().Methods[actor.Code()]; ok {
			var m vm.MethodMeta
			var found bool
			if m, found = MethodMetaMap[abi.MethodNum(method)]; found {
				paramType = reflect.New(m.Params.Elem()).Interface().(cbg.CBORUnmarshaler)
			}
		}
	}

	if paramType == nil {
		return nil, fmt.Errorf("unknown method %d", method)
	}

	if err := json.Unmarshal(json.RawMessage(params), &paramType); err != nil {
		return nil, xerrors.Errorf("json unmarshal: %w", err)
	}

	var cbb bytes.Buffer
	if err := paramType.(cbor.Marshaler).MarshalCBOR(&cbb); err != nil {
		return nil, xerrors.Errorf("cbor marshal: %w", err)
	}

	return cbb.Bytes(), nil
}

func DecodeParams(method abi.MethodNum, params []byte) ([]byte, error) {
	var paramType cbg.CBORUnmarshaler
	for _, actor := range exported7.BuiltinActors() {
		if MethodMetaMap, ok := filcns.NewActorRegistry().Methods[actor.Code()]; ok {
			var m vm.MethodMeta
			var found bool
			if m, found = MethodMetaMap[abi.MethodNum(method)]; found {
				paramType = reflect.New(m.Params.Elem()).Interface().(cbg.CBORUnmarshaler)
			}
		}
	}

	if paramType == nil {
		return nil, fmt.Errorf("unknown method %d", method)
	}

	if err := paramType.UnmarshalCBOR(bytes.NewReader(params)); err != nil {
		return nil, err
	}

	return json.MarshalIndent(paramType, "", "  ")
}
