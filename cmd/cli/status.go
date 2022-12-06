package main

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/urfave/cli/v2"
)

var statusCmd = &cli.Command{
	Name:  "status",
	Usage: "query wallet status",
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		si, err := walletAPI.Status()
		if err != nil {
			return err
		}

		fmt.Println("Wallet Lock:    ", si.Lock)
		fmt.Println("Wallet Offline: ", si.Offline)
		fmt.Println("Wallet Version: ", si.Version)
		return nil
	},
}
