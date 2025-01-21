package action

import (
	"context"
	"testing"

	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/Renal37/goph-keeper/internal/client/transport"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCreateCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockWalletStorage(ctrl)
	mockClient := transport.NewMockWalletClient(ctrl)

	ctx := context.Background()
	u := &storage.User{Token: "token"}
	c := &storage.BankCard{Number: "xxxx"}

	mustBeCryptedData := "crypted_data"

	mockStorage.EXPECT().CreateCard(ctx, u, c).Return(mustBeCryptedData, nil)
	mockClient.EXPECT().CreateCardData(ctx, u.Token, c.Number, mustBeCryptedData).Return(nil)

	err := CreateCardActionHandler(ctx, u, mockStorage, mockClient, c)
	require.NoError(t, err)
}

func TestDeleteCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockWalletStorage(ctrl)
	mockClient := transport.NewMockWalletClient(ctrl)

	ctx := context.Background()
	u := &storage.User{Token: "token"}
	c := &storage.BankCard{Number: "xxxx"}

	mockStorage.EXPECT().DeleteCard(ctx, u, c.Number).Return(nil)
	mockClient.EXPECT().DeleteCardData(ctx, u.Token, c.Number).Return(nil)

	err := DeleteCardActionHandler(ctx, u, mockStorage, mockClient, c.Number)
	require.NoError(t, err)
}

func TestListCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockWalletStorage(ctrl)
	mockClient := transport.NewMockWalletClient(ctrl)

	ctx := context.Background()
	u := &storage.User{Token: "token"}
	c := &storage.BankCard{Number: "xxxx"}

	mockStorage.EXPECT().ListCard(ctx, u).Return([]*storage.BankCard{c}, nil)

	err := ListCardActionHandler(ctx, u, mockStorage, mockClient)
	require.NoError(t, err)
}
