package client

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	lc "github.com/filecoin-project/lotus/api/client"
	"net/http"
)

type LotusClient struct {
	Api    api.FullNode
	Closer jsonrpc.ClientCloser
}

func NewLotusClient(nodeEndpoint, nodeToken string) (*LotusClient, error) {
	lotusApi, closer, err := newLotusAPI(nodeEndpoint, nodeToken)
	if err != nil {
		return nil, err
	}
	return &LotusClient{
		Api:    lotusApi,
		Closer: closer,
	}, err
}

func newLotusAPI(nodeEndpoint, nodeToken string) (api.FullNode, jsonrpc.ClientCloser, error) {
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")

	if nodeToken != "" {
		requestHeader.Set("Authorization", fmt.Sprintf("Bearer %s", nodeToken))
	}

	return lc.NewFullNodeRPCV1(context.Background(), nodeEndpoint, requestHeader)
}
