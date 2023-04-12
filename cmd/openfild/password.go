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
		updateRootCmd,
		updateLoginCmd,
	},
}

var updateRootCmd = &cli.Command{
	Name:  "update-root",
	Usage: "update root password",
	Action: func(cctx *cli.Context) error {
		db, closer, err := getWalletDB(cctx, false)
		if err != nil {
			return err
		}
		defer closer()

		if err := requirePassword(db); err != nil {
			return err
		}

		oldPassword, verified := verifyRootPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		fmt.Println("Please enter a new root password")
		rootPassword, err := app.Password(true)
		if err != nil {
			return err
		}

		rootScrypt := crypto.Scrypt(rootPassword)

		err = db.UpdateRootPassword(rootScrypt)
		if err != nil {
			return err
		}

		newPasswordKey := crypto.GenerateEncryptKey([]byte(rootPassword))
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

		fmt.Println("root password updated successfully")

		return nil
	},
}

var updateLoginCmd = &cli.Command{
	Name:  "update-login",
	Usage: "update login password, need root password",
	Action: func(cctx *cli.Context) error {
		db, closer, err := getWalletDB(cctx, false)
		if err != nil {
			return err
		}
		defer closer()

		if err := requirePassword(db); err != nil {
			return err
		}

		_, verified := verifyRootPassword(db)
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
