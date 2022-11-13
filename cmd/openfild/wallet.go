package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/account"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/filecoin-project/go-address"
	filcrypto "github.com/filecoin-project/go-state-types/crypto"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
)

var walletCmd = &cli.Command{
	Name:  "wallet",
	Usage: "OpenFilWallet wallet list / import / export",
	Subcommands: []*cli.Command{
		walletNew,
		walletListCmd,
		walletImportCmd,
		walletExportCmd,
		walletDeleteCmd,
	},
}

var walletNew = &cli.Command{
	Name:      "new",
	Usage:     "Generate a new key of the given type",
	ArgsUsage: "[bls|secp256k1 (default secp256k1)]",
	Action: func(cctx *cli.Context) error {
		db, closer, err := getWalletDB(cctx, false)
		if err != nil {
			return err
		}
		defer closer()

		if err := requirePassword(db); err != nil {
			return err
		}

		rootPassword, verified := verifyRootPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		mnemonic, err := account.LoadMnemonic(db, crypto.GenerateEncryptKey([]byte(rootPassword)))
		if err != nil {
			return err
		}

		t := cctx.Args().First()
		if t == "" {
			t = "secp256k1"
		}

		var sigType filcrypto.SigType
		switch t {
		case "secp256k1":
			sigType = filcrypto.SigTypeSecp256k1
		case "bls":
			sigType = filcrypto.SigTypeBLS
		default:
			return fmt.Errorf("KeyType: %s, TypeUnknown", t)
		}

		nk, err := account.GeneratePrivateKey(db, mnemonic, sigType, crypto.GenerateEncryptKey([]byte(rootPassword)))
		if err != nil {
			return err
		}

		afmt := app.NewAppFmt(cctx.App)
		afmt.Println(nk.Address.String())

		return nil
	},
}

var walletListCmd = &cli.Command{
	Name:  "list",
	Usage: "wallet list",
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

		rootPassword, verified := verifyRootPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		afmt := app.NewAppFmt(cctx.App)

		export := cctx.Bool("export")
		keys, err := account.LoadPrivateKeys(db, crypto.GenerateEncryptKey([]byte(rootPassword)))
		for _, key := range keys {
			afmt.Println("Address: ", key.Address)

			if export {
				b, err := json.Marshal(key.KeyInfo)
				if err != nil {
					return err
				}

				afmt.Println("Key:     ", hex.EncodeToString(b))
			}
		}

		return nil
	},
}

var walletImportCmd = &cli.Command{
	Name:      "import",
	Usage:     "wallet import",
	ArgsUsage: "[<path> (optional, will read from stdin if omitted)]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "format",
			Usage: "specify input format for key",
			Value: "hex-lotus",
		},
	},
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

		rootPassword, verified := verifyRootPassword(db)
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

		err = account.ImportPrivateKey(db, string(inpdata), cctx.String("format"), crypto.GenerateEncryptKey([]byte(rootPassword)))
		if err != nil {
			return err
		}

		log.Info("private key imported successfully")
		return nil
	},
}

var walletExportCmd = &cli.Command{
	Name:      "export",
	Usage:     "wallet export",
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

		rootPassword, verified := verifyRootPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		addrStr := cctx.Args().First()
		addr, err := address.NewFromString(addrStr)
		if err != nil {
			return err
		}
		key, err := account.GetPrivateKey(db, addr.String(), crypto.GenerateEncryptKey([]byte(rootPassword)))

		b, err := json.Marshal(key.KeyInfo)
		if err != nil {
			return err
		}

		afmt := app.NewAppFmt(cctx.App)
		afmt.Println("Address: ", key.Address)
		afmt.Println("Key:     ", hex.EncodeToString(b))

		return nil
	},
}

var walletDeleteCmd = &cli.Command{
	Name:      "delete",
	Usage:     "delete private key",
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

		_, verified := verifyRootPassword(db)
		if !verified {
			return errors.New("password verification failed")
		}

		addrStr := cctx.Args().First()
		addr, err := address.NewFromString(addrStr)
		if err != nil {
			return err
		}

		err = db.DeletePrivate(addr.String())
		if err != nil {
			return err
		}

		log.Info("private key deleted successfully")
		return nil
	},
}