package core

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Renal37/goph-keeper/internal/agent/client"
)

// Права доступа по умолчанию для файлов
var defaultPermition fs.FileMode = 0600

// Шаблон ошибки чтения из STDIN
var errorFailedReadSTDIN = "ошибка чтения из STDIN: %w"

// Run выполняет основную логику приложения в зависимости от команды
func Run(client *client.Client, command string) error {
	// В зависимости от команды выбираем логику поведения
	switch command {
	case "sign-up":
		fmt.Println("-> Создание нового аккаунта")

		// Получаем учетные данные пользователя из stdin
		ss, err := getUserCredentials()
		if err != nil {
			return fmt.Errorf("ошибка получения учетных данных: %w", err)
		}

		r, err := client.Register(ss.login, ss.password)
		if err != nil {
			return fmt.Errorf("ошибка регистрации пользователя: %w", err)
		}

		fmt.Printf("Токен: %s \n", r.Jwt)

		// Сохраняем токен
		err = saveAuthToken(r.Jwt)
		if err != nil {
			return fmt.Errorf("ошибка сохранения токена: %w", err)
		}
	case "sign-in":
		fmt.Println("-> Вход в аккаунт")

		ss, err := getUserCredentials()
		if err != nil {
			return fmt.Errorf("ошибка получения учетных данных: %w", err)
		}

		r, err := client.Login(ss.login, ss.password)
		if err != nil {
			return fmt.Errorf("ошибка входа пользователя: %w", err)
		}

		fmt.Printf("Токен: %s \n", r.Jwt)
		err = saveAuthToken(r.Jwt)
		if err != nil {
			return fmt.Errorf("ошибка сохранения токена: %w", err)
		}
	case "read-file":
		fmt.Println("-> Чтение файла")

		// Запрос на чтение всех файлов
		rAllFile, err := client.ReadAllFile()
		if err != nil {
			return fmt.Errorf("ошибка получения списка файлов: %w", err)
		}

		// Если файлов нет, выходим
		if len(rAllFile.Units) == 0 {
			fmt.Println("Файлы не найдены. До свидания!")
			return nil
		}

		// Показываем доступные файлы
		fmt.Println("Доступные файлы:")
		for _, v := range rAllFile.Units {
			if v.Id > 0 {
				fmt.Printf("[%v] - %s \n", v.Id, v.Name)
			}
		}

		// Выбор файла для загрузки
		i, err := selectReadFile()
		if err != nil {
			return fmt.Errorf("неверный ID файла: %w", err)
		}

		// Запрос на чтение файла
		rFile, err := client.ReadFile(int32(i))
		if err != nil {
			return fmt.Errorf("ошибка получения файла: %w", err)
		}

		// Если тип файла - файл
		if rFile.Type == "file" {
			err = saveFileInDisk(rFile.Name, rFile.Data)
			if err != nil {
				return fmt.Errorf("ошибка сохранения файла: %w", err)
			}
		} else {
			// Иначе тип - текст
			fmt.Println(string(rFile.Data))
		}
	case "write-file":
		fmt.Println("-> Запись файла")

		// Выбор типа файла и файла для сохранения
		err := selectWriteData(client)
		if err != nil {
			return fmt.Errorf("ошибка выбора данных для записи: %w", err)
		}
	case "delete-file":
		fmt.Println("-> Удаление файла")

		// Запрос на чтение всех файлов
		rAllFile, err := client.ReadAllFile()
		if err != nil {
			return fmt.Errorf("ошибка получения списка файлов: %w", err)
		}

		// Если файлов нет, выходим
		if len(rAllFile.Units) == 0 {
			fmt.Println("Файлы не найдены. До свидания!")
			return nil
		}

		// Показываем доступные файлы
		fmt.Println("Доступные файлы:")
		for _, v := range rAllFile.Units {
			if v.Id > 0 {
				fmt.Printf("[%v] - %s \n", v.Id, v.Name)
			}
		}

		// Выбор файла для удаления
		i, err := selectReadFile()
		if err != nil {
			return fmt.Errorf("неверный ID файла: %w", err)
		}

		// Запрос на удаление
		_, err = client.DeleteFile(int32(i))
		if err != nil {
			return fmt.Errorf("ошибка удаления файла: %w", err)
		}

		fmt.Println("Файл удален!")
	default:
		fmt.Printf("Команда:%s не найдена! \n", command)
	}

	fmt.Println("До свидания!")
	return nil
}

// УТИЛИТЫ ДЛЯ РАБОТЫ С ФАЙЛАМИ

// saveFileInDisk сохраняет файлы на диск
func saveFileInDisk(fileName string, data []byte) error {
	fmt.Println("Куда вы хотите сохранить файл?")
	fmt.Print("Введите путь к директории: ")

	reader := bufio.NewReader(os.Stdin)

	r, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf(errorFailedReadSTDIN, err)
	}

	dirPath := strings.TrimSpace(r)
	fullPath := filepath.Join(dirPath, fileName)

	err = os.WriteFile(fullPath, data, defaultPermition)
	if err != nil {
		return fmt.Errorf("ошибка записи данных: %w", err)
	}

	fmt.Printf("Файл сохранен в: %s \n", fullPath)

	return nil
}

