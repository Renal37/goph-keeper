// Package services contains the application services that implement
// business logic using the repository interfaces defined in the
// `ports` package. These services serve as an intermediary layer
// between the domain logic and the data layer, providing methods
// for operations such as finding, creating, updating, and deleting
// users and storage records.
//
//nolint:wrapcheck // This legal return
package services

import (
	"github.com/Renal37/goph-keeper/internal/server/core/domain"
	"github.com/Renal37/goph-keeper/internal/server/core/ports"
)

// UserService represents a service for user-related operations.
// It utilizes the `UserRepository` interface to interact with
// the data layer and perform business logic related to users.
type UserService struct {
	repo ports.UserRepository
}

// NewUserService creates a new instance of `UserService` with the given `UserRepository`.
func NewUserService(repo ports.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// FindUserByLogin retrieves a user by their login.
// It uses the `FindUserByLogin` method from the `UserRepository` interface.
func (u *UserService) FindUserByLogin(login string) (*domain.User, error) {
	return u.repo.FindUserByLogin(login)
}

// CreateUser creates a new user with the given login and hashed password.
// It uses the `CreateUser` method from the `UserRepository` interface.
func (u *UserService) CreateUser(login, hash string) (*domain.User, error) {
	return u.repo.CreateUser(login, hash)
}
