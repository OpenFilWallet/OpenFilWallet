package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/filecoin-project/go-address"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
)

var ethWalletCmd = &cli.Command{
	Name:  "fevm-wallet",
	Usage: "OpenFilWallet fevm wallet list / import / export",
	Subcommands: []*cli.Command{
		ethWalletNew,
		ethWalletListCmd,
		ethWalletImportCmd,
		ethWalletExportCmd,
		ethWalletDeleteCmd,
	},
}

var ethWalletNew = &cli.Command{
	Name:  "new",
	Usage: "Generate fevm wallets with the same index",
	Action: func(cctx *cli.Context) error {
		db, closer, err := getWalletDB(cctx, false)
		if err != nil {
			return err
		}
		defer closer()

		if err := requirePassword(db); err != nil {
			return err
		}

		masterPassword, verified := verifyMasterPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		mnemonic, err := account.LoadMnemonic(db, crypto.GenerateEncryptKey([]byte(masterPassword)))
		if err != nil {
			return err
		}

		ek, err := account.GenerateEthPrivateKeyFromMnemonicIndex(db, mnemonic, -1, crypto.GenerateEncryptKey([]byte(masterPassword)))
		if err != nil {
			return err
		}

		afmt := app.NewAppFmt(cctx.App)
		afmt.Printf("New Wallet: %s  Address: %s \n", "fevm-wallet", ek.Address.String())

		return nil
	},
}

var ethWalletListCmd = &cli.Command{
	Name:  "list",
	Usage: "fevm wallet list",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "export",
			Usage: "export private key",
			Value: false,
		},
	},
	Action: func(cctx *cli.Context) error {
		db, closer, err := getWalletDB(cctx, true)
		if err != nil {
			return err
		}
		defer closer()

		if err := requirePassword(db); err != nil {
			return err
		}

		masterPassword, verified := verifyMasterPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		afmt := app.NewAppFmt(cctx.App)

		export := cctx.Bool("export")
		keys, err := account.LoadEthPrivateKeys(db, crypto.GenerateEncryptKey([]byte(masterPassword)))
		if err != nil {
			return err
		}

		for _, key := range keys {
			afmt.Println("Address: ", key.Address)
			if export {
				pri := hex.EncodeToString(ethcrypto.FromECDSA(key.PriKey))
				afmt.Println("Key:     ", pri)
			}
		}

		return nil
	},
}

var ethWalletImportCmd = &cli.Command{
	Name:      "import",
	Usage:     "fevm wallet import",
	ArgsUsage: "[<path> (optional, will read from stdin if omitted)]",
	Action: func(cctx *cli.Context) error {
		db, closer, err := getWalletDB(cctx, false)
		if err != nil {
			return err
		}
		defer closer()

		if err := requirePassword(db); err != nil {
			return err
		}

		ok, err := db.HasMnemonic()
		if err != nil {
			return err
		}

		if !ok {
			return errors.New("mnemonic does not exist")
		}

		masterPassword, verified := verifyMasterPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		var inpdata []byte
		if !cctx.Args().Present() || cctx.Args().First() == "-" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter private key: ")
			indata, err := reader.ReadBytes('\n')
			if err != nil {
				return err
			}
			inpdata = indata

		} else {
			fdata, err := ioutil.ReadFile(cctx.Args().First())
			if err != nil {
				return err
			}
			inpdata = fdata
		}

		err = account.ImportEthPrivateKey(db, string(inpdata), crypto.GenerateEncryptKey([]byte(masterPassword)))
		if err != nil {
			return err
		}

		fmt.Println("fevm private key imported successfully")
		return nil
	},
}

var ethWalletExportCmd = &cli.Command{
	Name:      "export",
	Usage:     "fevm wallet export",
	ArgsUsage: "[address]",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must have address param")
		}

		db, closer, err := getWalletDB(cctx, true)
		if err != nil {
			return err
		}
		defer closer()

		if err := requirePassword(db); err != nil {
			return err
		}

		masterPassword, verified := verifyMasterPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		addrStr := cctx.Args().First()
		addr := common.HexToAddress(addrStr)
		key, err := account.GetEthPrivateKey(db, addr.String(), crypto.GenerateEncryptKey([]byte(masterPassword)))

		pri := hex.EncodeToString(ethcrypto.FromECDSA(key.PriKey))

		afmt := app.NewAppFmt(cctx.App)
		afmt.Println("Address: ", key.Address)
		afmt.Println("Key:     ", pri)

		return nil
	},
}

var ethWalletDeleteCmd = &cli.Command{
	Name:      "delete",
	Usage:     "delete fevm private key",
	ArgsUsage: "[address]",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must have address param")
		}

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

		addrStr := cctx.Args().First()
		addr, err := address.NewFromString(addrStr)
		if err != nil {
			return err
		}

		err = db.DeleteEthPrivate(addr.String())
		if err != nil {
			return err
		}

		fmt.Println("fevm private key deleted successfully")
		return nil
	},
}
