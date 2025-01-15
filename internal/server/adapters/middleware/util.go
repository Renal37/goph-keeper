package middleware

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey represents the type used for keys in the context package.
// Using a custom type helps avoid collisions with other context keys.
type contextKey int

// JWTclaims represents the claims from a JWT token, including the user ID,
// login, and standard JWT registered claims.
type JWTclaims struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	jwt.RegisteredClaims
}

// Enumeration of context keys used for storing values in context.
const (
	ContextKeyToken contextKey = iota
)

// GetTokenFromContext retrieves JWT claims from the given context.
// It returns the JWT claims and a boolean indicating whether the claims
// were successfully retrieved. If the claims are not found in the context,
// the function returns false.
func GetTokenFromContext(ctx context.Context) (JWTclaims, bool) {
	caller, ok := ctx.Value(ContextKeyToken).(JWTclaims)
	return caller, ok
}

// SetTokenToContext adds JWT claims to the given context and returns
// the new context. It associates the claims with the `ContextKeyToken` key.
func SetTokenToContext(ctx context.Context, pl JWTclaims) context.Context {
	return context.WithValue(ctx, ContextKeyToken, pl)
}
