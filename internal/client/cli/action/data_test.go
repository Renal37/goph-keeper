package action

import (
	"context"
	"os"
	"testing"

	"github.com/Renal37/goph-keeper/internal/client/gophcrypto"
	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/Renal37/goph-keeper/internal/client/storage/sqlstorage"
	"github.com/Renal37/goph-keeper/internal/client/transport"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

const testFileName = "./test.txt"

func TestCreateData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataStorage := storage.NewMockDataStorage(ctrl)
	mockClient := transport.NewMockBinaryDataClient(ctrl)

	key, data := getCryptoKeyAndData(t)

	user := &storage.User{Login: "user", Password: "pass", Token: "token", IsActive: true, CryptoKey: key}
	record := &storage.Record{Name: testFileName, Data: data, Revision: 1}

	ctx := context.Background()

	mockDataStorage.EXPECT().CreateData(ctx, user, record).Return(nil)
	mockClient.EXPECT().UploadBinaryData(ctx, user, record).Return(nil)

	err := CreateDataAction(ctx, user, mockDataStorage, mockClient, testFileName)
	require.NoError(t, err)
}

func TestGetExistingData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataStorage := storage.NewMockDataStorage(ctrl)
	mockClient := transport.NewMockBinaryDataClient(ctrl)

	key, data := getCryptoKeyAndData(t)

	user := &storage.User{Login: "user", Password: "pass", Token: "token", IsActive: true, CryptoKey: key}
	record := &storage.Record{Name: testFileName, Data: data, Revision: 1}

	ctx := context.Background()

	mockDataStorage.EXPECT().LoadData(ctx, user, testFileName).Return(record, nil)

	err := GetDataAction(ctx, user, mockDataStorage, mockClient, testFileName)
	require.NoError(t, err)
}

func TestGetDataFromServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataStorage := storage.NewMockDataStorage(ctrl)
	mockClient := transport.NewMockBinaryDataClient(ctrl)

	key, data := getCryptoKeyAndData(t)

	user := &storage.User{Login: "user", Password: "pass", Token: "token", IsActive: true, CryptoKey: key}
	record := &storage.Record{Name: testFileName, Data: data, Revision: 1}

	ctx := context.Background()

	mockDataStorage.EXPECT().LoadData(ctx, user, testFileName).Return(nil, sqlstorage.ErrDataNotExist)
	mockClient.EXPECT().DownloadBinaryData(ctx, user, testFileName).Return(record, nil)
	mockDataStorage.EXPECT().CreateData(ctx, user, record).Return(nil)

	err := GetDataAction(ctx, user, mockDataStorage, mockClient, testFileName)
	require.NoError(t, err)
}

func TestListDataAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataStorage := storage.NewMockDataStorage(ctrl)
	mockClient := transport.NewMockBinaryDataClient(ctrl)

	key, data := getCryptoKeyAndData(t)

	user := &storage.User{Login: "user", Password: "pass", Token: "token", IsActive: true, CryptoKey: key}
	records := []*storage.Record{
		{Name: testFileName, Data: data, Revision: 1},
	}

	ctx := context.Background()

	mockDataStorage.EXPECT().ListData(ctx, user).Return(records, nil)

	err := ListDataAction(ctx, user, mockDataStorage, mockClient)
	require.NoError(t, err)
}

func TestUpdateData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataStorage := storage.NewMockDataStorage(ctrl)
	mockClient := transport.NewMockBinaryDataClient(ctrl)

	key, data := getCryptoKeyAndData(t)

	user := &storage.User{Login: "user", Password: "pass", Token: "token", IsActive: true, CryptoKey: key}
	record := &storage.Record{
		Name: testFileName, Data: data, Revision: 1,
	}

	ctx := context.Background()

	mockDataStorage.EXPECT().UpdateData(ctx, user, record).Return(uint64(2), true, nil)
	sendingRecord := &storage.Record{
		Name: testFileName, Data: data, Revision: 2,
	}
	mockClient.EXPECT().UpdateBinaryData(ctx, user, sendingRecord).Return(nil)

	err := UpdateAction(ctx, user, mockDataStorage, mockClient, testFileName)
	require.NoError(t, err)
}

func TestUpdateDataNotNeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataStorage := storage.NewMockDataStorage(ctrl)
	mockClient := transport.NewMockBinaryDataClient(ctrl)

	key, data := getCryptoKeyAndData(t)

	user := &storage.User{Login: "user", Password: "pass", Token: "token", IsActive: true, CryptoKey: key}
	record := &storage.Record{
		Name: testFileName, Data: data, Revision: 1,
	}

	ctx := context.Background()

	mockDataStorage.EXPECT().UpdateData(ctx, user, record).Return(uint64(1), false, nil)

	err := UpdateAction(ctx, user, mockDataStorage, mockClient, testFileName)
	require.NoError(t, err)
}

func TestDeleteData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataStorage := storage.NewMockDataStorage(ctrl)
	mockClient := transport.NewMockBinaryDataClient(ctrl)

	key, _ := getCryptoKeyAndData(t)

	user := &storage.User{Login: "user", Password: "pass", Token: "token", IsActive: true, CryptoKey: key}
	ctx := context.Background()

	mockDataStorage.EXPECT().DeleteData(ctx, user, testFileName).Return(nil)
	mockClient.EXPECT().DeleteBinaryData(ctx, user, testFileName).Return(nil)

	err := DeleteBinaryDataAction(ctx, user, mockDataStorage, mockClient, testFileName)
	require.NoError(t, err)
}

func getCryptoKeyAndData(t *testing.T) ([]byte, string) {
	key, err := gophcrypto.GenerateCryptoKey()
	require.NoError(t, err)

	data, err := os.ReadFile(testFileName)
	require.NoError(t, err)

	c, err := gophcrypto.New(key)
	require.NoError(t, err)

	return key, c.Encrypt(data)
}
