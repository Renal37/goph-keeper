// Package core contains basic app logic.
package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	handler "github.com/Renal37/goph-keeper/internal/server/adapters/handler/grpc"
	interceptors "github.com/Renal37/goph-keeper/internal/server/adapters/middleware/grpc"
	repository "github.com/Renal37/goph-keeper/internal/server/adapters/repository/pg"
	"github.com/Renal37/goph-keeper/internal/server/core/domain/proto"
	"github.com/Renal37/goph-keeper/internal/server/core/services"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// RunGRPCserver run gRPC server.
func RunGRPCserver(
	lg *zap.Logger,
	host string,
	pathCert string,
	pathKey string,
	jwtKey string,
	mk string,
	repo *repository.DB,
) error {
	lg.Info("gRPC server start...", zap.String("address", host))

	// Load certificates
	tlsCredentials, err := loadTLSCredentials(pathCert, pathKey)
	if err != nil {
		return fmt.Errorf("failed load tls: %w", err)
	}

	// Listen port
	listen, err := net.Listen("tcp", host)
	if err != nil {
		return fmt.Errorf("failde listen grpc port: %w", err)
	}

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	// Create gRPC server
	s := grpc.NewServer(
		grpc.Creds(tlsCredentials),
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(interceptors.InterceptorLogger(lg), opts...),
			selector.UnaryServerInterceptor(
				auth.UnaryServerInterceptor(interceptors.GetAuthenticator(jwtKey)),
				selector.MatchFunc(interceptors.AuthMatcher),
			),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(interceptors.InterceptorLogger(lg), opts...),
			selector.StreamServerInterceptor(
				auth.StreamServerInterceptor(interceptors.GetAuthenticator(jwtKey)),
				selector.MatchFunc(interceptors.AuthMatcher),
			),
		),
	)

	// Create user service
	userSvc := services.NewUserService(repo)
	proto.RegisterUserServer(s, &handler.UserHandler{
		Svc:    *userSvc,
		Logger: lg,
		JWTkey: jwtKey,
	})

	// Create storage service
	storageSvc := services.NewStorageService(repo)
	proto.RegisterStorageServer(s, &handler.StorageHandler{
		Svc:       *storageSvc,
		Logger:    lg,
		MasterKey: mk,
	})

	// Graceful server
	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	errCh := make(chan error, 1)

	wg.Add(1)

	go func() {
		defer func() {
			if err := repo.Close(); err != nil {
				lg.Info(err.Error())
			}

			stop()
			wg.Done()
		}()

		<-ctx.Done()
	}()

	// Start gRPC server
	go func() {
		if err := s.Serve(listen); err != nil {
			errCh <- err
			return
		}
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failde create grpc server: %w", err)
		}
	case <-ctx.Done():
	}

	wg.Wait()
	return nil
}

// loadTLSCredentials loading cert.
func loadTLSCredentials(cert string, key string) (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("failde load file: %w", err)
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(config), nil
}
