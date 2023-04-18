package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/build"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/OpenFilWallet/OpenFilWallet/repo"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"os"
)

var log = logging.Logger("openfild")

func main() {
	app := &cli.App{
		Name:                 "openfild",
		Usage:                "open source hd wallet for Filecoin",
		Version:              build.Version(),
		EnableBashCompletion: true,
		Before:               before,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    repo.FlagWalletRepo,
				EnvVars: []string{"OPEN_FIL_WALLET_PATH"},
				Value:   "~/.openfilwallet",
				Usage:   fmt.Sprintf("Specify openfilwallet repo path. flag(--wallet-repo) or env(OPEN_FIL_WALLET_PATH)"),
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Value:   false,
				Aliases: []string{"vv"},
				Usage:   "enables very verbose mode",
			},
		},
		Commands: []*cli.Command{
			initCmd,
			authCmd,
			runCmd,
			mnemonicCmd,
			walletCmd,
			passwordCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString("Error: " + err.Error() + "\n")
	}
}

func verifyMasterPassword(db datastore.WalletDB) (string, bool) {
	fmt.Println("Please enter master password")
	for i := 0; i < 3; i++ {
		masterPassword, err := app.Password(false)
		if err != nil {
			continue
		}

		masterKey, err := db.GetMasterPassword()
		if err != nil {
			log.Warnw("walletDB load encrypted password failed", "err", err)
			return "", false
		}

		// check password
		isOk, err := crypto.VerifyScrypt(masterPassword, masterKey)
		if err != nil || !isOk {
			if i == 2 {
				fmt.Println("Incorrect password")
				continue
			}

			fmt.Printf("Incorrect password, please try again. You can retry %d times\n", 2-i)
			continue
		}

		return masterPassword, true
	}

	return "", false
}

func getWalletDB(cctx *cli.Context, readonly bool) (datastore.WalletDB, func(), error) {
	repoPath := cctx.String(repo.FlagWalletRepo)
	r, err := repo.NewFS(repoPath)
	if err != nil {
		return datastore.WalletDB{}, nil, err
	}

	ok, err := r.Exists()
	if err != nil {
		return datastore.WalletDB{}, nil, err
	}
	if !ok {
		return datastore.WalletDB{}, nil, xerrors.Errorf("repo at '%s' is not initialized, run 'openfild init' to set it up", repo.FlagWalletRepo)
	}

	var lr repo.LockedRepo
	if readonly {
		lr, err = r.LockRO()
		if err != nil {
			return datastore.WalletDB{}, nil, err
		}
	} else {
		lr, err = r.Lock()
		if err != nil {
			return datastore.WalletDB{}, nil, err
		}
	}

	ds, err := lr.Datastore(context.Background())
	if err != nil {
		return datastore.WalletDB{}, nil, err
	}

	return datastore.NewWalletDB(ds), func() {
		if readonly {
			err = lr.CloseRO()
		} else {
			err = lr.Close()
		}

		if err != nil {
			log.Warnw("DB close fail", "err", err.Error())
		}
	}, nil
}

func requirePassword(db datastore.WalletDB) error {
	ok, err := db.HasMasterPassword()
	if err != nil {
		return fmt.Errorf("master password check failed, err: %s", err.Error())
	}
	if !ok {
		return errors.New("master password does not exist")
	}

	ok, err = db.HasLoginPassword()
	if err != nil {
		return fmt.Errorf("login password check failed, err: %s", err.Error())
	}
	if !ok {
		return errors.New("login password does not exist")
	}

	return nil
}

func before(cctx *cli.Context) error {
	_ = logging.SetLogLevel("*", "INFO")

	if cctx.Bool("verbose") {
		_ = logging.SetLogLevel("*", "DEBUG")
	}

	return nil
}
