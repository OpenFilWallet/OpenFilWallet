package account

import (
	"crypto/ecdsa"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/OpenFilWallet/OpenFilWallet/lib/hd"
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	_ "github.com/filecoin-project/lotus/lib/sigs/bls"
	_ "github.com/filecoin-project/lotus/lib/sigs/secp"
	"strings"
)

type EthKey struct {
	PriKey  *ecdsa.PrivateKey
	Address common.Address
}

func GenerateEthPrivateKeyFromMnemonicIndex(walletDB datastore.WalletDB, mnemonic string, index int64, passwordKey []byte) (*EthKey, error) {
	seed, err := hd.GenerateSeedFromMnemonic(mnemonic, "")
	if err != nil {
		return nil, err
	}

	if index == -1 {
		i, err := walletDB.NextMnemonicEthIndex()
		if err != nil {
			return nil, err
		}
		index = int64(i)
	}

	log.Debugw("GenerateEthPrivateKeyFromMnemonicIndex", "index", index)

	path := hd.EthPath(uint64(index))
	extendSeed, err := hd.GetExtendSeedFromPath(path, seed)
	if err != nil {
		return nil, err
	}

	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), extendSeed)
	privateKeyECDSA := privateKey.ToECDSA()
	priKey := ethcrypto.FromECDSA(privateKeyECDSA)
	encryptedPrivateKey, err := crypto.Encrypt(priKey, passwordKey)
	if err != nil {
		return nil, err
	}

	addr := ethcrypto.PubkeyToAddress(privateKeyECDSA.PublicKey)

	err = walletDB.SetEthPrivate(&datastore.PrivateWallet{
		PriKey:  encryptedPrivateKey,
		Address: addr.String(),
		KeyHash: crypto.Hash256(encryptedPrivateKey),
		Path:    path,
	})

	if err != nil {
		return nil, err
	}

	return &EthKey{
		PriKey:  privateKeyECDSA,
		Address: addr,
	}, nil
}

func ImportEthPrivateKey(walletDB datastore.WalletDB, priKey string, passwordKey []byte) error {
	log.Debug("ImportEthPrivateKey")
	priKey = strings.Replace(priKey, "\n", "", -1)
	privateKeyECDSA, err := ethcrypto.HexToECDSA(priKey)
	if err != nil {
		return err
	}

	ethKey := ethcrypto.FromECDSA(privateKeyECDSA)

	encryptedPrivateKey, err := crypto.Encrypt(ethKey, passwordKey)
	if err != nil {
		return err
	}

	return walletDB.SetEthPrivate(&datastore.PrivateWallet{
		PriKey:  encryptedPrivateKey,
		Address: ethcrypto.PubkeyToAddress(privateKeyECDSA.PublicKey).String(),
		KeyHash: crypto.Hash256(encryptedPrivateKey),
		Path:    "Import",
	})
}

func UpdateEthPrivateKey(walletDB datastore.WalletDB, oldPasswordKey, newPasswordKey []byte) error {
	privateWallets, err := walletDB.EthWalletList()
	if err != nil {
		return err
	}

	for _, pri := range privateWallets {
		decryptKey, err := crypto.Decrypt(pri.PriKey, oldPasswordKey)
		if err != nil {
			return err
		}

		encryptedPrivateKey, err := crypto.Encrypt(decryptKey, newPasswordKey)
		if err != nil {
			return err
		}

		err = walletDB.UpdateEthPrivate(&datastore.PrivateWallet{
			PriKey:  encryptedPrivateKey,
			Address: pri.Address,
			KeyHash: crypto.Hash256(encryptedPrivateKey),
			Path:    pri.Path,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func GetEthPrivateKey(walletDB datastore.WalletDB, addr string, passwordKey []byte) (*EthKey, error) {
	pri, err := walletDB.GetEthPrivate(addr)
	if err != nil {
		return nil, err
	}

	decryptKey, err := crypto.Decrypt(pri.PriKey, passwordKey)
	if err != nil {
		return nil, err
	}

	privateKeyECDSA, err := ethcrypto.ToECDSA(decryptKey)
	if err != nil {
		return nil, err
	}

	return &EthKey{
		PriKey:  privateKeyECDSA,
		Address: ethcrypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
	}, nil
}

func LoadEthPrivateKeys(walletDB datastore.WalletDB, passwordKey []byte) ([]EthKey, error) {
	privateWallets, err := walletDB.EthWalletList()
	if err != nil {
		return nil, err
	}

	var keys = make([]EthKey, 0)
	for _, pri := range privateWallets {
		decryptKey, err := crypto.Decrypt(pri.PriKey, passwordKey)
		if err != nil {
			return nil, err
		}

		privateKeyECDSA, err := ethcrypto.ToECDSA(decryptKey)
		if err != nil {
			return nil, err
		}

		keys = append(keys, EthKey{
			PriKey:  privateKeyECDSA,
			Address: ethcrypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		})
	}

	return keys, nil
}
