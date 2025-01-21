package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Renal37/goph-keeper/internal/server/endpoint"
	"github.com/Renal37/goph-keeper/internal/server/handler"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAuthRegisterNewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChecker := NewMockUserChecker(ctrl)
	mockHandler := NewMockHTTPHandler(ctrl)

	authMiddleware := NewAuthMiddleware(mockChecker)

	r := httptest.NewRequest(http.MethodGet, endpoint.RegisterEndpoint, nil)
	r.Header.Add("login", "user")
	r.Header.Add("password", "1234")

	w := httptest.NewRecorder()

	wrappedHandler := authMiddleware.Middleware(mockHandler)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(func(w http.ResponseWriter, r *http.Request) {
		o := r.Context().Value(handler.AuthInfo("user"))
		user, ok := o.(*handler.User)
		require.True(t, ok)
		require.Equal(t, &handler.User{Login: "user", Password: "1234"}, user)
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestAuthRegisterWithouLoginHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChecker := NewMockUserChecker(ctrl)
	mockHandler := NewMockHTTPHandler(ctrl)

	authMiddleware := NewAuthMiddleware(mockChecker)

	r := httptest.NewRequest(http.MethodGet, endpoint.RegisterEndpoint, nil)
	r.Header.Add("password", "1234")

	w := httptest.NewRecorder()

	wrappedHandler := authMiddleware.Middleware(mockHandler)

	wrappedHandler.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthRegisterWithouPasswordHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChecker := NewMockUserChecker(ctrl)
	mockHandler := NewMockHTTPHandler(ctrl)

	authMiddleware := NewAuthMiddleware(mockChecker)

	r := httptest.NewRequest(http.MethodGet, endpoint.RegisterEndpoint, nil)
	r.Header.Add("login", "user")

	w := httptest.NewRecorder()

	wrappedHandler := authMiddleware.Middleware(mockHandler)

	wrappedHandler.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandledDataEndpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChecker := NewMockUserChecker(ctrl)
	mockHandler := NewMockHTTPHandler(ctrl)

	authMiddleware := NewAuthMiddleware(mockChecker)

	r := httptest.NewRequest(http.MethodGet, endpoint.BinaryDataEndpoint, nil)

	expectedToken := "ppopopopo"
	r.Header.Add("token", expectedToken)

	w := httptest.NewRecorder()

	wrappedHandler := authMiddleware.Middleware(mockHandler)

	mockChecker.EXPECT().Check(gomock.Any(), expectedToken).Return(nil)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(func(w http.ResponseWriter, r *http.Request) {
		o := r.Context().Value(handler.AuthInfo("token"))
		token, ok := o.(string)
		require.True(t, ok)
		require.Equal(t, expectedToken, token)
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandledDataEndpointWithoutToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChecker := NewMockUserChecker(ctrl)
	mockHandler := NewMockHTTPHandler(ctrl)

	authMiddleware := NewAuthMiddleware(mockChecker)

	r := httptest.NewRequest(http.MethodGet, endpoint.BinaryDataEndpoint, nil)

	w := httptest.NewRecorder()

	wrappedHandler := authMiddleware.Middleware(mockHandler)

	wrappedHandler.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandledDataEndpointUnknownUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChecker := NewMockUserChecker(ctrl)
	mockHandler := NewMockHTTPHandler(ctrl)

	authMiddleware := NewAuthMiddleware(mockChecker)

	r := httptest.NewRequest(http.MethodGet, endpoint.BinaryDataEndpoint, nil)

	expectedToken := "ppopopopo"
	r.Header.Add("token", expectedToken)

	w := httptest.NewRecorder()

	wrappedHandler := authMiddleware.Middleware(mockHandler)

	mockChecker.EXPECT().Check(gomock.Any(), expectedToken).Return(handler.ErrUnknownUser)

	wrappedHandler.ServeHTTP(w, r)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
