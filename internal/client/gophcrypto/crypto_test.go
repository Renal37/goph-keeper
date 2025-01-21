package gophcrypto

import (
	"crypto/aes"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrypto(t *testing.T) {

	data := []byte("12345")

	key, err := generateRandom(2 * aes.BlockSize)

	fmt.Println(base64.RawStdEncoding.EncodeToString(key))

	require.NoError(t, err)

	c, err := New(key)
	require.NoError(t, err)

	encrypted := c.Encrypt(data)
	decrypted, err := c.Decrypt([]byte(encrypted))
	require.NoError(t, err)

	fmt.Println("data", string(data))
	fmt.Println("decrypted", string(decrypted))

	require.Equal(t, data, decrypted)
}

func TestGenerateCryptoKey(t *testing.T) {
	key, err := GenerateCryptoKey()
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Len(t, key, aes.BlockSize)
}
