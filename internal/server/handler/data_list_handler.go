package handler

import (
	"net/http"
)

type ListDataHandler struct {
	storage DataStorage
}

func NewListDataHandler(storage DataStorage) *ListDataHandler {
	return &ListDataHandler{storage: storage}
}

type ListDataResponse struct {
	Data []*Record `json:"data"`
}

func (h *ListDataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	h.handleListData(w, r)
}

func (h *ListDataHandler) handleListData(w http.ResponseWriter, r *http.Request) {
	token := getTokenFromRequestContext(r)

	data, err := h.storage.ListData(r.Context(), token)
	if err != nil {
		responsestorageError(w, err)
		return
	}

	response := ListDataResponse{
		Data: data,
	}

	if err := writeResponse(w, response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
