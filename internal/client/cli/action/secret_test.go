package action

import (
	"context"
	"testing"

	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/Renal37/goph-keeper/internal/client/storage/sqlstorage"
	"github.com/Renal37/goph-keeper/internal/client/transport"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCreateSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockSecretStorage(ctrl)
	mockClient := transport.NewMockSecretDataClient(ctrl)

	key, data := getCryptoKeyAndData(t)

	ctx := context.Background()
	user := &storage.User{Login: "user", Password: "pass", Token: "token", IsActive: true, CryptoKey: key}
	secret := &storage.Secret{Name: "secret", Key: "key", Value: data}

	mustBeCryptedSecret := "crypted_data"

	mockStorage.EXPECT().CreateSecret(gomock.Any(), user, secret).Return(mustBeCryptedSecret, nil)
	mockClient.EXPECT().CreateSecret(ctx, user.Token, secret.Name, mustBeCryptedSecret).Return(nil)

	err := CreateSecretAction(ctx, user, mockStorage, mockClient, secret)
	require.NoError(t, err)
}

func TestGetExistingSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockSecretStorage(ctrl)
	mockClient := transport.NewMockSecretDataClient(ctrl)

	key, data := getCryptoKeyAndData(t)

	user := &storage.User{Login: "user", Password: "pass", Token: "token", IsActive: true, CryptoKey: key}
	secret := &storage.Secret{Name: "secret", Key: "key", Value: data}

	ctx := context.Background()

	mockStorage.EXPECT().GetSecret(ctx, user, "key").Return(secret, nil)

	err := GetSecretAction(ctx, user, mockStorage, mockClient, "key")
	require.NoError(t, err)
}

func TestGetDataFromSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockSecretStorage(ctrl)
	mockClient := transport.NewMockSecretDataClient(ctrl)

	key, data := getCryptoKeyAndData(t)

	user := &storage.User{Login: "user", Password: "pass", Token: "token", IsActive: true, CryptoKey: key}
	secret := &storage.Secret{Name: "secret", Key: "key", Value: data}

	ctx := context.Background()

	mockStorage.EXPECT().GetSecret(ctx, user, "key").Return(nil, sqlstorage.ErrDataNotExist)
	mockClient.EXPECT().GetSecret(ctx, user.Token, secret.Key).Return(secret, nil)
	mockStorage.EXPECT().CreateSecret(ctx, user, secret).Return("", nil)

	err := GetSecretAction(ctx, user, mockStorage, mockClient, "key")
	require.NoError(t, err)
}

func TestDeleteSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockSecretStorage(ctrl)
	mockClient := transport.NewMockSecretDataClient(ctrl)

	key, data := getCryptoKeyAndData(t)

	user := &storage.User{Login: "user", Password: "pass", Token: "token", IsActive: true, CryptoKey: key}
	secret := &storage.Secret{Name: "secret", Key: "key", Value: data}

	ctx := context.Background()

	mockStorage.EXPECT().DeleteSecret(ctx, user, "key").Return(nil)
	mockClient.EXPECT().DeleteSecret(ctx, user.Token, secret.Key).Return(nil)

	err := DeleteSecretAction(ctx, user, mockStorage, mockClient, "key")
	require.NoError(t, err)
}
