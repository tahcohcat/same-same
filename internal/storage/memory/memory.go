package memory

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github/tahcohcat/same-same/internal/models"

	"github.com/sirupsen/logrus"
)

type Storage struct {
	vectors map[string]*models.Vector
	mu      sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		vectors: make(map[string]*models.Vector),
	}
}

func (ms *Storage) Store(vector *models.Vector) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now()
	if vector.ID == "" {
		return fmt.Errorf("vector ID cannot be empty")
	}

	if _, exists := ms.vectors[vector.ID]; exists {
		vector.UpdatedAt = now
	} else {
		vector.CreatedAt = now
		vector.UpdatedAt = now
	}

	ms.vectors[vector.ID] = vector

	logrus.WithFields(logrus.Fields{
		"vector_id":  vector.ID,
		"created_at": vector.CreatedAt,
	}).Debug("vector stored")

	return nil
}

func (ms *Storage) Get(id string) (*models.Vector, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	vector, exists := ms.vectors[id]
	if !exists {
		return nil, fmt.Errorf("vector with ID %s not found", id)
	}

	logrus.WithFields(logrus.Fields{
		"vector_id":  vector.ID,
		"created_at": vector.CreatedAt,
		"updated_at": vector.UpdatedAt,
	}).Debug("vector found")

	return vector, nil
}

func (ms *Storage) Delete(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.vectors[id]; !exists {
		return fmt.Errorf("vector with ID %s not found", id)
	}

	delete(ms.vectors, id)
	return nil
}

func (ms *Storage) List() ([]*models.Vector, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	vectors := make([]*models.Vector, 0, len(ms.vectors))
	for _, vector := range ms.vectors {
		vectors = append(vectors, vector)
	}

	return vectors, nil
}

func (ms *Storage) Count() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return len(ms.vectors)
}

func (ms *Storage) Search(req *models.SearchByEmbbedingRequest) ([]*models.SearchResult, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []*models.SearchResult

	queryVector := &models.Vector{Embedding: req.Embedding}

	ctxLog := logrus.WithFields(logrus.Fields{
		"query_vector.lenght": len(queryVector.Embedding),
	})

	for _, vector := range ms.vectors {
		if len(vector.Embedding) != len(req.Embedding) {
			ctxLog.WithFields(logrus.Fields{
				"skipped_vector_id":     vector.ID,
				"skipped_vector_length": len(vector.Embedding),
			}).Warn("skipping vector due to embedding length mismatch")
			continue
		}

		if len(req.Filters) > 0 && !matchesAdvancedFilters(vector.Metadata, req.Filters) {
			ctxLog.WithFields(logrus.Fields{
				"skipped_vector_id":       vector.ID,
				"skipped_vector_metadata": vector.Metadata,
			}).Debug("skipping vector due to metadata mismatch")
			continue
		}

		score := queryVector.CosineSimilarity(vector)
		results = append(results, &models.SearchResult{
			Vector: vector,
			Score:  score,
		})
	}

	ctxLog.WithField("matched_vectors", len(results)).Debug("search completed")

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if req.TopK > 0 && len(results) > req.TopK {
		results = results[:req.TopK]
	}

	ctxLog.WithField("returned_vectors", len(results)).Debug("results limited")

	return results, nil
}

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
