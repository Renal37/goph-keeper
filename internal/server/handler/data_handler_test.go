package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Renal37/goph-keeper/internal/server/endpoint"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

const testToken = "user-token"

func TestDataHandlerGetData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockDataStorage(ctrl)
	h := NewDataHandler(mockStorage)

	req := GetDataRequest{Key: "key"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodGet, endpoint.BinaryDataEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	mockStorage.EXPECT().LoadData(gomock.Any(), testToken, req.Key).Return(&Record{Name: req.Key, Data: "user data", Revision: 1}, nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	answer := w.Body.Bytes()
	resp := &GetDataResponse{}
	err = json.Unmarshal(answer, resp)
	require.NoError(t, err)

	require.Equal(t, "key", resp.Key)
	require.Equal(t, "user data", resp.Data)
	require.Equal(t, uint64(1), resp.Revision)
}

func TestDataHandlerCreateData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockDataStorage(ctrl)
	h := NewDataHandler(mockStorage)

	req := &SaveDataRequest{Key: "key", Data: "user_data"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, endpoint.BinaryDataEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	rec := &Record{Name: req.Key, Data: req.Data}
	mockStorage.EXPECT().CreateData(gomock.Any(), testToken, rec).Return(nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestDataHandlerCreatedDataAlreadyExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockDataStorage(ctrl)
	h := NewDataHandler(mockStorage)

	req := &SaveDataRequest{Key: "key", Data: "user_data"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, endpoint.BinaryDataEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	rec := &Record{Name: req.Key, Data: req.Data}
	mockStorage.EXPECT().CreateData(gomock.Any(), testToken, rec).Return(ErrDataAlreadyExist)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestDataHandlerUpdateData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockDataStorage(ctrl)
	h := NewDataHandler(mockStorage)

	req := &UpdateDataRequest{Key: "key", Data: "user_data", Revision: 1}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPut, endpoint.BinaryDataEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	rec := &Record{Name: req.Key, Data: req.Data, Revision: 1}
	mockStorage.EXPECT().UpdateData(gomock.Any(), testToken, rec).Return(nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestDataHandlerDeleteData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockDataStorage(ctrl)
	h := NewDataHandler(mockStorage)

	req := &DeleteDataRequest{Key: "key"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodDelete, endpoint.BinaryDataEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	rec := &Record{Name: req.Key}
	mockStorage.EXPECT().DeleteData(gomock.Any(), testToken, rec).Return(nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestDataHandlerBadMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockDataStorage(ctrl)
	h := NewDataHandler(mockStorage)

	r := httptest.NewRequest(http.MethodOptions, endpoint.BinaryDataEndpoint, nil)
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestDataHandlerBadRevision(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockDataStorage(ctrl)
	h := NewDataHandler(mockStorage)

	req := &UpdateDataRequest{Key: "key", Data: "user_data", Revision: 1}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPut, endpoint.BinaryDataEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	rec := &Record{Name: req.Key, Data: req.Data, Revision: 1}
	mockStorage.EXPECT().UpdateData(gomock.Any(), testToken, rec).Return(ErrBadRevision)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestDataHandlerDataIsNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockDataStorage(ctrl)
	h := NewDataHandler(mockStorage)

	req := &UpdateDataRequest{Key: "key", Data: "user_data", Revision: 1}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPut, endpoint.BinaryDataEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	rec := &Record{Name: req.Key, Data: req.Data, Revision: 1}
	mockStorage.EXPECT().UpdateData(gomock.Any(), testToken, rec).Return(ErrDataNotFound)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusNotFound, w.Code)
}
