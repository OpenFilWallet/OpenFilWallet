package datastore

type HdWallet struct {
	Mnemonic     []byte `json:"mnemonic"`
	MnemonicHash []byte `json:"mnemonic_hash"`
}

type PrivateWallet struct {
	PriKey  []byte `json:"pri_key"`
	Address string `json:"address"`
	KeyHash []byte `json:"key_hash"`
	Path    string `json:"path"`
}

type MsigWallet struct {
	MsigAddr              string   `json:"msig_addr"`
	Signers               []string `json:"signers"`
	NumApprovalsThreshold uint64   `json:"num_approvals_threshold"`
	UnlockDuration        int64    `json:"unlock_duration"`
	StartEpoch            int64    `json:"start_epoch"`
}

type NodeInfo struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

type History struct {
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
