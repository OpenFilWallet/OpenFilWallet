package repo

import (
	"context"
	"github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRepo(t *testing.T) {
	r, err := NewFS("~/openfilwallet-test")
	require.NoError(t, err)

	require.NoError(t, r.Init())

	lr, err := r.Lock()
	require.NoError(t, err)

	ds, err := lr.Datastore(context.Background())
	require.NoError(t, err)

	require.NoError(t, ds.Put(context.Background(), datastore.NewKey("test"), []byte("test")))

	value, err := ds.Get(context.Background(), datastore.NewKey("test"))
	require.NoError(t, err)

	require.Equal(t, value, []byte("test"))
}
