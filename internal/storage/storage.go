package storage

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Storage структура для работы с базой данных
type Storage struct {
	db *gorm.DB
}

// NewStorage подключается к базе данных и выполняет миграцию
func NewStorage(dbPath string) (*Storage, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Автоматическая миграция (создание таблицы, если её нет)
	err = db.AutoMigrate(&Transaction{})
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

// SaveTransaction сохраняет новую транзакцию в базе данных
func (s *Storage) SaveTransaction(transaction *Transaction) error {
	result := s.db.Create(transaction)
	if result.Error != nil {
		log.Printf("Ошибка сохранения транзакции в базе данных: %v", result.Error)
	}
	return result.Error
}
