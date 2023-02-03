package chain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/chain/types"
)

// SignedMessage : Signature use CBOR encoding and conforms to lotus spec
//   sign := crypto.Signature{
//	   Type: crypto.SigTypeSecp256k1,
//	   Data: []byte{1, 2, 3, 4},
//   }
//
//   var buf bytes.Buffer
//   sign.MarshalCBOR(&buf)
//   signedMsg := SignedMessage{
//	   Signature: buf.String(),
//   }
type SignedMessage struct {
	Message   Message `json:"message"`
	Signature string  `json:"signature"`
}

func (m *SignedMessage) String() string {
	signedMessage, _ := json.Marshal(m)
	return string(signedMessage)
}

func BuildSignedMessage(msg *Message, signature crypto.Signature) (*SignedMessage, error) {
	sign, err := EncodeSignature(signature)
	if err != nil {
		return nil, err
	}

	return &SignedMessage{
		Message:   *msg,
		Signature: sign,
	}, nil
}

func DecodeSignedMessage(signedMsg *SignedMessage) (*types.SignedMessage, error) {
	tMsg, err := DecodeMessage(&signedMsg.Message)
	if err != nil {
		return nil, err
	}
	sign, err := DecodeSignature(signedMsg.Signature)
	if err != nil {
		return nil, err
	}

	return &types.SignedMessage{
		Message:   *tMsg,
		Signature: sign,
	}, nil
}

func DecodeSignature(signature string) (crypto.Signature, error) {
	var sign crypto.Signature
	signByte, err := hex.DecodeString(signature)
	if err != nil {
		return crypto.Signature{}, err
	}
	err = sign.UnmarshalCBOR(bytes.NewReader(signByte))
	if err != nil {
		return crypto.Signature{}, err
	}

	return sign, err
}

func EncodeSignature(signature crypto.Signature) (string, error) {
	var buf bytes.Buffer
	err := signature.MarshalCBOR(&buf)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(buf.Bytes()), nil
}
