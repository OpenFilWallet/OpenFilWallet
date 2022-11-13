package datastore

import (
	"context"
	"github.com/OpenFilWallet/OpenFilWallet/repo"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewWalletDB(t *testing.T) {
	r, err := repo.NewFS("~/openfilwallet-test")
	require.NoError(t, err)

	require.NoError(t, r.Init())

	lr, err := r.Lock()
	require.NoError(t, err)

	ds, err := lr.Datastore(context.Background())
	require.NoError(t, err)

	db := NewWalletDB(ds)

	require.NoError(t, db.DeleteRootPassword())
	require.NoError(t, db.DeleteLoginPassword())

	rootPassword, err := db.GetRootPassword()
	require.Equal(t, rootPassword, []byte{})
	loginPassword, err := db.GetLoginPassword()
	require.Equal(t, loginPassword, []byte{})

	require.NoError(t, db.SetRootPassword([]byte("root password")))
	require.NoError(t, db.SetLoginPassword([]byte("login password")))

	require.Equal(t, db.SetRootPassword([]byte("root password")).Error(), "scrypt already exist")
	require.Equal(t, db.SetLoginPassword([]byte("login password")).Error(), "scrypt already exist")

}
