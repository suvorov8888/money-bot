package bot

import (
	"log"

	"money-bot/internal/handlers" // Импортируем наши хендлеры

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot структура содержит ссылку на API и другие зависимости
type Bot struct {
	api *tgbotapi.BotAPI
}

// NewBot создает новый экземпляр бота
func NewBot(api *tgbotapi.BotAPI) *Bot {
	return &Bot{api: api}
}

// Run запускает бота и обрабатывает входящие сообщения
func (b *Bot) Run() {
	log.Printf("Авторизовались для аккаунта %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			// Логика обработки сообщений будет здесь
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			switch update.Message.Command() {
			case "start":
				handlers.HandleStart(b.api, update)
			default:
				// Обработка неизвестной команды (позже)
			}
		}
	}
}
