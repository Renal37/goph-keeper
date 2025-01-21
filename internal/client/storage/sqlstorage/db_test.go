package sqlstorage

import (
	"testing"
	"time"

	"github.com/Renal37/goph-keeper/internal/client/gophcrypto"
	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/stretchr/testify/require"
)

func TestSerializeCard(t *testing.T) {
	cryptoKey, err := gophcrypto.GenerateCryptoKey()
	require.NoError(t, err)

	c := &storage.BankCard{Number: "1234123412341234", ExpiryDate: time.Now().Truncate(time.Hour * 24), Owner: "IVAN PETROV", CvvCode: "123"}
	u := &storage.User{Login: "l", Password: "p", Token: "t", CryptoKey: cryptoKey}

	data, err := serializeUserBankCard(u, c)
	require.NoError(t, err)

	c2, err := desirializeUserBankCard(u, data)
	require.NoError(t, err)

	require.Equal(t, c, c2)
}

func TestSerializeSecret(t *testing.T) {
	cryptoKey, err := gophcrypto.GenerateCryptoKey()
	require.NoError(t, err)

	c := &storage.Secret{Name: "n", Key: "k", Value: "v"}
	u := &storage.User{Login: "l", Password: "p", Token: "t", CryptoKey: cryptoKey}

	data, err := serializeSecret(u, c)
	require.NoError(t, err)

	c2, err := deserializeSecret(u, data)
	require.NoError(t, err)

	require.Equal(t, c, c2)
}
