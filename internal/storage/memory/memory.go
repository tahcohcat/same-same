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

		//todo: rethink the criteria for metadata filtering
		// if req.Metadata != nil && !matchesMetadata(vector.Metadata, req.Metadata) {
		// 	ctxLog.WithFields(logrus.Fields{
		// 		"skipped_vector_id": vector.ID,
		// 		"skipped_vector_metadata": vector.Metadata,
		// 	}).Debug("skipping vector due to metadata mismatch")
		// 	continue
		// }

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

	if req.Limit > 0 && len(results) > req.Limit {
		results = results[:req.Limit]
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
