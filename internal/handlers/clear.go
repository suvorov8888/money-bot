package handlers

import (
	"errors"
	"fmt"
	"log"

	"money-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// HandleClearLast обрабатывает команду /clear_last
func HandleClearLast(bot *tgbotapi.BotAPI, update tgbotapi.Update, s *storage.Storage) {
	log.Printf("Обработка команды /clear_last от пользователя %s (ID: %d)", update.Message.From.UserName, update.Message.From.ID)

	deletedTransaction, err := s.DeleteLastTransaction(update.Message.From.ID)
	if err != nil {
		var responseText string
		if errors.Is(err, gorm.ErrRecordNotFound) {
			responseText = "Нет транзакций для удаления."
			log.Printf("Для пользователя %d нет транзакций для удаления.", update.Message.From.ID)
		} else {
			responseText = "Произошла ошибка при удалении последней транзакции."
			log.Printf("Ошибка при удалении последней транзакции для UserID %d: %v", update.Message.From.ID, err)
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
		bot.Send(msg)
		return
	}

	responseText := fmt.Sprintf(
		"✅ Последняя транзакция удалена:\n\nСумма: %.2f\nКомментарий: %s\nКатегория: %s",
		deletedTransaction.Amount,
		deletedTransaction.Comment,
		deletedTransaction.Category,
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
	bot.Send(msg)
	log.Printf("Последняя транзакция для пользователя %d успешно удалена.", update.Message.From.ID)
}

// HandleClearToday обрабатывает команду /clear_today
func HandleClearToday(bot *tgbotapi.BotAPI, update tgbotapi.Update, s *storage.Storage) {
	log.Printf("Обработка команды /clear_today от пользователя %s (ID: %d)", update.Message.From.UserName, update.Message.From.ID)

	count, err := s.DeleteTransactionsForToday(update.Message.From.ID)
	if err != nil {
		log.Printf("Ошибка при удалении транзакций за сегодня для UserID %d: %v", update.Message.From.ID, err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при удалении транзакций.")
		bot.Send(msg)
		return
	}

	responseText := "За сегодня не найдено транзакций для удаления."
	if count > 0 {
		responseText = fmt.Sprintf("✅ Удалено %d транзакций за сегодня.", count)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
	bot.Send(msg)
	log.Printf("Для пользователя %d удалено %d транзакций за сегодня.", update.Message.From.ID, count)
}
