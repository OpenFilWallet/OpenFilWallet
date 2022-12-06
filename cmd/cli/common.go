package main

import (
	"encoding/json"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/modules/buildmessage"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/urfave/cli/v2"
	"os"
)

func getBaseParams(cctx *cli.Context) (buildmessage.BaseParams, error) {
	var baseParams buildmessage.BaseParams
	if cctx.IsSet("gas-premium") {
		gasPremium, err := types.BigFromString(cctx.String("gas-premium"))
		if err != nil {
			return buildmessage.BaseParams{}, fmt.Errorf("parsing gas-premium : %s", err)
		}
		baseParams.GasPremium = gasPremium.String()
	}

	if cctx.IsSet("gas-feecap") {
		gasFeeCap, err := types.BigFromString(cctx.String("gas-feecap"))
		if err != nil {
			return buildmessage.BaseParams{}, fmt.Errorf("parsing gas-feecap: %w", err)
		}
		baseParams.GasFeeCap = gasFeeCap.String()
	}

	baseParams.GasLimit = cctx.Int64("gas-limit")
	baseParams.MaxFee = cctx.String("max-fee")

	return baseParams, nil
}

func printMessage(cctx *cli.Context, msg interface{}) error {
	v, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return err
	}

	if cctx.IsSet("output") {
		output := cctx.String("output")
		fi, err := os.Open(output)
		if err != nil {
			log.Warnf("open file (path: %s): %s \n", output, err)

			fmt.Println(string(v))
			return nil
		}

		defer fi.Close()

		_, err = fi.Write(v)
		if err != nil {
			log.Warnf("save message: %s \n", err)
			fmt.Println(string(v))
			return nil
		}

		return nil
	} else {
		fmt.Println(string(v))
		return nil
	}
}

func getLotusAPI(cctx *cli.Context) (*client.LotusClient, error) {
	walletAPI, err := client.GetOpenFilAPI(cctx)
	if err != nil {
		return nil, err
	}

	nodeInfo, err := walletAPI.NodeBest()
	if err != nil {
		return nil, err
	}

	return client.NewLotusClient(nodeInfo.Endpoint, nodeInfo.Token)
}
