package handler

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Renal37/goph-keeper/internal/server/adapters/middleware"
	"github.com/Renal37/goph-keeper/internal/server/core/domain"
	"github.com/Renal37/goph-keeper/internal/server/core/domain/proto"
	"github.com/Renal37/goph-keeper/internal/server/core/services"
	"go.uber.org/zap"
)

// StorageHandler обработчик для хранилища
type StorageHandler struct {
	proto.UnimplementedStorageServer
	Svc       services.StorageService
	Logger    *zap.Logger
	MasterKey string
}

var errorInvalidToken = "недействительный токен"
var errorCloseStream = "ошибка закрытия потока: %w"

// ReadAllRecord читает все записи из БД
func (s StorageHandler) ReadAllRecord(ctx context.Context, in *proto.ReadAllRecordRequest) (*proto.ReadAllRecordResponse, error) {
	var resp proto.ReadAllRecordResponse

	// Получаем токен из контекста
	token, ok := middleware.GetTokenFromContext(ctx)
	if !ok {
		s.Logger.Error(errorInvalidToken)
		resp.Error = errorInvalidToken
		return &resp, nil
	}

	// Получаем данные из БД
	rec, err := s.Svc.ReadAllRecord(token.ID)
	if err != nil {
		s.Logger.With(zap.Error(err)).Error("ошибка получения всех записей")
		resp.Error = "ошибка получения всех записей"
		return &resp, nil
	}

	// Подготовка ответа
	respSlice := make([]*proto.StorageUnit, 0, len(rec))
	for _, v := range rec {
		respSlice = append(respSlice, &proto.StorageUnit{
			Id:    int32(v.ID),
			Name:  v.Name,
			Type:  v.Type,
			Owner: int32(v.Owner),
		})
	}

	resp.Units = respSlice
	return &resp, nil
}

// ReadRecord читает одну запись из БД
func (s StorageHandler) ReadRecord(ctx context.Context, in *proto.ReadRecordRequest) (*proto.ReadRecordResponse, error) {
	var resp proto.ReadRecordResponse

	// Получаем токен из контекста
	token, ok := middleware.GetTokenFromContext(ctx)
	if !ok {
		s.Logger.Error(errorInvalidToken)
		resp.Error = errorInvalidToken
		return &resp, nil
	}

	// Получаем запись из БД
	rec, err := s.Svc.ReadRecord(int(in.Id), token.ID)
	if err != nil {
		s.Logger.With(zap.Error(err)).Error("ошибка чтения записи")
		resp.Error = "ошибка чтения записи"
		return &resp, nil
	}

	if rec == nil {
		resp.Error = "запись не найдена"
		return &resp, nil
	}

	// Расшифровка данных
	data, err := decryptionData(s.MasterKey, rec.Key, rec.Value)
	if err != nil {
		s.Logger.With(zap.Error(err)).Error("ошибка расшифровки данных")
		resp.Error = "ошибка расшифровки данных"
		return &resp, nil
	}

	resp.Name = rec.Name
	resp.Type = rec.Type
	resp.Data = data

	return &resp, nil
}

// WriteRecord записывает данные в БД
func (s StorageHandler) WriteRecord(stream proto.Storage_WriteRecordServer) error {
	var resp proto.WriteRecordResponse
	var fileName string
	var fileType string

	// Для чанков
	buffer := &bytes.Buffer{}

	// Получаем токен из контекста
	token, ok := middleware.GetTokenFromContext(stream.Context())
	if !ok {
		s.Logger.Error(errorInvalidToken)
		resp.Error = errorInvalidToken

		err := stream.SendAndClose(&resp)
		if err != nil {
			return fmt.Errorf(errorCloseStream, err)
		}
	}

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			s.Logger.With(zap.Error(err)).Error("ошибка получения чанка")
			resp.Error = "ошибка получения чанка"

			err := stream.SendAndClose(&resp)
			if err != nil {
				return fmt.Errorf(errorCloseStream, err)
			}
		}

		// Сохраняем имя файла из запроса
		if fileName == "" {
			fileName = chunk.GetName()
		}

		if fileType == "" {
			fileType = chunk.GetType()
		}

		// Записываем данные в буфер
		if _, err := buffer.Write(chunk.GetData()); err != nil {
			s.Logger.With(zap.Error(err)).Error("ошибка записи чанка в буфер")
			resp.Error = "ошибка записи чанка в буфер"

			err := stream.SendAndClose(&resp)
			if err != nil {
				return fmt.Errorf(errorCloseStream, err)
			}
		}
	}

	// Шифрование данных
	data, key, err := encryptionData(s.MasterKey, buffer.Bytes())
	if err != nil {
		s.Logger.With(zap.Error(err)).Error("ошибка шифрования данных")
		resp.Error = "ошибка шифрования данных"

		err := stream.SendAndClose(&resp)
		if err != nil {
			return fmt.Errorf(errorCloseStream, err)
		}
	}

	// Подготовка записи для сохранения
	var unit = domain.Storage{
		Name:  fileName,
		Type:  fileType,
		Value: data,
		Key:   key,
		Owner: token.ID,
	}

	// Запись в БД
	err = s.Svc.WriteRecord(unit)
	if err != nil {
		s.Logger.With(zap.Error(err)).Error("ошибка записи в БД")
		resp.Error = "ошибка записи в БД"

		err := stream.SendAndClose(&resp)
		if err != nil {
			return fmt.Errorf(errorCloseStream, err)
		}
	}

	// Закрываем поток
	err = stream.SendAndClose(&resp)
	if err != nil {
		return fmt.Errorf(errorCloseStream, err)
	}

	return nil
}

