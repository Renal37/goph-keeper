// Package handler contains gRPC handlers that implement the server-side logic for the application.
package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Renal37/goph-keeper/internal/server/adapters/middleware"
	"github.com/Renal37/goph-keeper/internal/server/core/domain/proto"
	"github.com/Renal37/goph-keeper/internal/server/core/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler is a gRPC handler that implements the `UserServer` interface
// defined in the `proto` package. It handles gRPC calls related to user
// operations such as registration and login. The handler relies on the
// `UserService` for the business logic and uses a `zap.Logger` for logging.
// It also uses a JWT key (`JWTkey`) for creating JWT tokens during user
// registration and login.
type UserHandler struct {
	proto.UnimplementedUserServer
	Svc    services.UserService
	Logger *zap.Logger
	JWTkey string
}

// Register handles the user registration gRPC call. It creates a new user
// with the provided login and hashed password using the `UserService`.
// If registration is successful, it generates a JWT token for the user.
// Errors during registration or token generation are logged and returned
// as error responses.
func (h UserHandler) Register(ctx context.Context, in *proto.RegiserRequest) (*proto.RegisterResponse, error) {
	var res proto.RegisterResponse

	if in.Login == "" || in.Password == "" {
		res.Error = "login or password incorrect"
		return &res, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed get hash from password")
		res.Error = "internal server error"
		return &res, nil
	}

	user, err := h.Svc.CreateUser(in.Login, string(hash))
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed create user")

		res.Error = "failed create user"
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			res.Error = "this user exists"
		}

		return &res, nil
	}

	token, err := getJWT(h.JWTkey, user.ID, user.Login)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed create jwt token")
		res.Error = "failed create jwt token"
		return &res, nil
	}

	res.Jwt = *token

	return &res, nil
}

// Login handles the user login gRPC call. It verifies the user's credentials
// using the `UserService`. If the credentials are valid, it generates a JWT token
// for the user. Errors during login verification or token generation are logged
// and returned as error responses.
func (h UserHandler) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	var res proto.LoginResponse
	user, err := h.Svc.FindUserByLogin(in.Login)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed get user")
		res.Error = "failed get user"
		return &res, nil
	}

	if user == nil {
		res.Error = "user not found"
		return &res, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(in.Password)); err != nil {
		res.Error = "login or password incorrect"
		//nolint:nilerr // This legal return
		return &res, nil
	}

	token, err := getJWT(h.JWTkey, user.ID, user.Login)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("failed create jwt token")
		res.Error = "failed create jwt token"

		return &res, nil
	}

	res.Jwt = *token

	return &res, nil
}

// getJWT generates a JWT token for the specified user ID and login using the
// provided JWT key. The token includes the user's ID, login, and expiration
// time (defaulting to 30 minutes). If token generation fails, it returns an error.
func getJWT(jwtKey string, id int, login string) (*string, error) {
	var DefaultSession = 30
	var DefaultExpTime = time.Now().Add(time.Duration(DefaultSession) * time.Minute)

	claims := &middleware.JWTclaims{
		ID:    id,
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(DefaultExpTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return nil, fmt.Errorf("failed signed jwt: %w", err)
	}

	return &tokenString, nil
}
