package main

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/urfave/cli/v2"
	"text/tabwriter"
)

var walletCmd = &cli.Command{
	Name:  "wallet",
	Usage: "OpenFilWallet wallet new / list",
	Subcommands: []*cli.Command{
		walletNewCmd,
		walletBalanceCmd,
		walletListCmd,
		walletHistoryCmd,
	},
}

var walletNewCmd = &cli.Command{
	Name:  "new",
	Usage: "Generate bls and secp256k1 wallets with the same index",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "index",
			Usage: "hd wallet index",
			Value: -1,
		},
	},
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		index := cctx.Int("index")

		r, err := walletAPI.WalletCreate(index)
		if err != nil {
			return err
		}

		for _, addr := range r.NewWalletAddrs {
			fmt.Println(addr)
		}

		return nil
	},
}

var walletBalanceCmd = &cli.Command{
	Name:  "balance",
	Usage: "request wallet balance",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "address",
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
		bi, err := walletAPI.Balance(address)
		if err != nil {
			return err
		}

		fmt.Printf("%s %s\n", bi.Address, bi.Amount)
		return nil
	},
}

var walletListCmd = &cli.Command{
	Name:  "list",
	Usage: "wallet list",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "balance",
			Usage: "request wallet balance",
			Value: false,
		},
	},
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		walletInfo, err := walletAPI.WalletList()
		if err != nil {
			return err
		}

		balance := cctx.Bool("balance")
		w := tabwriter.NewWriter(cctx.App.Writer, 8, 4, 2, ' ', 0)
		if balance {
			fmt.Fprintf(w, "ID\tWallet Type\tAddress\tPath\tBalance\n")
		} else {
			fmt.Fprintf(w, "ID\tWallet Type\tAddress\tPath\n")
		}

		i := 0
		for _, wallet := range walletInfo {
			if balance {
				bi, err := walletAPI.Balance(wallet.WalletAddress)
				if err != nil {
					log.Warnw("request wallet balance fail", "addr", wallet.WalletAddress)
					continue
				}
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", i, wallet.WalletType, wallet.WalletAddress, wallet.WalletPath, bi.Amount)
				continue
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", i, wallet.WalletType, wallet.WalletAddress, wallet.WalletPath)
			i++
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("flushing output: %+v", err)
		}
		return nil
	},
}

var walletHistoryCmd = &cli.Command{
	Name:  "history",
	Usage: "wallet tx history",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "address",
			Usage:    "request wallet tx history",
			Value:    "",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		addr := cctx.String("address")
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		txs, err := walletAPI.TxHistory(addr)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(cctx.App.Writer, 8, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tVersion\tTo\tFrom\tNonce\tValue\tGasLimit\tGasFeeCap\tGasPremium\tMethod\tParams\n")

		for i, tx := range txs {
			fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%s\n", i, tx.Version, tx.To, tx.From, tx.Nonce, tx.Value, tx.GasLimit, tx.GasFeeCap, tx.GasPremium, tx.Method, tx.Params)
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("flushing output: %+v", err)
		}

		return nil
	},
}
