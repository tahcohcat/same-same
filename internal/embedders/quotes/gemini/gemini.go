package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/tahcohcat/same-same/internal/embedders"

	"github.com/sirupsen/logrus"
)

type GeminiEmbedder struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

type EmbedRequest struct {
	Model   string  `json:"model"`
	Content Content `json:"content"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type EmbedResponse struct {
	Embedding Embedding `json:"embedding"`
}

type Embedding struct {
	Values []float64 `json:"values"`
}

func NewGeminiEmbedder(apiKey string) embedders.Embedder {
	return &GeminiEmbedder{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent",
	}
}

func (g *GeminiEmbedder) Embed(text string) ([]float64, error) {
	reqBody := EmbedRequest{
		Model: "models/embedding-001",
		Content: Content{
			Parts: []Part{
				{Text: text},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	u, err := url.Parse(g.baseURL)
	if err != nil {
		panic(err)
	}

	q := u.Query()
	q.Set("key", g.apiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	logrus.WithField("google-api-key", g.apiKey).Infof("Sending request to Gemini API: %s", u.String())

	//req.Header.Set("x-google-api-key", g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var embedResponse EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return embedResponse.Embedding.Values, nil
}

func (g *GeminiEmbedder) Name() string {
	return "gemini"
}
