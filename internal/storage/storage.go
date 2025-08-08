package storage

import (
	"log"
	"time"

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

// GetTransactionsByPeriod возвращает все транзакции пользователя за указанный период
func (s *Storage) GetTransactionsByPeriod(userID int64, from, to time.Time) ([]Transaction, error) {
	var transactions []Transaction
	result := s.db.Where("user_id = ? AND transaction_date BETWEEN ? AND ?", userID, from, to).Find(&transactions)
	return transactions, result.Error
}

// GetPeriodSummary рассчитывает сумму всех транзакций пользователя за указанный период
func (s *Storage) GetPeriodSummary(userID int64, from, to time.Time) (float64, error) {
	var total float64
	result := s.db.Model(&Transaction{}).Where("user_id = ? AND transaction_date BETWEEN ? AND ?", userID, from, to).Select("SUM(amount)").Row().Scan(&total)
	return total, result
}
