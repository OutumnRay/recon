package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type OpenAIEmbeddingsClient struct {
	apiKey string
	model  string
}

type embeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// NewOpenAIEmbeddingsClient creates a new OpenAI embeddings client
func NewOpenAIEmbeddingsClient() *OpenAIEmbeddingsClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("OPENAI_EMBEDDING_MODEL")
	if model == "" {
		model = "text-embedding-3-small"
	}

	return &OpenAIEmbeddingsClient{
		apiKey: apiKey,
		model:  model,
	}
}

// GetEmbedding generates an embedding for the given text
func (c *OpenAIEmbeddingsClient) GetEmbedding(text string) ([]float32, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	reqBody := embeddingRequest{
		Input: text,
		Model: c.model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return embResp.Data[0].Embedding, nil
}

// ChunkText splits text into chunks of approximately maxChunkSize characters
func ChunkText(text string, maxChunkSize int) []string {
	if maxChunkSize <= 0 {
		maxChunkSize = 1000
	}

	var chunks []string
	runes := []rune(text)

	for i := 0; i < len(runes); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}
