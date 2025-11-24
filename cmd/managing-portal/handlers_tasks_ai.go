package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/auth"
	"github.com/google/uuid"
)

// OpenAI API configuration
const (
	OpenAIAPIURL = "https://api.openai.com/v1/chat/completions"
)

// OpenAI Request/Response structures
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// @Summary Extract tasks using AI
// @Description Extract tasks from session transcription using OpenAI
// @Tags Tasks
// @Accept json
// @Produce json
// @Param session_id path string true "Session ID (UUID)"
// @Param request body models.ExtractTasksRequest true "AI extraction parameters"
// @Success 200 {object} models.ExtractTasksResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/sessions/{session_id}/tasks/extract [post]
func (mp *ManagingPortal) extractTasksHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		mp.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Get session_id from URL
	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		mp.respondWithError(w, http.StatusBadRequest, "session_id is required", "")
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid session_id format", err.Error())
		return
	}

	// Parse request body
	var req models.ExtractTasksRequest
	if err := parseJSONBody(r, &req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate provider (only OpenAI for now)
	if req.LLMProvider != "openai" {
		mp.respondWithError(w, http.StatusBadRequest, "Only 'openai' provider is supported", "")
		return
	}

	// Get transcription for session
	transcription, err := mp.getSessionTranscription(sessionID)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to get transcription", err.Error())
		return
	}

	if transcription == "" {
		mp.respondWithError(w, http.StatusNotFound, "No transcription found for this session", "")
		return
	}

	// Extract tasks using OpenAI
	extractedTasks, err := mp.extractTasksWithOpenAI(transcription, req.Model)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to extract tasks", err.Error())
		return
	}

	// Process extracted tasks
	savedTasks := []models.TaskWithDetails{}
	skippedCount := 0

	for _, extracted := range extractedTasks {
		// Skip if confidence is below threshold
		if extracted.Confidence < req.MinConfidence {
			skippedCount++
			continue
		}

		// Create task
		task := &models.Task{
			SessionID:     sessionID,
			Title:         extracted.Title,
			Description:   &extracted.Description,
			Hint:          extracted.Hint,
			Priority:      getPriorityFromString(extracted.Priority),
			Status:        models.TaskStatusPending,
			ExtractedByAI: true,
			AIConfidence:  &extracted.Confidence,
			SourceSegment: &extracted.SourceSegment,
		}

		// Auto-assign if username provided
		if req.AutoAssign && extracted.AssignedToUsername != nil {
			userID, err := mp.taskRepo.GetUserByUsername(*extracted.AssignedToUsername)
			if err == nil {
				task.AssignedTo = userID
			}
		}

		// Set assigned_by to current user
		task.AssignedBy = &claims.UserID

		// Save task
		if err := mp.taskRepo.CreateTask(task); err != nil {
			mp.logger.Errorf("Failed to save extracted task: %v", err)
			continue
		}

		// Get task with details
		taskWithDetails, err := mp.taskRepo.GetTaskWithDetails(task.ID)
		if err != nil {
			mp.logger.Errorf("Failed to get task details: %v", err)
			continue
		}

		savedTasks = append(savedTasks, *taskWithDetails)
	}

	response := models.ExtractTasksResponse{
		ExtractedCount: len(extractedTasks),
		SavedCount:     len(savedTasks),
		SkippedCount:   skippedCount,
		Tasks:          savedTasks,
	}

	mp.respondWithJSON(w, http.StatusOK, response)
}

// getSessionTranscription retrieves the full transcription text for a session
func (mp *ManagingPortal) getSessionTranscription(sessionID uuid.UUID) (string, error) {
	// Get transcription phrases from database
	var phrases []struct {
		Text      string
		StartTime float64
		Speaker   *string
	}

	err := mp.db.DB.Table("transcription_phrases").
		Select("text, start_time, speaker").
		Where("session_id = ?", sessionID).
		Order("start_time ASC").
		Find(&phrases).Error

	if err != nil {
		return "", err
	}

	if len(phrases) == 0 {
		return "", nil
	}

	// Build transcription text with speaker labels
	var builder strings.Builder
	var lastSpeaker string

	for _, phrase := range phrases {
		speaker := "Speaker"
		if phrase.Speaker != nil && *phrase.Speaker != "" {
			speaker = *phrase.Speaker
		}

		// Add speaker label if changed
		if speaker != lastSpeaker {
			if builder.Len() > 0 {
				builder.WriteString("\n\n")
			}
			builder.WriteString(fmt.Sprintf("%s: ", speaker))
			lastSpeaker = speaker
		} else {
			builder.WriteString(" ")
		}

		builder.WriteString(phrase.Text)
	}

	return builder.String(), nil
}

// extractTasksWithOpenAI calls OpenAI API to extract tasks from transcription
func (mp *ManagingPortal) extractTasksWithOpenAI(transcription string, model string) ([]models.ExtractedTask, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Build prompt
	systemPrompt := `You are an AI assistant that extracts action items and tasks from meeting transcriptions.

Analyze the transcription and identify:
1. Clear action items or tasks mentioned
2. Who is responsible (if mentioned by name/username)
3. Priority level (low, medium, high, urgent)
4. Any hints or context on how to complete the task

Return a JSON array of tasks with this exact structure:
[
  {
    "title": "Brief task title",
    "description": "Detailed description",
    "hint": "Optional hint on how to solve it",
    "assigned_to_username": "username or null",
    "priority": "low|medium|high|urgent",
    "confidence": 0.95,
    "source_segment": "Exact quote from transcription"
  }
]

Rules:
- Only extract explicit tasks or action items
- Confidence should be 0.0-1.0 (how sure you are this is a real task)
- Priority defaults to "medium" if unclear
- assigned_to_username should match names mentioned in the transcript (or null)
- source_segment should be the exact relevant quote

Return ONLY valid JSON array, no explanations.`

	userPrompt := fmt.Sprintf("Meeting transcription:\n\n%s", transcription)

	// Prepare OpenAI request
	openAIReq := OpenAIRequest{
		Model: model,
		Messages: []OpenAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3, // Lower temperature for more consistent output
		ResponseFormat: &struct {
			Type string `json:"type"`
		}{
			Type: "json_object",
		},
	}

	// Marshal request
	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", OpenAIAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse OpenAI response
	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	// Parse tasks from response
	content := openAIResp.Choices[0].Message.Content

	// Handle response_format json_object - might be wrapped
	var tasksWrapper struct {
		Tasks []models.ExtractedTask `json:"tasks"`
	}

	// Try parsing as wrapper first
	if err := json.Unmarshal([]byte(content), &tasksWrapper); err == nil && len(tasksWrapper.Tasks) > 0 {
		return tasksWrapper.Tasks, nil
	}

	// Try parsing as direct array
	var tasks []models.ExtractedTask
	if err := json.Unmarshal([]byte(content), &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks from response: %w\nContent: %s", err, content)
	}

	return tasks, nil
}

// getPriorityFromString converts string to TaskPriority
func getPriorityFromString(priority string) models.TaskPriority {
	switch strings.ToLower(priority) {
	case "low":
		return models.TaskPriorityLow
	case "medium":
		return models.TaskPriorityMedium
	case "high":
		return models.TaskPriorityHigh
	case "urgent":
		return models.TaskPriorityUrgent
	default:
		return models.TaskPriorityMedium
	}
}
