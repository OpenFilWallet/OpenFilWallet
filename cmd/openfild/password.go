package main

import (
	"errors"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/urfave/cli/v2"
)

var passwordCmd = &cli.Command{
	Name:  "password",
	Usage: "OpenFilWallet password settings",
	Subcommands: []*cli.Command{
		updateMasterCmd,
		updateLoginCmd,
	},
}

var updateMasterCmd = &cli.Command{
	Name:  "update-master",
	Usage: "update master password",
	Action: func(cctx *cli.Context) error {
		db, closer, err := getWalletDB(cctx, false)
		if err != nil {
			return err
		}
		defer closer()

		if err := requirePassword(db); err != nil {
			return err
		}

		oldPassword, verified := verifyMasterPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		fmt.Println("Please enter a new master password")
		masterPassword, err := app.Password(true)
		if err != nil {
			return err
		}

		masterScrypt := crypto.Scrypt(masterPassword)

		err = db.UpdateMasterPassword(masterScrypt)
		if err != nil {
			return err
		}

		newPasswordKey := crypto.GenerateEncryptKey([]byte(masterPassword))
		oldPasswordKey := crypto.GenerateEncryptKey([]byte(oldPassword))

		// Encrypt the mnemonic with the new key
		err = account.UpdateMnemonic(db, oldPasswordKey, newPasswordKey)
		if err != nil {
			return err
		}

		// Encrypt the private key with the new key
		err = account.UpdatePrivateKey(db, oldPasswordKey, newPasswordKey)
		if err != nil {
			return err
		}

		fmt.Println("master password updated successfully")

		return nil
	},
}

var updateLoginCmd = &cli.Command{
	Name:  "update-login",
	Usage: "update login password, need master password",
	Action: func(cctx *cli.Context) error {
		db, closer, err := getWalletDB(cctx, false)
		if err != nil {
			return err
		}
		defer closer()

		if err := requirePassword(db); err != nil {
			return err
		}

		_, verified := verifyMasterPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		fmt.Println("Please enter a new login password")
		loginPassword, err := app.Password(true)
		if err != nil {
			return err
		}

		loginScrypt := crypto.Scrypt(loginPassword)
		err = db.UpdateLoginPassword(loginScrypt)
		if err != nil {
			return err
		}

		return nil
	},
}
