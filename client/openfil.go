package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/chain"
	"github.com/OpenFilWallet/OpenFilWallet/modules/buildmessage"
	"github.com/OpenFilWallet/OpenFilWallet/repo"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type OpenFilAPI struct {
	endpoint string
	token    string
}

func GetOpenFilAPI(ctx *cli.Context) (*OpenFilAPI, error) {
	repoPath := ctx.String(repo.FlagWalletRepo)
	r, err := repo.NewFS(repoPath)
	if err != nil {
		return nil, err
	}

	ok, err := r.Exists()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("repo at '%s' is not initialized, run 'openfild init' to set it up", repo.FlagWalletRepo)
	}

	endpoint, err := r.APIEndpoint()
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	token, err := r.APIToken()
	if err != nil {
		return nil, err
	}

	return &OpenFilAPI{
		endpoint: u.String(),
		token:    string(token),
	}, nil
}

func (api *OpenFilAPI) Status() (*StatusInfo, error) {
	res, err := GetRequest(api.endpoint, "/status", api.token, nil)
	if err != nil {
		return nil, err
	}
	var si StatusInfo
	err = json.Unmarshal(res, &si)
	if err != nil {
		return nil, err
	}

	return &si, nil
}

func (api *OpenFilAPI) Login(loginPassword string) error {
	req := LoginRequest{
		loginPassword,
	}

	res, err := PostRequest(api.endpoint, "/login", api.token, req)
	if err != nil {
		return err
	}

	var li LoginInfo
	err = json.Unmarshal(res, &li)
	if err != nil {
		return err
	}

	return nil
}

func (api *OpenFilAPI) SignOut() error {
	_, err := PostRequest(api.endpoint, "/signout", api.token, nil)
	if err != nil {
		return err
	}

	return nil
}

