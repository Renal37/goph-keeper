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

func TestDataListHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockDataStorage(ctrl)
	h := NewListDataHandler(mockStorage)

	r := httptest.NewRequest(http.MethodGet, endpoint.BinariesDataEndpoint, nil)
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	expectedData := []*Record{
		{Name: "name1", Data: "data", Revision: 1},
		{Name: "name2", Data: "data2", Revision: 2},
	}
	mockStorage.EXPECT().ListData(gomock.Any(), testToken).Return(expectedData, nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	answer := w.Body.Bytes()
	resp := &ListDataResponse{}
	err := json.Unmarshal(answer, resp)
	require.NoError(t, err)

	require.Equal(t, expectedData, resp.Data)
}
