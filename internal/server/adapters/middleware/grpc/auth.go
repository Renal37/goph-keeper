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

// GetAuthenticator возвращает функцию для аутентификации gRPC-запросов с использованием JWT-токенов.
// Использует функцию `AuthFromMD` для извлечения токена из метаданных и проверяет
// токен с помощью `verifyJWTandGetPayload`. Если токен действителен, он устанавливает
// утверждения токена в контексте и возвращает расширенный контекст. Если возникает ошибка,
// возвращает ошибку неаутентифицированного доступа.
func GetAuthenticator(jwtKey string) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		token, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, fmt.Errorf("Ошибка AuthFromMD: %w", err)
		}

		pl, err := verifyJWTandGetPayload(jwtKey, token)
		if err != nil {
			//nolint:wrapcheck // Это допустимый возврат
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		enCtx := middleware.SetTokenToContext(ctx, pl)

		return enCtx, nil
	}
}

// AuthMatcher — это функция, которая определяет, должен ли данный gRPC-вызов
// требовать аутентификации. Возвращает `true`, если имя службы не совпадает
// с `User_ServiceDesc.ServiceName`, указывая, что аутентификация требуется.
func AuthMatcher(ctx context.Context, callMeta interceptors.CallMeta) bool {
	return proto.User_ServiceDesc.ServiceName != callMeta.Service
}

// verifyJWTandGetPayload проверяет JWT-токен и возвращает его утверждения как `JWTclaims`.
// Использует предоставленный `jwtKey` для разбора и проверки токена. Если токен
// действителен, возвращает утверждения. Если возникает ошибка во время разбора или проверки,
// возвращает ошибку.
func verifyJWTandGetPayload(jwtKey string, token string) (middleware.JWTclaims, error) {
	claims := &middleware.JWTclaims{}

	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return *claims, fmt.Errorf("Ошибка подписи JWT: %w", err)
		}
		return *claims, fmt.Errorf("Недействительный JWT-токен: %w", err)
	}

	if !tkn.Valid {
		return *claims, fmt.Errorf("JWT-токен недействителен: %w", err)
	}

	return *claims, nil
}
