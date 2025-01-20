// Package repository contains the data access layer for the application,
// providing functions to interact with the database and perform operations
// related to the domain entities such as `User` and `Storage`. This package
// serves as an interface between the application services and the database,
// utilizing an ORM (such as GORM) to execute queries and manage transactions.
package repository

import (
	"github.com/Renal37/goph-keeper/internal/server/core/domain"
)

// ReadAllRecord retrieves all storage records for a specific owner.
// It uses the `Find` method to query the database for storage records
// that match the specified owner. If no records are found, it returns
// nil for both the slice of records and the error. If an error occurs
// during the query, it returns the error.
func (s *DB) ReadAllRecord(owner int) ([]*domain.Storage, error) {
	docs := []*domain.Storage{}

	req := s.db.Select("id", "name", "owner").Find(&docs, "owner = ?", owner)
	if req.RowsAffected == 0 {
		return nil, nil
	}

	if req.Error != nil {
		return nil, req.Error
	}

	return docs, nil
}

// ReadRecord retrieves a specific storage record by its ID and owner.
// It uses the `First` method to query the database for a storage record
// that matches the specified ID and owner. If no record is found, it returns
// nil for both the record and the error. If an error occurs during the query,
// it returns the error.
func (s *DB) ReadRecord(id int, owner int) (*domain.Storage, error) {
	doc := domain.Storage{}

	req := s.db.First(&doc, "id = ? AND owner = ?", id, owner)
	if req.RowsAffected == 0 {
		//nolint:nilnil // This legal return
		return nil, nil
	}

	if req.Error != nil {
		return nil, req.Error
	}

	return &doc, nil
}

// WriteRecord adds a new storage record to the database.
// It uses the `Create` method to insert the record. If an error occurs
// during the insertion, it returns the error.
func (s *DB) WriteRecord(doc domain.Storage) error {
	req := s.db.Create(&doc)
	if req.Error != nil {
		return req.Error
	}

	return nil
}

// DeleteRecord removes a storage record from the database by its ID and owner.
// It uses the `Delete` method to remove the record. If an error occurs during
// the deletion, it returns the error.
func (s *DB) DeleteRecord(id int, owner int) error {
	doc := domain.Storage{}

	req := s.db.Delete(&doc, "id = ? AND owner = ?", id, owner)
	if req.Error != nil {
		return req.Error
	}

	return nil
}
