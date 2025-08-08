package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleStart - обрабатывает команду /start
func HandleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я твой бот-помощник для учёта финансов.\n\n Отправь мне число, чтобы записать доход, или число со знаком минус для расхода.\n\n После любого числа через пробел укажи комментарий (например -500 кофе).\n\n Чтобы получить итоги за день, неделю и месяц - воспользуйся командами /today, /week, /month.\n\n Чтобы выгрузить свои данные в exel - воспользуйся командой /export.")
	bot.Send(msg)
}
