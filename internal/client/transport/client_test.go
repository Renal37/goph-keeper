package transport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Renal37/goph-keeper/internal/client/config"
	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/Renal37/goph-keeper/internal/server/endpoint"
	"github.com/Renal37/goph-keeper/internal/server/handler"
	"github.com/stretchr/testify/require"
)

func TestUploadData(t *testing.T) {
	ctx := context.Background()
	user := &storage.User{Token: "token"}
	record := &storage.Record{Name: "n", Data: "d", Revision: 123}

	finished := false

	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.URL.Path == endpoint.BinaryDataEndpoint {
			w.WriteHeader(http.StatusInternalServerError)
		}

		req := parseRequest[handler.SaveDataRequest](t, r)
		require.Equal(t, record.Data, req.Data)
		require.Equal(t, record.Name, req.Key)
		require.Equal(t, "token", r.Header.Get("token"))

		w.WriteHeader(http.StatusOK)
		finished = true
	}))

	defer srvr.Close()

	cl := NewClient(&config.Config{Hostport: srvr.URL})

	err := cl.UploadBinaryData(ctx, user, record)
	require.NoError(t, err)
	require.True(t, finished)
}

func TestUpdateBinData(t *testing.T) {
	ctx := context.Background()
	user := &storage.User{Token: "token"}
	record := &storage.Record{Name: "n", Data: "d", Revision: 123}

	finished := false

	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut && r.URL.Path == endpoint.BinaryDataEndpoint {
			w.WriteHeader(http.StatusInternalServerError)
		}

		req := parseRequest[handler.UpdateDataRequest](t, r)
		require.Equal(t, record.Data, req.Data)
		require.Equal(t, record.Name, req.Key)
		require.Equal(t, record.Revision, req.Revision)
		require.Equal(t, "token", r.Header.Get("token"))

		w.WriteHeader(http.StatusOK)
		finished = true
	}))

	defer srvr.Close()

	cl := NewClient(&config.Config{Hostport: srvr.URL})

	err := cl.UpdateBinaryData(ctx, user, record)
	require.NoError(t, err)
	require.True(t, finished)
}

func TestDownloadBinData(t *testing.T) {
	ctx := context.Background()
	user := &storage.User{Token: "token"}
	record := &storage.Record{Name: "n", Data: "d", Revision: 123}

	finished := false

	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.URL.Path == endpoint.BinaryDataEndpoint {
			w.WriteHeader(http.StatusInternalServerError)
		}

		req := parseRequest[handler.GetDataRequest](t, r)
		require.Equal(t, record.Name, req.Key)
		require.Equal(t, "token", r.Header.Get("token"))

		data, err := json.Marshal(record)
		require.NoError(t, err)

		_, err = w.Write(data)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		finished = true
	}))

	defer srvr.Close()

	cl := NewClient(&config.Config{Hostport: srvr.URL})

	actual, err := cl.DownloadBinaryData(ctx, user, record.Name)
	require.NoError(t, err)
	require.Equal(t, record, actual)
	require.True(t, finished)
}

func TestDeleteData(t *testing.T) {
	ctx := context.Background()
	user := &storage.User{Token: "token"}
	record := &storage.Record{Name: "n", Data: "d", Revision: 123}

	finished := false

	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete && r.URL.Path == endpoint.BinaryDataEndpoint {
			w.WriteHeader(http.StatusInternalServerError)
		}

		req := parseRequest[handler.DeleteDataRequest](t, r)
		require.Equal(t, record.Name, req.Key)
		require.Equal(t, "token", r.Header.Get("token"))

		w.WriteHeader(http.StatusOK)
		finished = true
	}))

	defer srvr.Close()

	cl := NewClient(&config.Config{Hostport: srvr.URL})

	err := cl.DeleteBinaryData(ctx, user, record.Name)
	require.NoError(t, err)
	require.True(t, finished)
}

func TestRegister(t *testing.T) {
	ctx := context.Background()
	user := &storage.User{Login: "login", Password: "password"}

	finished := false

	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut && r.URL.Path == endpoint.RegisterEndpoint {
			w.WriteHeader(http.StatusInternalServerError)
		}

		require.Equal(t, "login", r.Header.Get("login"))
		require.Equal(t, "password", r.Header.Get("password"))

		resp := handler.RegistrationResponse{Token: "token"}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		_, err = w.Write(data)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		finished = true
	}))

	defer srvr.Close()

	cl := NewClient(&config.Config{Hostport: srvr.URL})

	token, err := cl.RegisterUser(ctx, user.Login, user.Password)
	require.NoError(t, err)
	require.Equal(t, "token", token)
	require.True(t, finished)
}

func TestCreateCard(t *testing.T) {
	ctx := context.Background()
	user := &storage.User{Token: "token"}
	card := &storage.BankCard{Number: "1234123412341234", ExpiryDate: time.Now().Truncate(time.Hour * 24), Owner: "IVAN PETOROV", CvvCode: "123"}

	finished := false

	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut && r.URL.Path == endpoint.WalletEndpoint {
			w.WriteHeader(http.StatusInternalServerError)
		}

		req := parseRequest[handler.SaveCardDataRequest](t, r)
		require.Equal(t, card.Number, req.CardNumber)
		require.Equal(t, "data", req.CardData)
		require.Equal(t, user.Token, r.Header.Get("token"))

		w.WriteHeader(http.StatusOK)
		finished = true
	}))

	defer srvr.Close()

	cl := NewClient(&config.Config{Hostport: srvr.URL})

	err := cl.CreateCardData(ctx, user.Token, card.Number, "data")
	require.NoError(t, err)
	require.True(t, finished)
}

