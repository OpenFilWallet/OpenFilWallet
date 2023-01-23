package main

import (
	"context"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/urfave/cli/v2"
	"text/tabwriter"
)

var nodeCmd = &cli.Command{
	Name:  "node",
	Usage: "The daemon node used by the OpenFilWallet wallet ",
	Subcommands: []*cli.Command{
		addNodeCmd,
		updateNodeCmd,
		deleteNodeCmd,
		useNodeCmd,
		nodeListCmd,
		bestNodeCmd,
	},
}

var addNodeCmd = &cli.Command{
	Name:  "add",
	Usage: "add a node",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Usage:    "node name",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "endpoint",
			Usage:    "node endpoint",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "token",
			Usage: "node token",
		},
	},
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		name := cctx.String("name")
		endpoint := cctx.String("endpoint")
		token := cctx.String("token")

		api, err := client.NewLotusClient(endpoint, token)
		if err != nil {
			return fmt.Errorf("unavailable node: %s", err.Error())
		}

		_, err = api.Api.ChainHead(context.Background())
		if err != nil {
			return fmt.Errorf("unavailable node: %s", err.Error())
		}

		err = walletAPI.NodeAdd(name, endpoint, token)
		if err != nil {
			return fmt.Errorf("failed to add node: %s", err.Error())
		}

		fmt.Println("node added successfully")
		return nil
	},
}

var updateNodeCmd = &cli.Command{
	Name:  "update",
	Usage: "update a node",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Aliases:  []string{"nm"},
			Usage:    "node name",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "endpoint",
			Aliases:  []string{"ep"},
			Usage:    "node endpoint",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "token",
			Aliases: []string{"t"},
			Usage:   "node token",
		},
	},
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		name := cctx.String("name")
		endpoint := cctx.String("endpoint")
		token := cctx.String("token")

		err = walletAPI.NodeUpdate(name, endpoint, token)
		if err != nil {
			return err
		}

		fmt.Println("node update successfully")
		return nil
	},
}

var deleteNodeCmd = &cli.Command{
	Name:  "delete",
	Usage: "delete a node",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Aliases:  []string{"nm"},
			Usage:    "node name",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		name := cctx.String("name")

		err = walletAPI.NodeDelete(name)
		if err != nil {
			return err
		}

		fmt.Println("node delete successfully")
		return nil
	},
}

var useNodeCmd = &cli.Command{
	Name:  "use",
	Usage: "use node",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Aliases:  []string{"nm"},
			Usage:    "node name",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		name := cctx.String("name")

		err = walletAPI.UseNode(name)
		if err != nil {
			return err
		}

		fmt.Println("node use successfully")
		return nil
	},
}

var nodeListCmd = &cli.Command{
	Name:  "list",
	Usage: "node list",
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		nodeInfos, err := walletAPI.NodeList()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(cctx.App.Writer, 8, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tEndpoint\tToken\n")

		for i, nodeInfo := range nodeInfos {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", i, nodeInfo.Name, nodeInfo.Endpoint, nodeInfo.Token)
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("flushing output: %+v", err)
		}

		return nil
	},
}

var bestNodeCmd = &cli.Command{
	Name:  "best",
	Usage: "best node",
	Action: func(cctx *cli.Context) error {
		walletAPI, err := client.GetOpenFilAPI(cctx)
		if err != nil {
			return err
		}

		nodeInfo, err := walletAPI.NodeBest()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(cctx.App.Writer, 8, 4, 2, ' ', 0)
		fmt.Fprintf(w, "Name\tEndpoint\tToken\n")

		fmt.Fprintf(w, "%s\t%s\t%s\n", nodeInfo.Name, nodeInfo.Endpoint, nodeInfo.Token)

		if err := w.Flush(); err != nil {
			return fmt.Errorf("flushing output: %+v", err)
		}
		return nil
	},
}
