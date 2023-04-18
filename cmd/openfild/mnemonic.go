package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/lib/hd"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
	"strings"
)

var mnemonicCmd = &cli.Command{
	Name:  "mnemonic",
	Usage: "OpenFilWallet mnemonic generate / import / export",
	Subcommands: []*cli.Command{
		mnemonicGenerateCmd,
		mnemonicImportCmd,
		mnemonicExportCmd,
		mnemonicDeleteCmd,
	},
}

var mnemonicGenerateCmd = &cli.Command{
	Name:      "generate",
	Usage:     "generate mnemonic",
	ArgsUsage: "[number of mnemonic words, 12 / 24]",
	Action: func(cctx *cli.Context) error {
		mType := hd.Mnemonic12
		if cctx.Args().Len() != 0 {
			mnemonicNumber := cctx.Args().Get(0)
			if mnemonicNumber == "24" {
				mType = hd.Mnemonic24
			}
		}

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

		if ok {
			return errors.New("mnemonic already exists")
		}

		masterPassword, verified := verifyMasterPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		err = account.GenerateMnemonic(db, mType, crypto.GenerateEncryptKey([]byte(masterPassword)))
		if err != nil {
			return err
		}

		fmt.Println("mnemonic generated successfully")
		return nil
	},
}

var mnemonicImportCmd = &cli.Command{
	Name:      "import",
	Usage:     "import mnemonic",
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

		if ok {
			return errors.New("mnemonic already exists")
		}

		masterPassword, verified := verifyMasterPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		var mnemonic []byte
		if !cctx.Args().Present() || cctx.Args().First() == "-" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter Mnemonic key: ")
			indata, err := reader.ReadBytes('\n')
			if err != nil {
				return err
			}
			mnemonic = indata
		} else {
			fdata, err := ioutil.ReadFile(cctx.Args().First())
			if err != nil {
				return err
			}
			mnemonic = fdata
		}

		err = account.ImportMnemonic(db, strings.Replace(string(mnemonic), "\n", "", -1), crypto.GenerateEncryptKey([]byte(masterPassword)))
		if err != nil {
			return err
		}

		fmt.Println("mnemonic imported successfully")
		return nil
	},
}

var mnemonicExportCmd = &cli.Command{
	Name:  "export",
	Usage: "export mnemonic",
	Action: func(cctx *cli.Context) error {
		db, closer, err := getWalletDB(cctx, true)
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

		mnemonic, err := account.LoadMnemonic(db, crypto.GenerateEncryptKey([]byte(masterPassword)))
		if err != nil {
			return err
		}

		afmt := app.NewAppFmt(cctx.App)
		fmt.Println("Be sure to save mnemonic. Losing mnemonic will cause all property damage!")
		fmt.Println()
		afmt.Println(mnemonic)
		fmt.Println()
		return nil
	},
}

var mnemonicDeleteCmd = &cli.Command{
	Name:  "delete",
	Usage: "delete mnemonic",
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

		_, verified := verifyMasterPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		fmt.Println("The mnemonic must have been backed up, deletion cannot be undone")
		isConfirm, err := app.Confirm("Already backed up the mnemonic?")
		if err != nil {
			return err
		}

		if !isConfirm {
			fmt.Println("please back up the mnemonic, run 'openfild mnemonic export'")
			return nil
		}

		err = db.DeleteMnemonic()
		if err != nil {
			return err
		}

		fmt.Println("mnemonic deleted successfully")
		return nil
	},
}
