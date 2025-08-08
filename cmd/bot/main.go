package main

import (
	"log"
	"os"
	"path/filepath"

	"money-bot/internal/bot"
	"money-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Загружаем переменные из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set in the .env file")
	}

	// 2. Инициализируем хранилище данных (базу)
	dbPath := "db/data.db"
	// Создаем директорию для базы данных, если она не существует
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Не удалось создать директорию для базы данных: %v", err)
	}

	dbStorage, err := storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// 3. Создаем новый экземпляр нашего бота
	tgBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// 4. Создаем наш собственный экземпляр бота, передавая ему токен и хранилище
	myBot := bot.NewBot(tgBot, dbStorage)

	// 5. Запускаем бота
	myBot.Run()
}
