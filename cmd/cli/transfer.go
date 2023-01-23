package main

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var transferCmd = &cli.Command{
	Name:  "transfer",
	Usage: "transfer amount",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:  "nonce",
			Usage: "specify the nonce to use",
			Value: 0,
		},
		&cli.StringFlag{
			Name:  "gas-premium",
			Usage: "specify gas price to use in AttoFIL",
			Value: "0",
		},
		&cli.StringFlag{
			Name:  "gas-feecap",
			Usage: "specify gas fee cap to use in AttoFIL",
			Value: "0",
		},
		&cli.Int64Flag{
			Name:  "gas-limit",
			Usage: "specify gas limit",
			Value: 0,
		},
		&cli.StringFlag{
			Name:  "max-fee",
			Usage: "the max tx fee allowed for this transaction",
			Value: "1 FIL",
		},
		&cli.StringFlag{
			Name:  "output",
			Usage: "save message to file",
			Value: "",
		},
	},
	ArgsUsage: "[from to amount (FIL)]",
	Action: func(cctx *cli.Context) error {
		from, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return fmt.Errorf("parsing address %s: %w", cctx.Args().Get(0), err)
		}

		to, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return fmt.Errorf("parsing address %s: %w", cctx.Args().Get(1), err)
		}

		amount := cctx.Args().Get(2)

		_, err = types.ParseFIL(amount)
		if err != nil {
			return xerrors.Errorf("parsing 'amount' argument: %w", err)
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		baseParams, err := getBaseParams(cctx)
		if err != nil {
			return err
		}

		msg, err := walletAPI.Transfer(baseParams, from.String(), to.String(), amount)
		if err != nil {
			return err // todo 错误提示
		}

		return printMessage(cctx, msg)
	},
}
