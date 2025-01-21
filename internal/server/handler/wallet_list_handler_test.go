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

func TestWalletListHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWalletStorage := NewMockWalletStorage(ctrl)
	h := NewWalletListHandler(mockWalletStorage)

	r := httptest.NewRequest(http.MethodGet, endpoint.WalletEndpoint, nil)
	r = r.WithContext(context.WithValue(r.Context(), AuthInfo("token"), testToken))
	w := httptest.NewRecorder()

	cards := []*CardData{
		{Number: "1234123412341234", Data: "secretdata1"},
		{Number: "5678567856785670", Data: "secretdata2"},
	}
	mockWalletStorage.EXPECT().ListCard(gomock.Any(), testToken).Return(cards, nil)

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	answer := w.Body.Bytes()
	resp := &GetCardsResponse{}
	err := json.Unmarshal(answer, resp)
	require.NoError(t, err)

	require.Equal(t, cards, resp.Cards)
}
