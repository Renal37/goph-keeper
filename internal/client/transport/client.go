package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Renal37/goph-keeper/internal/client/config"
	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/Renal37/goph-keeper/internal/server/endpoint"
	"github.com/Renal37/goph-keeper/internal/server/handler"
)

//go:generate mockgen -source=client.go -destination=./mock_client.go -package=transport
type BinaryDataClient interface {
	UploadBinaryData(ctx context.Context, u *storage.User, r *storage.Record) error
	UpdateBinaryData(ctx context.Context, u *storage.User, r *storage.Record) error
	DownloadBinaryData(ctx context.Context, u *storage.User, dataKey string) (*storage.Record, error)
	DeleteBinaryData(ctx context.Context, u *storage.User, dataKey string) error
}

type RegisterClient interface {
	RegisterUser(ctx context.Context, login string, password string) (string, error)
}

type SecretDataClient interface {
	CreateSecret(ctx context.Context, userToken string, secretName string, secretData string) error
	DeleteSecret(ctx context.Context, userToken string, secretKey string) error
	GetSecret(ctx context.Context, userToken string, secretName string) (*storage.Secret, error)
}

type WalletClient interface {
	CreateCardData(ctx context.Context, userToken string, cardNumber string, cardData string) error
	DeleteCardData(ctx context.Context, userToken string, cardNumber string) error
	ListCardData(ctx context.Context, userToken string) ([]*handler.CardData, error)
}

type Client struct {
	hostport string
	done     chan error
}

func NewClient(config *config.Config) *Client {
	return &Client{
		hostport: config.Hostport,
		done:     make(chan error),
	}
}

func (c *Client) UploadBinaryData(
	ctx context.Context,
	u *storage.User,
	r *storage.Record,
) error {
	saveDataRequest := handler.SaveDataRequest{
		Key:  r.Name,
		Data: string(r.Data),
	}

	uri := makeURI(c.hostport, endpoint.BinaryDataEndpoint)
	headers := map[string]string{
		"token": u.Token,
	}

	return requestAndHandle(ctx, uri, http.MethodPost, headers, saveDataRequest, func(r *http.Response) error {
		if r.StatusCode == http.StatusConflict {
			return errors.New("data already exist")
		}

		return defaultHttpResponseHandler(r)
	})
}

func (c *Client) UpdateBinaryData(
	ctx context.Context,
	u *storage.User,
	r *storage.Record,
) error {
	saveDataRequest := handler.UpdateDataRequest{
		Key:      r.Name,
		Data:     string(r.Data),
		Revision: r.Revision,
	}

	uri := makeURI(c.hostport, endpoint.BinaryDataEndpoint)
	headers := map[string]string{
		"token": u.Token,
	}

	return requestAndHandle(ctx, uri, http.MethodPut, headers, saveDataRequest, func(r *http.Response) error {
		if r.StatusCode == http.StatusNotFound {
			return errors.New("data isn't exist")
		}

		return defaultHttpResponseHandler(r)
	})
}

