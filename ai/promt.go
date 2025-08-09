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

// --- Структуры для работы с DeepSeek API (OpenAI-совместимый формат) ---

// DSMessage представляет одно сообщение в диалоге
type DSMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DSRequest представляет тело запроса к API
type DSRequest struct {
	Model    string      `json:"model"`
	Messages []DSMessage `json:"messages"`
}

// DSResponse представляет тело ответа от API
type DSResponse struct {
	Choices []struct {
		Message DSMessage `json:"message"`
	} `json:"choices"`
}

const (
	// URL для DeepSeek Chat API
	modelAPIURL = "https://api.deepseek.com/chat/completions"
)

var (
	// Эта переменная будет хранить API ключ для DeepSeek.
	deepseekAPIKey string
	// Создаем один HTTP-клиент на уровне пакета для переиспользования.
	// Это более эффективно, чем создавать нового клиента на каждый запрос.
	// Также добавляем таймаут, чтобы избежать "зависших" запросов.
	apiClient = &http.Client{Timeout: time.Second * 30}
)

// Init инициализирует пакет ai, устанавливая токен.
func Init(token string) {
	deepseekAPIKey = token
}

// ClassifyTransaction отправляет запрос к API DeepSeek для классификации транзакции.
func ClassifyTransaction(text string, categories []string) (string, error) {
	log.Printf("Начинаем классификацию текста через DeepSeek: \"%s\"", text)

	if deepseekAPIKey == "" {
		return "", fmt.Errorf("API ключ для DeepSeek не установлен")
	}

	// Формируем системный и пользовательский промпты
	systemPrompt := fmt.Sprintf(`Ты — ассистент для классификации трат. Твоя задача - проанализировать текст и определить наиболее подходящую категорию из списка.
Отвечай строго названием одной категории из списка, без лишних слов и знаков препинания.

Список категорий:
- %s`, strings.Join(categories, "\n- "))

	userPrompt := fmt.Sprintf(`Текст для анализа: "%s"`, text)

	// Создаем тело запроса
	requestPayload := DSRequest{
		Model: "deepseek-chat", // Используем основную чат-модель DeepSeek
		Messages: []DSMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	// Преобразование промта в JSON
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		log.Printf("Критическая ошибка при маршалинге JSON для DeepSeek запроса: %v", err)
		return "", fmt.Errorf("ошибка при маршалинге JSON: %w", err)
	}

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", modelAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Критическая ошибка при создании HTTP-запроса к DeepSeek: %v", err)
		return "", fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	// Установка заголовков
	req.Header.Set("Authorization", "Bearer "+deepseekAPIKey)
	req.Header.Set("Content-Type", "application/json")
	log.Println("Заголовки для DeepSeek запроса установлены.")

	log.Printf("Отправка запроса на URL: %s", modelAPIURL)
	resp, err := apiClient.Do(req)
	if err != nil {
		log.Printf("Ошибка при отправке HTTP-запроса к DeepSeek: %v", err)
		return "", fmt.Errorf("ошибка при отправке запроса: %w", err)
	}
	defer resp.Body.Close()
	log.Printf("Получен ответ от DeepSeek API со статусом: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка при чтении тела ответа от DeepSeek: %v", err)
		return "", fmt.Errorf("ошибка при чтении ответа: %w", err)
	}
	log.Printf("Тело ответа от DeepSeek: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API вернуло ошибку (статус %d): %s", resp.StatusCode, string(body))
	}

	var dsResp DSResponse
	if err := json.Unmarshal(body, &dsResp); err != nil {
		log.Printf("Ошибка демаршалинга JSON ответа от DeepSeek: %v. Ответ: %s", err, string(body))
		return "", fmt.Errorf("ошибка при демаршалинге JSON: %w. Ответ от API: %s", err, string(body))
	}

	// Извлечение категории из ответа
	if len(dsResp.Choices) > 0 && dsResp.Choices[0].Message.Content != "" {
		category := strings.TrimSpace(dsResp.Choices[0].Message.Content)
		log.Printf("Извлечена категория: \"%s\"", category)
		return category, nil
	} else {
		log.Println("Ответ от DeepSeek пустой или в некорректном формате.")
	}

	log.Println("Не удалось извлечь категорию из ответа DeepSeek.")
	return "", fmt.Errorf("не удалось получить корректный ответ от API")
}
