package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"Recontext.online/pkg/auth"
	"Recontext.online/pkg/database"
	"Recontext.online/pkg/llm"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssistantChatRequest — запрос к AI-ассистенту
type AssistantChatRequest struct {
	Message string `json:"message"`
	FileID  string `json:"file_id,omitempty"`
	// mode: "summary" | "definitions" | "find_videos" | "chat"
	Mode string `json:"mode"`
}

// AssistantFileRef — ссылка на файл в ответе ассистента
type AssistantFileRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// AssistantChatResponse — ответ AI-ассистента
type AssistantChatResponse struct {
	Response        string             `json:"response"`
	ReferencedFiles []AssistantFileRef `json:"referenced_files,omitempty"`
}

const (
	maxTranscriptChars = 3000
	maxPhrasesForLLM   = 150
)

// assistantChatHandler godoc
// @Summary AI-ассистент: чат
// @Description Отправить запрос AI-ассистенту. Режимы: summary (краткое содержание файла), definitions (ключевые определения), find_videos (поиск по видео), chat (свободный диалог)
// @Tags Assistant
// @Accept json
// @Produce json
// @Param request body AssistantChatRequest true "Запрос ассистенту"
// @Success 200 {object} AssistantChatResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/assistant/chat [post]
func (up *UserPortal) assistantChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		up.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	var req AssistantChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if strings.TrimSpace(req.Message) == "" && req.Mode != "summary" && req.Mode != "definitions" {
		up.respondWithError(w, http.StatusBadRequest, "Message is required", "")
		return
	}

	if up.llmClient == nil || !up.llmClient.IsConfigured() {
		up.respondWithError(w, http.StatusServiceUnavailable, "AI assistant not configured", "LLM service is unavailable")
		return
	}

	switch req.Mode {
	case "summary":
		up.handleAssistantSummary(w, claims.UserID, req)
	case "definitions":
		up.handleAssistantDefinitions(w, claims.UserID, req)
	case "find_videos":
		up.handleAssistantFindVideos(w, claims.UserID, req)
	default:
		up.handleAssistantChat(w, claims.UserID, req)
	}
}

// handleAssistantSummary — краткое содержание видео/аудио файла
func (up *UserPortal) handleAssistantSummary(w http.ResponseWriter, userID uuid.UUID, req AssistantChatRequest) {
	if req.FileID == "" {
		up.respondWithError(w, http.StatusBadRequest, "file_id is required for summary mode", "")
		return
	}

	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid file_id", err.Error())
		return
	}

	// Verify ownership and get file
	var file database.UploadedFile
	if err := up.db.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", fileID, userID).
		First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			up.respondWithError(w, http.StatusNotFound, "File not found", "")
			return
		}
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get file", err.Error())
		return
	}

	// Try to use cached summary from file_summaries
	var existingSummary database.FileSummary
	if err := up.db.DB.Where("file_id = ? AND status = 'completed'", fileID).
		First(&existingSummary).Error; err == nil {
		text := existingSummary.SummaryRu
		if text == "" {
			text = existingSummary.Summary
		}
		if text != "" {
			title := file.Title
			if title == "" {
				title = file.OriginalName
			}
			response := fmt.Sprintf("Краткое содержание файла «%s»:\n\n%s", title, text)
			if len(existingSummary.KeyTopics) > 0 {
				response += "\n\nКлючевые темы: " + strings.Join(existingSummary.KeyTopics, ", ")
			}
			respondJSON(w, AssistantChatResponse{Response: response})
			return
		}
	}

	// No cached summary — build from phrases
	transcript, err := up.buildTranscriptForFile(fileID, userID)
	if err != nil {
		up.logger.Errorf("[Assistant:summary] buildTranscript error for file %s: %v", fileID, err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to load transcript", err.Error())
		return
	}

	if transcript == "" {
		up.logger.Infof("[Assistant:summary] No phrases found for file %s (transcript not ready)", fileID)
		up.respondWithError(w, http.StatusUnprocessableEntity, "Transcript not ready", "Transcription is not yet completed for this file")
		return
	}

	title := file.Title
	if title == "" {
		title = file.OriginalName
	}

	up.logger.Infof("[Assistant:summary] Sending %d chars to LLM for file %s", len(transcript), fileID)

	messages := []llm.Message{
		{
			Role: "system",
			Content: "Ты — AI-ассистент платформы Recontext для работы с транскрипциями совещаний и лекций. " +
				"ВАЖНО: отвечай ТОЛЬКО на русском языке, не используй другие языки, латиницу или иероглифы. " +
				"Твоя задача — подробно и структурировано пересказать содержание записи. " +
				"Выдели: 1) Основные темы обсуждения, 2) Ключевые моменты и решения, 3) Выводы. " +
				"Пиши развёрнуто и информативно — минимум 3-5 предложений в каждом разделе.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Сделай подробное краткое содержание записи «%s».\n\nТранскрипция:\n%s", title, transcript),
		},
	}

	result, err := up.llmClient.GenerateChatCompletion(messages)
	if err != nil {
		up.logger.Errorf("[Assistant:summary] LLM error for file %s: %v", fileID, err)
		up.respondWithError(w, http.StatusInternalServerError, "LLM request failed", err.Error())
		return
	}

	respondJSON(w, AssistantChatResponse{Response: result})
}

