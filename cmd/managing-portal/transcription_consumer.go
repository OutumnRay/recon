package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"Recontext.online/pkg/database"
	amqp "github.com/rabbitmq/amqp091-go"
)

// TranscriptionResult represents the result message from transcription service
type TranscriptionResult struct {
	Event        string                 `json:"event"`
	TrackID      string                 `json:"track_id"`
	UserID       string                 `json:"user_id"`
	AudioURL     string                 `json:"audio_url"`
	JSONURL      string                 `json:"json_url"`
	Transcription TranscriptionData     `json:"transcription"`
	Timestamp    string                 `json:"timestamp"`
	Status       string                 `json:"status"`
}

// TranscriptionData contains the transcription details
type TranscriptionData struct {
	Phrases      []TranscriptionPhrase `json:"phrases"`
	PhraseCount  int                   `json:"phrase_count"`
	TotalDuration float64              `json:"total_duration"`
}

// TranscriptionPhrase represents a single transcribed phrase
type TranscriptionPhrase struct {
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Language   string  `json:"language"`
}

// StartTranscriptionConsumer запускает потребителя результатов транскрибации из RabbitMQ
// Функция подключается к RabbitMQ и обрабатывает результаты транскрибации в фоновом режиме
func StartTranscriptionConsumer() {
	// Получаем параметры подключения к RabbitMQ из переменных окружения
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://recontext:je9rO4k6CQ3M@5.129.227.21:5672/"
	}

	resultQueue := os.Getenv("RABBITMQ_RESULT_QUEUE")
	if resultQueue == "" {
		resultQueue = "transcription_results"
	}

	log.Printf("Starting transcription result consumer...")
	log.Printf("RabbitMQ URL: %s", rabbitmqURL)
	log.Printf("Result Queue: %s", resultQueue)

	// Подключаемся к RabbitMQ с повторными попытками
	var conn *amqp.Connection
	var err error

	for retries := 0; retries < 5; retries++ {
		conn, err = amqp.Dial(rabbitmqURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/5): %v", retries+1, err)
		time.Sleep(time.Duration(retries+1) * 2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after 5 attempts: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	// Объявляем очередь результатов
	q, err := ch.QueueDeclare(
		resultQueue, // name
		true,        // durable - очередь переживёт перезапуск брокера
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Устанавливаем QoS для обработки одного сообщения за раз
	err = ch.Qos(
		1,     // prefetch count - обрабатываем по одному сообщению
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	// Начинаем потребление сообщений
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack - используем ручное подтверждение для надёжности
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Printf("✅ Transcription consumer started, waiting for results...")

	// Обрабатываем сообщения в фоновой горутине
	go func() {
		for msg := range msgs {
			processTranscriptionResult(msg.Body)
			msg.Ack(false) // Подтверждаем обработку сообщения
		}
	}()

	// Блокируем выполнение навсегда
	select {}
}

// processTranscriptionResult обрабатывает сообщение с результатом транскрибации
// Извлекает данные из JSON и обновляет статус трека в базе данных
func processTranscriptionResult(body []byte) {
	var result TranscriptionResult

	err := json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Error parsing transcription result: %v", err)
		return
	}

	log.Print("\n" + strings.Repeat("=", 60))
	log.Printf("📥 Received transcription result:")
	log.Printf("  Track ID: %s", result.TrackID)
	log.Printf("  User ID: %s", result.UserID)
	log.Printf("  Status: %s", result.Status)
	log.Printf("  JSON URL: %s", result.JSONURL)
	log.Printf("  Phrases: %d", result.Transcription.PhraseCount)
	log.Printf("  Duration: %.2f seconds", result.Transcription.TotalDuration)
	log.Print(strings.Repeat("=", 60) + "\n")

	// Обновляем базу данных с результатами транскрибации
	updateTrackTranscriptionStatus(result.TrackID, result.UserID, result.JSONURL, result.Transcription.PhraseCount, result.Transcription.TotalDuration, result.Transcription.Phrases)
}

// updateTrackTranscriptionStatus обновляет трек информацией о транскрибации
// Сохраняет URL JSON-файла, количество фраз, длительность и сами фразы в базу данных
func updateTrackTranscriptionStatus(trackID string, userID string, jsonURL string, phraseCount int, duration float64, phrases []TranscriptionPhrase) {
	log.Printf("📝 Updating track %s with transcription data", trackID)
	log.Printf("   JSON URL: %s", jsonURL)
	log.Printf("   Phrases: %d", phraseCount)
	log.Printf("   Duration: %.2f", duration)

	// Получаем подключение к базе данных из глобального контекста
	// Используем конфигурацию из переменных окружения
	dbConfig := database.Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvIntOrDefault("DB_PORT", 5432),
		User:     getEnvOrDefault("DB_USER", "recontext"),
		Password: getEnvOrDefault("DB_PASSWORD", "recontext"),
		DBName:   getEnvOrDefault("DB_NAME", "recontext"),
		SSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
	}

	db, err := database.NewDB(dbConfig)
	if err != nil {
		log.Printf("❌ Failed to connect to database: %v", err)
		return
	}

	// Начинаем транзакцию для атомарного обновления
	tx := db.DB.Begin()
	if tx.Error != nil {
		log.Printf("❌ Failed to start transaction: %v", tx.Error)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Обновляем статус транскрибации в таблице livekit_tracks
	result := tx.Exec(`
		UPDATE livekit_tracks
		SET
			transcription_status = 'completed',
			transcription_url = $1,
			transcription_phrase_count = $2,
			transcription_duration = $3,
			updated_at = NOW()
		WHERE id = $4
	`, jsonURL, phraseCount, duration, trackID)

	if result.Error != nil {
		tx.Rollback()
		log.Printf("❌ Failed to update track %s: %v", trackID, result.Error)
		return
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		log.Printf("⚠️ Track %s not found in database", trackID)
		return
	}

	log.Printf("✅ Track %s updated successfully", trackID)

	// Удаляем старые фразы для этого трека (если есть)
	deleteResult := tx.Exec(`DELETE FROM transcription_phrases WHERE track_id = $1`, trackID)
	if deleteResult.Error != nil {
		tx.Rollback()
		log.Printf("❌ Failed to delete old phrases for track %s: %v", trackID, deleteResult.Error)
		return
	}
	if deleteResult.RowsAffected > 0 {
		log.Printf("🗑️ Deleted %d old phrases for track %s", deleteResult.RowsAffected, trackID)
	}

	// Получаем информацию о треке и комнате для расчета абсолютного времени
	// Get track and room information to calculate absolute time from meeting start
	var trackPublishedAt time.Time
	var roomStartedAt time.Time
	err = tx.Raw(`
		SELECT
			t.published_at as track_published_at,
			r.started_at as room_started_at
		FROM livekit_tracks t
		JOIN livekit_rooms r ON t.room_sid = r.sid
		WHERE t.id = $1
	`, trackID).Row().Scan(&trackPublishedAt, &roomStartedAt)

	if err != nil {
		tx.Rollback()
		log.Printf("❌ Failed to get track/room times for track %s: %v", trackID, err)
		return
	}

	// Вычисляем смещение трека относительно начала встречи (в секундах)
	// Calculate track offset from meeting start (in seconds)
	trackOffsetSeconds := trackPublishedAt.Sub(roomStartedAt).Seconds()
	log.Printf("📊 Track offset from meeting start: %.2f seconds", trackOffsetSeconds)

	// Сохраняем фразы в таблицу transcription_phrases
	if len(phrases) > 0 {
		log.Printf("💾 Saving %d transcription phrases to database...", len(phrases))

		// Подготавливаем пакетную вставку
		for i, phrase := range phrases {
			// Вычисляем абсолютное время относительно начала встречи
			// Calculate absolute time from meeting start
			absoluteStart := trackOffsetSeconds + phrase.Start
			absoluteEnd := trackOffsetSeconds + phrase.End

			insertResult := tx.Exec(`
				INSERT INTO transcription_phrases
				(track_id, user_id, phrase_index, start_time, end_time, absolute_start_time, absolute_end_time, text, confidence, language, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
			`, trackID, userID, i, phrase.Start, phrase.End, absoluteStart, absoluteEnd, phrase.Text, phrase.Confidence, phrase.Language)

			if insertResult.Error != nil {
				tx.Rollback()
				log.Printf("❌ Failed to insert phrase %d for track %s: %v", i, trackID, insertResult.Error)
				return
			}
		}

		log.Printf("✅ Saved %d phrases to database", len(phrases))
	}

	// Фиксируем транзакцию
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Failed to commit transaction: %v", err)
		return
	}

	log.Printf("✅ Transcription data saved successfully for track %s", trackID)
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault возвращает значение переменной окружения как int или значение по умолчанию
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
			return intVal
		}
	}
	return defaultValue
}
