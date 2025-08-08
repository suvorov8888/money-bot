package bot

import (
	"log"
	"regexp"
	"strconv"
	"strings"
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
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				handlers.HandleStart(b.api, update)
			case "today":
				handlers.HandleReport(b.api, update, b.storage, "today")
				log.Printf("Получена команда /today от пользователя %s", update.Message.From.UserName)
			case "week":
				handlers.HandleReport(b.api, update, b.storage, "week")
				log.Printf("Получена команда /week от пользователя %s", update.Message.From.UserName)
			case "month":
				handlers.HandleReport(b.api, update, b.storage, "month")
				log.Printf("Получена команда /month от пользователя %s", update.Message.From.UserName)
			case "export":
				handlers.HandleExport(b.api, update, b.storage)
				log.Printf("Получена команда /export от пользователя %s", update.Message.From.UserName)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я не знаю такой команды.")
				b.api.Send(msg)
			}
			continue
		}

		// Используем регулярное выражение для поиска числа в начале сообщения
		re := regexp.MustCompile(`^-?\d+(\.\d+)?`)
		matches := re.FindStringSubmatch(update.Message.Text)

		if len(matches) > 0 {
			amount, err := strconv.ParseFloat(matches[0], 64)
			if err != nil {
				log.Printf("Ошибка при парсинге числа: %v", err)
				continue
			}

			// Получаем комментарий, обрезая число
			comment := strings.TrimSpace(strings.Replace(update.Message.Text, matches[0], "", 1))

			b.saveTransaction(update, amount, comment)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, введите число (например, 1000 или -500 на кофе).")
			b.api.Send(msg)
		}
	}
}

// saveTransaction сохраняет транзакцию в базе данных
func (b *Bot) saveTransaction(update tgbotapi.Update, amount float64, comment string) {
	transaction := &storage.Transaction{
		UserID:          update.Message.From.ID,
		Amount:          amount,
		Comment:         comment,
		TransactionDate: time.Now(),
	}

	if err := b.storage.SaveTransaction(transaction); err != nil {
		log.Printf("Ошибка при сохранении транзакции: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при сохранении транзакции. Попробуйте еще раз.")
		b.api.Send(msg)
	} else {
		var responseText string
		if amount > 0 {
			responseText = "✅ Доход успешно сохранён!"
		} else {
			responseText = "✅ Расход успешно сохранён!"
		}
		if comment != "" {
			responseText += "\nКомментарий: " + comment
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
		b.api.Send(msg)
	}
}
