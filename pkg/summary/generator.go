package summary

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// TranscriptSegment представляет сегмент транскрипции
type TranscriptSegment struct {
	ParticipantID   string  `json:"participant_id"`
	ParticipantName string  `json:"participant_name"`
	StartTime       float64 `json:"start_time"`
	EndTime         float64 `json:"end_time"`
	Text            string  `json:"text"`
}

// Summary представляет сводку встречи на одном языке
type Summary struct {
	Language     string   `json:"language"`      // "en" или "ru"
	MainPoints   []string `json:"main_points"`   // Основные тезисы
	KeyDecisions []string `json:"key_decisions"` // Ключевые решения
	ActionItems  []string `json:"action_items"`  // План действий
	FullSummary  string   `json:"full_summary"`  // Полная сводка
}

// MeetingSummary содержит сводки на всех языках
type MeetingSummary struct {
	English *Summary `json:"en"`
	Russian *Summary `json:"ru"`
}

// SummaryGenerator генерирует сводки встреч
type SummaryGenerator struct {
	client *openai.Client
	model  string
}

// NewSummaryGenerator создает новый генератор сводок
func NewSummaryGenerator() (*SummaryGenerator, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini" // По умолчанию используем gpt-4o-mini для экономии
	}

	client := openai.NewClient(apiKey)

	return &SummaryGenerator{
		client: client,
		model:  model,
	}, nil
}

// GenerateSummaries генерирует сводки встречи на русском и английском
func (sg *SummaryGenerator) GenerateSummaries(ctx context.Context, meetingTitle string, segments []TranscriptSegment) (*MeetingSummary, error) {
	log.Printf("📝 Generating meeting summaries for: %s", meetingTitle)
	log.Printf("   Segments to process: %d", len(segments))

	if len(segments) == 0 {
		return nil, fmt.Errorf("no transcript segments provided")
	}

	// Формируем транскрипт
	transcript := sg.formatTranscript(segments)

	// Генерируем сводки на обоих языках параллельно
	englishChan := make(chan *Summary, 1)
	russianChan := make(chan *Summary, 1)
	errorChan := make(chan error, 2)

	// Генерируем английскую сводку
	go func() {
		summary, err := sg.generateSummaryInLanguage(ctx, meetingTitle, transcript, "English")
		if err != nil {
			errorChan <- fmt.Errorf("failed to generate English summary: %w", err)
			return
		}
		englishChan <- summary
	}()

	// Генерируем русскую сводку
	go func() {
		summary, err := sg.generateSummaryInLanguage(ctx, meetingTitle, transcript, "Russian")
		if err != nil {
			errorChan <- fmt.Errorf("failed to generate Russian summary: %w", err)
			return
		}
		russianChan <- summary
	}()

	// Ожидаем результаты
	result := &MeetingSummary{}
	for i := 0; i < 2; i++ {
		select {
		case summary := <-englishChan:
			result.English = summary
			log.Printf("✅ English summary generated")
		case summary := <-russianChan:
			result.Russian = summary
			log.Printf("✅ Russian summary generated")
		case err := <-errorChan:
			log.Printf("⚠️ Error generating summary: %v", err)
			// Продолжаем, даже если одна из сводок не удалась
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled")
		}
	}

	if result.English == nil && result.Russian == nil {
		return nil, fmt.Errorf("failed to generate any summaries")
	}

	log.Printf("✅ Meeting summaries generated successfully")
	return result, nil
}

// formatTranscript форматирует сегменты транскрипции в текст
func (sg *SummaryGenerator) formatTranscript(segments []TranscriptSegment) string {
	var builder strings.Builder

	for _, segment := range segments {
		// Форматируем время
		minutes := int(segment.StartTime) / 60
		seconds := int(segment.StartTime) % 60

		builder.WriteString(fmt.Sprintf("[%02d:%02d] %s: %s\n",
			minutes, seconds, segment.ParticipantName, segment.Text))
	}

	return builder.String()
}

