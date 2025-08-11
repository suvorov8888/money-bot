package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	"money-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleReport генерирует и отправляет отчет по транзакциям за определенный период
func HandleReport(bot *tgbotapi.BotAPI, update tgbotapi.Update, s *storage.Storage, period string) {
	log.Printf("Начало обработки отчета за период '%s' для пользователя %s (ID: %d)", period, update.Message.From.UserName, update.Message.From.ID)
	var (
		from, to    time.Time
		reportTitle string
	)

	switch period {
	case "today":
		from, to = GetStartAndEndOfDay()
		reportTitle = "Итоги за сегодня"
	case "week":
		from, to = GetStartAndEndOfWeek()
		reportTitle = "Итоги за неделю"
	case "month":
		from, to = GetStartAndEndOfMonth()
		reportTitle = "Итоги за месяц"
	default:
		log.Printf("Неизвестный период для отчета: %s", period)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестный период для отчета.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения о неизвестном периоде: %v", err)
		}
		return
	}
	log.Printf("Рассчитан временной интервал для отчета: с %s по %s", from.Format(time.RFC3339), to.Format(time.RFC3339))

	log.Printf("Запрос транзакций из БД для UserID: %d", update.Message.From.ID)
	transactions, err := s.GetTransactionsByPeriod(update.Message.From.ID, from, to)
	if err != nil {
		log.Printf("Ошибка при получении транзакций из БД: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении данных.")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения об ошибке получения данных: %v", err)
		}
		return
	}

	if len(transactions) == 0 {
		log.Printf("Транзакции за период не найдены для UserID: %d", update.Message.From.ID)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%s: транзакций не найдено.", reportTitle))
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка при отправке сообщения об отсутствии транзакций: %v", err)
		}
		return
	}
	log.Printf("Найдено %d транзакций. Начинаем формирование отчета.", len(transactions))

	var responseText strings.Builder
	// Заголовки отчетов не содержат спецсимволов, поэтому можно не экранировать.
	// Звёздочки для жирного шрифта — это часть нашей разметки.
	responseText.WriteString(fmt.Sprintf("📊 *%s* 📊\n\n", reportTitle))

	var totalIncome, totalExpense float64
	for _, tr := range transactions {
		if tr.Amount > 0 {
			totalIncome += tr.Amount
		} else {
			totalExpense += tr.Amount
		}
		sign := "➕"
		if tr.Amount < 0 {
			sign = "➖"
		}
		// Суммы в блоках `code` (обратные кавычки), их экранировать не нужно.
		amountStr := fmt.Sprintf("%.2f", tr.Amount)
		// Комментарий может содержать спецсимволы, его нужно экранировать.
		escapedComment := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, tr.Comment)
		// Добавляем категорию в отчет, чтобы было нагляднее
		escapedCategory := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, tr.Category)
		responseText.WriteString(fmt.Sprintf("%s `%s` руб\\. \\| %s \\(*%s*\\)\n", sign, amountStr, escapedComment, escapedCategory))
	}

	responseText.WriteString("\n\\-\\-\\-\n")
	responseText.WriteString(fmt.Sprintf("💰 *Доходы*: `%.2f` руб\\.\n", totalIncome))
	responseText.WriteString(fmt.Sprintf("💸 *Расходы*: `%.2f` руб\\.\n", totalExpense))
	responseText.WriteString(fmt.Sprintf("📈 *Баланс*: `%.2f` руб\\.", totalIncome+totalExpense))

	// Получаем и добавляем общий баланс за все время для контекста
	overallBalance, err := s.GetAllTimeSummary(update.Message.From.ID)
	if err != nil {
		log.Printf("Ошибка при получении общего баланса для UserID %d: %v", update.Message.From.ID, err)
		// Не прерываем отчет, просто не показываем общий баланс
	} else {
		responseText.WriteString(fmt.Sprintf("\n\n🏦 *Общий баланс*: `%.2f` руб\\.", overallBalance))
	}

	log.Printf("Отчет сформирован. Итоги: Доход=%.2f, Расход=%.2f, Баланс=%.2f", totalIncome, totalExpense, totalIncome+totalExpense)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText.String())
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	log.Println("Отправка отчета пользователю.")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке отчета: %v", err)
	}
}
