package handlers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleStart - обрабатывает команду /start
func HandleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Printf("Обработка команды /start от пользователя %s (ID: %d)", update.Message.From.UserName, update.Message.From.ID)

	// Используем один большой строковый литерал для ясности и надежности.
	// Все специальные символы для MarkdownV2 ('.', '!', '-') экранированы с помощью '\\'.
	text := "Привет\\! Я твой бот\\-помощник для учёта финансов\\.\n\n" +
		"*Основные команды:*\n" +
		"`1000`  \\- записать доход\n" +
		"`-500 кофе`  \\- записать расход с комментарием\n\n" +
		"*Отчёты:*\n" +
		"/today  \\- итоги за сегодня\n" +
		"/week  \\- итоги за неделю\n" +
		"/month  \\- итоги за месяц\n" +
		"/export  \\- выгрузить всё в CSV\n\n" +
		"*Управление данными:*\n" +
		"/clearlast \\- удалить последнюю запись\n" +
		"/cleartoday \\- удалить все записи за сегодня"

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке приветственного сообщения: %v", err)
	} else {
		log.Println("Приветственное сообщение отправлено.")
	}
}
