package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/OpenFilWallet/OpenFilWallet/repo"
	"github.com/OpenFilWallet/OpenFilWallet/wallet"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "Start OpenFilWallet process",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "wallet-api",
			Usage: "wallet api port",
			Value: "6678",
		},
		&cli.BoolFlag{
			Name:  "offline",
			Usage: "offline wallet",
			Value: false,
		},
	},
	Action: func(cctx *cli.Context) error {
		repoPath := cctx.String(repo.FlagWalletRepo)
		r, err := repo.NewFS(repoPath)
		if err != nil {
			return err
		}

		ok, err := r.Exists()
		if err != nil {
			return err
		}
		if !ok {
			return xerrors.Errorf("repo at '%s' is not initialized, run 'openfild init' to set it up", repo.FlagWalletRepo)
		}

		lr, err := r.Lock()
		if err != nil {
			return err
		}

		endpoint := "localhost:" + cctx.String("wallet-api")

		err = lr.SetAPIEndpoint("http://" + endpoint)
		if err != nil {
			return err
		}

		ds, err := lr.Datastore(context.Background())
		if err != nil {
			return err
		}

		db := datastore.NewWalletDB(ds)

		if err := requirePassword(db); err != nil {
			return err
		}

		loginScrypt, err := db.GetLoginPassword()
		if err != nil {
			return err
		}

		app.SetSecret(loginScrypt)
		token, err := app.AuthNew(app.AllPermissions)
		if err != nil {
			return err
		}

		err = lr.SetAPIToken(token)
		if err != nil {
			return err
		}

		hasMnemonic, err := db.HasMnemonic()
		if err != nil {
			return err
		}

		if !hasMnemonic {
			return errors.New("mnemonic does not exist")
		}

		rootPassword, verified := verifyRootPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		_, err = account.LoadMnemonic(db, crypto.GenerateEncryptKey([]byte(rootPassword)))
		if err != nil {
			return fmt.Errorf("failed to decrypt mnemonic, err: %s", err.Error())
		}

		var closeCh = make(chan struct{})
		// new server
		walletServer, err := wallet.NewWallet(cctx.Bool("offline"), rootPassword, db, closeCh)
		if err != nil {
			return fmt.Errorf("new Wallet fail: %s", err.Error())
		}

		router := walletServer.NewRouter()

		s := &http.Server{
			Addr:         endpoint,
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		log.Infow("start wallet server", "endpoint", endpoint)
		go func() {
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("s.ListenAndServe err: %v", err)
			}
		}()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		close(closeCh)

		log.Info("shutting down wallet server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			log.Fatal("server forced to shutdown:", err)
		}

		err = lr.Close()
		if err != nil {
			log.Warnw("wallet db close fail", "err", err.Error())
			return err
		}

		log.Info("wallet db close")

		log.Info("wallet server exit")

		return nil
	},
}
