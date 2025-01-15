package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Renal37/goph-keeper/internal/logger"
	repository "github.com/Renal37/goph-keeper/internal/server/adapters/repository/pg"
	"github.com/Renal37/goph-keeper/internal/server/config"
	"github.com/Renal37/goph-keeper/internal/server/core"
)

var (
	// Минимальное количество символов для мастер-ключа
	minimumCharMasterKey = 16
	// Версия сборки, по умолчанию "N/A"
	buildVersion string = "N/A"
	// Дата сборки, по умолчанию "N/A"
	buildDate string = "N/A"
)

func main() {
	// Получаем конфигурацию из окружения
	eCfg, err := config.GetConfig()
	if err != nil {
		log.Fatalln("Ошибка при загрузке конфигурации:", err)
	}

	// Инициализация логгера
	lg, err := logger.Init("info")
	if err != nil {
		log.Fatalln("Ошибка при инициализации логгера:", err)
	}

	// Логируем информацию о версии и дате сборки
	lg.Info(fmt.Sprintf("Версия сборки: %v", buildVersion))
	lg.Info(fmt.Sprintf("Дата сборки: %v", buildDate))

	// Проверка наличия мастер-ключа
	if eCfg.MasterKey == "" {
		lg.Fatal("Мастер-ключ не найден! Пожалуйста, используйте флаг -mk")
	}

	// Проверка длины мастер-ключа
	if len(eCfg.MasterKey) < minimumCharMasterKey {
		lg.Sugar().Fatalf("Минимальная длина мастер-ключа должна быть %v символов!", minimumCharMasterKey)
	}

	// Инициализация подключения к базе данных
	repo, err := repository.NewDB(context.Background(), lg, eCfg.DSN)
	if err != nil {
		lg.Fatal("Ошибка при подключении к базе данных: " + err.Error())
	}

	// Запуск GRPC сервера
	err = core.RunGRPCserver(lg, eCfg.Host, eCfg.CertificatePath, eCfg.CertificateKeyPath, eCfg.JWTkey, eCfg.MasterKey, repo)
	if err != nil {
		lg.Fatal("Ошибка при запуске GRPC сервера: " + err.Error())
	}
}
