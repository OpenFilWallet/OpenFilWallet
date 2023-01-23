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

var sendCmd = &cli.Command{
	Name:  "send",
	Usage: "send tx",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "tx-path",
			Aliases:  []string{"tp"},
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

		var msg chain.SignedMessage
		err = json.Unmarshal(content, &msg)
		if err != nil {
			return fmt.Errorf("failed to parse message: %s", err)
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		cid, err := walletAPI.Send(msg)
		if err != nil {
			return err
		}

		fmt.Println(cid)
		return nil
	},
}
