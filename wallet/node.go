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

var glifNodeInfo = &datastore.NodeInfo{
	Name:     "glif",
	Endpoint: "https://api.node.glif.io/rpc/v0",
	Token:    "",
}

type node struct {
	name         string
	nodeEndpoint string
	token        string
	height       string
	*client.LotusClient
}

func newNode(name, nodeEndpoint, nodeToken string) (*node, error) {
	lotusClient, err := client.NewLotusClient(nodeEndpoint, nodeToken)
	if err != nil {
		return nil, err
	}

	head, err := lotusClient.Api.ChainHead(context.Background())
	if err != nil {
		log.Warnw("newNode: ChainHead", "err", err)
		return nil, fmt.Errorf("nodeEndpoint: %s is bad", nodeEndpoint)
	}

	return &node{
		name,
		nodeEndpoint,
		nodeToken,
		head.Height().String(),
		lotusClient,
	}, nil
}

// NodeAdd Post
func (w *Wallet) NodeAdd(c *gin.Context) {
	param := client.NodeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("NodeAdd: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}

	_, err = newNode(param.Name, param.Endpoint, param.Token)
	if err != nil {
		log.Warnw("NodeAdd: newNode", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	err = w.db.SetNode(&datastore.NodeInfo{
		Name:     param.Name,
		Endpoint: param.Endpoint,
		Token:    param.Token,
	})
	if err != nil {
		log.Warnw("NodeAdd: SetNode", "err", err)
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
		log.Warnw("NodeUpdate: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}

	_, err = w.db.GetNode(param.Name)
	if err != nil {
		log.Warnw("NodeUpdate: GetNode", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	err = w.db.UpdateNode(&datastore.NodeInfo{
		Name:     param.Name,
		Endpoint: param.Endpoint,
		Token:    param.Token,
	})
	if err != nil {
		log.Warnw("NodeUpdate: UpdateNode", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, nil)
}

func (w *Wallet) NodeDelete(c *gin.Context) {
	param := client.NodeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("NodeDelete: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}

	if w.node.name == param.Name {
		ReturnError(c, NewError(500, "unable to delete node in use"))
		return
	}

	err = w.db.DeleteNode(param.Name)
	if err != nil {
		log.Warnw("NodeDelete: DeleteNode", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, nil)
}

func (w *Wallet) UseNode(c *gin.Context) {
	param := client.NodeRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("UseNode: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}

	if w.node.name == param.Name {
		ReturnError(c, NewError(500, "already in use"))
		return
	}

	nodeInfo := glifNodeInfo
	if param.Name != glifNodeInfo.Name {
		nodeInfo, err = w.db.GetNode(param.Name)
		if err != nil {
			log.Warnw("UseNode: GetNode", "err", err)
			ReturnError(c, NewError(500, err.Error()))
			return
		}
	}

	n, err := newNode(nodeInfo.Name, nodeInfo.Endpoint, nodeInfo.Token)
	if err != nil {
		log.Warnw("UseNode: newNode", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}
	w.node = n
	w.txTracker.node = n

	ReturnOk(c, nil)
}

// NodeList Get
func (w *Wallet) NodeList(c *gin.Context) {
	nodeInfos, err := w.db.NodeList()
	if err != nil {
		log.Warnw("NodeList: NodeList", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	// Add a default node
	nodeInfos = append(nodeInfos, *glifNodeInfo)

	var nis = []client.NodeInfo{}

	for _, ni := range nodeInfos {
		isUsing := false
		if ni.Name == w.node.name {
			isUsing = true
		}
		n, err := newNode(ni.Name, ni.Endpoint, ni.Token)
		if err != nil {
			n = &node{
				height: "0",
			}
		}
		nis = append(nis, client.NodeInfo{
			Name:        ni.Name,
			Endpoint:    ni.Endpoint,
			Token:       ni.Token,
			IsUsing:     isUsing,
			BlockHeight: n.height,
		})
	}

	ReturnOk(c, nis)
	return
}

// NodeBest Get
func (w *Wallet) NodeBest(c *gin.Context) {
	nodeInfo, err := w.getBestNode()
	if err != nil {
		log.Warnw("NodeBest: getBestNode", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	n, err := newNode(nodeInfo.Name, nodeInfo.Endpoint, nodeInfo.Token)
	if err != nil {
		log.Warnw("NodeBest: newNode", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}
	w.node = n

	ReturnOk(c, client.NodeInfo{
		Name:        nodeInfo.Name,
		Endpoint:    nodeInfo.Endpoint,
		Token:       nodeInfo.Token,
		BlockHeight: n.height,
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
