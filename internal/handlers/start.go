package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleStart - обрабатывает команду /start
func HandleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я твой бот-помощник для учёта финансов. Отправь мне число, чтобы записать доход, или число со знаком минус для расхода.")
	bot.Send(msg)
}
