package secp

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/lib/sigs"
	"github.com/btcsuite/btcd/btcec"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-crypto"
	crypto2 "github.com/filecoin-project/go-state-types/crypto"
	"github.com/minio/blake2b-simd"
)

type secpSigner struct{}

func (secpSigner) GenPrivate(seed []byte) ([]byte, error) {
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), seed)
	privateKeyECDSA := privateKey.ToECDSA()

	privkey := make([]byte, crypto.PrivateKeyBytes)
	blob := privateKeyECDSA.D.Bytes()

	// the length is guaranteed to be fixed, given the serialization rules for secp2561k curve points.
	copy(privkey[crypto.PrivateKeyBytes-len(blob):], blob)

	return privkey, nil
}

func (secpSigner) ToPublic(pk []byte) ([]byte, error) {
	return crypto.PublicKey(pk), nil
}

func (secpSigner) Sign(pk []byte, msg []byte) ([]byte, error) {
	b2sum := blake2b.Sum256(msg)
	sig, err := crypto.Sign(pk, b2sum[:])
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (secpSigner) Verify(sig []byte, a address.Address, msg []byte) error {
	b2sum := blake2b.Sum256(msg)
	pubk, err := crypto.EcRecover(b2sum[:], sig)
	if err != nil {
		return err
	}

	maybeaddr, err := address.NewSecp256k1Address(pubk)
	if err != nil {
		return err
	}

	if a != maybeaddr {
		return fmt.Errorf("signature did not match")
	}

	return nil
}

func init() {
	sigs.RegisterSignature(crypto2.SigTypeSecp256k1, secpSigner{})
}
