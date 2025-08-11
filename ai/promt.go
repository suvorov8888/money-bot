package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// AIMessage представляет одно сообщение в диалоге
type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIRequest представляет тело запроса к API
type AIRequest struct {
	Model    string      `json:"model"`
	Messages []AIMessage `json:"messages"`
}

// AIResponse представляет тело ответа от API
type AIResponse struct {
	Choices []struct {
		Message AIMessage `json:"message"`
	} `json:"choices"`
}

const (
	// URL для OpenRouter Chat API
	modelAPIURL = "https://openrouter.ai/api/v1/chat/completions"
)

var (
	// Эта переменная будет хранить API ключ для OpenRouter.
	openRouterAPIKey string
	// Создаем один HTTP-клиент на уровне пакета для переиспользования.
	// Это более эффективно, чем создавать нового клиента на каждый запрос.
	// Также добавляем таймаут, чтобы избежать "зависших" запросов.
	apiClient = &http.Client{Timeout: time.Second * 30}
)

// Init инициализирует пакет ai, устанавливая токен.
func Init(token string) {
	openRouterAPIKey = token
}

// ClassifyTransaction отправляет запрос к API OpenRouter для классификации транзакции.
func ClassifyTransaction(text string, categories []string) (string, error) {
	log.Printf("Начинаем классификацию текста через OpenRouter: \"%s\"", text)

	if openRouterAPIKey == "" {
		return "", fmt.Errorf("API ключ для OpenRouter не установлен")
	}

	// Формируем системный и пользовательский промпты.
	// Системный промпт задает "личность" и задачу для AI.
	systemPrompt := fmt.Sprintf(`Ты — ассистент для классификации трат. Твоя задача - проанализировать текст и определить наиболее подходящую категорию из списка.
Отвечай строго названием одной категории из списка, без лишних слов и знаков препинания.

Список категорий:
- %s`, strings.Join(categories, "\n- "))

	userPrompt := fmt.Sprintf(`Текст для анализа: "%s"`, text)

	// Создаем тело запроса
	requestPayload := AIRequest{
		Model: "mistralai/mistral-7b-instruct:free", // Используем надежную бесплатную модель от Mistral
		Messages: []AIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	// Преобразование промта в JSON
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		log.Printf("Критическая ошибка при маршалинге JSON для OpenRouter запроса: %v", err)
		return "", fmt.Errorf("ошибка при маршалинге JSON: %w", err)
	}

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", modelAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Критическая ошибка при создании HTTP-запроса к OpenRouter: %v", err)
		return "", fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	// Установка заголовков
	req.Header.Set("Authorization", "Bearer "+openRouterAPIKey)
	req.Header.Set("Content-Type", "application/json")
	// OpenRouter рекомендует добавлять эти заголовки для идентификации вашего проекта
	req.Header.Set("HTTP-Referer", "https://github.com/user/money-bot")
	req.Header.Set("X-Title", "Money Bot")

	log.Println("Заголовки для OpenRouter запроса установлены.")

	log.Printf("Отправка запроса на URL: %s", modelAPIURL)
	resp, err := apiClient.Do(req)
	if err != nil {
		log.Printf("Ошибка при отправке HTTP-запроса к OpenRouter: %v", err)
		return "", fmt.Errorf("ошибка при отправке запроса: %w", err)
	}
	defer resp.Body.Close()
	log.Printf("Получен ответ от OpenRouter API со статусом: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка при чтении тела ответа от OpenRouter: %v", err)
		return "", fmt.Errorf("ошибка при чтении ответа: %w", err)
	}
	log.Printf("Тело ответа от OpenRouter: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API вернуло ошибку (статус %d): %s", resp.StatusCode, string(body))
	}

	var aiResp AIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		log.Printf("Ошибка демаршалинга JSON ответа от OpenRouter: %v. Ответ: %s", err, string(body))
		return "", fmt.Errorf("ошибка при демаршалинге JSON: %w. Ответ от API: %s", err, string(body))
	}

	// Извлечение категории из ответа
	if len(aiResp.Choices) > 0 && aiResp.Choices[0].Message.Content != "" {
		category := strings.TrimSpace(aiResp.Choices[0].Message.Content)
		log.Printf("Извлечена категория от AI: \"%s\"", category)

		// Проверяем, есть ли полученная категория в нашем списке допустимых категорий.
		// Это защищает от "галлюцинаций" модели, когда она придумывает свою категорию.
		for _, validCat := range categories {
			// Сравниваем без учета регистра, на случай если модель вернет "продукты" вместо "Продукты"
			if strings.EqualFold(category, validCat) {
				log.Printf("Категория '%s' валидна. Возвращаем каноническое название: '%s'", category, validCat)
				return validCat, nil // Возвращаем категорию с правильным регистром из нашего списка
			}
		}

		// Если цикл завершился, а категория не найдена
		log.Printf("ВНИМАНИЕ: Модель вернула категорию '%s', которой нет в списке.", category)
		return "", fmt.Errorf("модель вернула невалидную категорию: %s", category)
	} else {
		log.Println("Ответ от OpenRouter пустой или в некорректном формате.")
	}

	log.Println("Не удалось извлечь категорию из ответа OpenRouter.")
	return "", fmt.Errorf("не удалось получить корректный ответ от API")
}
