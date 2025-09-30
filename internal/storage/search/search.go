package search

import (
	"fmt"
	"sort"

	"github.com/tahcohcat/same-same/internal/models"
)

// FilterAndScoreVectors applies advanced filtering and scoring to a slice of vectors.
// It returns the top N results sorted by score.
func FilterAndScoreVectors(vectors []*models.Vector, req *models.SearchByEmbbedingRequest) []*models.SearchResult {
	var results []*models.SearchResult
	queryVector := &models.Vector{Embedding: req.Embedding}

	for _, vector := range vectors {
		if len(vector.Embedding) != len(req.Embedding) {
			continue
		}
		// Advanced filters
		if len(req.Filters) > 0 && !matchesAdvancedFilters(vector.Metadata, req.Filters) {
			continue
		}
		score := queryVector.CosineSimilarity(vector)
		results = append(results, &models.SearchResult{
			Vector: vector,
			Score:  score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}
	if len(results) > topK {
		results = results[:topK]
	}
	return results
}

// matchesMetadata checks legacy metadata equality
func matchesMetadata(vectorMeta, queryMeta map[string]string) bool {
	for key, value := range queryMeta {
		if vectorMeta[key] != value {
			return false
		}
	}
	return true
}

// Advanced filter support
func matchesAdvancedFilters(vectorMeta map[string]string, filters []models.MetadataFilter) bool {
	for _, filter := range filters {
		val, ok := vectorMeta[filter.Field]
		if !ok {
			return false
		}
		switch filter.Operator {
		case "=":
			if val != fmt.Sprintf("%v", filter.Value) {
				return false
			}
		case "in":
			arr, ok := filter.Value.([]interface{})
			if !ok {
				return false
			}
			found := false
			for _, v := range arr {
				if val == fmt.Sprintf("%v", v) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		case "not_in":
			arr, ok := filter.Value.([]interface{})
			if !ok {
				return false
			}
			for _, v := range arr {
				if val == fmt.Sprintf("%v", v) {
					return false
				}
			}
		case ">=":
			if !compareNumeric(val, filter.Value, ">=") {
				return false
			}
		case "<=":
			if !compareNumeric(val, filter.Value, "<=") {
				return false
			}
		case ">":
			if !compareNumeric(val, filter.Value, ">") {
				return false
			}
		case "<":
			if !compareNumeric(val, filter.Value, "<") {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func compareNumeric(a string, b interface{}, op string) bool {
	var af, bf float64
	_, err := fmt.Sscanf(a, "%f", &af)
	if err != nil {
		return false
	}
	switch v := b.(type) {
	case float64:
		bf = v
	case int:
		bf = float64(v)
	case string:
		_, err := fmt.Sscanf(v, "%f", &bf)
		if err != nil {
			return false
		}
	default:
		return false
	}
	switch op {
	case ">=":
		return af >= bf
	case "<=":
		return af <= bf
	case ">":
		return af > bf
	case "<":
		return af < bf
	}
	return false
}
