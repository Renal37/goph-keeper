package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Renal37/goph-keeper/internal/server/config"
	"github.com/Renal37/goph-keeper/internal/server/endpoint"
	"github.com/Renal37/goph-keeper/internal/server/handler"
	"github.com/Renal37/goph-keeper/internal/server/middleware"
	"github.com/Renal37/goph-keeper/internal/server/sql"
	"github.com/Renal37/goph-keeper/internal/zlog"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	httpServer http.Server
	storage    *sql.Storage

	wait chan struct{}
}

func StartNew(config *config.Config) (*Server, error) {
	storage, err := sql.StartNewStorage(config.DataSourceName)
	if err != nil {
		return nil, fmt.Errorf("start db storage, err=%w", err)
	}

	router := chi.NewRouter()

	authMiddleware := middleware.NewAuthMiddleware(storage)

	router.Use(authMiddleware.Middleware)

	router.Handle(endpoint.RegisterEndpoint, handler.NewRegistrationHandler(storage))

	router.Handle(endpoint.BinaryDataEndpoint, handler.NewDataHandler(storage))
	router.Handle(endpoint.BinariesDataEndpoint, handler.NewListDataHandler(storage))

	router.Handle(endpoint.WalletEndpoint, handler.NewWalletHandler(storage))
	router.Handle(endpoint.WalletsEndpoint, handler.NewWalletListHandler(storage))

	router.Handle(endpoint.SecretEndpoint, handler.NewSecretDataHandler(storage))
	router.Handle(endpoint.SecretsEndpoint, handler.NewSecretListHandler(storage))

	server := &Server{
		httpServer: http.Server{Addr: config.Hostport, Handler: router},
		storage:    storage,
		wait:       make(chan struct{}),
	}

	server.start()

	return server, nil
}

func (s *Server) start() {
	go func() {
		defer close(s.wait)

		err := s.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zlog.Logger().Errorf("server failed with err=%s", err)

			return
		}
	}()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return errors.Join(
		s.storage.Stop(),
		s.httpServer.Shutdown(ctx),
	)
}

func (s *Server) WaitStop() <-chan struct{} {
	return s.wait
}