// handleAssistantDefinitions — ключевые определения и термины из записи
func (up *UserPortal) handleAssistantDefinitions(w http.ResponseWriter, userID uuid.UUID, req AssistantChatRequest) {
	if req.FileID == "" {
		up.respondWithError(w, http.StatusBadRequest, "file_id is required for definitions mode", "")
		return
	}

	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid file_id", err.Error())
		return
	}

	// Verify ownership and get file
	var file database.UploadedFile
	if err := up.db.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", fileID, userID).
		First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			up.respondWithError(w, http.StatusNotFound, "File not found", "")
			return
		}
		up.respondWithError(w, http.StatusInternalServerError, "Failed to get file", err.Error())
		return
	}

	transcript, err := up.buildTranscriptForFile(fileID, userID)
	if err != nil {
		up.logger.Errorf("[Assistant:definitions] buildTranscript error for file %s: %v", fileID, err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to load transcript", err.Error())
		return
	}

	if transcript == "" {
		up.logger.Infof("[Assistant:definitions] No phrases found for file %s (transcript not ready)", fileID)
		up.respondWithError(w, http.StatusUnprocessableEntity, "Transcript not ready", "Transcription is not yet completed for this file")
		return
	}

	title := file.Title
	if title == "" {
		title = file.OriginalName
	}

	up.logger.Infof("[Assistant:definitions] Sending %d chars to LLM for file %s", len(transcript), fileID)

	messages := []llm.Message{
		{
			Role: "system",
			Content: "Ты — AI-ассистент платформы Recontext для работы с транскрипциями совещаний и лекций. " +
				"ВАЖНО: отвечай ТОЛЬКО на русском языке, не используй другие языки, латиницу или иероглифы. " +
				"Твоя задача — извлечь ключевые термины, определения и понятия из транскрипции. " +
				"Для каждого термина дай развёрнутое объяснение (2-3 предложения) в контексте записи. " +
				"Термины пиши по-русски, технические названия можно указать в скобках на языке оригинала.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Извлеки ключевые термины и определения из записи «%s».\n\nТранскрипция:\n%s", title, transcript),
		},
	}

	result, err := up.llmClient.GenerateChatCompletion(messages)
	if err != nil {
		up.logger.Errorf("[Assistant:definitions] LLM error for file %s: %v", fileID, err)
		up.respondWithError(w, http.StatusInternalServerError, "LLM request failed", err.Error())
		return
	}

	respondJSON(w, AssistantChatResponse{Response: result})
}