func (api *OpenFilAPI) Decode(to string, method uint64, params string, encoding string) (string, error) {
	req := DecodeRequest{
		ToAddr:   to,
		Method:   method,
		Params:   params,
		Encoding: encoding,
	}

	res, err := PostRequest(api.endpoint, "/chain/decode", api.token, req)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func (api *OpenFilAPI) Encode(dest string, method uint64, params string, encoding string) (string, error) {
	req := EncodeRequest{
		Dest:     dest,
		Method:   method,
		Params:   params,
		Encoding: encoding,
	}

	res, err := PostRequest(api.endpoint, "/chain/encode", api.token, req)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func (api *OpenFilAPI) NodeAdd(name, endpoint, token string) error {
	req := NodeRequest{
		Name:     name,
		Endpoint: endpoint,
		Token:    token,
	}

	res, err := PostRequest(api.endpoint, "/node/add", api.token, req)
	if err != nil {
		return err
	}

	var r Response
	err = json.Unmarshal(res, &r)
	if err != nil {
		return err
	}

	if r.Code != 200 {
		return errors.New(r.Message)
	}

	return nil
}

func (api *OpenFilAPI) NodeUpdate(name, endpoint, token string) error {
	req := NodeRequest{
		Name:     name,
		Endpoint: endpoint,
		Token:    token,
	}

	res, err := PostRequest(api.endpoint, "/node/update", api.token, req)
	if err != nil {
		return err
	}

	var r Response
	err = json.Unmarshal(res, &r)
	if err != nil {
		return err
	}

	if r.Code != 200 {
		return errors.New(r.Message)
	}

	return nil
}

func (api *OpenFilAPI) UseNode(name string) error {
	req := NodeRequest{
		Name: name,
	}

	res, err := PostRequest(api.endpoint, "/node/use_node", api.token, req)
	if err != nil {
		return err
	}

	var r Response
	err = json.Unmarshal(res, &r)
	if err != nil {
		return err
	}

	if r.Code != 200 {
		return errors.New(r.Message)
	}

	return nil
}

func (api *OpenFilAPI) NodeDelete(name string) error {
	req := NodeRequest{
		Name: name,
	}

	res, err := PostRequest(api.endpoint, "/node/delete", api.token, req)
	if err != nil {
		return err
	}

	var r Response
	err = json.Unmarshal(res, &r)
	if err != nil {
		return err
	}

	if r.Code != 200 {
		return errors.New(r.Message)
	}

	return nil
}

func (api *OpenFilAPI) NodeList() ([]NodeInfo, error) {
	res, err := GetRequest(api.endpoint, "/node/list", api.token, nil)
	if err != nil {
		return nil, err
	}
	var nis []NodeInfo
	err = json.Unmarshal(res, &nis)
	if err != nil {
		return nil, err
	}

	return nis, nil
}

func (api *OpenFilAPI) NodeBest() (*NodeInfo, error) {
	res, err := GetRequest(api.endpoint, "/node/best", api.token, nil)
	if err != nil {
		return nil, err
	}
	var ni NodeInfo
	err = json.Unmarshal(res, &ni)
	if err != nil {
		return nil, err
	}

	return &ni, nil
}

func (api *OpenFilAPI) WalletCreate(index int) (*CreateWalletResponse, error) {
	req := CreateWalletRequest{
		Index: index,
	}

	res, err := PostRequest(api.endpoint, "/wallet/create", api.token, req)
	if err != nil {
		return nil, err
	}

	var r CreateWalletResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) WalletList() ([]WalletListInfo, error) {
	res, err := GetRequest(api.endpoint, "/wallet/list", api.token, nil)
	if err != nil {
		return nil, err
	}

	var r = make([]WalletListInfo, 0)
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (api *OpenFilAPI) MsigWalletList() ([]MsigWalletListInfo, error) {
	res, err := GetRequest(api.endpoint, "/msig/list", api.token, nil)
	if err != nil {
		return nil, err
	}

	var r = make([]MsigWalletListInfo, 0)
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (api *OpenFilAPI) Balance(addr string) (*BalanceInfo, error) {
	res, err := GetRequest(api.endpoint, "/balance", api.token, map[string]string{"address": addr})
	if err != nil {
		return nil, err
	}

	var r BalanceInfo
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) Transfer(baseParams buildmessage.BaseParams, from, to, amount string) (*chain.Message, error) {
	req := TransferRequest{
		BaseParams: baseParams,
		From:       from,
		To:         to,
		Amount:     amount,
	}

	res, err := PostRequest(api.endpoint, "/transfer", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) Send(req chain.SignedMessage) (string, error) {
	res, err := PostRequest(api.endpoint, "/send", api.token, req)
	if err != nil {
		return "", err
	}

	var r Response
	err = json.Unmarshal(res, &r)
	if err != nil {
		return "", err
	}

	if r.Code != 200 {
		return "", errors.New(r.Message)
	}

	return r.Message, nil
}

func (api *OpenFilAPI) TxHistory(addr string) ([]HistoryResponse, error) {
	res, err := GetRequest(api.endpoint, "/tx_history", api.token, map[string]string{"address": addr})
	if err != nil {
		return nil, err
	}

	var r []HistoryResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (api *OpenFilAPI) Sign(req chain.Message) (*chain.SignedMessage, error) {
	res, err := PostRequest(api.endpoint, "/sign", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.SignedMessage
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) SignMsg(from string, msg string) (string, error) {
	req := SingRequest{
		From:       from,
		HexMessage: msg,
	}
	res, err := PostRequest(api.endpoint, "/sign_msg", api.token, req)
	if err != nil {
		return "", err
	}

	var r Response
	err = json.Unmarshal(res, &r)
	if err != nil {
		return "", err
	}

	return r.Message, nil
}

func (api *OpenFilAPI) SignAndSend(req chain.Message) (string, error) {
	res, err := PostRequest(api.endpoint, "/sign_send", api.token, req)
	if err != nil {
		return "", err
	}

	var r Response
	err = json.Unmarshal(res, &r)
	if err != nil {
		return "", err
	}

	return r.Message, nil
}

func (api *OpenFilAPI) Withdraw(baseParams buildmessage.BaseParams, minerId, amount string) (*chain.Message, error) {
	req := WithdrawRequest{
		BaseParams: baseParams,
		MinerId:    minerId,
		Amount:     amount,
	}
	res, err := PostRequest(api.endpoint, "/miner/withdraw", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) ChangeOwner(baseParams buildmessage.BaseParams, minerId string, newOwner string, from string) (*chain.Message, error) {
	req := ChangeOwnerRequest{
		BaseParams: baseParams,
		MinerId:    minerId,
		NewOwner:   newOwner,
		From:       from,
	}
	res, err := PostRequest(api.endpoint, "/miner/change_owner", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) ChangeWorker(baseParams buildmessage.BaseParams, minerId, newWorker string) (*chain.Message, error) {
	req := ChangeWorkerRequest{
		BaseParams: baseParams,
		MinerId:    minerId,
		NewWorker:  newWorker,
	}

	res, err := PostRequest(api.endpoint, "/miner/change_worker", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) ConfirmChangeWorker(baseParams buildmessage.BaseParams, minerId, newWorker string) (*chain.Message, error) {
	req := ConfirmChangeWorkerRequest{
		BaseParams: baseParams,
		MinerId:    minerId,
		NewWorker:  newWorker,
	}

	res, err := PostRequest(api.endpoint, "/miner/confirm_change_worker", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) ChangeBeneficiary(baseParams buildmessage.BaseParams, minerId, beneficiaryAddress, quota, expiration string, OverwritePendingChange bool) (*chain.Message, error) {
	req := ChangeBeneficiaryRequest{
		BaseParams:             baseParams,
		MinerId:                minerId,
		BeneficiaryAddress:     beneficiaryAddress,
		Quota:                  quota,
		Expiration:             expiration,
		OverwritePendingChange: OverwritePendingChange,
	}

	res, err := PostRequest(api.endpoint, "/miner/change_beneficiary", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) ConfirmChangeBeneficiary(baseParams buildmessage.BaseParams, minerId string, existingBeneficiary, newBeneficiary bool) (*chain.Message, error) {
	req := ConfirmChangeBeneficiaryRequest{
		BaseParams:          baseParams,
		MinerId:             minerId,
		ExistingBeneficiary: existingBeneficiary,
		NewBeneficiary:      newBeneficiary,
	}

	res, err := PostRequest(api.endpoint, "/miner/confirm_change_beneficiary", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) ChangeControl(baseParams buildmessage.BaseParams, minerId string, controlAddrs []string) (*chain.Message, error) {
	req := ChangeWorkerRequest{
		BaseParams:      baseParams,
		MinerId:         minerId,
		NewControlAddrs: controlAddrs,
	}

	res, err := PostRequest(api.endpoint, "/miner/change_control", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) ControlList(minerId string) (*MinerControl, error) {
	res, err := GetRequest(api.endpoint, "/miner/control_list", api.token, map[string]string{"miner_id": minerId})
	if err != nil {
		return nil, err
	}

	var r MinerControl
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigInspect(msigAddress string) (*MsigInspect, error) {
	res, err := GetRequest(api.endpoint, "/msig/inspect", api.token, map[string]string{"msig_address": msigAddress})
	if err != nil {
		return nil, err
	}

	var r MsigInspect
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigCreate(baseParams buildmessage.BaseParams, from string, required uint64, duration uint64, value string, signer ...string) (*chain.Message, error) {
	req := MsigCreateRequest{
		BaseParams: baseParams,
		From:       from,
		Required:   required,
		Duration:   duration,
		Value:      value,
		Signers:    signer,
	}

	res, err := PostRequest(api.endpoint, "/msig/create", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigApprove(baseParams buildmessage.BaseParams, from string, msigAddress string, txId string) (*chain.Message, error) {
	req := MsigBaseRequest{
		BaseParams:  baseParams,
		From:        from,
		MsigAddress: msigAddress,
		TxId:        txId,
	}

	res, err := PostRequest(api.endpoint, "/msig/approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigCancel(baseParams buildmessage.BaseParams, from string, msigAddress string, txId string) (*chain.Message, error) {
	req := MsigBaseRequest{
		BaseParams:  baseParams,
		From:        from,
		MsigAddress: msigAddress,
		TxId:        txId,
	}

	res, err := PostRequest(api.endpoint, "/msig/cancel", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigTransferPropose(baseParams buildmessage.BaseParams, from string, msigAddress, destinationAddress string, amount string) (*chain.Message, error) {
	req := MsigTransferProposeRequest{
		BaseParams:         baseParams,
		From:               from,
		MsigAddress:        msigAddress,
		DestinationAddress: destinationAddress,
		Amount:             amount,
	}

	res, err := PostRequest(api.endpoint, "/msig/transfer_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigTransferApprove(baseParams buildmessage.BaseParams, from string, msigAddress string, txId string) (*chain.Message, error) {
	req := MsigBaseRequest{
		BaseParams:  baseParams,
		From:        from,
		MsigAddress: msigAddress,
		TxId:        txId,
	}

	res, err := PostRequest(api.endpoint, "/msig/transfer_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigTransferCancel(baseParams buildmessage.BaseParams, from string, msigAddress string, txId string) (*chain.Message, error) {
	req := MsigBaseRequest{
		BaseParams:  baseParams,
		From:        from,
		MsigAddress: msigAddress,
		TxId:        txId,
	}

	res, err := PostRequest(api.endpoint, "/msig/transfer_cancel", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigAddPropose(baseParams buildmessage.BaseParams, from string, msigAddress string, newSigner string, inc bool) (*chain.Message, error) {
	req := MsigAddSignerProposeRequest{
		BaseParams:        baseParams,
		From:              from,
		MsigAddress:       msigAddress,
		SignerAddress:     newSigner,
		IncreaseThreshold: inc,
	}

	res, err := PostRequest(api.endpoint, "/msig/add_signer_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigAddApprove(baseParams buildmessage.BaseParams, from string, msigAddress, proposerAddress, txId, newSigner string, inc bool) (*chain.Message, error) {
	req := MsigAddSignerApprovRequest{
		BaseParams:        baseParams,
		From:              from,
		MsigAddress:       msigAddress,
		ProposerAddress:   proposerAddress,
		TxId:              txId,
		SignerAddress:     newSigner,
		IncreaseThreshold: inc,
	}

	res, err := PostRequest(api.endpoint, "/msig/add_signer_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigAddCancel(baseParams buildmessage.BaseParams, from string, msigAddress, txId, newSigner string, inc bool) (*chain.Message, error) {
	req := MsigAddSignerCancelRequest{
		BaseParams:        baseParams,
		From:              from,
		MsigAddress:       msigAddress,
		TxId:              txId,
		SignerAddress:     newSigner,
		IncreaseThreshold: inc,
	}

	res, err := PostRequest(api.endpoint, "/msig/add_signer_cancel", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigSwapPropose(baseParams buildmessage.BaseParams, from string, msigAddress, oldAddress, newAddress string) (*chain.Message, error) {
	req := MsigSwapProposeRequest{
		BaseParams:  baseParams,
		From:        from,
		MsigAddress: msigAddress,
		OldAddress:  oldAddress,
		NewAddress:  newAddress,
	}

	res, err := PostRequest(api.endpoint, "/msig/swap_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigSwapApprove(baseParams buildmessage.BaseParams, from string, msigAddress, proposerAddress, txId, oldAddress, newAddress string) (*chain.Message, error) {
	req := MsigSwapApproveRequest{
		BaseParams:      baseParams,
		From:            from,
		ProposerAddress: proposerAddress,
		TxId:            txId,
		MsigAddress:     msigAddress,
		OldAddress:      oldAddress,
		NewAddress:      newAddress,
	}

	res, err := PostRequest(api.endpoint, "/msig/swap_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigSwapCancel(baseParams buildmessage.BaseParams, from string, msigAddress, txId, oldAddress, newAddress string) (*chain.Message, error) {
	req := MsigSwapCancelRequest{
		BaseParams:  baseParams,
		From:        from,
		TxId:        txId,
		MsigAddress: msigAddress,
		OldAddress:  oldAddress,
		NewAddress:  newAddress,
	}

	res, err := PostRequest(api.endpoint, "/msig/swap_cancel", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigLockPropose(baseParams buildmessage.BaseParams, from string, msigAddress, startEpoch, unlockDuration, amount string) (*chain.Message, error) {
	req := MsigLockProposeRequest{
		BaseParams:     baseParams,
		From:           from,
		MsigAddress:    msigAddress,
		StartEpoch:     startEpoch,
		UnlockDuration: unlockDuration,
		Amount:         amount,
	}

	res, err := PostRequest(api.endpoint, "/msig/lock_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigLockApprove(baseParams buildmessage.BaseParams, from string, msigAddress, proposerAddress, txId, startEpoch, unlockDuration, amount string) (*chain.Message, error) {
	req := MsigLockApproveRequest{
		BaseParams:      baseParams,
		From:            from,
		MsigAddress:     msigAddress,
		ProposerAddress: proposerAddress,
		TxId:            txId,
		StartEpoch:      startEpoch,
		UnlockDuration:  unlockDuration,
		Amount:          amount,
	}

	res, err := PostRequest(api.endpoint, "/msig/lock_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigLockCancel(baseParams buildmessage.BaseParams, from string, msigAddress, txId, startEpoch, unlockDuration, amount string) (*chain.Message, error) {
	req := MsigLockCancelRequest{
		BaseParams:     baseParams,
		From:           from,
		MsigAddress:    msigAddress,
		TxId:           txId,
		StartEpoch:     startEpoch,
		UnlockDuration: unlockDuration,
		Amount:         amount,
	}

	res, err := PostRequest(api.endpoint, "/msig/lock_cancel", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigThresholdPropose(baseParams buildmessage.BaseParams, from string, msigAddress, newThreshold string) (*chain.Message, error) {
	req := MsigThresholdProposeRequest{
		BaseParams:   baseParams,
		From:         from,
		MsigAddress:  msigAddress,
		NewThreshold: newThreshold,
	}

	res, err := PostRequest(api.endpoint, "/msig/threshold_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigThresholdApprove(baseParams buildmessage.BaseParams, from string, msigAddress, proposerAddress, txId, newThreshold string) (*chain.Message, error) {
	req := MsigThresholdApproveRequest{
		BaseParams:      baseParams,
		From:            from,
		ProposerAddress: proposerAddress,
		TxId:            txId,
		MsigAddress:     msigAddress,
		NewThreshold:    newThreshold,
	}

	res, err := PostRequest(api.endpoint, "/msig/threshold_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigThresholdCancel(baseParams buildmessage.BaseParams, from string, msigAddress, txId, newThreshold string) (*chain.Message, error) {
	req := MsigThresholdCancelRequest{
		BaseParams:   baseParams,
		From:         from,
		TxId:         txId,
		MsigAddress:  msigAddress,
		NewThreshold: newThreshold,
	}

	res, err := PostRequest(api.endpoint, "/msig/threshold_cancel", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigChangeOwnerPropose(baseParams buildmessage.BaseParams, from string, msigAddress, minerId, newOwner string) (*chain.Message, error) {
	req := MsigChangeOwnerProposeRequest{
		BaseParams:  baseParams,
		From:        from,
		MsigAddress: msigAddress,
		MinerId:     minerId,
		NewOwner:    newOwner,
	}

	res, err := PostRequest(api.endpoint, "/msig/change_owner_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigChangeOwnerApprove(baseParams buildmessage.BaseParams, from string, msigAddress, proposerAddress, txId, minerId, newOwner string) (*chain.Message, error) {
	req := MsigChangeOwnerApproveRequest{
		BaseParams:      baseParams,
		From:            from,
		MsigAddress:     msigAddress,
		ProposerAddress: proposerAddress,
		TxId:            txId,
		MinerId:         minerId,
		NewOwner:        newOwner,
	}

	res, err := PostRequest(api.endpoint, "/msig/change_owner_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigWithdrawPropose(baseParams buildmessage.BaseParams, from string, msigAddress, minerId, amount string) (*chain.Message, error) {
	req := MsigWithdrawProposeRequest{
		BaseParams:  baseParams,
		From:        from,
		MsigAddress: msigAddress,
		MinerId:     minerId,
		Amount:      amount,
	}

	res, err := PostRequest(api.endpoint, "/msig/withdraw_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigWithdrawApprove(baseParams buildmessage.BaseParams, from string, msigAddress, proposerAddress, txId, minerId, amount string) (*chain.Message, error) {
	req := MsigWithdrawApproveRequest{
		BaseParams:      baseParams,
		From:            from,
		ProposerAddress: proposerAddress,
		TxId:            txId,
		MsigAddress:     msigAddress,
		MinerId:         minerId,
		Amount:          amount,
	}

	res, err := PostRequest(api.endpoint, "/msig/withdraw_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigChangeWorkerPropose(baseParams buildmessage.BaseParams, from string, msigAddress, minerId, newWorker string) (*chain.Message, error) {
	req := MsigChangeWorkerProposeRequest{
		BaseParams:  baseParams,
		From:        from,
		MsigAddress: msigAddress,
		MinerId:     minerId,
		NewWorker:   newWorker,
	}

	res, err := PostRequest(api.endpoint, "/msig/change_worker_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigChangeWorkerApprove(baseParams buildmessage.BaseParams, from string, msigAddress, proposerAddress, txId, minerId, newWorker string) (*chain.Message, error) {
	req := MsigChangeWorkerApproveRequest{
		BaseParams:      baseParams,
		From:            from,
		ProposerAddress: proposerAddress,
		TxId:            txId,
		MsigAddress:     msigAddress,
		MinerId:         minerId,
		NewWorker:       newWorker,
	}

	res, err := PostRequest(api.endpoint, "/msig/change_worker_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigConfirmChangeWorkerPropose(baseParams buildmessage.BaseParams, from string, msigAddress, minerId, newWorker string) (*chain.Message, error) {
	req := MsigChangeWorkerProposeRequest{
		BaseParams:  baseParams,
		From:        from,
		MsigAddress: msigAddress,
		MinerId:     minerId,
		NewWorker:   newWorker,
	}

	res, err := PostRequest(api.endpoint, "/msig/confirm_change_worker_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigConfirmChangeWorkerApprove(baseParams buildmessage.BaseParams, from string, msigAddress, proposerAddress, txId, minerId, newWorker string) (*chain.Message, error) {
	req := MsigChangeWorkerApproveRequest{
		BaseParams:      baseParams,
		From:            from,
		ProposerAddress: proposerAddress,
		TxId:            txId,
		MsigAddress:     msigAddress,
		MinerId:         minerId,
		NewWorker:       newWorker,
	}

	res, err := PostRequest(api.endpoint, "/msig/confirm_change_worker_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigChangeBeneficiaryPropose(baseParams buildmessage.BaseParams, from string, msigAddress, minerId string, beneficiaryAddress, quota, expiration string, OverwritePendingChange bool) (*chain.Message, error) {
	req := MsigChangeBeneficiaryProposeRequest{
		BaseParams:             baseParams,
		From:                   from,
		MsigAddress:            msigAddress,
		MinerId:                minerId,
		BeneficiaryAddress:     beneficiaryAddress,
		Quota:                  quota,
		Expiration:             expiration,
		OverwritePendingChange: OverwritePendingChange,
	}

	res, err := PostRequest(api.endpoint, "/msig/change_beneficiary_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigChangeBeneficiaryApprove(baseParams buildmessage.BaseParams, from string, msigAddress, proposerAddress, txId, minerId string, beneficiaryAddress, quota, expiration string) (*chain.Message, error) {
	req := MsigChangeBeneficiaryApproveRequest{
		BaseParams:         baseParams,
		From:               from,
		ProposerAddress:    proposerAddress,
		TxId:               txId,
		MsigAddress:        msigAddress,
		MinerId:            minerId,
		BeneficiaryAddress: beneficiaryAddress,
		Quota:              quota,
		Expiration:         expiration,
	}

	res, err := PostRequest(api.endpoint, "/msig/change_beneficiary_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigConfirmChangeBeneficiaryPropose(baseParams buildmessage.BaseParams, from string, msigAddress, minerId string) (*chain.Message, error) {
	req := MsigConfirmChangeBeneficiaryProposeRequest{
		BaseParams:  baseParams,
		From:        from,
		MsigAddress: msigAddress,
		MinerId:     minerId,
	}

	res, err := PostRequest(api.endpoint, "/msig/confirm_change_beneficiary_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigConfirmChangeBeneficiaryApprove(baseParams buildmessage.BaseParams, from string, msigAddress, proposerAddress, txId, minerId string) (*chain.Message, error) {
	req := MsigConfirmChangeBeneficiaryApproveRequest{
		BaseParams:      baseParams,
		From:            from,
		ProposerAddress: proposerAddress,
		TxId:            txId,
		MsigAddress:     msigAddress,
		MinerId:         minerId,
	}

	res, err := PostRequest(api.endpoint, "/msig/confirm_change_beneficiary_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigSetControlPropose(baseParams buildmessage.BaseParams, from string, msigAddress, minerId string, controlAddrs []string) (*chain.Message, error) {
	req := MsigSetControlProposeRequest{
		BaseParams:   baseParams,
		From:         from,
		MsigAddress:  msigAddress,
		MinerId:      minerId,
		ControlAddrs: controlAddrs,
	}

	res, err := PostRequest(api.endpoint, "/msig/set_control_propose", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (api *OpenFilAPI) MsigSetControlApprove(baseParams buildmessage.BaseParams, from string, msigAddress, proposerAddress, txId, minerId string, controlAddrs []string) (*chain.Message, error) {
	req := MsigSetControlApproveRequest{
		BaseParams:      baseParams,
		From:            from,
		ProposerAddress: proposerAddress,
		TxId:            txId,
		MsigAddress:     msigAddress,
		MinerId:         minerId,
		ControlAddrs:    controlAddrs,
	}

	res, err := PostRequest(api.endpoint, "/msig/set_control_approve", api.token, req)
	if err != nil {
		return nil, err
	}

	var r chain.Message
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func ifServerErr(res []byte) error {
	var r Response
	err := json.Unmarshal(res, &r)
	if err != nil {
		return nil
	}

	if r.Code == 0 || r.Code == 200 {
		return nil
	}

	if r.Code != 200 {
		return errors.New(r.Message)
	}

	return nil
}

// PostRequest http post request
func PostRequest(endpoint, relativePath string, token string, params interface{}) ([]byte, error) {
	dataByte, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", urlJoin(endpoint, relativePath), bytes.NewBuffer(dataByte))
	if err != nil {
		return nil, err
	}

	return Call(req, token)
}

// GetRequest http get request
func GetRequest(endpoint, relativePath string, token string, params map[string]string) ([]byte, error) {
	u, err := url.Parse(urlJoin(endpoint, relativePath))
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	for key, param := range params {
		values.Set(key, param)
	}

	u.RawQuery = values.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	return Call(req, token)
}

func Call(req *http.Request, token string) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")

	if len(strings.Trim(token, " ")) > 0 {
		req.Header.Set("Authorization", token)
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

	if err := ifServerErr(body); err != nil {
		return nil, err
	}

	return body, nil
}

func urlJoin(endpoint string, relativePath string) string {
	u := strings.TrimRight(endpoint, "/") + "/" + strings.TrimLeft(relativePath, "/")
	return strings.TrimRight(u, "/")
}
