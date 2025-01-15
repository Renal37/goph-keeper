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

// NewDB инициализирует новую сессию базы данных, используя предоставленный DSN (имя источника данных).
// Он подключается к базе данных PostgreSQL с использованием GORM и настраивает логгер для работы в тихом режиме.
// Если подключение успешно, он продолжает миграцию схемы с использованием
// AutoMigrate для моделей домена `User` и `Storage`. Если возникает ошибка во время
// инициализации или миграции, возвращается ошибка вместе с частично инициализированным экземпляром `DB`.
func NewDB(ctx context.Context, lg *zap.Logger, dsn string) (*DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return &DB{}, fmt.Errorf("не удалось инициализировать сессию базы данных: %w", err)
	}

	// Миграция схемы
	err = db.AutoMigrate(&domain.User{}, &domain.Storage{})
	if err != nil {
		return &DB{}, fmt.Errorf("не удалось мигрировать модели: %w", err)
	}

	lg.Info("Подключение к PostgreSQL: успешно")

	return &DB{
		db: db,
	}, nil
}

// Close закрывает соединение с базой данных.
func (s DB) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("не удалось получить SQL DB: %w", err)
	}

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("не удалось закрыть базу данных: %w", err)
	}

	return nil
}
