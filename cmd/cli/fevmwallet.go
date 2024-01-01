package main

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/urfave/cli/v2"
	"text/tabwriter"
)

var fevmWalletCmd = &cli.Command{
	Name:  "fevm-wallet",
	Usage: "OpenFilWallet fevm wallet new / list",
	Subcommands: []*cli.Command{
		fevmWalletNewCmd,
		fevmWalletBalanceCmd,
		fevmWalletListCmd,
		walletHistoryCmd,
	},
}

var fevmWalletNewCmd = &cli.Command{
	Name:  "new",
	Usage: "Generate fevm wallets with the same index",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    "index",
			Aliases: []string{"i"},
			Usage:   "hd wallet index",
			Value:   -1,
		},
	},
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		index := cctx.Int("index")

		r, err := walletAPI.FevmWalletCreate(index)
		if err != nil {
			return err
		}

		for _, addr := range r.NewWalletAddrs {
			fmt.Println(addr)
		}

		return nil
	},
}

var fevmWalletBalanceCmd = &cli.Command{
	Name:  "balance",
	Usage: "request fevm wallet balance",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "address",
			Aliases:  []string{"addr"},
			Usage:    "wallet address",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		address := cctx.String("address")
		bi, err := walletAPI.FevmBalance(address)
		if err != nil {
			return err
		}

		fmt.Printf("Address:%s \nFIL Address: %s \nBalance:%s\n", bi.Address, bi.FilAddress, bi.Amount)
		return nil
	},
}

var fevmWalletListCmd = &cli.Command{
	Name:  "list",
	Usage: "fevm wallet list",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "balance",
			Aliases: []string{"b"},
			Usage:   "request wallet balance",
			Value:   false,
		},
	},
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}
		balance := cctx.Bool("balance")

		walletInfo, err := walletAPI.FevmWalletList(balance)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(cctx.App.Writer, 8, 4, 2, ' ', 0)
		if balance {
			fmt.Fprintf(w, "ID\tWallet Type\tAddress\tFIL Address\tPath\tBalance\n")
		} else {
			fmt.Fprintf(w, "ID\tWallet Type\tAddress\tFIL Address\tPath\n")
		}

		for _, wallet := range walletInfo {
			if balance {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", wallet.WalletId, wallet.WalletType, wallet.WalletAddress, wallet.FilAddress, wallet.WalletPath, wallet.Balance)
				continue
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", wallet.WalletId, wallet.WalletType, wallet.WalletAddress, wallet.FilAddress, wallet.WalletPath)
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("flushing output: %+v", err)
		}
		return nil
	},
}
