package main

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/lotus/api"
	"github.com/urfave/cli/v2"
)

var authCmd = &cli.Command{
	Name:  "auth",
	Usage: "Manage RPC permissions",
	Subcommands: []*cli.Command{
		AuthCreateAdminToken,
	},
}

var AuthCreateAdminToken = &cli.Command{
	Name:  "create-token",
	Usage: "Create token",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "perm",
			Usage: "permission to assign to the token, one of: read, write, sign, admin",
			Value: "read",
		},
	},

	Action: func(cctx *cli.Context) error {
		db, closer, err := getWalletDB(cctx, true)
		if err != nil {
			return err
		}
		defer closer()

		if err := requirePassword(db); err != nil {
			return err
		}

		loginScrypt, err := db.GetLoginPassword()
		if err != nil {
			return err
		}

		app.SetSecret(loginScrypt)

		perm := cctx.String("perm")
		idx := 0
		for i, p := range api.AllPermissions {
			if auth.Permission(perm) == p {
				idx = i + 1
			}
		}

		if idx == 0 {
			return fmt.Errorf("--perm flag has to be one of: %s", api.AllPermissions)
		}

		token, err := app.AuthNew(app.AllPermissions[:idx])
		if err != nil {
			return err
		}

		fmt.Println(string(token))
		return nil
	},
}
