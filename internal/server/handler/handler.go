package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Renal37/goph-keeper/internal/zlog"
)

type Validaitable interface {
	Validate() bool
}

var ErrIsNotValid = errors.New("is not valid")

func readRequest[T Validaitable](r *http.Request) (T, error) {
	var parsedRequest T

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return parsedRequest, fmt.Errorf("read all err=%w", err)
	}

	zlog.Logger().Debugf("request data: %s", string(data))

	if err := json.Unmarshal(data, &parsedRequest); err != nil {
		return parsedRequest, fmt.Errorf("unmarshal data=%s err=%w", string(data), err)
	}

	if !parsedRequest.Validate() {
		return parsedRequest, ErrIsNotValid
	}

	return parsedRequest, nil
}

func writeResponse[T any](w http.ResponseWriter, response T) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal err=%w", err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return fmt.Errorf("write data  err=%w", err)
	}

	return nil
}

func getUserFromRequestContext(r *http.Request) *User {
	return (r.Context().Value(AuthInfo("user"))).(*User)
}

func getTokenFromRequestContext(r *http.Request) string {
	return (r.Context().Value(AuthInfo("token"))).(string)
}

type AuthInfo string