// generateSummaryInLanguage генерирует сводку на указанном языке
func (sg *SummaryGenerator) generateSummaryInLanguage(ctx context.Context, meetingTitle, transcript, language string) (*Summary, error) {
	prompt := sg.buildPrompt(meetingTitle, transcript, language)

	resp, err := sg.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: sg.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: sg.getSystemPrompt(language),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3, // Низкая температура для более точных результатов
		MaxTokens:   2000,
	})

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Парсим ответ
	summary := sg.parseResponse(resp.Choices[0].Message.Content, language)
	return summary, nil
}

// getSystemPrompt возвращает системный промпт для генерации сводки
func (sg *SummaryGenerator) getSystemPrompt(language string) string {
	if language == "Russian" {
		return `Ты - профессиональный ассистент для создания сводок совещаний. Твоя задача - проанализировать транскрипцию встречи и создать структурированную сводку на русском языке.

Создай сводку в следующем формате:

## Основные тезисы
- [тезис 1]
- [тезис 2]
- [тезис 3]

## Ключевые решения
- [решение 1]
- [решение 2]

## План действий
- [действие 1] - [ответственный]
- [действие 2] - [ответственный]

## Полная сводка
[2-3 абзаца с подробным описанием встречи]`
	}

	return `You are a professional meeting summary assistant. Your task is to analyze the meeting transcript and create a structured summary in English.

Create a summary in the following format:

## Main Points
- [point 1]
- [point 2]
- [point 3]

## Key Decisions
- [decision 1]
- [decision 2]

## Action Items
- [action 1] - [responsible person]
- [action 2] - [responsible person]

## Full Summary
[2-3 paragraphs with detailed description of the meeting]`
}

// buildPrompt создает промпт для генерации сводки
func (sg *SummaryGenerator) buildPrompt(meetingTitle, transcript, language string) string {
	if language == "Russian" {
		return fmt.Sprintf(`Проанализируй транскрипцию встречи и создай структурированную сводку на русском языке.

Название встречи: %s

Транскрипция:
%s

Создай сводку в указанном формате.`, meetingTitle, transcript)
	}

	return fmt.Sprintf(`Analyze this meeting transcript and create a structured summary in English.

Meeting title: %s

Transcript:
%s

Create a summary in the specified format.`, meetingTitle, transcript)
}

// parseResponse парсит ответ от OpenAI в структуру Summary
func (sg *SummaryGenerator) parseResponse(content, language string) *Summary {
	summary := &Summary{
		Language:     strings.ToLower(language[:2]), // "en" или "ru"
		MainPoints:   []string{},
		KeyDecisions: []string{},
		ActionItems:  []string{},
		FullSummary:  "",
	}

	// Разбиваем на секции
	sections := strings.Split(content, "##")

	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		lines := strings.Split(section, "\n")
		if len(lines) == 0 {
			continue
		}

		header := strings.ToLower(strings.TrimSpace(lines[0]))
		items := []string{}

		// Собираем элементы списка
		for i := 1; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}

			// Удаляем маркер списка
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "* ")
			line = strings.TrimSpace(line)

			if line != "" {
				items = append(items, line)
			}
		}

		// Распределяем по секциям
		if strings.Contains(header, "main points") || strings.Contains(header, "основные тезисы") {
			summary.MainPoints = items
		} else if strings.Contains(header, "key decisions") || strings.Contains(header, "ключевые решения") {
			summary.KeyDecisions = items
		} else if strings.Contains(header, "action items") || strings.Contains(header, "план действий") {
			summary.ActionItems = items
		} else if strings.Contains(header, "full summary") || strings.Contains(header, "полная сводка") {
			// Для полной сводки берем весь текст после заголовка
			summary.FullSummary = strings.Join(items, "\n\n")
		}
	}

	return summary
}
