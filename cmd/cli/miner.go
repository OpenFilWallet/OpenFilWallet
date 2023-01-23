package main

import (
	"context"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/fatih/color"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/lib/tablewriter"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"os"
	"strings"
)

var minerCmd = &cli.Command{
	Name:  "miner",
	Usage: "miner control",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:    "nonce",
			Aliases: []string{"n"},
			Usage:   "specify the nonce to use",
			Value:   0,
		},
		&cli.StringFlag{
			Name:    "gas-premium",
			Aliases: []string{"gp"},
			Usage:   "specify gas price to use in AttoFIL",
			Value:   "0",
		},
		&cli.StringFlag{
			Name:    "gas-feecap",
			Aliases: []string{"gf"},
			Usage:   "specify gas fee cap to use in AttoFIL",
			Value:   "0",
		},
		&cli.Int64Flag{
			Name:    "gas-limit",
			Aliases: []string{"gl"},
			Usage:   "specify gas limit",
			Value:   0,
		},
		&cli.StringFlag{
			Name:    "max-fee",
			Aliases: []string{"mf"},
			Usage:   "the max tx fee allowed for this transaction",
			Value:   "1 FIL",
		},
	},
	Subcommands: []*cli.Command{
		actorWithdrawCmd,
		actorSetOwnerCmd,
		actorControl,
		actorProposeChangeWorker,
		actorConfirmChangeWorker,
	},
}