func (c *Client) DownloadBinaryData(
	ctx context.Context,
	u *storage.User,
	dataKey string,
) (*storage.Record, error) {
	uri := makeURI(c.hostport, endpoint.BinaryDataEndpoint)
	headers := map[string]string{
		"token": u.Token,
	}

	getDataRequest := &handler.GetDataRequest{
		Key: dataKey,
	}

	resp, err := requestHandleAndParse[handler.GetDataResponse](ctx, uri, http.MethodGet, headers, getDataRequest, func(r *http.Response) error {
		if r.StatusCode == http.StatusNotFound {
			return errors.New("data isn't exist")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &storage.Record{Name: dataKey, Data: resp.Data, Revision: resp.Revision}, nil
}

func (c *Client) RegisterUser(
	ctx context.Context,
	login string,
	password string,
) (string, error) {
	uri := makeURI(c.hostport, endpoint.RegisterEndpoint)

	headers := map[string]string{
		"login":    login,
		"password": password,
	}

	resp, err := requestAndParse[handler.RegistrationResponse](ctx, uri, http.MethodPut, headers, nil)
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}

func (c *Client) DeleteBinaryData(
	ctx context.Context,
	u *storage.User,
	dataKey string,
) error {
	uri := makeURI(c.hostport, endpoint.BinaryDataEndpoint)

	deleteRequest := &handler.DeleteDataRequest{
		Key: dataKey,
	}

	headers := map[string]string{
		"token": u.Token,
	}

	return request(ctx, uri, http.MethodDelete, headers, deleteRequest)
}

func (c *Client) CreateCardData(
	ctx context.Context,
	userToken string,
	cardNumber string,
	cardData string,
) error {
	uri := makeURI(c.hostport, endpoint.WalletEndpoint)

	saveRequest := &handler.SaveCardDataRequest{CardNumber: cardNumber, CardData: cardData}

	return request(ctx, uri, http.MethodPut, map[string]string{"token": userToken}, saveRequest)
}

func (c *Client) DeleteCardData(
	ctx context.Context,
	userToken string,
	cardNumber string,
) error {
	uri := makeURI(c.hostport, endpoint.WalletEndpoint)

	deleteRequest := &handler.DeleteCardDataRequest{CardNumber: cardNumber}

	return request(ctx, uri, http.MethodDelete, map[string]string{"token": userToken}, deleteRequest)
}

func (c *Client) ListCardData(
	ctx context.Context,
	userToken string,
) ([]*handler.CardData, error) {
	uri := makeURI(c.hostport, endpoint.WalletEndpoint)

	resp, err := requestAndParse[handler.GetCardsResponse](ctx, uri, http.MethodGet, map[string]string{"token": userToken}, nil)
	if err != nil {
		return nil, err
	}

	return resp.Cards, nil
}

func (c *Client) CreateSecret(
	ctx context.Context,
	userToken string,
	secretName string,
	secretData string,
) error {
	uri := makeURI(c.hostport, endpoint.SecretEndpoint)

	saveRequest := &handler.SaveSecretRequest{Key: secretName, Value: secretData}

	return request(ctx, uri, http.MethodPut, map[string]string{"token": userToken}, saveRequest)
}

func (c *Client) DeleteSecret(
	ctx context.Context,
	userToken string,
	secretKey string,
) error {
	uri := makeURI(c.hostport, endpoint.SecretEndpoint)

	saveRequest := &handler.DeleteSecretRequest{Key: secretKey}

	return request(ctx, uri, http.MethodDelete, map[string]string{"token": userToken}, saveRequest)
}

func (c *Client) GetSecret(
	ctx context.Context,
	userToken string,
	secretName string,
) (*storage.Secret, error) {
	uri := makeURI(c.hostport, endpoint.SecretEndpoint)

	request := handler.GetSecretDataRequest{Key: secretName}

	resp, err := requestAndParse[handler.GetSecretResponse](ctx, uri, http.MethodGet, map[string]string{"token": userToken}, request)
	if err != nil {
		return nil, err
	}

	return &storage.Secret{Key: resp.Key, Value: resp.Data}, nil
}

type httpResponseHandler func(*http.Response) error

func defaultHttpResponseHandler(r *http.Response) error {
	switch r.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return errors.New("user must be registered")
	default:
		return statusCodeToError(r.StatusCode)
	}
}

func request(
	ctx context.Context,
	uri string,
	method string,
	headers map[string]string,
	request any,
) error {
	return requestAndHandle(ctx, uri, method, headers, request, defaultHttpResponseHandler)
}

func statusCodeToError(statusCode int) error {
	return fmt.Errorf("request failed code=%d error=%s", statusCode, http.StatusText(statusCode))
}

func requestAndHandle(
	ctx context.Context,
	uri string,
	method string,
	headers map[string]string,
	request any,
	handler httpResponseHandler,
) error {
	req, err := makeRequest(ctx, uri, method, headers, request)
	if err != nil {
		return err
	}

	resp, err := doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := handler(resp); err != nil {
		return err
	}

	return nil
}

func requestAndParse[T any](
	ctx context.Context,
	uri string,
	method string,
	headers map[string]string,
	request any,
) (*T, error) {
	return requestHandleAndParse[T](ctx, uri, method, headers, request, defaultHttpResponseHandler)
}

func requestHandleAndParse[T any](
	ctx context.Context,
	uri string,
	method string,
	headers map[string]string,
	request any,
	handler httpResponseHandler,
) (*T, error) {
	req, err := makeRequest(ctx, uri, method, headers, request)
	if err != nil {
		return nil, err
	}

	resp, err := doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := handler(resp); err != nil {
		return nil, err
	}

	return parseResponse[T](resp)
}

func parseResponse[T any](resp *http.Response) (*T, error) {
	bin, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	obj := new(T)

	err = json.Unmarshal(bin, obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func makeURI(hostport string, endpoint string) string {
	return hostport + endpoint
}

func makeRequest(
	ctx context.Context,
	uri string,
	method string,
	additionalHeaders map[string]string,
	msg any,
) (*http.Request, error) {
	var body io.Reader

	if msg != nil {
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	for header, v := range additionalHeaders {
		req.Header.Set(header, v)
	}

	return req, nil
}

var tryingIntervals []time.Duration = []time.Duration{
	time.Millisecond * 100,
	time.Millisecond * 300,
	time.Millisecond * 500,
}

func doRequest(req *http.Request) (*http.Response, error) {
	maxTryingsNum := len(tryingIntervals)

	var err error

	for trying := 0; trying <= maxTryingsNum; trying++ {
		var resp *http.Response

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			if trying < maxTryingsNum {
				time.Sleep(tryingIntervals[trying])
			}
		} else {
			return resp, nil
		}
	}

	return nil, fmt.Errorf("request error: %w", err)
}
