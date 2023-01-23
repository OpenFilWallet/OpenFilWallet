package main

import (
	"context"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"strconv"
	"text/tabwriter"
)

var multisigCmd = &cli.Command{
	Name:  "msig",
	Usage: "Interact with a multisig wallet",
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
		msigCreateCmd,
		msigWalletListCmd,
		msigInspectCmd,
		msigApproveCmd,
		msigCancelCmd,
		msigTransferProposeCmd,
		msigTransferApproveCmd,
		msigTransferCancelCmd,
		msigAddProposeCmd,
		msigAddApproveCmd,
		msigAddCancelCmd,
		msigSwapProposeCmd,
		msigSwapApproveCmd,
		msigSwapCancelCmd,
		msigLockProposeCmd,
		msigLockApproveCmd,
		msigLockCancelCmd,
		msigThresholdProposeCmd,
		msigThresholdApproveCmd,
		msigThresholdCancelCmd,
		msigChangeOwnerProposeCmd,
		msigChangeOwnerApproveCmd,
		msigWithdrawBalanceProposeCmd,
		msigWithdrawBalanceApproveCmd,
		msigChangeWorkerProposeCmd,
		msigChangeWorkerApproveCmd,
		msigConfirmChangeWorkerProposeCmd,
		msigConfirmChangeWorkerApproveCmd,
		msigSetControlProposeCmd,
		msigSetControlApproveCmd,
		msigChangeBeneficiaryProposeCmd,
		msigChangeBeneficiaryApproveCmd,
		msigConfirmChangeBeneficiaryProposeCmd,
		msigConfirmChangeBeneficiaryApproveCmd,
	},
}

var msigCreateCmd = &cli.Command{
	Name:      "create",
	Usage:     "Create a new multisig wallet",
	ArgsUsage: "[address1 address2 ...]",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:    "required",
			Aliases: []string{"r"},
			Usage:   "number of required approvals (uses number of signers provided if omitted)",
		},
		&cli.StringFlag{
			Name:    "value",
			Aliases: []string{"v"},
			Usage:   "initial funds to give to multisig",
			Value:   "0",
		},
		&cli.StringFlag{
			Name:    "duration",
			Aliases: []string{"d"},
			Usage:   "length of the period over which funds unlock",
			Value:   "0",
		},
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the create message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() < 1 {
			return fmt.Errorf("multisigs must have at least one signer")
		}

		var addrs []address.Address
		for _, a := range cctx.Args().Slice() {
			addr, err := address.NewFromString(a)
			if err != nil {
				return err
			}
			addrs = append(addrs, addr)
		}

		var sendAddr address.Address
		addr, err := address.NewFromString(cctx.String("from"))
		if err != nil {
			return err
		}

		sendAddr = addr

		val := cctx.String("value")

		required := cctx.Uint64("required")
		if required == 0 {
			required = uint64(len(addrs))
		}

		d := cctx.Uint64("duration")

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		baseParams, err := getBaseParams(cctx)
		if err != nil {
			return err
		}

		msg, err := walletAPI.MsigCreate(baseParams, sendAddr.String(), required, d, val, cctx.Args().Slice()...)
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigWalletListCmd = &cli.Command{
	Name:  "list",
	Usage: "msig wallet list",
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		lotusAPI, err := getLotusAPI(cctx)
		if err != nil {
			return err
		}
		defer lotusAPI.Closer()

		walletInfo, err := walletAPI.MsigWalletList()
		if err != nil {
			return err
		}

		ctx := context.Background()
		for _, wallet := range walletInfo {
			bi, err := walletAPI.Balance(wallet.MsigAddr)
			if err != nil {
				log.Warnw("request wallet balance fail", "addr", wallet.MsigAddr)
				continue
			}
			fmt.Fprintf(cctx.App.Writer, "Msig address: %s \n", wallet.MsigAddr)
			fmt.Fprintf(cctx.App.Writer, "Balance: %s\n", bi.Amount)
			fmt.Fprintf(cctx.App.Writer, "Threshold: %d / %d\n", wallet.NumApprovalsThreshold, len(wallet.Signers))
			fmt.Fprintln(cctx.App.Writer, "Signers:")

			signerTable := tabwriter.NewWriter(cctx.App.Writer, 8, 4, 2, ' ', 0)
			fmt.Fprintf(signerTable, "ID\tAddress\n")
			for _, s := range wallet.Signers {
				addr, _ := address.NewFromString(s)
				signerActor, err := lotusAPI.Api.StateAccountKey(ctx, addr, types.EmptyTSK)
				if err != nil {
					fmt.Fprintf(signerTable, "%s\t%s\n", s, "N/A")
				} else {
					fmt.Fprintf(signerTable, "%s\t%s\n", s, signerActor)
				}
			}
			if err := signerTable.Flush(); err != nil {
				return xerrors.Errorf("flushing output: %+v", err)
			}
			fmt.Fprintln(cctx.App.Writer, "")
		}

		return nil
	},
}

