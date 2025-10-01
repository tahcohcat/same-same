package memory

import (
	"sort"
	"time"

	"github.com/tahcohcat/same-same/internal/models"

	"github.com/sirupsen/logrus"
)

// TemporalSearch performs vector search with temporal decay
func (ms *Storage) TemporalSearch(req *models.TemporalSearchRequest, queryEmbedding []float64) ([]*models.TemporalSearchResult, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	config := req.GetTemporalConfig()
	scorer := models.NewTemporalScorer(config)
	queryVector := &models.Vector{Embedding: queryEmbedding}

	ctxLog := logrus.WithFields(logrus.Fields{
		"query_length":   len(queryEmbedding),
		"temporal_decay": req.TemporalDecay,
		"lambda":         config.Lambda,
		"reference_time": config.ReferenceTime,
	})

	var results []*models.TemporalSearchResult

	// Apply metadata filters if present
	evaluator := models.NewFilterEvaluator()

	for _, vector := range ms.vectors {
		// Check embedding dimension
		if len(vector.Embedding) != len(queryEmbedding) {
			continue
		}

		// Apply metadata filters
		if len(req.Filters) > 0 {
			if !evaluator.Evaluate(vector.Metadata, req.Filters) {
				continue
			}
		}

		// Calculate base cosine similarity
		baseScore := queryVector.CosineSimilarity(vector)

		// Get document time from metadata
		documentTime := ms.getDocumentTime(vector, config.TimeField)

		// Apply temporal decay
		finalScore := scorer.ApplyDecay(baseScore, documentTime)
		decayFactor := scorer.GetDecayFactor(documentTime)

		results = append(results, &models.TemporalSearchResult{
			Vector:       vector,
			Score:        finalScore,
			BaseScore:    baseScore,
			DecayFactor:  decayFactor,
			DocumentTime: documentTime,
			Age:          models.CalculateAge(documentTime, config.ReferenceTime),
		})
	}

	ctxLog.WithField("matched_vectors", len(results)).Debug("temporal search completed")

	// Sort by final score (with decay applied)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if req.TopK > 0 && len(results) > req.TopK {
		results = results[:req.TopK]
	}

	ctxLog.WithField("returned_vectors", len(results)).Debug("results limited")

	return results, nil
}

// getDocumentTime extracts timestamp from metadata
func (ms *Storage) getDocumentTime(vector *models.Vector, timeField string) time.Time {
	// Try the specified time field
	if timeStr, ok := vector.Metadata[timeField]; ok {
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			return t
		}
	}

	// Fallback to created_at
	if !vector.CreatedAt.IsZero() {
		return vector.CreatedAt
	}

	// Fallback to updated_at
	if !vector.UpdatedAt.IsZero() {
		return vector.UpdatedAt
	}

	// Default to current time (no decay)
	return time.Now()
}
