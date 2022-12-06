package main

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"strconv"
)

var chainCmd = &cli.Command{
	Name:  "chain",
	Usage: "Interact with filecoin blockchain",
	Subcommands: []*cli.Command{
		decodeCmd,
		encodeCmd,
	},
}

var decodeCmd = &cli.Command{
	Name:  "decode",
	Usage: "decode various types",
	Subcommands: []*cli.Command{
		decodeParamsCmd,
	},
}

var decodeParamsCmd = &cli.Command{
	Name:      "params",
	Usage:     "Decode message params",
	ArgsUsage: "[toAddr method params]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "encoding",
			Value: "base64",
			Usage: "specify input encoding to parse",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 3 {
			return fmt.Errorf("incorrect number of arguments")
		}

		var params = cctx.Args().Get(2)
		var err error
		switch cctx.String("encoding") {
		case "base64":
		case "hex":
		default:
			return xerrors.Errorf("unrecognized encoding: %s", cctx.String("encoding"))
		}

		method, err := strconv.ParseInt(cctx.Args().Get(1), 10, 64)
		if err != nil {
			return xerrors.Errorf("parsing method id: %w", err)
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		decParams, err := walletAPI.Decode(cctx.Args().Get(0), uint64(method), params, cctx.String("encoding"))
		if err != nil {
			return err
		}

		fmt.Println(decParams)

		return nil
	},
}

var encodeCmd = &cli.Command{
	Name:  "encode",
	Usage: "encode various types",
	Subcommands: []*cli.Command{
		encodeParamsCmd,
	},
}

var encodeParamsCmd = &cli.Command{
	Name:      "params",
	Usage:     "Encodes the given JSON params, encoding: hex",
	ArgsUsage: "[dest method params]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "encoding",
			Value: "base64",
			Usage: "specify input encoding to parse",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 3 {
			return fmt.Errorf("incorrect number of arguments")
		}

		method, err := strconv.ParseInt(cctx.Args().Get(1), 10, 64)
		if err != nil {
			return xerrors.Errorf("parsing method id: %w", err)
		}

		switch cctx.String("encoding") {
		case "base64", "b64":
		case "hex":
		default:
			return xerrors.Errorf("unknown encoding")
		}

		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		encParams, err := walletAPI.Encode(cctx.Args().Get(0), uint64(method), cctx.Args().Get(2), cctx.String("encoding"))
		if err != nil {
			return err
		}

		fmt.Println(encParams)
		return nil
	},
}
