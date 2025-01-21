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

func TestSecretHandlerCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretStorage := NewMockSecretStorage(ctrl)
	h := NewSecretDataHandler(mockSecretStorage)

	req := SaveSecretRequest{Key: "key", Value: "value"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPut, endpoint.WalletEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	secret := &Secret{Key: "key", Value: "value"}
	mockSecretStorage.EXPECT().CreateSecret(gomock.Any(), testToken, secret).Return(nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestSecretHandlerCreateBadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretStorage := NewMockSecretStorage(ctrl)
	h := NewSecretDataHandler(mockSecretStorage)

	// request without body
	r := httptest.NewRequest(http.MethodPut, endpoint.WalletEndpoint, nil)
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecretHandlerGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretStorage := NewMockSecretStorage(ctrl)
	h := NewSecretDataHandler(mockSecretStorage)

	req := SaveSecretRequest{Key: "key", Value: "value"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodGet, endpoint.WalletEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	secret := &Secret{Key: "key", Value: "value"}
	mockSecretStorage.EXPECT().GetSecret(gomock.Any(), testToken, "key").Return(secret, nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	answer := w.Body.Bytes()
	resp := &GetSecretResponse{}
	err = json.Unmarshal(answer, resp)
	require.NoError(t, err)

	require.Equal(t, "key", resp.Key)
	require.Equal(t, "value", resp.Data)
}

func TestSecretHandlerGetBadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretStorage := NewMockSecretStorage(ctrl)
	h := NewSecretDataHandler(mockSecretStorage)

	// request without body
	r := httptest.NewRequest(http.MethodGet, endpoint.WalletEndpoint, nil)
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecretHandlerDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretStorage := NewMockSecretStorage(ctrl)
	h := NewSecretDataHandler(mockSecretStorage)

	req := DeleteSecretRequest{Key: "key"}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodDelete, endpoint.WalletEndpoint, bytes.NewBuffer(data))
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	mockSecretStorage.EXPECT().DeleteSecret(gomock.Any(), testToken, "key").Return(nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestSecretHandlerDeleteBadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretStorage := NewMockSecretStorage(ctrl)
	h := NewSecretDataHandler(mockSecretStorage)

	// request without body
	r := httptest.NewRequest(http.MethodDelete, endpoint.WalletEndpoint, nil)
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadRequest, w.Code)
}
