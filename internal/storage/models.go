package storage

import (
	"time"

	"gorm.io/gorm"
)

// Transaction модель для хранения финансовой операции
type Transaction struct {
	gorm.Model              // Включает поля ID, CreatedAt, UpdatedAt, DeletedAt
	UserID          int64   // ID пользователя Telegram
	Amount          float64 // Сумма операции (положительная для дохода, отрицательная для расхода)
	Category        string  // Категория (пока не используем, но оставим на будущее)
	Comment         string  // Комментарий к операции
	TransactionDate time.Time
}
