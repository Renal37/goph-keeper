package handler

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/Renal37/goph-keeper/internal/client/gophcrypto"
	"github.com/Renal37/goph-keeper/internal/zlog"
)

//go:generate mockgen -source=register_handler.go -destination=./mock_registrator.go -package=handler
type Registrator interface {
	Register(ctx context.Context, user *User) error
}

type RegisterHandler struct {
	registrator Registrator
}

func NewRegistrationHandler(registrator Registrator) *RegisterHandler {
	return &RegisterHandler{
		registrator: registrator,
	}
}

type RegistrationResponse struct {
	Token string `json:"token"`
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	user := getUserFromRequestContext(r)

	token, err := makeUserToken(user.Login, user.Password)
	if err != nil {
		zlog.Logger().Errorf("can't make user token")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	user.Token = token

	if err := h.registrator.Register(r.Context(), user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if err := writeResponse(w, &RegistrationResponse{Token: token}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func makeUserToken(login, password string) (string, error) {
	const key = "gPPDZtkm/8d4jg4tJ6DpprjFINUz34SklFsrRJg8Dts"

	var err error

	cryptoKey, err := base64.RawStdEncoding.DecodeString(key)
	if err != nil {
		panic(err)
	}

	crypto, err := gophcrypto.New(cryptoKey)
	if err != nil {
		panic(err)
	}

	token := crypto.Encrypt([]byte(login + password))

	return token, nil
}
