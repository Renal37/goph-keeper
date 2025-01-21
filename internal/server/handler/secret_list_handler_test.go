package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Renal37/goph-keeper/internal/server/endpoint"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestListSecretHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretStorage := NewMockSecretStorage(ctrl)
	h := NewSecretListHandler(mockSecretStorage)

	r := httptest.NewRequest(http.MethodGet, endpoint.WalletsEndpoint, nil)
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	expected := []*Secret{
		{Key: "key_1", Value: "val_2"},
		{Key: "key_2", Value: "val_2"},
	}
	mockSecretStorage.EXPECT().ListSecret(gomock.Any(), testToken).Return(expected, nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	answer := w.Body.Bytes()
	resp := &SecretListResponse{}
	err := json.Unmarshal(answer, resp)
	require.NoError(t, err)

	require.Equal(t, expected, resp.Data)
}

func TestListSecretHandlerBadMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSecretStorage := NewMockSecretStorage(ctrl)
	h := NewSecretListHandler(mockSecretStorage)

	r := httptest.NewRequest(http.MethodDelete, endpoint.WalletsEndpoint, nil)
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadRequest, w.Code)
}
