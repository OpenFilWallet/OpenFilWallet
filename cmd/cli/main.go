package main

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/build"
	"github.com/OpenFilWallet/OpenFilWallet/repo"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"os"
)

var log = logging.Logger("openfil-cli")

func main() {
	_ = logging.SetLogLevel("*", "INFO")

	app := &cli.App{
		Name:                 "openfil-cli",
		Usage:                "open source hd wallet client for Filecoin",
		Version:              build.Version(),
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    repo.FlagWalletRepo,
				EnvVars: []string{"OPEN_FIL_WALLET_PATH"},
				Value:   "~/.openfilwallet",
				Usage:   fmt.Sprintf("Specify openfilwallet repo path. flag(--wallet-repo) or env(OPEN_FIL_WALLET_PATH)"),
			},
		},
		Commands: []*cli.Command{
			statusCmd,
			loginCmd,
			nodeCmd,
			chainCmd,
			sendCmd,
			signCmd,
			walletCmd,
			minerCmd,
			multisigCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString("Error: " + err.Error() + "\n")
	}
}
