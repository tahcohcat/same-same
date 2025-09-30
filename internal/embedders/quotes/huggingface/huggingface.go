package huggingface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tahcohcat/same-same/internal/embedders"
)

type EmbeddingRequest struct {
	Inputs Input `json:"inputs"`
}

type Input struct {
	Source    string   `json:"source_sentence"`
	Sentences []string `json:"sentences"`
}

type Embedder struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
	model      string
}

func NewHuggingFaceEmbedder(apiKey string) embedders.Embedder {
	return &Embedder{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api-inference.huggingface.co/models",
		model:   "sentence-transformers/all-MiniLM-L6-v2",
	}
}

func (h *Embedder) Embed(text string) ([]float64, error) {
	reqBody := EmbeddingRequest{
		Inputs: Input{
			Source:    text,
			Sentences: []string{text},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s", h.baseURL, h.model)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.apiKey))

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var embeddings []float64
	if err := json.NewDecoder(resp.Body).Decode(&embeddings); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embeddings, nil
}

func (h *Embedder) Name() string {
	return "huggingface"
}
