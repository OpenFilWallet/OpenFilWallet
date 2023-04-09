package wallet

import (
	"context"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/OpenFilWallet/OpenFilWallet/repo"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWallet_TxHistory(t *testing.T) {
	r, err := repo.NewFS("~/openfilwallet-test2")
	require.NoError(t, err)

	require.NoError(t, r.Init())

	lr, err := r.Lock()
	require.NoError(t, err)

	ds, err := lr.Datastore(context.Background())
	require.NoError(t, err)

	db := datastore.NewWalletDB(ds)

	n, err := newNode("glif", "https://api.node.glif.io/rpc/v0", "")
	require.NoError(t, err)

	txTracker := newTxTracker(n, db, nil)

	txTracker.trackTx(&datastore.History{
		Version:    0,
		To:         "f01",
		From:       "f1e3fkjzjm7wio6bzec5eqesp6khn25smsrvrv2ea",
		Nonce:      10,
		Value:      0,
		GasLimit:   26682752,
		GasFeeCap:  102100,
		GasPremium: 100161,
		Method:     2,
		Params:     "{\"Signers\":[\"f1e3fkjzjm7wio6bzec5eqesp6khn25smsrvrv2ea\",\"f3v4kunmpw5wxpc62lhwf57puurye5artjsqmdufmeo3r43tmqkpjkqmwmpfexcjdutowp5a6auhl7u3gzb27a\",\"f3qsjierxyqj2ej4uj2ioe7awin63undwb3uyyic6dztvcfumfmjiufnjkjd7q2ohj6hgtcnvqikytzve75zpq\"],\"NumApprovalsThreshold\":2,\"UnlockDuration\":0,\"StartEpoch\":0}",
		ParamName:  "ConstructorParams",
		TxCid:      "bafy2bzacedqkl2ljnksemlxxfu25d57oduzi4byclwk5hmdrstuw64gdulxiy",
		TxState:    datastore.Pending,
		Detail:     "",
	})

	for {
		msigs, err := db.MsigWalletList()
		require.NoError(t, err)
		if len(msigs) == 0 {
			time.Sleep(10 * time.Second)
			continue
		}

		for _, msig := range msigs {
			t.Log("msig", msig.MsigAddr)
		}
		return
	}
}
