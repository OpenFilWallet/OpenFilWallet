package messagesigner

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/lib/sigs"
	_ "github.com/OpenFilWallet/OpenFilWallet/lib/sigs/bls"
	_ "github.com/OpenFilWallet/OpenFilWallet/lib/sigs/secp"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/wallet/key"
	_ "github.com/filecoin-project/lotus/lib/sigs/bls"
	_ "github.com/filecoin-project/lotus/lib/sigs/secp"
	"golang.org/x/xerrors"
	"sync"
)

type Signer interface {
	RegisterSigner(...key.Key) error
	SignMsg(msg *types.Message) (*types.SignedMessage, error)
	Sign(from string, data []byte) (*crypto.Signature, error)
}

type SignerHouse struct {
	signers map[string]key.Key // key is address
	lk      sync.Mutex
}

func NewSigner() Signer {
	return &SignerHouse{
		signers: map[string]key.Key{},
	}
}

func (s *SignerHouse) RegisterSigner(keys ...key.Key) error {
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

func (s *SignerHouse) SignMsg(msg *types.Message) (*types.SignedMessage, error) {
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

	sig, err := sigs.Sign(key.ActSigType(signer.Type), signer.PrivateKey, mb.Cid().Bytes())
	if err != nil {
		return nil, xerrors.Errorf("failed to sign message: %w", err)
	}

	return &types.SignedMessage{
		Message:   *msg,
		Signature: *sig,
	}, nil
}

func (s *SignerHouse) Sign(from string, data []byte) (*crypto.Signature, error) {
	s.lk.Lock()
	defer s.lk.Unlock()

	signer, ok := s.signers[from]
	if !ok {
		return nil, fmt.Errorf("wallet: %s does not exist", from)
	}

	sig, err := sigs.Sign(key.ActSigType(signer.Type), signer.PrivateKey, data)
	if err != nil {
		return nil, xerrors.Errorf("failed to sign message: %w", err)
	}

	return sig, nil
}
