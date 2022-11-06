package crypto

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestScrypt(t *testing.T) {
	key := Scrypt("hello world")
	t.Log(hex.EncodeToString(key))

	ok, err := VerifyScrypt("hello world", key)
	if err != nil {
		t.Fatal(err)
	}

	require.True(t, ok)
}
