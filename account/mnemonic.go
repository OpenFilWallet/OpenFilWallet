package account

import (
	"crypto/subtle"
	"errors"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/OpenFilWallet/OpenFilWallet/lib/hd"
)

func GenerateMnemonic(walletDB datastore.WalletDB, mType hd.MnemonicType, passwordKey []byte) error {
	mnemonic, err := hd.NewMnemonic(mType)
	if err != nil {
		return err
	}

	encryptMnemonic, err := crypto.Encrypt([]byte(mnemonic), passwordKey)
	if err != nil {
		return err
	}

	err = walletDB.SetMnemonic(&datastore.HdWallet{
		Mnemonic:     encryptMnemonic,
		MnemonicHash: crypto.Hash256(encryptMnemonic),
	})
	if err != nil {
		return err
	}

	return nil
}

func ImportMnemonic(walletDB datastore.WalletDB, mnemonic string, passwordKey []byte) error {
	// check mnemonic
	good := hd.CheckMnemonic(mnemonic)
	if !good {
		return errors.New("invalid mnemonic")
	}

	encryptedMnemonic, err := crypto.Encrypt([]byte(mnemonic), passwordKey)
	if err != nil {
		return err
	}

	return walletDB.SetMnemonic(&datastore.HdWallet{
		Mnemonic:     encryptedMnemonic,
		MnemonicHash: crypto.Hash256(encryptedMnemonic),
	})
}

func UpdateMnemonic(walletDB datastore.WalletDB, oldPasswordKey, newPasswordKey []byte) error {
	hdWallet, err := walletDB.GetMnemonic()
	if err != nil {
		return err
	}

	mnemonic, err := crypto.Decrypt(hdWallet.Mnemonic, oldPasswordKey)
	if err != nil {
		return err
	}

	if subtle.ConstantTimeCompare(hdWallet.MnemonicHash, crypto.Hash256(hdWallet.Mnemonic)) != 1 {
		err = errors.New("warning, abnormal mnemonic check, possible data corruption")
	}

	encryptMnemonic, err := crypto.Encrypt([]byte(mnemonic), newPasswordKey)
	if err != nil {
		return err
	}

	err = walletDB.UpdateMnemonic(&datastore.HdWallet{
		Mnemonic:     encryptMnemonic,
		MnemonicHash: crypto.Hash256(encryptMnemonic),
	})
	if err != nil {
		return err
	}

	return nil
}

func LoadMnemonic(walletDB datastore.WalletDB, passwordKey []byte) (string, error) {
	hdWallet, err := walletDB.GetMnemonic()
	if err != nil {
		return "", err
	}

	mnemonic, err := crypto.Decrypt(hdWallet.Mnemonic, passwordKey)
	if err != nil {
		return "", err
	}

	if subtle.ConstantTimeCompare(hdWallet.MnemonicHash, crypto.Hash256(hdWallet.Mnemonic)) != 1 {
		err = errors.New("warning, abnormal mnemonic check, possible data corruption")
	}

	return string(mnemonic), err
}
