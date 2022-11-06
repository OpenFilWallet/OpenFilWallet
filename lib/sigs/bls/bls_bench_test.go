package bls

import (
	"crypto/rand"
	"testing"

	"github.com/OpenFilWallet/OpenFilWallet/lib/hd"
	"github.com/filecoin-project/go-address"
)

func BenchmarkBLSSign(b *testing.B) {
	signer := blsSigner{}
	mnemonic, err := hd.NewMnemonic(hd.Mnemonic12)
	if err != nil {
		b.Fatal(err)
	}
	seed, err := hd.GenerateSeedFromMnemonic(mnemonic, "")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		pk, _ := signer.GenPrivate(seed)
		randMsg := make([]byte, 32)
		_, _ = rand.Read(randMsg)
		b.StartTimer()

		_, _ = signer.Sign(pk, randMsg)
	}
}

func BenchmarkBLSVerify(b *testing.B) {
	signer := blsSigner{}
	mnemonic, err := hd.NewMnemonic(hd.Mnemonic24)
	if err != nil {
		b.Fatal(err)
	}
	seed, err := hd.GenerateSeedFromMnemonic(mnemonic, "")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		randMsg := make([]byte, 32)
		_, _ = rand.Read(randMsg)

		priv, _ := signer.GenPrivate(seed)
		pk, _ := signer.ToPublic(priv)
		addr, _ := address.NewBLSAddress(pk)
		sig, _ := signer.Sign(priv, randMsg)

		b.StartTimer()

		_ = signer.Verify(sig, addr, randMsg)
	}
}
