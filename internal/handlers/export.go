package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"time"

	"money-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleExport создает и отправляет CSV-файл с транзакциями
func HandleExport(bot *tgbotapi.BotAPI, update tgbotapi.Update, s *storage.Storage) {
	log.Printf("Начало обработки экспорта для пользователя %s (ID: %d)", update.Message.From.UserName, update.Message.From.ID)
	transactions, err := s.GetAllTransactions(update.Message.From.ID)
	if err != nil {
		log.Printf("Ошибка при получении всех транзакций из БД для экспорта: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении данных для экспорта.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения об ошибке экспорта: %v", err)
		}
		return
	}

	if len(transactions) == 0 {
		log.Printf("Нет транзакций для экспорта для UserID: %d", update.Message.From.ID)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет транзакций для экспорта.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения об отсутствии транзакций для экспорта: %v", err)
		}
		return
	}
	log.Printf("Найдено %d транзакций для экспорта. Начинаем генерацию CSV.", len(transactions))

	// Создаем буфер для записи CSV-файла
	var b bytes.Buffer
	w := csv.NewWriter(&b)

	// Записываем заголовок
	header := []string{"ID", "Дата", "Сумма", "Комментарий", "Категория"}
	if err := w.Write(header); err != nil {
		log.Printf("Ошибка при записи заголовка в CSV: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при создании CSV-файла.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения об ошибке создания CSV: %v", err)
		}
		return
	}

	// Записываем данные из транзакций
	for _, tr := range transactions {
		record := []string{
			fmt.Sprintf("%d", tr.ID),
			tr.TransactionDate.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.2f", tr.Amount),
			tr.Comment,
			tr.Category,
		}
		if err := w.Write(record); err != nil {
			log.Printf("Ошибка при записи строки %d в CSV: %v", tr.ID, err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при создании CSV-файла.")
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения об ошибке создания CSV: %v", err)
			}
			return
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		log.Printf("Ошибка при сбросе буфера CSV: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при создании CSV-файла.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения об ошибке создания CSV: %v", err)
		}
		return
	}
	log.Println("CSV-данные успешно сгенерированы.")

	// Создаем и отправляем файл
	fileName := fmt.Sprintf("transactions_%s.csv", time.Now().Format("2006-01-02"))
	log.Printf("Подготовка файла для отправки: %s", fileName)
	file := tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: b.Bytes(),
	}
	doc := tgbotapi.NewDocument(update.Message.Chat.ID, file)
	log.Println("Отправка CSV-файла пользователю.")
	if _, err := bot.Send(doc); err != nil {
		log.Printf("Ошибка при отправке CSV-файла: %v", err)
	}
}