var msigInspectCmd = &cli.Command{
	Name:      "inspect",
	Usage:     "Inspect a multisig wallet",
	ArgsUsage: "[address]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "vesting",
			Aliases: []string{"v"},
			Usage:   "Include vesting details",
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must specify address of multisig to inspect")
		}

		lotusAPI, err := getLotusAPI(cctx)
		if err != nil {
			return err
		}
		defer lotusAPI.Closer()

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		maddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		inspect, err := walletAPI.MsigInspect(maddr.String())
		if err != nil {
			return err
		}

		fmt.Fprintf(cctx.App.Writer, "Balance: %s\n", inspect.Balance)
		fmt.Fprintf(cctx.App.Writer, "Spendable: %s\n", inspect.Spendable)

		if cctx.Bool("vesting") {
			fmt.Fprintf(cctx.App.Writer, "InitialBalance: %s\n", inspect.Lock.InitialBalance)
			fmt.Fprintf(cctx.App.Writer, "LockAmount: %s\n", inspect.Lock.LockAmount)
			fmt.Fprintf(cctx.App.Writer, "StartEpoch: %d\n", inspect.Lock.StartEpoch)
			fmt.Fprintf(cctx.App.Writer, "UnlockDuration: %d\n", inspect.Lock.UnlockDuration)
		}

		fmt.Fprintf(cctx.App.Writer, "Threshold: %d / %d\n", inspect.Threshold, len(inspect.Signers))
		fmt.Fprintln(cctx.App.Writer, "Signers:")

		ctx := context.Background()

		signerTable := tabwriter.NewWriter(cctx.App.Writer, 8, 4, 2, ' ', 0)
		fmt.Fprintf(signerTable, "ID\tAddress\n")
		for _, s := range inspect.Signers {
			addr, _ := address.NewFromString(s)
			signerActor, err := lotusAPI.Api.StateAccountKey(ctx, addr, types.EmptyTSK)
			if err != nil {
				fmt.Fprintf(signerTable, "%s\t%s\n", s, "N/A")
			} else {
				fmt.Fprintf(signerTable, "%s\t%s\n", s, signerActor)
			}
		}
		if err := signerTable.Flush(); err != nil {
			return xerrors.Errorf("flushing output: %+v", err)
		}

		pending := inspect.Transactions
		fmt.Fprintln(cctx.App.Writer, "Transactions: ", len(pending))
		if len(pending) > 0 {
			w := tabwriter.NewWriter(cctx.App.Writer, 8, 4, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tState\tApprovals\tTo\tValue\tMethod\tParams\n")
			for i, tx := range pending {
				fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\t%s\t%s\n", i, "pending", len(tx.Approved), tx.To, tx.Value, tx.Method, tx.Params)
			}
			if err := w.Flush(); err != nil {
				return xerrors.Errorf("flushing output: %+v", err)
			}
		}

		return nil
	},
}

