package action

import (
	"context"
	"testing"

	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/Renal37/goph-keeper/internal/client/transport"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockUserStorage(ctrl)
	mockClient := transport.NewMockRegisterClient(ctrl)

	ctx := context.Background()

	mockClient.EXPECT().RegisterUser(ctx, "login", gomock.Any()).Return("token", nil)
	mockStorage.EXPECT().Register(ctx, "login", gomock.Any(), "token", gomock.Any()).Return(nil)

	err := RegisterAction(ctx, mockStorage, mockClient, "login", "pass")
	require.NoError(t, err)
}
