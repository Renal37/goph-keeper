// Package ports defines the interfaces for repository ports that serve as
// the boundary between the application's core domain logic and the data
// layer. These interfaces outline the methods for interacting with the
// storage repositories for `User` and `Storage` domain entities.
package ports

import "github.com/Renal37/goph-keeper/internal/server/core/domain"

// UserRepository represents the interface for user-related data storage.
// It provides methods for finding a user by login and creating a new user.
type UserRepository interface {
	FindUserByLogin(login string) (*domain.User, error)
	CreateUser(login, hash string) (*domain.User, error)
}

// StorageRepository represents the interface for storage-related data storage.
// It provides methods for reading, writing, and deleting storage records.
type StorageRepository interface {
	ReadRecord(id int, owner int) (*domain.Storage, error)
	ReadAllRecord(owner int) ([]*domain.Storage, error)
	WriteRecord(doc domain.Storage) error
	DeleteRecord(id int, owner int) error
}
