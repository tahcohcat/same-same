package models

import (
	"fmt"
	"math"
	"time"

	"github.com/pborman/uuid"
)

type Quote struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}

type Vector struct {
	ID        string            `json:"id"`
	Embedding []float64         `json:"embedding,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func (v *Vector) Validate() error {

	if len(v.Embedding) == 0 {
		return fmt.Errorf("embedding cannot be empty")
	}

	if v.ID == "" {
		v.ID = uuid.New()
	}

	return nil
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
