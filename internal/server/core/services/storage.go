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

// StorageService represents a service for storage-related operations.
// It uses the `StorageRepository` interface to interact with the
// storage data layer and perform business logic related to storage.
type StorageService struct {
	repo ports.StorageRepository
}

// NewStorageService creates a new instance of `StorageService`
// with the given `StorageRepository`.
func NewStorageService(repo ports.StorageRepository) *StorageService {
	return &StorageService{
		repo: repo,
	}
}

// ReadAllRecord retrieves all storage records for the specified owner.
// It uses the `ReadAllRecord` method from the `StorageRepository` interface.
func (s *StorageService) ReadAllRecord(owner int) ([]*domain.Storage, error) {
	return s.repo.ReadAllRecord(owner)
}

// ReadRecord retrieves a specific storage record by ID and owner.
// It uses the `ReadRecord` method from the `StorageRepository` interface.
func (s *StorageService) ReadRecord(id int, owner int) (*domain.Storage, error) {
	return s.repo.ReadRecord(id, owner)
}

// WriteRecord adds a new storage record.
// It uses the `WriteRecord` method from the `StorageRepository` interface.
func (s *StorageService) WriteRecord(doc domain.Storage) error {
	return s.repo.WriteRecord(doc)
}

// DeleteRecord removes a storage record by ID and owner.
// It uses the `DeleteRecord` method from the `StorageRepository` interface.
func (s *StorageService) DeleteRecord(id int, owner int) error {
	return s.repo.DeleteRecord(id, owner)
}
