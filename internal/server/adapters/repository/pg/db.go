// Package repository contains the data access layer for the application,
// providing functions to interact with the database and perform operations
// related to the domain entities such as `User` and `Storage`. This package
// serves as an interface between the application services and the database,
// utilizing an ORM (such as GORM) to execute queries and manage transactions.
package repository

import (
	"context"
	"fmt"

	"github.com/Renal37/goph-keeper/internal/server/core/domain"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	db *gorm.DB
}

// NewDB initializes a new database session using the given DSN (Data Source Name).
// It connects to the PostgreSQL database using GORM and configures the logger to operate in silent mode.
// If the connection is successful, it proceeds to migrate the schema using
// AutoMigrate for the `User` and `Storage` domain models. If an error occurs during
// initialization or migration, an error is returned along with a partially initialized `DB` instance.
func NewDB(ctx context.Context, lg *zap.Logger, dsn string) (*DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return &DB{}, fmt.Errorf("failed init db session: %w", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&domain.User{}, &domain.Storage{})
	if err != nil {
		return &DB{}, fmt.Errorf("failed migrate models: %w", err)
	}

	lg.Info(("Connection to postgre: success"))

	return &DB{
		db: db,
	}, nil
}

// Close close database connection.
func (s DB) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed get sql db: %w", err)
	}

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("failed close db: %w", err)
	}

	return nil
}
