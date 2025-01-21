package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Renal37/goph-keeper/internal/server/endpoint"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegistrator := NewMockRegistrator(ctrl)
	h := NewRegistrationHandler(mockRegistrator)

	r := httptest.NewRequest(http.MethodPut, endpoint.RegisterEndpoint, nil)
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("user"), &User{Login: "user", Password: "1234"}))
	w := httptest.NewRecorder()

	mockRegistrator.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	resp := &RegistrationResponse{}
	err := json.Unmarshal(w.Body.Bytes(), resp)
	require.NoError(t, err)
}

func TestRegisterStorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegistrator := NewMockRegistrator(ctrl)
	h := NewRegistrationHandler(mockRegistrator)

	r := httptest.NewRequest(http.MethodPut, endpoint.RegisterEndpoint, nil)
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("user"), &User{Login: "user", Password: "1234"}))
	w := httptest.NewRecorder()

	mockRegistrator.EXPECT().Register(gomock.Any(), gomock.Any()).Return(errors.New("internal error"))

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
