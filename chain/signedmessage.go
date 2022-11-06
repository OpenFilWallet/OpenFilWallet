package chain

import (
	"bytes"
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

func BuildSignedMessage(signedMsg *SignedMessage) (*types.SignedMessage, error) {
	tMsg, err := BuildMessage(&signedMsg.Message)
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
	buf := bytes.NewBufferString(signature)
	err := sign.UnmarshalCBOR(buf)
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

	return buf.String(), nil
}
