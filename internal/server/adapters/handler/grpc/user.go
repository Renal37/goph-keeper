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

// UserHandler обработчик GRPC, реализующий интерфейс `UserServer`
// из пакета `proto`. Обрабатывает GRPC вызовы, связанные с операциями
// пользователя, такими как регистрация и вход. Обработчик использует
// `UserService` для бизнес-логики и `zap.Logger` для логирования.
// Также использует JWT ключ (`JWTkey`) для создания JWT токенов при
// регистрации и входе пользователя.
type UserHandler struct {
	proto.UnimplementedUserServer
	Svc    services.UserService
	Logger *zap.Logger
	JWTkey string
}

// Register обрабатывает GRPC вызов регистрации пользователя. Создает нового
// пользователя с предоставленным логином и хэшированным паролем, используя
// `UserService`. Если регистрация успешна, генерирует JWT токен для пользователя.
// Ошибки во время регистрации или генерации токена логируются и возвращаются
// как ответы с ошибками.
func (h UserHandler) Register(ctx context.Context, in *proto.RegiserRequest) (*proto.RegisterResponse, error) {
	var res proto.RegisterResponse

	if in.Login == "" || in.Password == "" {
		res.Error = "логин или пароль некорректны"
		return &res, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("ошибка получения хэша пароля")
		res.Error = "внутренняя ошибка сервера"
		return &res, nil
	}

	user, err := h.Svc.CreateUser(in.Login, string(hash))
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("ошибка создания пользователя")

		res.Error = "ошибка создания пользователя"
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			res.Error = "такой пользователь уже существует"
		}

		return &res, nil
	}

	token, err := getJWT(h.JWTkey, user.ID, user.Login)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("ошибка создания jwt токена")
		res.Error = "ошибка создания jwt токена"
		return &res, nil
	}

	res.Jwt = *token

	return &res, nil
}

// Login обрабатывает GRPC вызов входа пользователя. Проверяет учетные данные
// пользователя с помощью `UserService`. Если учетные данные верны, генерирует
// JWT токен для пользователя. Ошибки во время проверки или генерации токена
// логируются и возвращаются как ответы с ошибками.
func (h UserHandler) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	var res proto.LoginResponse
	user, err := h.Svc.FindUserByLogin(in.Login)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("ошибка получения пользователя")
		res.Error = "ошибка получения пользователя"
		return &res, nil
	}

	if user == nil {
		res.Error = "пользователь не найден"
		return &res, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(in.Password)); err != nil {
		res.Error = "логин или пароль некорректны"
		//nolint:nilerr // Это легальный возврат
		return &res, nil
	}

	token, err := getJWT(h.JWTkey, user.ID, user.Login)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("ошибка создания jwt токена")
		res.Error = "ошибка создания jwt токена"

		return &res, nil
	}

	res.Jwt = *token

	return &res, nil
}

// getJWT генерирует JWT токен для указанного ID пользователя и логина,
// используя предоставленный JWT ключ. Токен включает ID пользователя,
// логин и время истечения (по умолчанию 30 минут). Если генерация токена
// не удалась, возвращает ошибку.
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
		return nil, fmt.Errorf("ошибка подписи jwt: %w", err)
	}

	return &tokenString, nil
}
