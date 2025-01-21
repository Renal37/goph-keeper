package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Renal37/goph-keeper/internal/server/endpoint"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockHandler := NewMockHTTPHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	r := httptest.NewRequest(http.MethodGet, endpoint.BinaryDataEndpoint, nil)
	w := httptest.NewRecorder()

	LoggingHTTPHandler(mockHandler).ServeHTTP(w, r)

	require.Equal(t, http.StatusAccepted, w.Code)
}
