package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	env "github.com/caarlos0/env/v6"
)

// ConfigENV содержит настройки приложения.
type ConfigENV struct {
	JWTkey             string `json:"jwt_key" env:"JWT_KEY"`
	Host               string `json:"host" env:"HOST"`
	DSN                string `json:"dsn" env:"DSN"`
	CertificatePath    string `json:"certificate"`
	CertificateKeyPath string `json:"certificate_key"`
	MasterKey          string
}

// GetConfig получает настройки приложения.
func GetConfig() (*ConfigENV, error) {
	var eCfg ConfigENV
	configPath := "config/server.json"

	flag.StringVar(&eCfg.MasterKey, "mk", "", "мастер-ключ для ключей шифрования")
	flag.Parse()

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл конфигурации: %w", err)
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&eCfg); err != nil {
		return nil, fmt.Errorf("не удалось декодировать файл конфигурации: %w", err)
	}

	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("не удалось закрыть файл конфигурации: %w", err)
	}

	err = env.Parse(&eCfg)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать переменные окружения: %w", err)
	}

	return &eCfg, nil
}
