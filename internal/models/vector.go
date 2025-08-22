package models

import (
	"fmt"
	"math"
	"time"
)

type Vector struct {
	ID          string            `json:"id"`
	Embedding   []float64         `json:"embedding"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func (v *Vector) CosineSimilarity(other *Vector) float64 {
	if len(v.Embedding) != len(other.Embedding) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range v.Embedding {
		dotProduct += v.Embedding[i] * other.Embedding[i]
		normA += v.Embedding[i] * v.Embedding[i]
		normB += other.Embedding[i] * other.Embedding[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (v *Vector) EuclideanDistance(other *Vector) float64 {
	if len(v.Embedding) != len(other.Embedding) {
		return math.Inf(1)
	}

	var sum float64
	for i := range v.Embedding {
		diff := v.Embedding[i] - other.Embedding[i]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

type SearchResult struct {
	Vector *Vector `json:"vector"`
	Score  float64 `json:"score"`
}

type SearchRequest struct {
	Embedding []float64 `json:"embedding"`
	Limit     int       `json:"limit,omitempty"`
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