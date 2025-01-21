package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/Renal37/goph-keeper/internal/server/endpoint"
	"github.com/Renal37/goph-keeper/internal/server/handler"
	"github.com/Renal37/goph-keeper/internal/zlog"
)

type AuthModdleware struct {
	checker UserChecker
}

func NewAuthMiddleware(c UserChecker) *AuthModdleware {
	return &AuthModdleware{
		checker: c,
	}
}

//go:generate mockgen -source=auth.go -destination=./mock_user_checker.go -package=middleware
type UserChecker interface {
	Check(ctx context.Context, token string) error
}

func (a *AuthModdleware) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == endpoint.RegisterEndpoint {
			login := r.Header.Get("login")
			if len(login) == 0 {
				zlog.Logger().Debug("headers don't have login field")

				w.WriteHeader(http.StatusBadRequest)

				return
			}

			password := r.Header.Get("password")
			if len(password) == 0 {
				zlog.Logger().Debug("headers don't have password field")

				w.WriteHeader(http.StatusBadRequest)

				return
			}

			ctxWithAuthInfo := context.WithValue(r.Context(), handler.AuthInfo("user"), &handler.User{Login: login, Password: password})

			r = r.WithContext(ctxWithAuthInfo)

			h.ServeHTTP(w, r)

			return
		}

		token := r.Header.Get("token")
		if len(token) == 0 {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		ctxWithAuthInfo := context.WithValue(r.Context(), handler.AuthInfo("token"), token)
		r = r.WithContext(ctxWithAuthInfo)

		if err := a.checker.Check(r.Context(), token); err != nil {
			if errors.Is(err, handler.ErrUnknownUser) {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			return
		}

		h.ServeHTTP(w, r)
	})
}
