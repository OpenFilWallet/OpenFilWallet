package crypto

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateEncryptKey(t *testing.T) {
	key1 := GenerateEncryptKey([]byte("hello world"))
	key2 := GenerateEncryptKey([]byte("hello world"))
	require.Equal(t, key1, key2)
}

func TestEncryptAndDecrypt(t *testing.T) {
	key := Hash256([]byte("hello world"))

	t.Log(hex.EncodeToString(key))

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
