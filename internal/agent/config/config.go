package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"

	env "github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

// Права доступа по умолчанию для файлов
var defaultPermition fs.FileMode = 0600

// ConfigENV содержит настройки приложения
type ConfigENV struct {
	Command     string // Команда для хранилища GophKeeper
	JWT         string `env:"JWT"`                            // JWT токен для авторизации
	ServerAddr  string `json:"server_addr" env:"SERVER_ADDR"` // Адрес сервера
	Certificate string `json:"certificate"`                   // Путь к сертификату
}

// GetConfig получает настройки приложения из конфигурационных файлов и переменных окружения
func GetConfig() (*ConfigENV, error) {
	var eCfg ConfigENV
	configPath := "config/agent.json"

	// Парсим флаги командной строки
	flag.StringVar(&eCfg.Command, "c", "", "команда для хранилища GophKeeper")
	flag.Parse()

	// Открываем конфигурационный файл
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия конфигурационного файла: %w", err)
	}

	// Декодируем JSON конфигурацию
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&eCfg); err != nil {
		return nil, fmt.Errorf("ошибка декодирования конфигурационного файла: %w", err)
	}

	// Закрываем файл конфигурации
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("ошибка закрытия конфигурационного файла: %w", err)
	}

	// Создаем .env файл, если он не существует
	file, err = os.OpenFile(".env", os.O_CREATE, defaultPermition)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания файла .env: %w", err)
	}
	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("ошибка закрытия файла .env: %w", err)
	}

	// Загружаем переменные окружения из .env файла
	err = godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки файла .env: %w", err)
	}

	// Парсим переменные окружения в структуру
	err = env.Parse(&eCfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга переменных окружения: %w", err)
	}

	return &eCfg, nil
}
