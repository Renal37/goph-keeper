package repository

import (
	"github.com/Renal37/goph-keeper/internal/server/core/domain"
)

// FindUserByLogin находит пользователя по его логину. Использует метод ORM `First`
// для поиска пользователя с указанным логином в базе данных. Если пользователь
// не найден, возвращает `nil` для пользователя и ошибки. Если возникает ошибка
// во время операции с базой данных, возвращает `nil` для пользователя и ошибку.
func (s *DB) FindUserByLogin(login string) (*domain.User, error) {
	user := domain.User{}

	req := s.db.First(&user, "login = ?", login)
	if req.RowsAffected == 0 {
		//nolint:nilnil // Это допустимый возврат
		return nil, nil
	}

	if req.Error != nil {
		return nil, req.Error
	}

	return &user, nil
}

// CreateUser создает нового пользователя с указанным логином и хешированным паролем.
// Использует метод ORM `Create` для добавления нового пользователя в базу данных.
// Если возникает ошибка во время операции с базой данных, возвращает `nil` для
// пользователя и ошибку. Если успешно, возвращает указатель на созданного пользователя.
func (s *DB) CreateUser(login, hash string) (*domain.User, error) {
	user := domain.User{
		Login: login,
		Hash:  hash,
	}

	req := s.db.Create(&user)
	if req.Error != nil {
		return nil, req.Error
	}

	return &user, nil
}
