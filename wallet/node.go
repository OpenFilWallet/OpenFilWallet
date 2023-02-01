package wallet

import (
	"context"
	"errors"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/gin-gonic/gin"
	"time"
)

type node struct {
	name         string
	nodeEndpoint string
	token        string
	*client.LotusClient
}

func newNode(name, nodeEndpoint, nodeToken string) (*node, error) {
	lotusClient, err := client.NewLotusClient(nodeEndpoint, nodeToken)
	if err != nil {
		return nil, err
	}

	_, err = lotusClient.Api.ChainHead(context.Background())
	if err != nil {
		return nil, fmt.Errorf("nodeEndpoint: %s is bad", nodeEndpoint)
	}

	return &node{
		name,
		nodeEndpoint,
		nodeToken,
		lotusClient,
	}, nil
}

// NodeAdd Post
func (w *Wallet) NodeAdd(c *gin.Context) {
	param := client.NodeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	_, err = newNode(param.Name, param.Endpoint, param.Token)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	err = w.db.SetNode(&datastore.NodeInfo{
		Name:     param.Name,
		Endpoint: param.Endpoint,
		Token:    param.Token,
	})
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, nil)
}

// NodeUpdate Post
func (w *Wallet) NodeUpdate(c *gin.Context) {
	param := client.NodeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	_, err = w.db.GetNode(param.Name)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	err = w.db.UpdateNode(&datastore.NodeInfo{
		Name:     param.Name,
		Endpoint: param.Endpoint,
		Token:    param.Token,
	})
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, nil)
}

func (w *Wallet) NodeDelete(c *gin.Context) {
	param := client.NodeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	if w.node.name == param.Name {
		ReturnError(c, NewError(500, "unable to delete node in use"))
		return
	}

	err = w.db.DeleteNode(param.Name)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, nil)
}

func (w *Wallet) UseNode(c *gin.Context) {
	param := client.NodeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	if w.node.name == param.Name {
		ReturnError(c, NewError(500, "already in use"))
		return
	}

	nodeInfo, err := w.db.GetNode(param.Name)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	node, err := newNode(nodeInfo.Name, nodeInfo.Endpoint, nodeInfo.Token)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}
	w.node = node
	w.txTracker.node = node

	ReturnOk(c, nil)
}

// NodeList Get
func (w *Wallet) NodeList(c *gin.Context) {
	nodeInfos, err := w.db.NodeList()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	// Add a default node
	nodeInfos = append(nodeInfos, datastore.NodeInfo{
		Name:     "glif",
		Endpoint: "https://api.node.glif.io/rpc/v0",
		Token:    "",
	})
	var nis = []client.NodeInfo{}

	for _, ni := range nodeInfos {
		nis = append(nis, client.NodeInfo{
			Name:     ni.Name,
			Endpoint: ni.Endpoint,
			Token:    ni.Token,
		})
	}

	ReturnOk(c, nis)
	return
}

// NodeBest Get
func (w *Wallet) NodeBest(c *gin.Context) {
	nodeInfo, err := w.getBestNode()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	node, err := newNode(nodeInfo.Name, nodeInfo.Endpoint, nodeInfo.Token)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}
	w.node = node

	ReturnOk(c, client.NodeInfo{
		Name:     nodeInfo.Name,
		Endpoint: nodeInfo.Endpoint,
		Token:    nodeInfo.Token,
	})
}

func (w *Wallet) getBestNode() (*datastore.NodeInfo, error) {
	nodeInfos, err := w.db.NodeList()
	if err != nil {
		return nil, err
	}

	// Add a default node
	nodeInfos = append(nodeInfos, datastore.NodeInfo{
		Name:     "glif",
		Endpoint: "https://api.node.glif.io/rpc/v0",
		Token:    "",
	})

	bestIndex := -1
	bestElapsed := time.Duration(0)
	for i, nodeInfo := range nodeInfos {
		start := time.Now()
		_, err = newNode(nodeInfo.Name, nodeInfo.Endpoint, nodeInfo.Token)
		if err != nil {
			continue
		}

		elapsed := time.Since(start)
		if bestIndex == -1 {
			bestIndex = i
			bestElapsed = elapsed
			continue
		}

		if elapsed < bestElapsed {
			bestIndex = i
			bestElapsed = elapsed
		}
	}

	if bestIndex == -1 {
		return nil, errors.New("no node available")
	}

	return &nodeInfos[bestIndex], nil
}
