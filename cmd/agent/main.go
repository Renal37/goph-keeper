package main

import (
	"fmt"
	"log"

	"github.com/Renal37/goph-keeper/internal/agent/client"
	"github.com/Renal37/goph-keeper/internal/agent/config"
	"github.com/Renal37/goph-keeper/internal/agent/core"
	"github.com/Renal37/goph-keeper/internal/logger"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
)

func main() {
	// Получение конфигурации
	eCfg, err := config.GetConfig()
	if err != nil {
		log.Fatalln(err)
	}

	// Инициализация логгера
	lg, err := logger.Init("info")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("*************************************")
	fmt.Println("Добро пожаловать в GophKeeper клиент")
	fmt.Printf("Версия сборки: %v \n", buildVersion)
	fmt.Printf("Дата сборки: %v \n", buildDate)
	fmt.Println("*************************************")

	// Вывод справки по доступным командам
	if eCfg.Command == "" {
		fmt.Println("Поддерживаемые команды -c:")
		fmt.Println("sign-up - создать новый аккаунт")
		fmt.Println("sign-in - войти в существующий аккаунт")
		fmt.Println("read-file - прочитать все файлы в вашем аккаунте")
		fmt.Println("write-file - записать файл в ваш аккаунт")
		fmt.Println("delete-file - удалить файл из вашего аккаунта")
		fmt.Println("*************************************")
	}

	// Создание нового клиента
	cl, err := client.NewClient(eCfg.ServerAddr, eCfg.Certificate, eCfg.JWT)
	if err != nil {
		lg.Sugar().Fatalf("ошибка создания клиента: %s", err.Error())
	}

	// Выполнение команды
	err = core.Run(cl, eCfg.Command)
	if err != nil {
		lg.Sugar().Fatalf("ошибка выполнения команды клиента: %s", err.Error())
	}

	// Закрытие клиента
	err = cl.Close()
	if err != nil {
		lg.Sugar().Fatalf("ошибка закрытия клиента: %s", err.Error())
	}
}
