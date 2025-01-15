// Package logger contains a function for logging based on the Zap logging library.
package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Init initializes the logger.
func Init(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed parse error level %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed build zap config %w", err)
	}

	return zl, nil
}
