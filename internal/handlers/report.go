package handlers

import (
	"fmt"
	"strings"
	"time"

	"money-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleReport генерирует и отправляет отчет по транзакциям за определенный период
func HandleReport(bot *tgbotapi.BotAPI, update tgbotapi.Update, s *storage.Storage, period string) {
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
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестный период для отчета.")
		bot.Send(msg)
		return
	}

	transactions, err := s.GetTransactionsByPeriod(update.Message.From.ID, from, to)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении данных.")
		bot.Send(msg)
		return
	}

	if len(transactions) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%s: транзакций не найдено.", reportTitle))
		bot.Send(msg)
		return
	}

	var responseText strings.Builder
	// Экранируем специальные символы для MarkdownV2
	title := strings.ReplaceAll(reportTitle, "_", "\\_")
	responseText.WriteString(fmt.Sprintf("📊 *%s* 📊\n\n", title))

	var totalIncome, totalExpense float64
	for _, tr := range transactions {
		if tr.Amount > 0 {
			totalIncome += tr.Amount
		} else {
			totalExpense += tr.Amount
		}
		// Форматируем каждую транзакцию для вывода
		sign := ""
		if tr.Amount < 0 {
			sign = "➖"
		} else {
			sign = "➕"
		}
		// Экранируем `.` в сумме и другие символы
		amountStr := strings.ReplaceAll(fmt.Sprintf("%.2f", tr.Amount), ".", "\\.")
		comment := strings.ReplaceAll(tr.Comment, "_", "\\_")
		responseText.WriteString(fmt.Sprintf("%s `%s` руб\\. \\| %s\n", sign, amountStr, comment))
	}

	responseText.WriteString("\n\\-\\-\\-\n")
	responseText.WriteString(fmt.Sprintf("💰 *Доходы*: `%s` руб\\.", strings.ReplaceAll(fmt.Sprintf("%.2f", totalIncome), ".", "\\.")))
	responseText.WriteString(fmt.Sprintf("\n💸 *Расходы*: `%s` руб\\.", strings.ReplaceAll(fmt.Sprintf("%.2f", totalExpense), ".", "\\.")))
	responseText.WriteString(fmt.Sprintf("\n📈 *Баланс*: `%s` руб\\.", strings.ReplaceAll(fmt.Sprintf("%.2f", totalIncome+totalExpense), ".", "\\.")))

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText.String())
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	bot.Send(msg)
}
