package account

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/OpenFilWallet/OpenFilWallet/lib/hd"
	"github.com/OpenFilWallet/OpenFilWallet/lib/sigs"
	_ "github.com/OpenFilWallet/OpenFilWallet/lib/sigs/bls"
	_ "github.com/OpenFilWallet/OpenFilWallet/lib/sigs/secp"
	filcrypto "github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/wallet"

	"golang.org/x/xerrors"
	"strings"
)

func GeneratePrivateKey(walletDB datastore.WalletDB, mnemonic string, sigType filcrypto.SigType, passwordKey []byte) (*wallet.Key, error) {
	keyType, err := sigType.Name()
	if err != nil {
		return nil, err
	}

	seed, err := hd.GenerateSeedFromMnemonic(mnemonic, "")
	if err != nil {
		return nil, err
	}

	index, err := walletDB.NextMnemonicIndex()
	if err != nil {
		return nil, err
	}

	path := hd.FILPath(index)
	extendSeed, err := hd.GetExtendSeedFromPath(path, seed)
	if err != nil {
		return nil, err
	}

	pk, err := sigs.Generate(sigType, extendSeed)
	if err != nil {
		return nil, err
	}

	ki := types.KeyInfo{
		Type:       types.KeyType(keyType),
		PrivateKey: pk,
	}

	privateKey, err := json.Marshal(ki)
	if err != nil {
		return nil, err
	}

	encryptedPrivateKey, err := crypto.Encrypt(privateKey, passwordKey)
	if err != nil {
		return nil, err
	}

	nk, err := wallet.NewKey(ki)
	if err != nil {
		return nil, err
	}

	err = walletDB.SetPrivate(&datastore.PrivateWallet{
		PriKey:  encryptedPrivateKey,
		Address: nk.Address.String(),
		KeyHash: crypto.Hash256(encryptedPrivateKey),
	})

	if err != nil {
		return nil, err
	}

	return nk, nil
}

func GeneratePrivateKeyFromIndex(walletDB datastore.WalletDB, mnemonic string, index uint64, sigType filcrypto.SigType, passwordKey []byte) (*wallet.Key, error) {
	keyType, err := sigType.Name()
	if err != nil {
		return nil, err
	}

	seed, err := hd.GenerateSeedFromMnemonic(mnemonic, "")
	if err != nil {
		return nil, err
	}

	path := hd.FILPath(index)
	extendSeed, err := hd.GetExtendSeedFromPath(path, seed)
	if err != nil {
		return nil, err
	}

	pk, err := sigs.Generate(sigType, extendSeed)
	if err != nil {
		return nil, err
	}

	ki := types.KeyInfo{
		Type:       types.KeyType(keyType),
		PrivateKey: pk,
	}

	privateKey, err := json.Marshal(ki)
	if err != nil {
		return nil, err
	}

	encryptedPrivateKey, err := crypto.Encrypt(privateKey, passwordKey)
	if err != nil {
		return nil, err
	}

	nk, err := wallet.NewKey(ki)
	if err != nil {
		return nil, err
	}

	err = walletDB.SetPrivate(&datastore.PrivateWallet{
		PriKey:  encryptedPrivateKey,
		Address: nk.Address.String(),
		KeyHash: crypto.Hash256(encryptedPrivateKey),
	})

	if err != nil {
		return nil, err
	}

	return nk, nil
}

func ImportPrivateKey(walletDB datastore.WalletDB, priKey, keyFormat string, passwordKey []byte) error {
	ki, err := GenerateKeyInfoFromPriKey(priKey, keyFormat)
	if err != nil {
		return err
	}

	privateKey, err := json.Marshal(ki)
	if err != nil {
		return err
	}

	encryptedPrivateKey, err := crypto.Encrypt(privateKey, passwordKey)
	if err != nil {
		return err
	}

	nk, err := wallet.NewKey(*ki)
	if err != nil {
		return err
	}

	return walletDB.SetPrivate(&datastore.PrivateWallet{
		PriKey:  encryptedPrivateKey,
		Address: nk.Address.String(),
		KeyHash: crypto.Hash256(encryptedPrivateKey),
	})
}

func UpdatePrivateKey(walletDB datastore.WalletDB, oldPasswordKey, newPasswordKey []byte) error {
	privateWallets, err := walletDB.WalletList()
	if err != nil {
		return err
	}

	for _, pri := range privateWallets {
		var ki types.KeyInfo

		decryptKey, err := crypto.Decrypt(pri.PriKey, oldPasswordKey)
		if err != nil {
			return err
		}

		err = json.Unmarshal(decryptKey, &ki)
		if err != nil {
			return err
		}

		encryptedPrivateKey, err := crypto.Encrypt(decryptKey, newPasswordKey)
		if err != nil {
			return err
		}

		nk, err := wallet.NewKey(ki)
		if err != nil {
			return err
		}

		err = walletDB.UpdatePrivate(&datastore.PrivateWallet{
			PriKey:  encryptedPrivateKey,
			Address: nk.Address.String(),
			KeyHash: crypto.Hash256(encryptedPrivateKey),
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func GetPrivateKey(walletDB datastore.WalletDB, addr string, passwordKey []byte) (wallet.Key, error) {
	pri, err := walletDB.GetPrivate(addr)
	if err != nil {
		return wallet.Key{}, err
	}

	var ki types.KeyInfo

	decryptKey, err := crypto.Decrypt(pri.PriKey, passwordKey)
	if err != nil {
		return wallet.Key{}, err
	}

	err = json.Unmarshal(decryptKey, &ki)
	if err != nil {
		return wallet.Key{}, err
	}

	nk, err := wallet.NewKey(ki)
	if err != nil {
		return wallet.Key{}, err
	}

	return *nk, nil
}

func LoadPrivateKeys(walletDB datastore.WalletDB, passwordKey []byte) ([]wallet.Key, error) {
	privateWallets, err := walletDB.WalletList()
	if err != nil {
		return nil, err
	}

	var keys = make([]wallet.Key, 0)
	for _, pri := range privateWallets {
		var ki types.KeyInfo

		decryptKey, err := crypto.Decrypt(pri.PriKey, passwordKey)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(decryptKey, &ki)
		if err != nil {
			return nil, err
		}

		nk, err := wallet.NewKey(ki)
		if err != nil {
			return nil, err
		}

		keys = append(keys, *nk)
	}

	return keys, nil
}

func GenerateKeyInfoFromPriKey(priKey, keyFormat string) (*types.KeyInfo, error) {
	var ki types.KeyInfo
	switch keyFormat {
	case "hex-lotus":
		data, err := hex.DecodeString(strings.TrimSpace(priKey))
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, &ki); err != nil {
			return nil, err
		}
	case "json-lotus":
		if err := json.Unmarshal([]byte(priKey), &ki); err != nil {
			return nil, err
		}
	case "gfc-json":
		var f struct {
			KeyInfo []struct {
				PrivateKey []byte
				SigType    int
			}
		}
		if err := json.Unmarshal([]byte(priKey), &f); err != nil {
			return nil, xerrors.Errorf("failed to parse go-filecoin key: %s", err)
		}

		gk := f.KeyInfo[0]
		ki.PrivateKey = gk.PrivateKey
		switch gk.SigType {
		case 1:
			ki.Type = types.KTSecp256k1
		case 2:
			ki.Type = types.KTBLS
		default:
			return nil, fmt.Errorf("unrecognized key type: %d", gk.SigType)
		}
	default:
		return nil, fmt.Errorf("unrecognized format: %s", keyFormat)
	}

	return &ki, nil
}
