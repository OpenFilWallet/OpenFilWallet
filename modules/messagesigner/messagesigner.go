package messagesigner

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/OpenFilWallet/OpenFilWallet/lib/sigs"
	_ "github.com/OpenFilWallet/OpenFilWallet/lib/sigs/bls"
	_ "github.com/OpenFilWallet/OpenFilWallet/lib/sigs/secp"
	"github.com/OpenFilWallet/OpenFilWallet/modules/buildmessage"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/wallet/key"
	_ "github.com/filecoin-project/lotus/lib/sigs/bls"
	_ "github.com/filecoin-project/lotus/lib/sigs/secp"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"
	"math/big"
	"sync"
)

var log = logging.Logger("buildmessage")

type Signer interface {
	RegisterSigner(...key.Key) error
	RegisterEthSigner(...account.EthKey) error
	SignMsg(msg *types.Message) (*types.SignedMessage, error)
	SignTx(sender string, tx *ethtypes.Transaction) (*ethtypes.Transaction, error)
	Sign(from string, data []byte) ([]byte, error)
	HasSigner(addr string) bool
}

type SignerHouse struct {
	signers    map[string]key.Key // key is address
	ethSigners map[string]account.EthKey
	lk         sync.Mutex
}

func NewSigner() Signer {
	return &SignerHouse{
		signers:    map[string]key.Key{},
		ethSigners: map[string]account.EthKey{},
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
		log.Infow("RegisterSigner", "address", key.Address.String())
	}

	return nil
}

func (s *SignerHouse) RegisterEthSigner(keys ...account.EthKey) error {
	s.lk.Lock()
	defer s.lk.Unlock()

	for _, key := range keys {
		if _, ok := s.ethSigners[key.Address.String()]; ok {
			return fmt.Errorf("wallet: %s already exist", key.Address.String())
		}

		s.ethSigners[key.Address.String()] = key
		log.Infow("RegisterEthSigner", "address", key.Address.String())
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

	log.Infow("SignMsg", "message", buildmessage.LotusMessageToString(msg))
	return &types.SignedMessage{
		Message:   *msg,
		Signature: *sig,
	}, nil
}

func (s *SignerHouse) SignTx(sender string, transaction *ethtypes.Transaction) (*ethtypes.Transaction, error) {
	s.lk.Lock()
	defer s.lk.Unlock()
	sender = common.HexToAddress(sender).String()
	key, ok := s.ethSigners[sender]
	if !ok {
		return nil, fmt.Errorf("wallet: %s does not exist", sender)
	}

	// filecoin mainnet chainid 314
	signer := ethtypes.NewLondonSigner(big.NewInt(314))

	var err error
	transaction, err = ethtypes.SignTx(transaction, signer, key.PriKey)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (s *SignerHouse) Sign(from string, data []byte) ([]byte, error) {
	s.lk.Lock()
	defer s.lk.Unlock()
	var sigBytes []byte
	signer, ok := s.signers[from]
	if ok {
		sig, err := sigs.Sign(key.ActSigType(signer.Type), signer.PrivateKey, data)
		if err != nil {
			return nil, xerrors.Errorf("failed to sign message: %w", err)
		}
		sigBytes = append([]byte{byte(sig.Type)}, sig.Data...)
	} else {
		ethSigner, ok := s.ethSigners[from]
		if ok {
			prefix := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(data)))
			prefixPack := [][]byte{prefix, data}
			msg := crypto.Keccak256(bytes.Join(prefixPack, nil))

			sig, err := crypto.Sign(msg, ethSigner.PriKey)
			if err != nil {
				return nil, err
			}

			v := new(big.Int).SetBytes([]byte{sig[64]})
			v = big.NewInt(0).Add(v, big.NewInt(27))
			sig[64] = v.Bytes()[0]
			copy(sigBytes, sig)
		} else {
			return nil, fmt.Errorf("wallet: %s does not exist", from)
		}
	}

	log.Infow("Sign", "data", hex.EncodeToString(sigBytes))
	return sigBytes, nil
}

func (s *SignerHouse) HasSigner(addr string) bool {
	s.lk.Lock()
	defer s.lk.Unlock()

	_, ok := s.signers[addr]
	_, ok2 := s.ethSigners[addr]
	if ok || ok2 {
		return true
	}

	return false
}
