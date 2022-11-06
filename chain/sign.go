package chain

import (
	"github.com/OpenFilWallet/OpenFilWallet/lib/sigs"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/wallet/key"
	"golang.org/x/xerrors"
)

func SignMessage(account key.Key, msg *types.Message) (*types.SignedMessage, error) {
	mb, err := msg.ToStorageBlock()
	if err != nil {
		return nil, xerrors.Errorf("serializing message: %w", err)
	}

	sig, err := sigs.Sign(key.ActSigType(account.Type), account.PrivateKey, mb.Cid().Bytes())
	if err != nil {
		return nil, xerrors.Errorf("failed to sign message: %w", err)
	}

	return &types.SignedMessage{
		Message:   *msg,
		Signature: *sig,
	}, nil
}
