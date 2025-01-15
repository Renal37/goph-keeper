package repository

import (
	"github.com/Renal37/goph-keeper/internal/server/core/domain"
)

// ReadAllRecord извлекает все записи хранилища для конкретного владельца.
// Использует метод `Find` для запроса базы данных на наличие записей хранилища,
// которые соответствуют указанному владельцу. Если записи не найдены, возвращает
// nil для обоих: среза записей и ошибки. Если возникает ошибка во время запроса,
// возвращает ошибку.
func (s *DB) ReadAllRecord(owner int) ([]*domain.Storage, error) {
	docs := []*domain.Storage{}

	req := s.db.Select("id", "name", "owner").Find(&docs, "owner = ?", owner)
	if req.RowsAffected == 0 {
		return nil, nil
	}

	if req.Error != nil {
		return nil, req.Error
	}

	return docs, nil
}

// ReadRecord извлекает конкретную запись хранилища по её ID и владельцу.
// Использует метод `First` для запроса базы данных на наличие записи хранилища,
// которая соответствует указанным ID и владельцу. Если запись не найдена, возвращает
// nil для обоих: записи и ошибки. Если возникает ошибка во время запроса,
// возвращает ошибку.
func (s *DB) ReadRecord(id int, owner int) (*domain.Storage, error) {
	doc := domain.Storage{}

	req := s.db.First(&doc, "id = ? AND owner = ?", id, owner)
	if req.RowsAffected == 0 {
		//nolint:nilnil // Это допустимый возврат
		return nil, nil
	}

	if req.Error != nil {
		return nil, req.Error
	}

	return &doc, nil
}

// WriteRecord добавляет новую запись хранилища в базу данных.
// Использует метод `Create` для вставки записи. Если возникает ошибка
// во время вставки, возвращает ошибку.
func (s *DB) WriteRecord(doc domain.Storage) error {
	req := s.db.Create(&doc)
	if req.Error != nil {
		return req.Error
	}

	return nil
}

// DeleteRecord удаляет запись хранилища из базы данных по её ID и владельцу.
// Использует метод `Delete` для удаления записи. Если возникает ошибка во время
// удаления, возвращает ошибку.
func (s *DB) DeleteRecord(id int, owner int) error {
	doc := domain.Storage{}

	req := s.db.Delete(&doc, "id = ? AND owner = ?", id, owner)
	if req.Error != nil {
		return req.Error
	}

	return nil
}
