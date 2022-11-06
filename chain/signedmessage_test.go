package chain

import (
	"bytes"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSignedMessage(t *testing.T) {
	a := crypto.Signature{
		Type: crypto.SigTypeSecp256k1,
		Data: []byte{1, 2, 3, 4},
	}

	var buf bytes.Buffer
	require.NoError(t, a.MarshalCBOR(&buf))

	buf2 := bytes.NewBufferString(buf.String())

	t.Log(buf.String())

	var out crypto.Signature
	require.NoError(t, out.UnmarshalCBOR(buf2))

	require.True(t, out.Equals(&a))
}
