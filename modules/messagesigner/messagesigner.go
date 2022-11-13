package messagesigner

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/lib/sigs"
	_ "github.com/OpenFilWallet/OpenFilWallet/lib/sigs/bls"
	_ "github.com/OpenFilWallet/OpenFilWallet/lib/sigs/secp"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/wallet"
	"golang.org/x/xerrors"
	"sync"
)

type Signer interface {
	RegisterSigner(...wallet.Key) error
	Sign(msg *types.Message) (*types.SignedMessage, error)
}

type SignerHouse struct {
	signers map[string]wallet.Key // key is address
	lk      sync.Mutex
}

func NewSigner() Signer {
	return &SignerHouse{
		signers: map[string]wallet.Key{},
	}
}

func (s *SignerHouse) RegisterSigner(keys ...wallet.Key) error {
	s.lk.Lock()
	defer s.lk.Unlock()

	for _, key := range keys {
		if _, ok := s.signers[key.Address.String()]; ok {
			return fmt.Errorf("wallet: %s already exist", key.Address.String())
		}

		s.signers[key.Address.String()] = key
	}

	return nil
}

func (s *SignerHouse) Sign(msg *types.Message) (*types.SignedMessage, error) {
	s.lk.Lock()
	defer s.lk.Unlock()

	signer, ok := s.signers[msg.From.String()]
	if !ok {
		return nil, fmt.Errorf("wallet: %s does not exist", msg.From.String())
	}

	mb, err := msg.ToStorageBlock()
	if err != nil {
		return nil, xerrors.Errorf("serializing message: %w", err)
	}

	sig, err := sigs.Sign(wallet.ActSigType(signer.Type), signer.PrivateKey, mb.Cid().Bytes())
	if err != nil {
		return nil, xerrors.Errorf("failed to sign message: %w", err)
	}

	return &types.SignedMessage{
		Message:   *msg,
		Signature: *sig,
	}, nil
}
