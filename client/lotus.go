package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	lc "github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
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

type Method string

const StateSearchMsg Method = "Filecoin.StateSearchMsg"

type client struct {
	rpcAddr string
	token   string
	JsonRpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int         `json:"id"`
}

type LotusResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Id      int         `json:"id"`
	Error   interface{} `json:"error"`
}

type MsgLookup struct {
	Message   cid.Cid
	Receipt   types.MessageReceipt
	ReturnDec interface{}
	TipSet    types.TipSetKey
	Height    abi.ChainEpoch
}

func newClient(rpcAddr, token string, method Method, params interface{}) *client {
	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(100)
	return &client{
		rpcAddr: rpcAddr,
		token:   token,
		JsonRpc: "2.0",
		Method:  string(method),
		Params:  params,
		Id:      id,
	}
}

func LotusStateSearchMsg(rpcAddr, token string, msgCidStr string) (*MsgLookup, error) {
	var params []interface{}
	msgCid, err := cid.Parse(msgCidStr)
	if err != nil {
		return nil, err
	}
	params = append(params, &msgCid)

	result, err := newClient(rpcAddr, token, StateSearchMsg, params).Call()
	if err != nil {
		return nil, err
	}

	r := LotusResponse{Result: &MsgLookup{}}
	err = json.Unmarshal(result, &r)
	if err != nil {
		return nil, err
	}
	if r.Error != nil {
		return nil, xerrors.Errorf("error: %s", r.Error.(map[string]interface{})["message"])
	}

	if r.Result != nil {
		return r.Result.(*MsgLookup), nil
	}

	return nil, nil
}

func (c *client) Call() ([]byte, error) {
	dataByte, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.rpcAddr, bytes.NewBuffer(dataByte))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if len(strings.Trim(c.token, " ")) > 0 {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, xerrors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
