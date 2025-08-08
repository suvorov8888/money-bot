package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"money-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleExport создает и отправляет CSV-файл с транзакциями
func HandleExport(bot *tgbotapi.BotAPI, update tgbotapi.Update, s *storage.Storage) {
	transactions, err := s.GetAllTransactions(update.Message.From.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении данных для экспорта.")
		bot.Send(msg)
		return
	}

	if len(transactions) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет транзакций для экспорта.")
		bot.Send(msg)
		return
	}

	// Создаем буфер для записи CSV-файла
	var b bytes.Buffer
	w := csv.NewWriter(&b)

	// Записываем заголовок
	header := []string{"ID", "Дата", "Сумма", "Комментарий"}
	w.Write(header)

	// Записываем данные из транзакций
	for _, tr := range transactions {
		record := []string{
			fmt.Sprintf("%d", tr.ID),
			tr.CreatedAt.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.2f", tr.Amount),
			tr.Comment,
		}
		w.Write(record)
	}

	w.Flush()
	if err := w.Error(); err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при создании CSV-файла.")
		bot.Send(msg)
		return
	}

	// Создаем и отправляем файл
	fileName := fmt.Sprintf("transactions_%s.csv", time.Now().Format("2006-01-02"))
	file := tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: b.Bytes(),
	}
	doc := tgbotapi.NewDocument(update.Message.Chat.ID, file)
	bot.Send(doc)
}
