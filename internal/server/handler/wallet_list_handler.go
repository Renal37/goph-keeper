package handler

import (
	"net/http"
)

type WalletListHandler struct {
	storage WalletStorage
}

func NewWalletListHandler(storage WalletStorage) *WalletListHandler {
	return &WalletListHandler{storage: storage}
}

func (h *WalletListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	h.handleGetData(w, r)
}

type GetCardsResponse struct {
	Cards []*CardData `json:"cards"`
}

func (h *WalletListHandler) handleGetData(w http.ResponseWriter, r *http.Request) {
	cards, err := h.storage.ListCard(r.Context(), getTokenFromRequestContext(r))
	if err != nil {
		responseErrorCardDataRequest(w, err)
	}

	response := GetCardsResponse{
		Cards: cards,
	}

	if err := writeResponse(w, response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
