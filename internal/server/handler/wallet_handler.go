package handler

import (
	"context"
	"net/http"

	"github.com/Renal37/goph-keeper/internal/zlog"
)

type CardData struct {
	Number string
	Data   string
}

//go:generate mockgen -source=wallet_handler.go -destination=./mock_wallet_handler.go -package=handler
type WalletStorage interface {
	CreateCard(ctx context.Context, userToken string, d *CardData) error
	ListCard(ctx context.Context, userToken string) ([]*CardData, error)
	DeleteCard(ctx context.Context, userToken string, d *CardData) error
}

type WalletHandler struct {
	storage WalletStorage
}

func NewWalletHandler(storage WalletStorage) *WalletHandler {
	return &WalletHandler{storage: storage}
}

func (h *WalletHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	switch r.Method {
	case http.MethodPut:
		err = h.handleSaveData(w, r)
	case http.MethodDelete:
		err = h.handleDeleteData(w, r)
	default:
		zlog.Logger().Infof("unhandled method %s", r.Method)

		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	if err != nil {
		zlog.Logger().Infof("handle error: %s", err)
	}
}

type SaveCardDataRequest struct {
	CardNumber string `json:"number"`
	CardData   string `json:"data"`
}

func (r *SaveCardDataRequest) Validate() bool {
	return len(r.CardNumber) > 0 && len(r.CardData) > 0
}

func (h *WalletHandler) handleSaveData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*SaveCardDataRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	card := &CardData{Number: req.CardNumber, Data: req.CardData}

	if err := h.storage.CreateCard(r.Context(), getTokenFromRequestContext(r), card); err != nil {

		responseErrorCardDataRequest(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

type DeleteCardDataRequest struct {
	CardNumber string `json:"card_number"`
}

func (r *DeleteCardDataRequest) Validate() bool {
	return len(r.CardNumber) > 0
}

func (h *WalletHandler) handleDeleteData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*DeleteCardDataRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	data := &CardData{Number: req.CardNumber}

	if err := h.storage.DeleteCard(r.Context(), getTokenFromRequestContext(r), data); err != nil {
		responseErrorCardDataRequest(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func responseErrorCardDataRequest(w http.ResponseWriter, err error) {
	zlog.Logger().Infof("storage err=%s", err)

	w.WriteHeader(http.StatusInternalServerError)
}
