package main

import (
	"encoding/json"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/chain"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
)

var signCmd = &cli.Command{
	Name:  "sign",
	Usage: "sign with the private key at the specified address",
	Subcommands: []*cli.Command{
		signMsgCmd,
		signTxCmd,
		signTxAndSendCmd,
	},
}

var signTxCmd = &cli.Command{
	Name:  "sign-tx",
	Usage: "sign a transaction",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "tx-path",
			Usage:    "path to file containing transaction information",
			Value:    "",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		path := cctx.String("tx-path")
		fi, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("fail to open the file (path: %s): %s", path, err)
		}
		defer fi.Close()
		content, err := ioutil.ReadAll(fi)
		if err != nil {
			return err
		}

		var msg chain.Message
		err = json.Unmarshal(content, &msg)
		if err != nil {
			return fmt.Errorf("failed to parse message: %s", err)
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		signedMessage, err := walletAPI.Sign(msg)
		if err != nil {
			return err
		}

		return printMessage(cctx, signedMessage)
	},
}

var signTxAndSendCmd = &cli.Command{
	Name:  "sign-tx-send",
	Usage: "sign a transaction and send tx",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "tx-path",
			Usage:    "path to file containing transaction information",
			Value:    "",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		path := cctx.String("tx-path")
		fi, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("fail to open the file (path: %s): %s", path, err)
		}
		defer fi.Close()

		content, err := ioutil.ReadAll(fi)
		if err != nil {
			return err
		}

		var msg chain.Message
		err = json.Unmarshal(content, &msg)
		if err != nil {
			return fmt.Errorf("failed to parse message: %s", err)
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		cid, err := walletAPI.SignAndSend(msg)
		if err != nil {
			return err
		}

		fmt.Println(cid)
		return nil
	},
}

var signMsgCmd = &cli.Command{
	Name:      "sign-msg",
	Usage:     "sign msg",
	ArgsUsage: "<signing address> <hexMessage>",
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("incorrect number of arguments")
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		sign, err := walletAPI.SignMsg(cctx.Args().Get(0), cctx.Args().Get(1))
		if err != nil {
			return err
		}

		fmt.Println(sign)
		return nil
	},
}
