package middleware

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey представляет тип, используемый для ключей в пакете context.
// Использование пользовательского типа помогает избежать коллизий с другими ключами контекста.
type contextKey int

// JWTclaims представляет утверждения из JWT-токена, включая идентификатор пользователя,
// логин и стандартные зарегистрированные утверждения JWT.
type JWTclaims struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	jwt.RegisteredClaims
}

// Перечисление ключей контекста, используемых для хранения значений в контексте.
const (
	ContextKeyToken contextKey = iota
)

// GetTokenFromContext извлекает утверждения JWT из данного контекста.
// Возвращает утверждения JWT и булево значение, указывающее, были ли утверждения
// успешно извлечены. Если утверждения не найдены в контексте, функция возвращает false.
func GetTokenFromContext(ctx context.Context) (JWTclaims, bool) {
	caller, ok := ctx.Value(ContextKeyToken).(JWTclaims)
	return caller, ok
}

// SetTokenToContext добавляет утверждения JWT в данный контекст и возвращает
// новый контекст. Он связывает утверждения с ключом `ContextKeyToken`.
func SetTokenToContext(ctx context.Context, pl JWTclaims) context.Context {
	return context.WithValue(ctx, ContextKeyToken, pl)
}
