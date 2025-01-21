package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Renal37/goph-keeper/internal/zlog"
)

var (
	ErrDataAlreadyExist = errors.New("data alreadt exist")
	ErrInternalProblem  = errors.New("storage internal error")
	ErrDataNotFound     = errors.New("storage doesn't have data")
	ErrBadRevision      = errors.New("bad revision error")
	ErrUnknownUser      = errors.New("unknown user")
	ErrBadPassword      = errors.New("bad password")
)

//go:generate mockgen -source=data_handler.go -destination=./mock_data_storage.go -package=handler
type DataStorage interface {
	CreateData(ctx context.Context, userToken string, r *Record) error
	UpdateData(ctx context.Context, userToken string, r *Record) error
	LoadData(ctx context.Context, userToken string, name string) (*Record, error)
	ListData(ctx context.Context, userToken string) ([]*Record, error)
	DeleteData(ctx context.Context, userToken string, r *Record) error
}

type Record struct {
	Name     string
	Data     string
	Revision uint64
}

type User struct {
	Login    string
	Password string
	Token    string
}

type DataHandler struct {
	storage DataStorage
}

func NewDataHandler(storage DataStorage) *DataHandler {
	return &DataHandler{storage: storage}
}

func (h *DataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	switch r.Method {
	case http.MethodGet:
		err = h.handleGetData(w, r)
	case http.MethodPost:
		err = h.handleSaveData(w, r)
	case http.MethodPut:
		err = h.handleUpdateData(w, r)
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

type GetDataRequest struct {
	Key string `json:"key"`
}

func (r *GetDataRequest) Validate() bool {
	return len(r.Key) > 0
}

type GetDataResponse struct {
	Key      string `json:"key"`
	Data     string `json:"data"`
	Revision uint64 `json:"revision"`
}

func (h *DataHandler) handleGetData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*GetDataRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	token := getTokenFromRequestContext(r)

	data, err := h.storage.LoadData(r.Context(), token, req.Key)
	if err != nil {
		responsestorageError(w, err)
		return err
	}

	response := GetDataResponse{
		Key:      data.Name,
		Data:     data.Data,
		Revision: data.Revision,
	}

	if err := writeResponse(w, response); err != nil {
		return err
	}

	return nil
}

type SaveDataRequest struct {
	Key  string `json:"key"`
	Data string `json:"data"`
}

func (r *SaveDataRequest) Validate() bool {
	return len(r.Key) > 0 && len(r.Data) > 0
}

func (h *DataHandler) handleSaveData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*SaveDataRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	token := getTokenFromRequestContext(r)
	data := &Record{Name: req.Key, Data: req.Data}

	if err = h.storage.CreateData(r.Context(), token, data); err != nil {

		responsestorageError(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

type UpdateDataRequest struct {
	Key      string `json:"key"`
	Data     string `json:"data"`
	Revision uint64 `json:"revision"`
}

func (r *UpdateDataRequest) Validate() bool {
	return len(r.Key) > 0 && len(r.Data) > 0 && r.Revision != 0
}

func (h *DataHandler) handleUpdateData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*UpdateDataRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	token := getTokenFromRequestContext(r)
	data := &Record{Name: req.Key, Data: req.Data, Revision: req.Revision}

	if err := h.storage.UpdateData(r.Context(), token, data); err != nil {
		responsestorageError(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

type DeleteDataRequest struct {
	Key string `json:"key"`
}

func (r *DeleteDataRequest) Validate() bool {
	return len(r.Key) > 0
}

func (h *DataHandler) handleDeleteData(w http.ResponseWriter, r *http.Request) error {
	req, err := readRequest[*DeleteDataRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return err
	}

	token := getTokenFromRequestContext(r)

	data := &Record{Name: req.Key}

	if err := h.storage.DeleteData(r.Context(), token, data); err != nil {
		responsestorageError(w, err)

		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func responsestorageError(w http.ResponseWriter, err error) {
	zlog.Logger().Infof("storage err=%s", err)

	if errors.Is(err, ErrBadRevision) {
		w.WriteHeader(http.StatusConflict)
	} else if errors.Is(err, ErrDataNotFound) {
		w.WriteHeader(http.StatusNotFound)
	} else if errors.Is(err, ErrDataAlreadyExist) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
