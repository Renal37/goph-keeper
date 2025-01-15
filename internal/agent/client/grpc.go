package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Renal37/goph-keeper/internal/server/core/domain/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Максимальный размер сообщения
var maxMsgSize = 100000648
// Шаблоны сообщений об ошибках
var errorResponseFinished = "ошибка завершения ответа: %w"
var errorEesponseReturn = "ошибка возврата ответа: %w"

// Client представляет GRPC клиент
type Client struct {
	Conn  *grpc.ClientConn // Соединение GRPC
	Token string           // Токен авторизации
}

// NewClient создает новый экземпляр GRPC клиента
func NewClient(addr string, certPath string, token string) (*Client, error) {
	// Получаем TLS сертификат
	tlsCredentials, err := loadTLSCredentials(certPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить TLS сертификат: %w", err)
	}

	// Подключаемся к GRPC серверу
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize), grpc.MaxCallSendMsgSize(maxMsgSize)),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка запуска GRPC сервера: %w", err)
	}

	return &Client{
		Conn:  conn,
		Token: token,
	}, nil
}

// Close закрывает соединение с сервером
func (c Client) Close() error {
	err := c.Conn.Close()
	if err != nil {
		return fmt.Errorf("ошибка закрытия GRPC клиента: %w", err)
	}
	return nil
}

// Register регистрирует нового пользователя
func (c Client) Register(login string, password string) (*proto.RegisterResponse, error) {
	// Создаем клиент
	client := proto.NewUserClient(c.Conn)
	resp, err := client.Register(context.Background(), &proto.RegiserRequest{
		Login:    login,
		Password: password,
	})

	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
	}

	return resp, nil
}

// Login выполняет вход пользователя
func (c Client) Login(login string, password string) (*proto.LoginResponse, error) {
	// Создаем клиент
	client := proto.NewUserClient(c.Conn)
	resp, err := client.Login(context.Background(), &proto.LoginRequest{
		Login:    login,
		Password: password,
	})

	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
	}

	return resp, nil
}

// ReadAllFile читает все записи
func (c Client) ReadAllFile() (*proto.ReadAllRecordResponse, error) {
	// Устанавливаем авторизацию в метаданных GRPC
	md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", c.Token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Создаем клиент
	client := proto.NewStorageClient(c.Conn)
	resp, err := client.ReadAllRecord(ctx, &proto.ReadAllRecordRequest{})

	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
	}

	return resp, nil
}

// ReadFile читает одну запись по ID
func (c Client) ReadFile(id int32) (*proto.ReadRecordResponse, error) {
	// Устанавливаем авторизацию в метаданных GRPC
	md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", c.Token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Создаем клиент
	client := proto.NewStorageClient(c.Conn)
	resp, err := client.ReadRecord(ctx, &proto.ReadRecordRequest{
		Id: id,
	})

	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
	}

	return resp, nil
}

// WriteFile записывает данные (текст или файл)
func (c Client) WriteFile(typ string, name string, data string) (*proto.WriteRecordResponse, error) {
	// Устанавливаем авторизацию в метаданных GRPC
	md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", c.Token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Создаем клиент
	client := proto.NewStorageClient(c.Conn)
	stream, err := client.WriteRecord(ctx)
	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}

	var resp *proto.WriteRecordResponse
	switch typ {
	case "text":
		// Отправляем данные через GRPC
		err = stream.Send(&proto.WriteRecordRequest{Name: name, Data: []byte(data), Type: "text"})
		if err != nil {
			return nil, fmt.Errorf("ошибка отправки потока: %w", err)
		}

		// Закрываем поток и получаем ответ
		resp, err = stream.CloseAndRecv()
		if err != nil {
			return nil, fmt.Errorf("ошибка закрытия потока: %w", err)
		}
		if resp.Error != "" {
			return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
		}
	case "file":
		file, err := os.Open(data)
		if err != nil {
			return nil, fmt.Errorf("ошибка открытия файла: %w", err)
		}

		fi, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения информации о файле: %w", err)
		}

		if fi.Size() > int64(maxMsgSize) {
			return nil, fmt.Errorf("максимальный размер файла должен быть меньше: %v байт", maxMsgSize)
		}

		// Читаем файл частями и отправляем
		chunkSize := 4096
		buf := make([]byte, chunkSize)
		for {
			n, err := file.Read(buf)
			if errors.Is(err, io.EOF) {
				// Конец файла, закрываем поток
				resp, err = stream.CloseAndRecv()
				if err != nil {
					return nil, fmt.Errorf("ошибка CloseAndRecv: %w", err)
				}
				if resp.Error != "" {
					return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
				}
				break
			}
			if err != nil {
				return nil, fmt.Errorf("ошибка чтения файла: %w", err)
			}

			// Отправляем часть данных
			err = stream.Send(&proto.WriteRecordRequest{Name: name, Data: buf[:n], Type: "file"})
			if err != nil {
				return nil, fmt.Errorf("ошибка отправки потока: %w", err)
			}
		}

		err = file.Close()
		if err != nil {
			return nil, fmt.Errorf("ошибка закрытия файла: %w", err)
		}
	}

	return resp, nil
}

// DeleteFile удаляет запись по ID
func (c Client) DeleteFile(id int32) (*proto.DeleteRecordResponse, error) {
	// Устанавливаем авторизацию в метаданных GRPC
	md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", c.Token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Создаем клиент
	client := proto.NewStorageClient(c.Conn)
	resp, err := client.DeleteRecord(ctx, &proto.DeleteRecordRequest{
		Id: id,
	})

	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
	}

	return resp, nil
}

// loadTLSCredentials загружает сертификаты
func loadTLSCredentials(cert string) (credentials.TransportCredentials, error) {
	// Загружаем сертификат CA, который подписал сертификат сервера
	pemServerCA, err := os.ReadFile(cert)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки файла: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("не удалось добавить сертификат CA сервера")
	}

	// Создаем и возвращаем учетные данные
	config := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	return credentials.NewTLS(config), nil
}