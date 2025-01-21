package handler

import (
	"net/http"
)

type SecretListHandler struct {
	storage SecretStorage
}

func NewSecretListHandler(storage SecretStorage) *SecretListHandler {
	return &SecretListHandler{storage: storage}
}

type SecretListResponse struct {
	Data []*Secret `json:"data"`
}

func (h *SecretListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	h.handleListData(w, r)
}

func (h *SecretListHandler) handleListData(w http.ResponseWriter, r *http.Request) {
	token := getTokenFromRequestContext(r)

	data, err := h.storage.ListSecret(r.Context(), token)
	if err != nil {
		responsestorageError(w, err)
		return
	}

	response := SecretListResponse{
		Data: data,
	}

	if err := writeResponse(w, response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
