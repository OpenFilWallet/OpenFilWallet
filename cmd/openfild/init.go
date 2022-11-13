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

		repoPath := cctx.String(flagWalletRepo)
		r, err := repo.NewFS(repoPath)
		if err != nil {
			return err
		}

		ok, err := r.Exists()
		if err != nil {
			return err
		}
		if ok {
			return xerrors.Errorf("repo at '%s' is already initialized", cctx.String(flagWalletRepo))
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

		rootPasswordExist, err := db.HasRootPassword()
		if err != nil {
			return err
		}

		if rootPasswordExist {
			return errors.New("root password exists")
		}

		loginPasswordExist, err := db.HasLoginPassword()
		if err != nil {
			return err
		}

		if loginPasswordExist {
			return errors.New("login password exists")
		}

		fmt.Println("Please enter the root password to encrypt the mnemonic and private key")
		rootPassword, err := app.Password(true)
		if err != nil {
			return err
		}

		rootScrypt := crypto.Scrypt(rootPassword)

		fmt.Println("Please enter the login password to login to the wallet")
		loginPassword, err := app.Password(true)
		if err != nil {
			return err
		}

		loginScrypt := crypto.Scrypt(loginPassword)

		err = db.SetRootPassword(rootScrypt)
		if err != nil {
			return err
		}
		err = db.SetLoginPassword(loginScrypt)
		if err != nil {
			return err
		}

		log.Info("openFilWallet initialized successfully, you can now start it with 'openfild mnemonic generate or openfild mnemonic import'")

		return nil
	},
}
