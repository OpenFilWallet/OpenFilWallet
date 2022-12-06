package main

import (
	"errors"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/urfave/cli/v2"
)

var loginCmd = &cli.Command{
	Name:  "login",
	Usage: "login openfil wallet",
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		password, err := loginPassword()
		if err != nil {
			return err
		}

		err = walletAPI.Login(password)
		if err != nil {
			return err
		}

		fmt.Println("login successful")
		return nil
	},
}

func loginPassword() (string, error) {
	fmt.Println("Please enter login password")
	for i := 0; i < 3; i++ {
		loginPassword, err := app.Password(false)
		if err != nil {
			continue
		}

		return loginPassword, nil
	}

	return "", errors.New("failed to get password")
}