// DeleteRecord удаляет запись из БД
func (s StorageHandler) DeleteRecord(ctx context.Context, in *proto.DeleteRecordRequest) (*proto.DeleteRecordResponse, error) {
	var resp proto.DeleteRecordResponse

	// Получаем токен из контекста
	token, ok := middleware.GetTokenFromContext(ctx)
	if !ok {
		s.Logger.Error(errorInvalidToken)
		resp.Error = errorInvalidToken
		return &resp, nil
	}

	// Удаляем запись
	err := s.Svc.DeleteRecord(int(in.Id), token.ID)
	if err != nil {
		s.Logger.With(zap.Error(err)).Error("ошибка удаления записи")
		resp.Error = "ошибка удаления записи"
		return &resp, nil
	}

	return &resp, nil
}

/* УТИЛИТЫ */

var sizeRandomKey = 16

// encryptionData шифрует данные с помощью мастер-ключа
func encryptionData(mk string, data []byte) (string, string, error) {
	key, err := generateRandom(sizeRandomKey)
	if err != nil {
		return "", "", fmt.Errorf("ошибка генерации случайных байтов: %w", err)
	}

	encKey, err := encrypt([]byte(mk), key)
	if err != nil {
		return "", "", fmt.Errorf("ошибка шифрования ключа: %w", err)
	}

	encData, err := encrypt(key, data)
	if err != nil {
		return "", "", fmt.Errorf("ошибка шифрования данных: %w", err)
	}

	return encData, encKey, nil
}

// decryptionData расшифровывает данные с помощью мастер-ключа
func decryptionData(mk string, key string, data string) ([]byte, error) {
	decKey, err := decrypt([]byte(mk), key)
	if err != nil {
		return []byte{}, fmt.Errorf("ошибка расшифровки ключа: %w", err)
	}

	decData, err := decrypt(decKey, data)
	if err != nil {
		return []byte{}, fmt.Errorf("ошибка расшифровки данных: %w", err)
	}

	return decData, nil
}

// encrypt шифрует данные с помощью AES-GCM
func encrypt(key []byte, plaintext []byte) (string, error) {
	keyBytes := adjustKeySize(key, sizeRandomKey)
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("ошибка создания AES шифра: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("ошибка создания шифра: %w", err)
	}

	nonce, err := generateRandom(aesgcm.NonceSize())
	if err != nil {
		return "", fmt.Errorf("ошибка генерации случайных байтов: %w", err)
	}

	dst := aesgcm.Seal(nil, nonce, plaintext, nil)
	encString := base64.StdEncoding.EncodeToString(nonce) + "*" + base64.StdEncoding.EncodeToString(dst)

	return encString, nil
}

// decrypt расшифровывает данные с помощью AES-GCM
func decrypt(key []byte, plaintext string) ([]byte, error) {
	splStr := strings.Split(plaintext, "*")

	decNonce, err := base64.StdEncoding.DecodeString(splStr[0])
	if err != nil {
		return []byte{}, fmt.Errorf("ошибка декодирования base64: %w", err)
	}

	decString, err := base64.StdEncoding.DecodeString(splStr[1])
	if err != nil {
		return []byte{}, fmt.Errorf("ошибка декодирования base64: %w", err)
	}

	keyBytes := adjustKeySize(key, sizeRandomKey)
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return []byte{}, fmt.Errorf("ошибка создания AES шифра: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, fmt.Errorf("ошибка создания шифра: %w", err)
	}

	dst, err := aesgcm.Open(nil, decNonce, decString, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("ошибка расшифровки: %w", err)
	}

	return dst, nil
}

// adjustKeySize корректирует размер ключа до нужной длины
func adjustKeySize(originalKey []byte, desiredSize int) []byte {
	if len(originalKey) > desiredSize {
		return originalKey[:desiredSize]
	}
	return originalKey
}

// generateRandom генерирует случайные байты заданного размера
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации байтов: %w", err)
	}
	return b, nil
}
