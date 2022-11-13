package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
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
	},
	Action: func(cctx *cli.Context) error {
		repoPath := cctx.String(flagWalletRepo)
		r, err := repo.NewFS(repoPath)
		if err != nil {
			return err
		}

		ok, err := r.Exists()
		if err != nil {
			return err
		}
		if !ok {
			return xerrors.Errorf("repo at '%s' is not initialized, run 'openfild init' to set it up", flagWalletRepo)
		}

		lr, err := r.Lock()
		if err != nil {
			return err
		}

		endpoint := "127.0.0.1:" + cctx.String("wallet-api")

		err = lr.SetAPIEndpoint(endpoint)
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

		// new server
		walletServer := wallet.NewWallet(rootPassword, db)
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
