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

// GetAllTransactions возвращает все транзакции пользователя
func (s *Storage) GetAllTransactions(userID int64) ([]Transaction, error) {
	var transactions []Transaction
	result := s.db.Where("user_id = ?", userID).Find(&transactions)
	return transactions, result.Error
}

// DeleteLastTransaction находит и удаляет последнюю транзакцию пользователя.
// Возвращает удаленную транзакцию или ошибку, если транзакций нет.
func (s *Storage) DeleteLastTransaction(userID int64) (*Transaction, error) {
	var lastTransaction Transaction
	// Ищем последнюю транзакцию по ID, так как это самый надежный способ найти последнюю запись
	if err := s.db.Where("user_id = ?", userID).Order("id desc").First(&lastTransaction).Error; err != nil {
		// Возвращаем ошибку, если ничего не найдено (gorm.ErrRecordNotFound)
		return nil, err
	}

	// Удаляем найденную транзакцию
	if err := s.db.Delete(&lastTransaction).Error; err != nil {
		return nil, err
	}

	return &lastTransaction, nil
}

// DeleteTransactionsForToday удаляет все транзакции пользователя за сегодняшний день.
// Возвращает количество удаленных транзакций.
func (s *Storage) DeleteTransactionsForToday(userID int64) (int64, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24*time.Hour - 1) // Конец дня

	result := s.db.Where("user_id = ? AND transaction_date BETWEEN ? AND ?", userID, startOfDay, endOfDay).Delete(&Transaction{})
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
