package crypto

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEncryptAndDecrypt(t *testing.T) {
	key := Hash256("hello world")

	data := "OpenFilWallet Encrypt"
	encryptedData, err := Encrypt([]byte(data), key)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(encryptedData))

	decryptData, err := Decrypt(encryptedData, key)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, data, string(decryptData))

}