// selectWriteData выбор файла для загрузки
func selectWriteData(client *client.Client) error {
	fmt.Println("Что вы хотите отправить на сервер?")
	fmt.Println("[1] - Текст")
	fmt.Println("[2] - Файл")
	fmt.Print("Введите номер: ")

	reader := bufio.NewReader(os.Stdin)

	r, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf(errorFailedReadSTDIN, err)
	}

	r = strings.TrimSpace(r)

	i, err := strconv.Atoi(r)
	if err != nil {
		return fmt.Errorf("ошибка преобразования в число: %w", err)
	}

	switch i {
	case 1:
		fmt.Println("Что вы хотите сохранить?")
		fmt.Println("[1] - Произвольный текст")
		fmt.Println("[2] - Логин | Пароль")
		fmt.Println("[3] - Кредитная карта")
		fmt.Print("Введите номер: ")

		r, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf(errorFailedReadSTDIN, err)
		}

		r = strings.TrimSpace(r)

		i, err := strconv.Atoi(r)
		if err != nil {
			return fmt.Errorf("ошибка преобразования в число: %w", err)
		}

		fmt.Print("Введите имя: ")

		fileName, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf(errorFailedReadSTDIN, err)
		}

		fileName = strings.TrimSpace(fileName)

		switch i {
		case 1:
			fmt.Println("Введите текст:")
		case 2:
			fmt.Println("Введите логин и пароль:")
		case 3:
			fmt.Println("Введите номер, имя, дату и CVV:")
		}

		data, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf(errorFailedReadSTDIN, err)
		}

		data = strings.TrimSpace(data)

		// Отправляем данные через gRPC
		_, err = client.WriteFile("text", fileName, data)
		if err != nil {
			return fmt.Errorf("ошибка записи файла: %w", err)
		}

	case 2:
		fmt.Print("Введите путь к файлу: ")

		filePath, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf(errorFailedReadSTDIN, err)
		}

		filePath = strings.TrimSpace(filePath)

		// Получаем имя файла
		baseName := filepath.Base(filePath)

		// Отправляем данные через gRPC
		_, err = client.WriteFile("file", baseName, filePath)
		if err != nil {
			return fmt.Errorf("ошибка записи файла: %w", err)
		}
	}

	fmt.Println("Файл записан!")

	return nil
}

// УТИЛИТЫ ДЛЯ ЧТЕНИЯ ФАЙЛОВ

// selectReadFile выбор файла для чтения
func selectReadFile() (int, error) {
	fmt.Print("Выберите ID файла: ")

	reader := bufio.NewReader(os.Stdin)

	response, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения из stdin: %w", err)
	}

	response = strings.TrimSpace(response)

	i, err := strconv.Atoi(response)
	if err != nil {
		return 0, fmt.Errorf("ошибка преобразования в число: %w", err)
	}

	return i, nil
}

// УТИЛИТЫ ДЛЯ РЕГИСТРАЦИИ И ВХОДА

// saveAuthToken сохранение токена в файл .env
func saveAuthToken(token string) error {
	fmt.Print("Хотите сохранить токен в .env? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)

	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf(errorFailedReadSTDIN, err)
	}

	response = strings.TrimSpace(response)

	if strings.ToLower(response) == "y" {
		file, err := os.OpenFile(".env", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultPermition)
		if err != nil {
			return fmt.Errorf("ошибка открытия файла .env: %w", err)
		}

		_, err = file.WriteString(fmt.Sprintf("JWT=%s\n", token))
		if err != nil {
			return fmt.Errorf("ошибка записи токена в файл .env: %w", err)
		}

		err = file.Close()
		if err != nil {
			return fmt.Errorf("ошибка закрытия файла: %w", err)
		}

		fmt.Println("Токен сохранен в файле .env.")
	}

	return nil
}

// userCredentials структура для хранения учетных данных пользователя
type userCredentials struct {
	login    string
	password string
}

// getUserCredentials получение пары логин/пароль от пользователя
func getUserCredentials() (userCredentials, error) {
	fmt.Print("Введите ваш логин: ")

	reader := bufio.NewReader(os.Stdin)

	loginResp, err := reader.ReadString('\n')
	if err != nil {
		return userCredentials{}, fmt.Errorf("ошибка чтения логина из stdin: %w", err)
	}

	fmt.Print("Введите ваш пароль: ")
	passwordResp, err := reader.ReadString('\n')
	if err != nil {
		return userCredentials{}, fmt.Errorf("ошибка чтения пароля из stdin: %w", err)
	}

	return userCredentials{
		login:    strings.TrimSpace(loginResp),
		password: strings.TrimSpace(passwordResp),
	}, nil
}
