// Package repository contains the data access layer for the application,
// providing functions to interact with the database and perform operations
// related to the domain entities such as `User` and `Storage`. This package
// serves as an interface between the application services and the database,
// utilizing an ORM (such as GORM) to execute queries and manage transactions.
package repository

import (
	"github.com/Renal37/goph-keeper/internal/server/core/domain"
)

// FindUserByLogin retrieves a user by their login. It uses the ORM `First` method
// to find a user with the specified login in the database. If the user is not found,
// it returns `nil` for both the user and error. If an error occurs during the
// database operation, it returns `nil` for the user and the error.
func (s *DB) FindUserByLogin(login string) (*domain.User, error) {
	user := domain.User{}

	req := s.db.First(&user, "login = ?", login)
	if req.RowsAffected == 0 {
		//nolint:nilnil // This legal return
		return nil, nil
	}

	if req.Error != nil {
		return nil, req.Error
	}

	return &user, nil
}

// CreateUser creates a new user with the given login and hashed password.
// It uses the ORM `Create` method to add the new user to the database.
// If an error occurs during the database operation, it returns `nil` for
// the user and the error. If successful, it returns a pointer to the created user.
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
