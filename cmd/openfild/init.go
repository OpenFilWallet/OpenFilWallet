package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/OpenFilWallet/OpenFilWallet/repo"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var initCmd = &cli.Command{
	Name:  "init",
	Usage: "Initialize a OpenFilWallet repo",
	Action: func(cctx *cli.Context) error {
		log.Info("Initializing OpenFilWallet")

		repoPath := cctx.String(repo.FlagWalletRepo)
		r, err := repo.NewFS(repoPath)
		if err != nil {
			return err
		}

		ok, err := r.Exists()
		if err != nil {
			return err
		}
		if ok {
			return xerrors.Errorf("repo at '%s' is already initialized", cctx.String(repo.FlagWalletRepo))
		}

		log.Info("Initializing repo")

		if err := r.Init(); err != nil {
			return err
		}

		lr, err := r.Lock()
		if err != nil {
			return err
		}

		ds, err := lr.Datastore(context.Background())
		if err != nil {
			return err
		}

		db := datastore.NewWalletDB(ds)

		masterPasswordExist, err := db.HasMasterPassword()
		if err != nil {
			return err
		}

		if masterPasswordExist {
			return errors.New("master password exists")
		}

		loginPasswordExist, err := db.HasLoginPassword()
		if err != nil {
			return err
		}

		if loginPasswordExist {
			return errors.New("login password exists")
		}

		fmt.Println("Please enter the master password to encrypt the mnemonic and private key")
		masterPassword, err := app.Password(true)
		if err != nil {
			return err
		}

		masterScrypt := crypto.Scrypt(masterPassword)

		fmt.Println("Please enter the login password to login to the wallet")
		loginPassword, err := app.Password(true)
		if err != nil {
			return err
		}

		loginScrypt := crypto.Scrypt(loginPassword)

		err = db.SetMasterPassword(masterScrypt)
		if err != nil {
			return err
		}
		err = db.SetLoginPassword(loginScrypt)
		if err != nil {
			return err
		}

		fmt.Println("openFilWallet initialized successfully, you can now start it with 'openfild mnemonic generate or openfild mnemonic import'")

		return nil
	},
}
