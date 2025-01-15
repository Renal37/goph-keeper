package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Init инициализирует логгер с указанным уровнем логирования
func Init(level string) (*zap.Logger, error) {
	// Парсим уровень логирования
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге уровня логирования: %w", err)
	}

	// Создаем конфигурацию для продакшн окружения
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	// Создаем логгер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании конфигурации zap: %w", err)
	}

	return zl, nil
}
