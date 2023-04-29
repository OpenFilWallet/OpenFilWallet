package wallet

import (
	"context"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"testing"
)

func TestBeneficiary(t *testing.T) {
	node := datastore.NodeInfo{
		Name:     "glif",
		Endpoint: "https://api.node.glif.io/rpc/v0",
		Token:    "",
	}
	n, err := newNode(node.Name, node.Endpoint, "")
	if err != nil {
		t.Fatal(err)
	}

	minerAddr, _ := address.NewFromString("f02104084")
	mi, err := n.LotusClient.Api.StateMinerInfo(context.Background(), minerAddr, types.EmptyTSK)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(mi.Owner.String())
	t.Log(mi.Beneficiary.String())
	t.Log(mi.BeneficiaryTerm.Quota)
	t.Log(mi.BeneficiaryTerm.Expiration)
	t.Log(mi.BeneficiaryTerm.UsedQuota)
	t.Log(mi.PendingBeneficiaryTerm)

}