func TestDeleteCard(t *testing.T) {
	ctx := context.Background()
	user := &storage.User{Token: "token"}
	card := &storage.BankCard{Number: "1234123412341234", ExpiryDate: time.Now().Truncate(time.Hour * 24), Owner: "IVAN PETOROV", CvvCode: "123"}

	finished := false

	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete && r.URL.Path == endpoint.WalletEndpoint {
			w.WriteHeader(http.StatusInternalServerError)
		}

		req := parseRequest[handler.DeleteCardDataRequest](t, r)
		require.Equal(t, card.Number, req.CardNumber)
		require.Equal(t, user.Token, r.Header.Get("token"))

		w.WriteHeader(http.StatusOK)
		finished = true
	}))

	defer srvr.Close()

	cl := NewClient(&config.Config{Hostport: srvr.URL})

	err := cl.DeleteCardData(ctx, user.Token, card.Number)
	require.NoError(t, err)
	require.True(t, finished)
}

func TestListCard(t *testing.T) {
	ctx := context.Background()
	user := &storage.User{Token: "token"}
	cards := []*handler.CardData{{Number: "n", Data: "d"}}

	finished := false

	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.URL.Path == endpoint.WalletEndpoint {
			w.WriteHeader(http.StatusInternalServerError)
		}

		require.Equal(t, user.Token, r.Header.Get("token"))

		l := &handler.GetCardsResponse{
			Cards: cards,
		}

		data, err := json.Marshal(l)
		require.NoError(t, err)

		_, err = w.Write(data)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)

		finished = true
	}))

	defer srvr.Close()

	cl := NewClient(&config.Config{Hostport: srvr.URL})

	data, err := cl.ListCardData(ctx, user.Token)
	require.NoError(t, err)
	require.Equal(t, cards, data)
	require.True(t, finished)
}

func TestCreateSecret(t *testing.T) {
	ctx := context.Background()
	user := &storage.User{Token: "token"}
	secret := &storage.Secret{Name: "s", Key: "k", Value: "v"}

	finished := false

	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut && r.URL.Path == endpoint.SecretEndpoint {
			w.WriteHeader(http.StatusInternalServerError)
		}

		req := parseRequest[handler.SaveSecretRequest](t, r)

		require.Equal(t, secret.Name, req.Key)
		require.Equal(t, "data", req.Value)
		require.Equal(t, user.Token, r.Header.Get("token"))

		w.WriteHeader(http.StatusOK)
		finished = true
	}))

	defer srvr.Close()

	cl := NewClient(&config.Config{Hostport: srvr.URL})

	err := cl.CreateSecret(ctx, user.Token, secret.Name, "data")
	require.NoError(t, err)
	require.True(t, finished)
}

func TestDeleteSecret(t *testing.T) {
	ctx := context.Background()
	user := &storage.User{Token: "token"}
	secret := &storage.Secret{Name: "s", Key: "k", Value: "v"}

	finished := false

	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete && r.URL.Path == endpoint.SecretEndpoint {
			w.WriteHeader(http.StatusInternalServerError)
		}

		req := parseRequest[handler.DeleteSecretRequest](t, r)

		require.Equal(t, secret.Name, req.Key)
		require.Equal(t, user.Token, r.Header.Get("token"))

		w.WriteHeader(http.StatusOK)
		finished = true
	}))

	defer srvr.Close()

	cl := NewClient(&config.Config{Hostport: srvr.URL})

	err := cl.DeleteSecret(ctx, user.Token, secret.Name)
	require.NoError(t, err)
	require.True(t, finished)
}

func TestGetSecret(t *testing.T) {
	ctx := context.Background()
	user := &storage.User{Token: "token"}
	secret := &storage.Secret{Name: "s", Key: "k", Value: "v"}

	finished := false

	srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.URL.Path == endpoint.SecretEndpoint {
			w.WriteHeader(http.StatusInternalServerError)
		}

		req := parseRequest[handler.GetSecretDataRequest](t, r)

		require.Equal(t, secret.Name, req.Key)
		require.Equal(t, user.Token, r.Header.Get("token"))

		resp := &handler.GetSecretResponse{
			Key:  "s",
			Data: "data",
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		_, err = w.Write(data)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		finished = true
	}))

	defer srvr.Close()

	cl := NewClient(&config.Config{Hostport: srvr.URL})

	actual, err := cl.GetSecret(ctx, user.Token, secret.Name)
	require.NoError(t, err)
	require.Equal(t, "s", actual.Key)
	require.Equal(t, "data", actual.Value)
	require.True(t, finished)
}

func parseRequest[T any](t *testing.T, r *http.Request) *T {
	bin, err := io.ReadAll(r.Body)
	require.NoError(t, err)

	obj := new(T)

	err = json.Unmarshal(bin, obj)
	require.NoError(t, err)

	return obj
}
