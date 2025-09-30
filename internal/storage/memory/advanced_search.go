package memory

import (
	"github/tahcohcat/same-same/internal/models"
	"sort"

	"github.com/sirupsen/logrus"
)

// AdvancedSearch performs filtered vector search with metadata filtering
func (ms *Storage) AdvancedSearch(req *models.AdvancedSearchRequest, queryEmbedding []float64) ([]*models.SearchResult, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []*models.SearchResult
	evaluator := models.NewFilterEvaluator()
	queryVector := &models.Vector{Embedding: queryEmbedding}

	ctxLog := logrus.WithFields(logrus.Fields{
		"query_length": len(queryEmbedding),
		"filters":      len(req.Filters),
	})

	for _, vector := range ms.vectors {
		// Check embedding dimension compatibility
		if len(vector.Embedding) != len(queryEmbedding) {
			ctxLog.WithFields(logrus.Fields{
				"skipped_vector_id":     vector.ID,
				"skipped_vector_length": len(vector.Embedding),
			}).Warn("skipping vector due to embedding length mismatch")
			continue
		}

		// Apply metadata filters
		if !evaluator.Evaluate(vector.Metadata, req.Filters) {
			ctxLog.WithFields(logrus.Fields{
				"skipped_vector_id":       vector.ID,
				"skipped_vector_metadata": vector.Metadata,
			}).Debug("skipping vector due to metadata filter mismatch")
			continue
		}

		// Calculate similarity score
		vectorScore := queryVector.CosineSimilarity(vector)

		// Apply hybrid weighting if specified
		finalScore := vectorScore
		if req.Options != nil && req.Options.HybridWeight != nil {
			hw := req.Options.HybridWeight
			metadataScore := ms.calculateMetadataScore(vector.Metadata, req.Filters)
			finalScore = (hw.Vector * vectorScore) + (hw.Metadata * metadataScore)
		}

		results = append(results, &models.SearchResult{
			Vector: vector,
			Score:  finalScore,
		})
	}

	ctxLog.WithField("matched_vectors", len(results)).Debug("advanced search completed")

	// Sort by score descending
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

// calculateMetadataScore provides a simple metadata matching score
// Returns 1.0 if all filters match perfectly, 0.0 otherwise
func (ms *Storage) calculateMetadataScore(metadata map[string]string, filters map[string]models.FilterExpr) float64 {
	if len(filters) == 0 {
		return 1.0
	}

	evaluator := models.NewFilterEvaluator()
	if evaluator.Evaluate(metadata, filters) {
		return 1.0
	}

	return 0.0
}
