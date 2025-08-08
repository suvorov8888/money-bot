package bot

import (
	"log"
	"strconv"
	"time"

	"money-bot/internal/handlers" // Импортируем наши хендлеры
	"money-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot структура содержит ссылку на API и другие зависимости
type Bot struct {
	api     *tgbotapi.BotAPI
	storage *storage.Storage // Добавляем поле для хранилища
}

// NewBot создает новый экземпляр бота
func NewBot(api *tgbotapi.BotAPI, s *storage.Storage) *Bot {
	return &Bot{api: api, storage: s}
}

// Run запускает бота и обрабатывает входящие сообщения
func (b *Bot) Run() {
	log.Printf("Авторизовались для аккаунта %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // Игнорируем все обновления, которые не являются сообщениями
			continue
		}

		// Сначала проверяем, является ли сообщение командой
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				handlers.HandleStart(b.api, update)
			default:
				// Обработка неизвестной команды
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я не знаю такой команды.")
				b.api.Send(msg)
			}
		} else {
			// Если это не команда, пробуем обработать как число
			amount, err := strconv.ParseFloat(update.Message.Text, 64)
			if err == nil {
				// Если это число, сохраняем его
				b.saveTransaction(update, amount)
			} else {
				// Если это не число, отправляем сообщение об ошибке
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, введите число (например, 1000 для дохода или -500 для расхода).")
				b.api.Send(msg)
			}
		}
	}
}

// saveTransaction сохраняет транзакцию в базе данных
func (b *Bot) saveTransaction(update tgbotapi.Update, amount float64) {
	transaction := &storage.Transaction{
		UserID:          update.Message.From.ID,
		Amount:          amount,
		Comment:         "", // Пока без комментариев
		TransactionDate: time.Now(),
	}

	if err := b.storage.SaveTransaction(transaction); err != nil {
		log.Printf("Error saving transaction: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при сохранении транзакции. Попробуйте еще раз.")
		b.api.Send(msg)
	} else {
		var responseText string
		if amount > 0 {
			responseText = "✅ Доход успешно сохранён!"
		} else {
			responseText = "✅ Расход успешно сохранён!"
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
		b.api.Send(msg)
	}
}
