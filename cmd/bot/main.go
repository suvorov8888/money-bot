package main

import (
	"log"
	"money-bot/ai"
	"os"
	"path/filepath"

	"money-bot/internal/bot"
	"money-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Запуск бота...")
	// 1. Загружаем переменные из .env файла
	// Используем Overload, чтобы переменные из .env файла
	// имели приоритет над системными переменными.
	log.Println("Загрузка переменных окружения из .env файла...")
	err := godotenv.Overload()
	if err != nil {
		log.Fatalf("Ошибка загрузки файла .env file: %v", err)
	}
	log.Println("Переменные окружения успешно загружены.")

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не найден в .env file")
	}
	log.Println("TELEGRAM_BOT_TOKEN успешно загружен.")

	deepseekAPIKey := os.Getenv("DEEPSEEK_API_KEY")
	if deepseekAPIKey == "" {
		// Можно сделать эту ошибку не фатальной, если AI не является основной функцией
		log.Println("ВНИМАНИЕ: DEEPSEEK_API_KEY не найден в .env file. Функция классификации будет использовать категорию 'Прочее'.")
	} else {
		log.Println("DEEPSEEK_API_KEY успешно загружен.")
	}

	// 2. Инициализируем хранилище данных (базу)
	dbPath := "db/data.db"
	log.Printf("Инициализация хранилища данных по пути: %s", dbPath)
	// Создаем директорию для базы данных, если она не существует
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Не удалось создать директорию для базы данных: %v", err)
	}

	dbStorage, err := storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	log.Println("Хранилище данных успешно инициализировано.")

	// Инициализируем пакет AI с токеном
	log.Println("Инициализация пакета AI...")
	ai.Init(deepseekAPIKey)
	log.Println("Пакет AI успешно инициализирован.")

	// 3. Создаем новый экземпляр нашего бота
	log.Println("Создание экземпляра Telegram Bot API...")
	tgBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Экземпляр Telegram Bot API успешно создан.")

	// 4. Создаем наш собственный экземпляр бота, передавая ему токен и хранилище
	log.Println("Создание кастомного экземпляра бота...")
	myBot := bot.NewBot(tgBot, dbStorage)
	log.Println("Кастомный экземпляр бота успешно создан.")

	// 5. Запускаем бота
	log.Println("Запуск основного цикла обработки сообщений...")
	myBot.Run()
}
