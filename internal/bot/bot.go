package bot

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"money-bot/ai"
	"money-bot/internal/handlers" // Импортируем наши хендлеры
	"money-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot структура содержит ссылку на API и другие зависимости
type Bot struct {
	api        *tgbotapi.BotAPI
	storage    *storage.Storage // Добавляем поле для хранилища
	categories []string         // Добавляем поле для категорий
}

// NewBot создает новый экземпляр бота
func NewBot(api *tgbotapi.BotAPI, s *storage.Storage) *Bot {
	// В будущем этот список можно будет загружать из файла конфигурации или базы данных
	defaultCategories := []string{
		"Автомобиль",           // Бензин, страховка, ремонт
		"Еда вне дома",         // Рестораны, кафе, доставка
		"Здоровье",             // Аптеки, врачи, страховка
		"Коммунальные платежи", // Аренда, ЖКУ, интернет
		"Одежда и обувь",
		"Образование", // Курсы, книги, обучение
		"Питомцы",     // Корм, игрушки, ветеринар
		"Подарки",
		"Продукты",         // Покупки в супермаркетах
		"Путешествия",      // Билеты, отели, расходы в отпуске
		"Развлечения",      // Кино, концерты, хобби
		"Связь и подписки", // Мобильная связь, стриминговые сервисы
		"Спорт и фитнес",   // Абонемент в зал, спорттовары
		"Товары для дома",  // Мебель, бытовая химия, декор
		"Транспорт",        // Общественный транспорт, такси
		"Уход за собой",    // Косметика, парикмахерская, спа
		"Прочее",           // Другие расходы
	}
	return &Bot{
		api:        api,
		storage:    s,
		categories: defaultCategories,
	}
}

// Run запускает бота и обрабатывает входящие сообщения
func (b *Bot) Run() {
	log.Printf("Авторизовались для аккаунта %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	log.Println("Начинаем прослушивание обновлений...")

	for update := range updates {
		log.Printf("Получено новое обновление. UpdateID: %d", update.UpdateID)

		if update.Message == nil {
			log.Println("Обновление не содержит сообщения, пропускаем.")
			continue
		}

		log.Printf("Получено сообщение от пользователя %s (ID: %d) в чате %d: \"%s\"", update.Message.From.UserName, update.Message.From.ID, update.Message.Chat.ID, update.Message.Text)

		// блок обработки команд от бота
		if update.Message.IsCommand() {
			command := update.Message.Command()
			log.Printf("Сообщение является командой: /%s", command)
			switch command {
			case "start":
				handlers.HandleStart(b.api, update)
			case "today":
				handlers.HandleReport(b.api, update, b.storage, "today")
			case "week":
				handlers.HandleReport(b.api, update, b.storage, "week")
			case "month":
				handlers.HandleReport(b.api, update, b.storage, "month")
			case "export":
				handlers.HandleExport(b.api, update, b.storage)
			case "clear_last", "clearlast": // Принимаем оба варианта
				handlers.HandleClearLast(b.api, update, b.storage)
			case "clear_today", "cleartoday": // Принимаем оба варианта
				handlers.HandleClearToday(b.api, update, b.storage)
			default:
				log.Printf("Неизвестная команда: /%s", command)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я не знаю такой команды.")
				if _, err := b.api.Send(msg); err != nil {
					log.Printf("Ошибка при отправке сообщения о неизвестной команде: %v", err)
				}
			}
			continue
		}

		// Используем регулярное выражение для поиска числа в начале сообщения
		log.Println("Сообщение не является командой, попытка обработать как транзакцию.")
		// этот код предназначен для извлечения числа в начале текстового сообщения
		re := regexp.MustCompile(`^-?\d+(\.\d+)?`)
		matches := re.FindStringSubmatch(update.Message.Text)

		if len(matches) > 0 {
			log.Printf("Найдено число в сообщении: %s", matches[0])
			amount, err := strconv.ParseFloat(matches[0], 64)
			if err != nil {
				log.Printf("Критическая ошибка: не удалось спарсить число '%s' после проверки регулярным выражением: %v", matches[0], err)
				continue
			}

			comment := strings.TrimSpace(strings.Replace(update.Message.Text, matches[0], "", 1))
			log.Printf("Извлечена сумма: %.2f, комментарий: \"%s\"", amount, comment)

			// === Изменения начинаются здесь ===
			var category string
			if amount < 0 { // Это расход, определяем категорию
				if comment != "" {
					log.Printf("Комментарий не пустой, начинаем классификацию транзакции...")
					// Вызываем нашу функцию для классификации
					category, err = ai.ClassifyTransaction(comment, b.categories)
					if err != nil {
						log.Printf("Ошибка при классификации транзакции: %v", err)
						category = "Прочее" // Если произошла ошибка, используем категорию по умолчанию
						log.Println("Установлена категория по умолчанию: 'Прочее'")
					} else {
						log.Printf("Транзакция успешно классифицирована. Категория: %s", category)
					}
				} else {
					category = "Прочее" // Категория по умолчанию, если комментария нет
					log.Println("Комментарий пустой, установлена категория по умолчанию: 'Прочее'")
				}
			} else {
				// Для доходов устанавливаем категорию "Доход" без анализа
				category = "Доход"
				log.Printf("Транзакция является доходом, установлена категория: '%s'", category)
			}

			// Передаем категорию в функцию saveTransaction
			log.Println("Вызов функции сохранения транзакции...")
			b.saveTransaction(update, amount, comment, category)

		} else {
			log.Printf("Сообщение не соответствует формату транзакции. Отправка подсказки пользователю.")
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, введите число (например, 1000 или -500 на кофе).")
			if _, err := b.api.Send(msg); err != nil {
				log.Printf("Ошибка при отправке подсказки: %v", err)
			}
		}
	}
}

// saveTransaction сохраняет транзакцию в базе данных
func (b *Bot) saveTransaction(update tgbotapi.Update, amount float64, comment, category string) {
	log.Printf("Подготовка к сохранению транзакции: UserID=%d, Amount=%.2f, Comment='%s', Category='%s'", update.Message.From.ID, amount, comment, category)
	transaction := &storage.Transaction{
		UserID:          update.Message.From.ID,
		Amount:          amount,
		Comment:         comment,
		Category:        category, // Убедитесь, что поле Category добавлено в структуру storage.Transaction
		TransactionDate: time.Now(),
	}

	if err := b.storage.SaveTransaction(transaction); err != nil {
		log.Printf("Ошибка при сохранении транзакции в БД: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при сохранении транзакции. Попробуйте еще раз.")
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения об ошибке сохранения: %v", err)
		}
	} else {
		log.Printf("Транзакция успешно сохранена в БД. ID транзакции: %d", transaction.ID)
		var responseText string
		if amount > 0 {
			responseText = "✅ Доход успешно сохранён!"
		} else {
			responseText = "✅ Расход успешно сохранён!"
		}
		// Добавляем сумму в ответ для наглядности
		responseText += "\nСумма: " + strconv.FormatFloat(amount, 'f', 2, 64)

		if comment != "" {
			responseText += "\nКомментарий: " + comment
		}

		// === Добавляем категорию в ответное сообщение ===
		responseText += "\nКатегория: " + category

		log.Printf("Отправка подтверждения пользователю: \"%s\"", responseText)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("Ошибка при отправке подтверждения о сохранении: %v", err)
		}
	}
}
