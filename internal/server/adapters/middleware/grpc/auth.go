package middleware

import (
	"context"
	"errors"
	"fmt"

	"github.com/Renal37/goph-keeper/internal/server/adapters/middleware"
	"github.com/Renal37/goph-keeper/internal/server/core/domain/proto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetAuthenticator returns a function for authenticating gRPC requests using JWT tokens.
// It uses the `AuthFromMD` function to extract the token from the metadata and verifies
// the token using `verifyJWTandGetPayload`. If the token is valid, it sets the token's
// claims in the context and returns the enhanced context. If an error occurs, it returns
// an unauthenticated error.
func GetAuthenticator(jwtKey string) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		token, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, fmt.Errorf("AuthFromMD has error: %w", err)
		}

		pl, err := verifyJWTandGetPayload(jwtKey, token)
		if err != nil {
			//nolint:wrapcheck // This legal return
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		enCtx := middleware.SetTokenToContext(ctx, pl)

		return enCtx, nil
	}
}

// AuthMatcher is a function that determines whether a given gRPC call should
// require authentication. It returns `true` if the service name does not match
// the `User_ServiceDesc.ServiceName`, indicating that authentication is required.
func AuthMatcher(ctx context.Context, callMeta interceptors.CallMeta) bool {
	return proto.User_ServiceDesc.ServiceName != callMeta.Service
}

// verifyJWTandGetPayload verifies a JWT token and returns its claims as `JWTclaims`.
// It uses the provided `jwtKey` to parse and validate the token. If the token
// is valid, it returns the claims. If an error occurs during parsing or verification,
// it returns the error.
func verifyJWTandGetPayload(jwtKey string, token string) (middleware.JWTclaims, error) {
	claims := &middleware.JWTclaims{}

	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return *claims, fmt.Errorf("failed signature from jwt: %w", err)
		}
		return *claims, fmt.Errorf("invalid jwt token: %w", err)
	}

	if !tkn.Valid {
		return *claims, fmt.Errorf("jwt token not valid: %w", err)
	}

	return *claims, nil
}
