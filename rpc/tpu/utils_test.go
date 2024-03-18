package tpu

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCert(t *testing.T) {
	cert, err := CreateTlsCertificate()
	require.NoError(t, err)
	println(hex.EncodeToString(cert.Certificate[0]))
	println(hex.EncodeToString(cert.PrivateKey.(ed25519.PrivateKey)))
}
