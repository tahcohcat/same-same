package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/tahcohcat/same-same/internal/models"
	"github.com/tahcohcat/same-same/internal/storage/search"

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

	// Use shared search utility
	vectors := make([]*models.Vector, 0, len(ms.vectors))
	for _, v := range ms.vectors {
		vectors = append(vectors, v)
	}
	results := search.FilterAndScoreVectors(vectors, req)
	return results, nil
}