// handleAssistantFindVideos — поиск записей, в которых упоминается тема
func (up *UserPortal) handleAssistantFindVideos(w http.ResponseWriter, userID uuid.UUID, req AssistantChatRequest) {
	query := strings.TrimSpace(req.Message)
	if query == "" {
		up.respondWithError(w, http.StatusBadRequest, "Search query is required", "")
		return
	}

	// Full-text search across all user's phrases
	type PhraseMatch struct {
		FileID       uuid.UUID
		FileTitle    string
		OriginalName string
		PhraseText   string
		PhraseIndex  int
	}

	var matches []PhraseMatch
	err := up.db.DB.Table("file_transcription_phrases").
		Select("file_transcription_phrases.file_id, uploaded_files.title as file_title, uploaded_files.original_name, file_transcription_phrases.text as phrase_text, file_transcription_phrases.phrase_index").
		Joins("JOIN uploaded_files ON uploaded_files.id = file_transcription_phrases.file_id").
		Where("uploaded_files.user_id = ? AND uploaded_files.deleted_at IS NULL", userID).
		Where("file_transcription_phrases.text ILIKE ?", "%"+query+"%").
		Order("file_transcription_phrases.file_id, file_transcription_phrases.phrase_index").
		Limit(150).
		Scan(&matches).Error

	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Search failed", err.Error())
		return
	}

	if len(matches) == 0 {
		respondJSON(w, AssistantChatResponse{
			Response: fmt.Sprintf("По запросу «%s» в ваших записях ничего не найдено.", query),
		})
		return
	}

	// Group matches by file and collect unique files
	type FileGroup struct {
		ID           string
		Title        string
		OriginalName string
		Snippets     []string
	}
	fileMap := make(map[string]*FileGroup)
	fileOrder := make([]string, 0)

	for _, m := range matches {
		id := m.FileID.String()
		if _, exists := fileMap[id]; !exists {
			name := m.FileTitle
			if name == "" {
				name = m.OriginalName
			}
			fileMap[id] = &FileGroup{ID: id, Title: name}
			fileOrder = append(fileOrder, id)
		}
		g := fileMap[id]
		if len(g.Snippets) < 3 {
			g.Snippets = append(g.Snippets, m.PhraseText)
		}
	}

	// Build context for LLM
	var contextBuilder strings.Builder
	var refFiles []AssistantFileRef
	for _, id := range fileOrder {
		g := fileMap[id]
		refFiles = append(refFiles, AssistantFileRef{ID: g.ID, Title: g.Title})
		contextBuilder.WriteString(fmt.Sprintf("Запись «%s»:\n", g.Title))
		for _, s := range g.Snippets {
			contextBuilder.WriteString("  — " + s + "\n")
		}
		contextBuilder.WriteString("\n")
	}

	messages := []llm.Message{
		{
			Role: "system",
			Content: "Ты — AI-ассистент платформы Recontext для работы с транскрипциями совещаний и лекций. " +
				"ВАЖНО: отвечай ТОЛЬКО на русском языке, не используй другие языки, латиницу или иероглифы. " +
				"Опиши, в каких записях упоминается запрашиваемая тема и в каком контексте. " +
				"Для каждой записи укажи название и кратко объясни, как тема раскрывается в этой записи.",
		},
		{
			Role: "user",
			Content: fmt.Sprintf("В каких записях упоминается «%s»?\n\nНайденные фрагменты:\n%s", query, contextBuilder.String()),
		},
	}

	result, err := up.llmClient.GenerateChatCompletion(messages)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "LLM request failed", err.Error())
		return
	}

	respondJSON(w, AssistantChatResponse{
		Response:        result,
		ReferencedFiles: refFiles,
	})
}

// handleAssistantChat — свободный диалог с ассистентом
func (up *UserPortal) handleAssistantChat(w http.ResponseWriter, userID uuid.UUID, req AssistantChatRequest) {
	var contextText string

	if req.FileID != "" {
		fileID, err := uuid.Parse(req.FileID)
		if err == nil {
			transcript, _ := up.buildTranscriptForFile(fileID, userID)
			if transcript != "" {
				contextText = "Контекст из записи:\n" + transcript + "\n\n"
			}
		}
	}

	messages := []llm.Message{
		{
			Role: "system",
			Content: "Ты — AI-ассистент платформы Recontext для работы с транскрипциями совещаний и лекций. " +
				"ВАЖНО: отвечай ТОЛЬКО на русском языке, не используй другие языки, латиницу или иероглифы. " +
				"Отвечай развёрнуто, чётко и по существу.",
		},
		{
			Role:    "user",
			Content: contextText + req.Message,
		},
	}

	result, err := up.llmClient.GenerateChatCompletion(messages)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "LLM request failed", err.Error())
		return
	}

	respondJSON(w, AssistantChatResponse{Response: result})
}

// buildTranscriptForFile возвращает текст транскрипции файла для передачи в LLM
func (up *UserPortal) buildTranscriptForFile(fileID uuid.UUID, userID uuid.UUID) (string, error) {
	// Verify ownership
	var count int64
	if err := up.db.DB.Model(&database.UploadedFile{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", fileID, userID).
		Count(&count).Error; err != nil {
		return "", err
	}
	if count == 0 {
		return "", fmt.Errorf("file not found or access denied")
	}

	var phrases []database.FileTranscriptionPhrase
	if err := up.db.DB.Where("file_id = ?", fileID).
		Order("phrase_index ASC").
		Limit(maxPhrasesForLLM).
		Find(&phrases).Error; err != nil {
		return "", err
	}

	if len(phrases) == 0 {
		return "", nil
	}

	var sb strings.Builder
	total := 0
	for _, p := range phrases {
		var line string
		if p.Speaker != "" {
			line = fmt.Sprintf("[%s]: %s\n", p.Speaker, p.Text)
		} else {
			line = p.Text + "\n"
		}
		if total+len(line) > maxTranscriptChars {
			sb.WriteString("...[транскрипция обрезана]")
			break
		}
		sb.WriteString(line)
		total += len(line)
	}

	return sb.String(), nil
}

// respondJSON is a helper to write a JSON response
func respondJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}
