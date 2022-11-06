package chain

import (
	"bytes"
	"encoding/json"
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/actors/builtin/power"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/wallet"
	power8 "github.com/filecoin-project/specs-actors/v8/actors/builtin/power"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestSignMessage(t *testing.T) {
	workerAddr, _ := address.NewFromString("f13p72btfd5ielrdibduudppjhrvg2ahuecd6xapy")
	ownerKey := "7b2254797065223a22736563703235366b31222c22507269766174654b6579223a22384c7331545157743063625574553270636450436d623371596d7959696274714153365a65664f456a4b343d227d"
	ki, err := account.GenerateKeyInfoFromPriKey(ownerKey, "hex-lotus")
	require.NoError(t, err)
	nk, err := wallet.NewKey(*ki)
	require.NoError(t, err)

	param := power8.CreateMinerParams{
		Owner:               nk.Address,
		Worker:              workerAddr,
		WindowPoStProofType: abi.RegisteredPoStProof_StackedDrgWindow64GiBV1,
		Peer:                abi.PeerID("not really a peer id"),
		Multiaddrs:          []abi.Multiaddrs{{1}},
	}
	sp, err := actors.SerializeParams(&param)
	require.NoError(t, err)

	msg := &types.Message{
		Version:    0,
		To:         power.Address,
		From:       nk.Address,
		Nonce:      0,
		Value:      abi.NewTokenAmount(0),
		GasLimit:   56518036,
		GasFeeCap:  abi.NewTokenAmount(1238542683),
		GasPremium: abi.NewTokenAmount(99967),
		Method:     power.Methods.CreateMiner,
		Params:     sp,
	}

	signedMsg, err := SignMessage(*nk, msg)
	require.NoError(t, err)
	t.Log(signedMsg)

	var signedMsgBuf bytes.Buffer
	err = signedMsg.MarshalCBOR(&signedMsgBuf)
	require.NoError(t, err)

	paramsInfo, err := EncodeParams(&param)
	require.NoError(t, err)

	sign, err := EncodeSignature(signedMsg.Signature)
	require.NoError(t, err)

	mySignedMsg := SignedMessage{
		Message: Message{
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
		},
		Signature: sign,
	}
	b, err := json.Marshal(mySignedMsg)
	require.NoError(t, err)
	t.Log(string(b))

	tSignMsg, err := BuildSignedMessage(&mySignedMsg)
	require.NoError(t, err)

	var mySignedMsgBuf bytes.Buffer
	err = tSignMsg.MarshalCBOR(&mySignedMsgBuf)
	require.NoError(t, err)

	require.True(t, reflect.DeepEqual(mySignedMsgBuf.Bytes(), signedMsgBuf.Bytes()))
}