var actorWithdrawCmd = &cli.Command{
	Name:      "withdraw",
	Usage:     "withdraw available balance",
	ArgsUsage: "[amount (FIL)]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "actor",
			Aliases:  []string{"a"},
			Usage:    "specify the address of miner actor",
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
		act := cctx.String("actor")
		maddr, err := address.NewFromString(act)
		if err != nil {
			return fmt.Errorf("parsing address %s: %w", act, err)
		}

		f, err := types.ParseFIL(cctx.Args().First())
		if err != nil {
			return xerrors.Errorf("parsing 'amount' argument: %w", err)
		}

		amount := abi.TokenAmount(f)

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		baseParams, err := getBaseParams(cctx)
		if err != nil {
			return err
		}

		msg, err := walletAPI.Withdraw(baseParams, maddr.String(), amount.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var actorSetOwnerCmd = &cli.Command{
	Name:      "set-owner",
	Usage:     "Set owner address (this command should be invoked twice, first with the old owner as the senderAddress, and then with the new owner)",
	ArgsUsage: "[newOwnerAddress senderAddress]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "actor",
			Aliases:  []string{"a"},
			Usage:    "specify the address of miner actor",
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
		if cctx.NArg() != 2 {
			return fmt.Errorf("must pass new owner address and sender address")
		}

		act := cctx.String("actor")
		maddr, err := address.NewFromString(act)
		if err != nil {
			return fmt.Errorf("parsing address %s: %w", act, err)
		}

		na, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		fa, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		baseParams, err := getBaseParams(cctx)
		if err != nil {
			return err
		}

		msg, err := walletAPI.ChangeOwner(baseParams, maddr.String(), na.String(), fa.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var actorControl = &cli.Command{
	Name:  "control",
	Usage: "Manage control addresses",
	Subcommands: []*cli.Command{
		actorControlList,
		actorControlSet,
	},
}

var actorControlList = &cli.Command{
	Name:  "list",
	Usage: "Get currently set control addresses",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "actor",
			Aliases:  []string{"a"},
			Usage:    "specify the address of miner actor",
			Required: true,
		},
		&cli.BoolFlag{
			Name:        "color",
			Aliases:     []string{"c"},
			Usage:       "use color in display output",
			DefaultText: "depends on output being a TTY",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.IsSet("color") {
			color.NoColor = !cctx.Bool("color")
		}

		act := cctx.String("actor")
		lotusAPI, err := getLotusAPI(cctx)
		if err != nil {
			return err
		}
		defer lotusAPI.Closer()
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		mc, err := walletAPI.ControlList(act)
		if err != nil {
			return err
		}

		tw := tablewriter.New(
			tablewriter.Col("name"),
			tablewriter.Col("ID"),
			tablewriter.Col("key"),
			tablewriter.Col("balance"),
		)

		printKey := func(name string, a string) {
			balance, err := walletAPI.Balance(a)
			if err != nil {
				fmt.Printf("%s\t%s: error getting balance: %s\n", name, a, err)
				return
			}

			b, err := big.FromString(balance.Amount)
			if err != nil {
				fmt.Printf("%s\t%s: error parsing balance: %s\n", name, a, err)
				return
			}

			addr, err := address.NewFromString(a)
			if err != nil {
				fmt.Printf("%s\t%s: error parsing address: %s\n", name, a, err)
				return
			}

			k, err := lotusAPI.Api.StateAccountKey(context.Background(), addr, types.EmptyTSK)
			if err != nil {
				if strings.Contains(err.Error(), "multisig") {
					fmt.Printf("%s\t%s (multisig) \n", name, a)
					return
				}

				fmt.Printf("%s\t%s: error getting account key: %s\n", name, a, err)
				return
			}

			bstr := types.FIL(b).String()
			switch {
			case b.LessThan(types.FromFil(10)):
				bstr = color.RedString(bstr)
			case b.LessThan(types.FromFil(50)):
				bstr = color.YellowString(bstr)
			default:
				bstr = color.GreenString(bstr)
			}

			tw.Write(map[string]interface{}{
				"name":    name,
				"ID":      a,
				"key":     k,
				"balance": bstr,
			})
		}

		printKey("owner", mc.Owner)
		printKey("worker", mc.Worker)
		for i, ca := range mc.ControlAddresses {
			printKey(fmt.Sprintf("control-%d", i), ca)
		}

		return tw.Flush(os.Stdout)
	},
}

var actorControlSet = &cli.Command{
	Name:      "set",
	Usage:     "Set control address(-es)",
	ArgsUsage: "[...address]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "actor",
			Aliases:  []string{"a"},
			Usage:    "specify the address of miner actor",
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
		act := cctx.String("actor")
		maddr, err := address.NewFromString(act)
		if err != nil {
			return fmt.Errorf("parsing address %s: %w", act, err)
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		baseParams, err := getBaseParams(cctx)
		if err != nil {
			return err
		}

		msg, err := walletAPI.ChangeControl(baseParams, maddr.String(), cctx.Args().Slice())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var actorProposeChangeWorker = &cli.Command{
	Name:      "propose-change-worker",
	Usage:     "Propose a worker address change",
	ArgsUsage: "[address]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "actor",
			Aliases:  []string{"a"},
			Usage:    "specify the address of miner actor",
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
		if !cctx.Args().Present() {
			return fmt.Errorf("must pass address of new worker address")
		}

		act := cctx.String("actor")
		maddr, err := address.NewFromString(act)
		if err != nil {
			return fmt.Errorf("parsing address %s: %w", act, err)
		}

		na, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		baseParams, err := getBaseParams(cctx)
		if err != nil {
			return err
		}

		msg, err := walletAPI.ChangeWorker(baseParams, maddr.String(), na.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var actorConfirmChangeWorker = &cli.Command{
	Name:      "confirm-change-worker",
	Usage:     "Confirm a worker address change",
	ArgsUsage: "[address]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "actor",
			Aliases:  []string{"a"},
			Usage:    "specify the address of miner actor",
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
		if !cctx.Args().Present() {
			return fmt.Errorf("must pass address of new worker address")
		}

		act := cctx.String("actor")
		maddr, err := address.NewFromString(act)
		if err != nil {
			return fmt.Errorf("parsing address %s: %w", act, err)
		}

		na, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		baseParams, err := getBaseParams(cctx)
		if err != nil {
			return err
		}

		msg, err := walletAPI.ConfirmChangeWorker(baseParams, maddr.String(), na.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var actorProposeChangeBeneficiary = &cli.Command{
	Name:      "propose-change-beneficiary",
	Usage:     "Propose a beneficiary address change",
	ArgsUsage: "[beneficiaryAddress quota expiration]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "overwrite-pending-change",
			Aliases: []string{"opc"},
			Usage:   "Overwrite the current beneficiary change proposal",
			Value:   false,
		},
		&cli.StringFlag{
			Name:     "actor",
			Aliases:  []string{"a"},
			Usage:    "specify the address of miner actor",
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
		if cctx.NArg() != 3 {
			return fmt.Errorf("must have [beneficiaryAddress quota expiration]")
		}

		act := cctx.String("actor")
		maddr, err := address.NewFromString(act)
		if err != nil {
			return fmt.Errorf("parsing address %s: %w", act, err)
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		baseParams, err := getBaseParams(cctx)
		if err != nil {
			return err
		}

		msg, err := walletAPI.ChangeBeneficiary(baseParams, maddr.String(), cctx.Args().Get(0), cctx.Args().Get(1), cctx.Args().Get(2), cctx.Bool("overwrite-pending-change"))
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var actorConfirmChangeBeneficiary = &cli.Command{
	Name:      "confirm-change-beneficiary",
	Usage:     "Confirm a beneficiary address change",
	ArgsUsage: "[minerAddress]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "existing-beneficiary",
			Aliases: []string{"eb"},
			Usage:   "send confirmation from the existing beneficiary address",
		},
		&cli.BoolFlag{
			Name:    "new-beneficiary",
			Aliases: []string{"nb"},
			Usage:   "send confirmation from the new beneficiary address",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() != 1 {
			return fmt.Errorf("must have [minerAddress]")
		}

		act := cctx.Args().Get(0)
		maddr, err := address.NewFromString(act)
		if err != nil {
			return fmt.Errorf("parsing address %s: %w", act, err)
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		baseParams, err := getBaseParams(cctx)
		if err != nil {
			return err
		}

		msg, err := walletAPI.ConfirmChangeBeneficiary(baseParams, maddr.String(), cctx.Bool("existing-beneficiary"), cctx.Bool("new-beneficiary"))
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}