var msigApproveCmd = &cli.Command{
	Name:      "approve",
	Usage:     "Approve a multisig message",
	ArgsUsage: "[multisigAddress txId]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the approve message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must pass at least multisig address and message ID")
		}

		_, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		_, err = strconv.ParseUint(cctx.Args().Get(1), 10, 64)
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigApprove(baseParams, from.String(), cctx.Args().Get(0), cctx.Args().Get(1))
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigCancelCmd = &cli.Command{
	Name:      "cancel",
	Usage:     "Cancel a multisig message",
	ArgsUsage: "[multisigAddress txId]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the cancel message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must pass at least multisig address and message ID")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		_, err = strconv.ParseUint(cctx.Args().Get(1), 10, 64)
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigCancel(baseParams, from.String(), msig.String(), cctx.Args().Get(1))
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigTransferProposeCmd = &cli.Command{
	Name:      "transfer-propose",
	Usage:     "Propose a multisig transaction",
	ArgsUsage: "[multisigAddress destinationAddress value]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the propose message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 3 {
			return fmt.Errorf("must have multisig address, destination, and value")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		dest, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		value, err := types.ParseFIL(cctx.Args().Get(2))
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigTransferPropose(baseParams, from.String(), msig.String(), dest.String(), value.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigTransferApproveCmd = &cli.Command{
	Name:      "transfer-approve",
	Usage:     "Approve a multisig message",
	ArgsUsage: "[multisigAddress txId]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the approve message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must have multisig address and message ID")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		_, err = strconv.ParseUint(cctx.Args().Get(1), 10, 64)
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigTransferApprove(baseParams, from.String(), msig.String(), cctx.Args().Get(1))
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigTransferCancelCmd = &cli.Command{
	Name:      "transfer-cancel",
	Usage:     "Cancel transfer multisig message",
	ArgsUsage: "[multisigAddress txId]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the cancel message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must have multisig address and txId")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		_, err = strconv.ParseUint(cctx.Args().Get(1), 10, 64)
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigTransferCancel(baseParams, from.String(), msig.String(), cctx.Args().Get(1))
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigAddProposeCmd = &cli.Command{
	Name:      "add-propose",
	Usage:     "Propose to add a signer",
	ArgsUsage: "[multisigAddress signer]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "increase-threshold",
			Aliases: []string{"it"},
			Usage:   "whether the number of required signers should be increased",
		},
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the propose message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must pass multisig address and signer address")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		addr, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigAddPropose(baseParams, from.String(), msig.String(), addr.String(), cctx.Bool("increase-threshold"))
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigAddApproveCmd = &cli.Command{
	Name:      "add-approve",
	Usage:     "Approve a message to add a signer",
	ArgsUsage: "[multisigAddress proposerAddress txId newAddress increaseThreshold]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the approve message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 5 {
			return fmt.Errorf("must pass multisig address, proposer address, transaction id, new signer address, whether to increase threshold")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		prop, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		_, err = strconv.ParseUint(cctx.Args().Get(2), 10, 64)
		if err != nil {
			return err
		}

		newAdd, err := address.NewFromString(cctx.Args().Get(3))
		if err != nil {
			return err
		}

		inc, err := strconv.ParseBool(cctx.Args().Get(4))
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigAddApprove(baseParams, from.String(), msig.String(), prop.String(), cctx.Args().Get(2), newAdd.String(), inc)
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigAddCancelCmd = &cli.Command{
	Name:      "add-cancel",
	Usage:     "Cancel a message to add a signer",
	ArgsUsage: "[multisigAddress txId newAddress increaseThreshold]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the approve message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 4 {
			return fmt.Errorf("must pass multisig address, transaction id, new signer address, whether to increase threshold")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		_, err = strconv.ParseUint(cctx.Args().Get(1), 10, 64)
		if err != nil {
			return err
		}

		newAdd, err := address.NewFromString(cctx.Args().Get(2))
		if err != nil {
			return err
		}

		inc, err := strconv.ParseBool(cctx.Args().Get(3))
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigAddCancel(baseParams, from.String(), msig.String(), cctx.Args().Get(2), newAdd.String(), inc)
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigSwapProposeCmd = &cli.Command{
	Name:      "swap-propose",
	Usage:     "Propose to swap signers",
	ArgsUsage: "[multisigAddress oldAddress newAddress]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the approve message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 3 {
			return fmt.Errorf("must pass multisig address, old signer address, new signer address")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		oldAddr, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		newAddr, err := address.NewFromString(cctx.Args().Get(2))
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigSwapPropose(baseParams, from.String(), msig.String(), oldAddr.String(), newAddr.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigSwapApproveCmd = &cli.Command{
	Name:      "swap-approve",
	Usage:     "Approve a message to swap signers",
	ArgsUsage: "[multisigAddress proposerAddress txId oldAddress newAddress]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the approve message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 5 {
			return fmt.Errorf("must pass multisig address, proposer address, transaction id, old signer address, new signer address")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		prop, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		_, err = strconv.ParseUint(cctx.Args().Get(2), 10, 64)
		if err != nil {
			return err
		}

		oldAddr, err := address.NewFromString(cctx.Args().Get(3))
		if err != nil {
			return err
		}

		newAddr, err := address.NewFromString(cctx.Args().Get(4))
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigSwapApprove(baseParams, from.String(), msig.String(), prop.String(), cctx.Args().Get(2), oldAddr.String(), newAddr.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigSwapCancelCmd = &cli.Command{
	Name:      "swap-cancel",
	Usage:     "Cancel a message to swap signers",
	ArgsUsage: "[multisigAddress txId oldAddress newAddress]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the approve message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 4 {
			return fmt.Errorf("must pass multisig address, transaction id, old signer address, new signer address")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		_, err = strconv.ParseUint(cctx.Args().Get(1), 10, 64)
		if err != nil {
			return err
		}

		oldAddr, err := address.NewFromString(cctx.Args().Get(2))
		if err != nil {
			return err
		}

		newAddr, err := address.NewFromString(cctx.Args().Get(3))
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigSwapCancel(baseParams, from.String(), msig.String(), cctx.Args().Get(2), oldAddr.String(), newAddr.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigLockProposeCmd = &cli.Command{
	Name:      "lock-propose",
	Usage:     "Propose to lock up some balance",
	ArgsUsage: "[multisigAddress startEpoch unlockDuration amount]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the propose message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 4 {
			return fmt.Errorf("must pass multisig address, start epoch, unlock duration, and amount")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		start := cctx.Args().Get(1)
		_, err = strconv.ParseUint(start, 10, 64)
		if err != nil {
			return err
		}

		duration := cctx.Args().Get(2)
		_, err = strconv.ParseUint(duration, 10, 64)
		if err != nil {
			return err
		}

		amount := cctx.Args().Get(3)
		_, err = types.ParseFIL(amount)
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigLockPropose(baseParams, from.String(), msig.String(), start, duration, amount)
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigLockApproveCmd = &cli.Command{
	Name:      "lock-approve",
	Usage:     "Approve a message to lock up some balance",
	ArgsUsage: "[multisigAddress proposerAddress txId startEpoch unlockDuration amount]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the approve message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 6 {
			return fmt.Errorf("must pass multisig address, proposer address, tx id, start epoch, unlock duration, and amount")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		prop, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		txid := cctx.Args().Get(2)
		_, err = strconv.ParseUint(txid, 10, 64)
		if err != nil {
			return err
		}

		start := cctx.Args().Get(3)
		_, err = strconv.ParseUint(start, 10, 64)
		if err != nil {
			return err
		}

		duration := cctx.Args().Get(4)
		_, err = strconv.ParseUint(duration, 10, 64)
		if err != nil {
			return err
		}

		amount := cctx.Args().Get(5)
		_, err = types.ParseFIL(amount)
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigLockApprove(baseParams, from.String(), msig.String(), prop.String(), txid, start, duration, amount)
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigLockCancelCmd = &cli.Command{
	Name:      "lock-cancel",
	Usage:     "Cancel a message to lock up some balance",
	ArgsUsage: "[multisigAddress txId startEpoch unlockDuration amount]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the cancel message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 5 {
			return fmt.Errorf("must pass multisig address, tx id, start epoch, unlock duration, and amount")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		txid := cctx.Args().Get(1)
		_, err = strconv.ParseUint(txid, 10, 64)
		if err != nil {
			return err
		}

		start := cctx.Args().Get(2)
		_, err = strconv.ParseUint(start, 10, 64)
		if err != nil {
			return err
		}

		duration := cctx.Args().Get(3)
		_, err = strconv.ParseUint(duration, 10, 64)
		if err != nil {
			return err
		}

		amount := cctx.Args().Get(4)
		_, err = types.ParseFIL(amount)
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigLockCancel(baseParams, from.String(), msig.String(), txid, start, duration, amount)
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigThresholdProposeCmd = &cli.Command{
	Name:      "threshold-propose",
	Usage:     "Propose setting a different signing threshold on the account",
	ArgsUsage: "[multisigAddress newThreshold]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the proposal from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must pass multisig address and new threshold value")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		newThreshold := cctx.Args().Get(1)
		_, err = strconv.ParseUint(newThreshold, 10, 64)
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigThresholdPropose(baseParams, from.String(), msig.String(), newThreshold)
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigThresholdApproveCmd = &cli.Command{
	Name:      "threshold-approve",
	Usage:     "Approve a message to setting a different signing threshold on the account",
	ArgsUsage: "[multisigAddress proposerAddress txId newThreshold]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the approve message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 4 {
			return fmt.Errorf("must pass multisig address, proposer address, transaction id, newM")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		prop, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		txid := cctx.Args().Get(2)
		_, err = strconv.ParseUint(txid, 10, 64)
		if err != nil {
			return err
		}

		newThreshold := cctx.Args().Get(3)
		_, err = strconv.ParseUint(newThreshold, 10, 64)
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigThresholdApprove(baseParams, from.String(), msig.String(), prop.String(), txid, newThreshold)
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigThresholdCancelCmd = &cli.Command{
	Name:      "threshold-cancel",
	Usage:     "Cancel a message to setting a different signing threshold on the account",
	ArgsUsage: "[multisigAddress txId newThreshold]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "from",
			Aliases: []string{"f"},
			Usage:   "account to send the approve message from",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 3 {
			return fmt.Errorf("must pass multisig address, proposer address, transaction id, newM")
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		txid := cctx.Args().Get(1)
		_, err = strconv.ParseUint(txid, 10, 64)
		if err != nil {
			return err
		}

		newThreshold := cctx.Args().Get(2)
		_, err = strconv.ParseUint(newThreshold, 10, 64)
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
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

		msg, err := walletAPI.MsigThresholdCancel(baseParams, from.String(), msig.String(), txid, newThreshold)
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigWithdrawBalanceProposeCmd = &cli.Command{
	Name:  "withdraw-propose",
	Usage: "Propose to withdraw FIL from the miner",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[amount]",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must pass amount to withdraw")
		}

		multisigAddr, from, minerAddr, err := getInputs(cctx)
		if err != nil {
			return err
		}

		val, err := types.ParseFIL(cctx.Args().First())
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

		msg, err := walletAPI.MsigWithdrawPropose(baseParams, from.String(), multisigAddr.String(), minerAddr.String(), val.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigWithdrawBalanceApproveCmd = &cli.Command{
	Name:  "withdraw-approve",
	Usage: "Approve to withdraw FIL from the miner",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[amount txnId proposer]",
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() != 3 {
			return fmt.Errorf("must pass amount, txn Id, and proposer address")
		}

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
		if err != nil {
			return err
		}

		val, err := types.ParseFIL(cctx.Args().First())
		if err != nil {
			return err
		}

		prop, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		txid := cctx.Args().Get(2)
		_, err = strconv.ParseUint(txid, 10, 64)
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

		msg, err := walletAPI.MsigWithdrawApprove(baseParams, sender.String(), multisigAddr.String(), prop.String(), txid, minerAddr.String(), val.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigChangeOwnerProposeCmd = &cli.Command{
	Name:  "change-owner-propose",
	Usage: "Propose an owner address change",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[newOwner]",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must pass new owner address")
		}

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
		if err != nil {
			return err
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

		msg, err := walletAPI.MsigChangeOwnerPropose(baseParams, sender.String(), multisigAddr.String(), minerAddr.String(), na.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigChangeOwnerApproveCmd = &cli.Command{
	Name:  "change-owner-approve",
	Usage: "Approve an owner address change",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[newOwner txnId proposer]",
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() != 3 {
			return fmt.Errorf("must pass new owner address, txn Id, and proposer address")
		}

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
		if err != nil {
			return err
		}

		na, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		txid := cctx.Args().Get(1)
		_, err = strconv.ParseUint(txid, 10, 64)
		if err != nil {
			return err
		}

		prop, err := address.NewFromString(cctx.Args().Get(2))
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

		msg, err := walletAPI.MsigChangeOwnerApprove(baseParams, sender.String(), multisigAddr.String(), prop.String(), txid, minerAddr.String(), na.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigChangeWorkerProposeCmd = &cli.Command{
	Name:  "change-worker-propose",
	Usage: "Propose an worker address change",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[newWorker]",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must pass new worker address")
		}

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
		if err != nil {
			return err
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

		msg, err := walletAPI.MsigChangeWorkerPropose(baseParams, sender.String(), multisigAddr.String(), minerAddr.String(), na.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigChangeWorkerApproveCmd = &cli.Command{
	Name:  "change-worker-approve",
	Usage: "Approve an owner address change",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[newWorker txnId proposer]",
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() != 3 {
			return fmt.Errorf("must have newWorker, txn Id, and proposer address")
		}

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
		if err != nil {
			return err
		}

		na, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		txid := cctx.Args().Get(1)
		_, err = strconv.ParseUint(txid, 10, 64)
		if err != nil {
			return err
		}

		proposer, err := address.NewFromString(cctx.Args().Get(2))
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

		msg, err := walletAPI.MsigChangeWorkerApprove(baseParams, sender.String(), multisigAddr.String(), proposer.String(), txid, minerAddr.String(), na.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigConfirmChangeWorkerProposeCmd = &cli.Command{
	Name:  "confirm-change-worker-propose",
	Usage: "Confirm an worker address change",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[newWorker]",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must pass new worker address")
		}

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
		if err != nil {
			return err
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

		msg, err := walletAPI.MsigConfirmChangeWorkerPropose(baseParams, sender.String(), multisigAddr.String(), minerAddr.String(), na.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigConfirmChangeWorkerApproveCmd = &cli.Command{
	Name:  "confirm-change-worker-approve",
	Usage: "Confirm an worker address change",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[newWorker txnId proposer]",
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() != 3 {
			return fmt.Errorf("must have newWorker, txn Id, and proposer address")
		}

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
		if err != nil {
			return err
		}

		na, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		txid := cctx.Args().Get(1)
		_, err = strconv.ParseUint(txid, 10, 64)
		if err != nil {
			return err
		}

		proposer, err := address.NewFromString(cctx.Args().Get(2))
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

		msg, err := walletAPI.MsigConfirmChangeWorkerApprove(baseParams, sender.String(), multisigAddr.String(), proposer.String(), txid, minerAddr.String(), na.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigSetControlProposeCmd = &cli.Command{
	Name:  "set-control-propose",
	Usage: "set control address(-es) propose",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[...address]",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must pass new owner address")
		}

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
		if err != nil {
			return err
		}

		for _, addr := range cctx.Args().Slice() {
			_, err = address.NewFromString(addr)
			if err != nil {
				return err
			}
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		baseParams, err := getBaseParams(cctx)
		if err != nil {
			return err
		}

		msg, err := walletAPI.MsigSetControlPropose(baseParams, sender.String(), multisigAddr.String(), minerAddr.String(), cctx.Args().Slice())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigSetControlApproveCmd = &cli.Command{
	Name:  "set-control-approve",
	Usage: "set control address(-es) approve",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[txnId proposer ...address]",
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() == 0 {
			return fmt.Errorf("must have txn Id, and proposer address and ...address")
		}

		txid := cctx.Args().Get(0)

		_, err := strconv.ParseUint(txid, 10, 64)
		if err != nil {
			return err
		}

		proposer, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
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

		msg, err := walletAPI.MsigSetControlApprove(baseParams, sender.String(), multisigAddr.String(), proposer.String(), txid, minerAddr.String(), cctx.Args().Slice())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigChangeBeneficiaryProposeCmd = &cli.Command{
	Name:  "change-beneficiary-propose",
	Usage: "change beneficiary propose",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.BoolFlag{
			Name:    "overwrite-pending-change",
			Aliases: []string{"opc"},
			Usage:   "Overwrite the current beneficiary change proposal",
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[beneficiaryAddress quota expiration]",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must beneficiaryAddress quota expiration")
		}

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
		if err != nil {
			return err
		}

		overwritePendingChange := cctx.Bool("overwrite-pending-change")
		beneficiaryAddress := cctx.Args().Get(0)
		quota := cctx.Args().Get(1)
		expiration := cctx.Args().Get(2)

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		baseParams, err := getBaseParams(cctx)
		if err != nil {
			return err
		}

		msg, err := walletAPI.MsigChangeBeneficiaryPropose(baseParams, sender.String(), multisigAddr.String(), minerAddr.String(), beneficiaryAddress, quota, expiration, overwritePendingChange)
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigChangeBeneficiaryApproveCmd = &cli.Command{
	Name:  "change-beneficiary-approve",
	Usage: "change beneficiary approve",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[txnId proposer beneficiaryAddress quota expiration]",
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() == 0 {
			return fmt.Errorf("must have txn Id, and proposer address and beneficiaryAddress quota expiration")
		}

		txid := cctx.Args().Get(0)

		_, err := strconv.ParseUint(txid, 10, 64)
		if err != nil {
			return err
		}

		proposer, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		beneficiaryAddress := cctx.Args().Get(2)
		quota := cctx.Args().Get(3)
		expiration := cctx.Args().Get(4)

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
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

		msg, err := walletAPI.MsigChangeBeneficiaryApprove(baseParams, sender.String(), multisigAddr.String(), proposer.String(), txid, minerAddr.String(), beneficiaryAddress, quota, expiration)
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigConfirmChangeBeneficiaryProposeCmd = &cli.Command{
	Name:  "confirm-change-beneficiary-propose",
	Usage: "confirm change beneficiary propose",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.BoolFlag{
			Name:    "overwrite-pending-change",
			Aliases: []string{"opc"},
			Usage:   "Overwrite the current beneficiary change proposal",
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	Action: func(cctx *cli.Context) error {
		multisigAddr, sender, minerAddr, err := getInputs(cctx)
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

		msg, err := walletAPI.MsigConfirmChangeBeneficiaryPropose(baseParams, sender.String(), multisigAddr.String(), minerAddr.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

var msigConfirmChangeBeneficiaryApproveCmd = &cli.Command{
	Name:  "confirm-change-beneficiary-approve",
	Usage: "confirm change beneficiary approve",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Aliases:  []string{"f"},
			Usage:    "specify address to send message from",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "multisig",
			Aliases:  []string{"msig"},
			Usage:    "specify multisig that will receive the message",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "miner",
			Aliases:  []string{"m"},
			Usage:    "specify miner being acted upon",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "a path to output tx message",
			Value:   "",
		},
	},
	ArgsUsage: "[txnId proposer]",
	Action: func(cctx *cli.Context) error {
		if cctx.NArg() == 0 {
			return fmt.Errorf("must have txn Id, and proposer address")
		}

		txid := cctx.Args().Get(0)

		_, err := strconv.ParseUint(txid, 10, 64)
		if err != nil {
			return err
		}

		proposer, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		multisigAddr, sender, minerAddr, err := getInputs(cctx)
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

		msg, err := walletAPI.MsigConfirmChangeBeneficiaryApprove(baseParams, sender.String(), multisigAddr.String(), proposer.String(), txid, minerAddr.String())
		if err != nil {
			return err
		}

		return printMessage(cctx, msg)
	},
}

func getInputs(cctx *cli.Context) (address.Address, address.Address, address.Address, error) {
	multisigAddr, err := address.NewFromString(cctx.String("multisig"))
	if err != nil {
		return address.Undef, address.Undef, address.Undef, err
	}

	sender, err := address.NewFromString(cctx.String("from"))
	if err != nil {
		return address.Undef, address.Undef, address.Undef, err
	}

	minerAddr, err := address.NewFromString(cctx.String("miner"))
	if err != nil {
		return address.Undef, address.Undef, address.Undef, err
	}

	return multisigAddr, sender, minerAddr, nil
}
