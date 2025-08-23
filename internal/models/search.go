package models

import "fmt"

type SearchResult struct {
	Vector *Vector `json:"vector"`
	Score  float64 `json:"score"`
}

type SearchRequest struct {
	Embedding []float64         `json:"embedding"`
	Limit     int               `json:"limit,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func (sr *SearchRequest) Validate() error {
	if len(sr.Embedding) == 0 {
		return fmt.Errorf("embedding cannot be empty")
	}
	if sr.Limit <= 0 {
		sr.Limit = 10
	}
	return nil
}
