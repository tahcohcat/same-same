package storage

import "github.com/tahcohcat/same-same/internal/models"

// Storage is the interface for vector storage backends
// Both memory and local file storage should implement this
// You can extend this interface as needed

type Storage interface {
	Store(vector *models.Vector) error
	Get(id string) (*models.Vector, error)
	List() ([]*models.Vector, error)
	Delete(id string) error
	Count() int
	Search(req *models.SearchByEmbbedingRequest) ([]*models.SearchResult, error)
	AdvancedSearch(req *models.AdvancedSearchRequest, queryEmbedding []float64) ([]*models.SearchResult, error)
}
