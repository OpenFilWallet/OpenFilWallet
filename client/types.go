package client

import (
	"github.com/OpenFilWallet/OpenFilWallet/modules/buildmessage"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type LoginRequest struct {
	LoginPassword string `json:"login_password"`
}

type NodeRequest struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

type NodeInfo struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

type CreateWalletRequest struct {
	Index int `json:"index"`
}

type CreateWalletResponse struct {
	NewWalletAddrs []string `json:"new_wallet_addrs"`
}

type TransferRequest struct {
	BaseParams buildmessage.BaseParams `json:"base_params"`
	From       string                  `json:"from"`
	To         string                  `json:"to"`
	Amount     string                  `json:"amount"`
}

type HistoryResponse struct {
	Version    uint64 `json:"version"`
	To         string `json:"to"`
	From       string `json:"from"`
	Nonce      uint64 `json:"nonce"`
	Value      int64  `json:"value"`
	GasLimit   int64  `json:"gas_limit"`
	GasFeeCap  int64  `json:"gas_feecap"`
	GasPremium int64  `json:"gas_premium"`
	Method     uint64 `json:"method"`
	Params     string `json:"params"`
}

type WithdrawRequest struct {
	BaseParams buildmessage.BaseParams `json:"base_params"`
	MinerId    string                  `json:"miner_id"`
	Amount     string                  `json:"amount"`
}

type ChangeOwnerRequest struct {
	BaseParams buildmessage.BaseParams `json:"base_params"`
	MinerId    string                  `json:"miner_id"`
	NewOwner   string                  `json:"new_owner"`
	From       string                  `json:"from"`
}

type ChangeWorkerRequest struct {
	BaseParams      buildmessage.BaseParams `json:"base_params"`
	MinerId         string                  `json:"miner_id"`
	NewWorker       string                  `json:"new_worker"`
	NewControlAddrs []string                `json:"new_controlAddrs"`
}

type ConfirmChangeWorkerRequest struct {
	BaseParams buildmessage.BaseParams `json:"base_params"`
	MinerId    string                  `json:"miner_id"`
	NewWorker  string                  `json:"new_worker"`
}

type SingRequest struct {
	From       string `json:"from"`
	HexMessage string `json:"hex_message"`
}

type BalanceInfo struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

type MinerControl struct {
	Owner            string   `json:"owner"`
	Worker           string   `json:"worker"`
	NewWorker        string   `json:"new_worker"`
	ControlAddresses []string `json:"control_addresses"`
}

type StatusInfo struct {
	Lock    bool   `json:"lock"`
	Offline bool   `json:"offline"`
	Version string `json:"version"`
}

type LoginInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

type DecodeRequest struct {
	ToAddr   string `json:"to_addr"`
	Method   uint64 `json:"method"`
	Params   string `json:"params"`
	Encoding string `json:"encoding"`
}

type EncodeRequest struct {
	Dest     string `json:"dest"`
	Method   uint64 `json:"method"`
	Params   string `json:"params"`
	Encoding string `json:"encoding"`
}

type DecodeResponse struct {
	DecodeMsg string `json:"decode_msg"`
}

type EncodeResponse struct {
	EncodeMsg string `json:"encode_msg"`
}

type MsigCreateRequest struct {
	BaseParams buildmessage.BaseParams `json:"base_params"`
	From       string                  `json:"from"`
	Required   uint64                  `json:"required"`
	Duration   uint64                  `json:"duration"`
	Value      string                  `json:"value"`
	Signers    []string                `json:"signers"`
}

type MsigBaseRequest struct {
	BaseParams  buildmessage.BaseParams `json:"base_params"`
	From        string                  `json:"from"`
	MsigAddress string                  `json:"msig_address"`
	TxId        string                  `json:"tx_id"`
}

type MsigTransferProposeRequest struct {
	BaseParams         buildmessage.BaseParams `json:"base_params"`
	From               string                  `json:"from"`
	MsigAddress        string                  `json:"msig_address"`
	DestinationAddress string                  `json:"destination_address"`
	Amount             string                  `json:"amount"`
}

type MsigAddSignerProposeRequest struct {
	BaseParams        buildmessage.BaseParams `json:"base_params"`
	From              string                  `json:"from"`
	MsigAddress       string                  `json:"msig_address"`
	SignerAddress     string                  `json:"signer_address"`
	IncreaseThreshold bool                    `json:"increase_threshold"`
}

type MsigAddSignerApprovRequest struct {
	BaseParams        buildmessage.BaseParams `json:"base_params"`
	From              string                  `json:"from"`
	MsigAddress       string                  `json:"msig_address"`
	ProposerAddress   string                  `json:"proposer_address"`
	TxId              string                  `json:"tx_id"`
	SignerAddress     string                  `json:"signer_address"`
	IncreaseThreshold bool                    `json:"increase_threshold"`
}

type MsigAddSignerCancelRequest struct {
	BaseParams        buildmessage.BaseParams `json:"base_params"`
	From              string                  `json:"from"`
	MsigAddress       string                  `json:"msig_address"`
	TxId              string                  `json:"tx_id"`
	SignerAddress     string                  `json:"signer_address"`
	IncreaseThreshold bool                    `json:"increase_threshold"`
}

type MsigSwapProposeRequest struct {
	BaseParams  buildmessage.BaseParams `json:"base_params"`
	From        string                  `json:"from"`
	MsigAddress string                  `json:"msig_address"`
	OldAddress  string                  `json:"old_address"`
	NewAddress  string                  `json:"new_address"`
}

type MsigSwapApproveRequest struct {
	BaseParams      buildmessage.BaseParams `json:"base_params"`
	From            string                  `json:"from"`
	MsigAddress     string                  `json:"msig_address"`
	ProposerAddress string                  `json:"proposer_address"`
	TxId            string                  `json:"tx_id"`
	OldAddress      string                  `json:"old_address"`
	NewAddress      string                  `json:"new_address"`
}

type MsigSwapCancelRequest struct {
	BaseParams  buildmessage.BaseParams `json:"base_params"`
	From        string                  `json:"from"`
	MsigAddress string                  `json:"msig_address"`
	TxId        string                  `json:"tx_id"`
	OldAddress  string                  `json:"old_address"`
	NewAddress  string                  `json:"new_address"`
}

type MsigLockProposeRequest struct {
	BaseParams     buildmessage.BaseParams `json:"base_params"`
	From           string                  `json:"from"`
	MsigAddress    string                  `json:"msig_address"`
	StartEpoch     string                  `json:"start_epoch"`
	UnlockDuration string                  `json:"unlock_duration"`
	Amount         string                  `json:"amount"`
}

type MsigLockApproveRequest struct {
	BaseParams      buildmessage.BaseParams `json:"base_params"`
	From            string                  `json:"from"`
	MsigAddress     string                  `json:"msig_address"`
	ProposerAddress string                  `json:"proposer_address"`
	TxId            string                  `json:"tx_id"`
	StartEpoch      string                  `json:"start_epoch"`
	UnlockDuration  string                  `json:"unlock_duration"`
	Amount          string                  `json:"amount"`
}

type MsigLockCancelRequest struct {
	BaseParams     buildmessage.BaseParams `json:"base_params"`
	From           string                  `json:"from"`
	MsigAddress    string                  `json:"msig_address"`
	TxId           string                  `json:"tx_id"`
	StartEpoch     string                  `json:"start_epoch"`
	UnlockDuration string                  `json:"unlock_duration"`
	Amount         string                  `json:"amount"`
}

type MsigThresholdProposeRequest struct {
	BaseParams   buildmessage.BaseParams `json:"base_params"`
	From         string                  `json:"from"`
	MsigAddress  string                  `json:"msig_address"`
	NewThreshold string                  `json:"new_threshold"`
}

type MsigThresholdApproveRequest struct {
	BaseParams      buildmessage.BaseParams `json:"base_params"`
	From            string                  `json:"from"`
	MsigAddress     string                  `json:"msig_address"`
	ProposerAddress string                  `json:"proposer_address"`
	TxId            string                  `json:"tx_id"`
	NewThreshold    string                  `json:"new_threshold"`
}

type MsigThresholdCancelRequest struct {
	BaseParams   buildmessage.BaseParams `json:"base_params"`
	From         string                  `json:"from"`
	MsigAddress  string                  `json:"msig_address"`
	TxId         string                  `json:"tx_id"`
	NewThreshold string                  `json:"new_threshold"`
}

type MsigChangeOwnerProposeRequest struct {
	BaseParams  buildmessage.BaseParams `json:"base_params"`
	From        string                  `json:"from"`
	MsigAddress string                  `json:"msig_address"`
	MinerId     string                  `json:"miner_id"`
	NewOwner    string                  `json:"new_owner"`
}

type MsigChangeOwnerApproveRequest struct {
	BaseParams      buildmessage.BaseParams `json:"base_params"`
	From            string                  `json:"from"`
	MsigAddress     string                  `json:"msig_address"`
	ProposerAddress string                  `json:"proposer_address"`
	TxId            string                  `json:"tx_id"`
	MinerId         string                  `json:"miner_id"`
	NewOwner        string                  `json:"new_owner"`
}

type MsigWithdrawProposeRequest struct {
	BaseParams  buildmessage.BaseParams `json:"base_params"`
	From        string                  `json:"from"`
	MsigAddress string                  `json:"msig_address"`
	MinerId     string                  `json:"miner_id"`
	Amount      string                  `json:"amount"`
}

type MsigWithdrawApproveRequest struct {
	BaseParams      buildmessage.BaseParams `json:"base_params"`
	From            string                  `json:"from"`
	MsigAddress     string                  `json:"msig_address"`
	ProposerAddress string                  `json:"proposer_address"`
	TxId            string                  `json:"tx_id"`
	MinerId         string                  `json:"miner_id"`
	Amount          string                  `json:"amount"`
}

type MsigChangeWorkerProposeRequest struct {
	BaseParams  buildmessage.BaseParams `json:"base_params"`
	From        string                  `json:"from"`
	MsigAddress string                  `json:"msig_address"`
	MinerId     string                  `json:"miner_id"`
	NewWorker   string                  `json:"new_worker"`
}

type MsigChangeWorkerApproveRequest struct {
	BaseParams      buildmessage.BaseParams `json:"base_params"`
	From            string                  `json:"from"`
	MsigAddress     string                  `json:"msig_address"`
	ProposerAddress string                  `json:"proposer_address"`
	TxId            string                  `json:"tx_id"`
	MinerId         string                  `json:"miner_id"`
	NewWorker       string                  `json:"new_worker"`
}

type MsigSetControlProposeRequest struct {
	BaseParams   buildmessage.BaseParams `json:"base_params"`
	From         string                  `json:"from"`
	MsigAddress  string                  `json:"msig_address"`
	MinerId      string                  `json:"miner_id"`
	ControlAddrs []string                `json:"control_addrs"`
}

type MsigSetControlApproveRequest struct {
	BaseParams      buildmessage.BaseParams `json:"base_params"`
	From            string                  `json:"from"`
	MsigAddress     string                  `json:"msig_address"`
	ProposerAddress string                  `json:"proposer_address"`
	TxId            string                  `json:"tx_id"`
	MinerId         string                  `json:"miner_id"`
	ControlAddrs    []string                `json:"control_addrs"`
}

type MsigInspect struct {
	MsigAddr     string            `json:"msig_addr"`
	Threshold    uint64            `json:"threshold"`
	Signers      []string          `json:"signers"`
	Balance      string            `json:"balance"`
	Spendable    string            `json:"spendable"`
	Lock         MsigLockInfo      `json:"lock"`
	Transactions []MsigTransaction `json:"transactions"`
}

type MsigLockInfo struct {
	InitialBalance string `json:"initial_balance"`
	LockAmount     string `json:"lock_amount"`
	StartEpoch     uint64 `json:"start_epoch"`
	UnlockDuration uint64 `json:"unlock_duration"`
}

type MsigTransaction struct {
	To       string   `json:"to"`
	Value    string   `json:"value"`
	Method   string   `json:"method"`
	Params   string   `json:"params"`
	Approved []string `json:"approved"`
}
