package local

import (
	"fmt"
	"time"

	"github.com/tahcohcat/same-same/internal/models"
	"github.com/tahcohcat/same-same/internal/storage/search"
)

// VectorStorageAdapter adapts LocalStorage to work with existing Vector storage interface
type VectorStorageAdapter struct {
	localStorage *LocalStorage
	collection   string
}

// NewVectorStorageAdapter creates an adapter for vector storage
func NewVectorStorageAdapter(basePath, collectionName string) (*VectorStorageAdapter, error) {
	localStorage, err := NewLocalStorage(basePath)
	if err != nil {
		return nil, err
	}

	// Create default vector collection if it doesn't exist
	if _, err := localStorage.GetCollection(collectionName); err != nil {
		schema := &CollectionSchema{
			Fields: map[string]FieldDefinition{
				"type":          {Type: "string", Indexed: true},
				"author":        {Type: "string", Indexed: true},
				"text":          {Type: "string", Indexed: false},
				"embedder.name": {Type: "string", Indexed: true},
			},
			VectorConfig: &VectorConfig{
				Dimension:    768, // Default dimension
				EmbedderType: "local",
				Metric:       "cosine",
			},
		}

		if _, err := localStorage.CreateCollection(collectionName, "Vector embeddings collection", schema); err != nil {
			return nil, err
		}
	}

	return &VectorStorageAdapter{
		localStorage: localStorage,
		collection:   collectionName,
	}, nil
}

// Store stores a vector using the local storage
func (vsa *VectorStorageAdapter) Store(vector *models.Vector) error {
	doc := &Document{
		ID:        vector.ID,
		Type:      TypeText,
		CreatedAt: vector.CreatedAt,
		UpdatedAt: vector.UpdatedAt,
		Metadata:  convertMetadataToInterface(vector.Metadata),
		Embedding: &EmbeddingData{
			Vector:    vector.Embedding,
			Dimension: len(vector.Embedding),
			Model:     getEmbedderName(vector.Metadata),
			CreatedAt: time.Now(),
		},
		Tags: extractTags(vector.Metadata),
	}

	// Extract text content if available
	if text, ok := vector.Metadata["text"]; ok {
		doc.Content = &ContentData{
			Type: TypeText,
			Text: &TextContent{
				Raw:    text,
				Format: "plain",
			},
		}
	}

	return vsa.localStorage.StoreDocument(vsa.collection, doc)
}

// Get retrieves a vector by ID
func (vsa *VectorStorageAdapter) Get(id string) (*models.Vector, error) {
	doc, err := vsa.localStorage.GetDocument(vsa.collection, id)
	if err != nil {
		return nil, err
	}

	return documentToVector(doc), nil
}

// Delete deletes a vector by ID
func (vsa *VectorStorageAdapter) Delete(id string) error {
	return vsa.localStorage.DeleteDocument(vsa.collection, id)
}

// List returns all vectors in the collection
func (vsa *VectorStorageAdapter) List() ([]*models.Vector, error) {
	collection, err := vsa.localStorage.GetCollection(vsa.collection)
	if err != nil {
		return nil, err
	}

	vectors := make([]*models.Vector, 0, len(collection.Documents))
	for _, doc := range collection.Documents {
		vectors = append(vectors, documentToVector(doc))
	}

	return vectors, nil
}

// Count returns the number of vectors
func (vsa *VectorStorageAdapter) Count() int {
	collection, err := vsa.localStorage.GetCollection(vsa.collection)
	if err != nil {
		return 0
	}

	return collection.Stats.DocumentCount
}

// Search performs vector similarity search
func (vsa *VectorStorageAdapter) Search(req *models.SearchByEmbbedingRequest) ([]*models.SearchResult, error) {
	collection, err := vsa.localStorage.GetCollection(vsa.collection)
	if err != nil {
		return nil, err
	}

	queryVector := &models.Vector{Embedding: req.Embedding}
	results := make([]*models.SearchResult, 0)

	for _, doc := range collection.Documents {
		if doc.Embedding == nil {
			continue
		}

		// Load embedding if stored separately
		if len(doc.Embedding.Vector) == 0 && doc.Embedding.Path != "" {
			embedding, err := vsa.localStorage.loadEmbedding(vsa.collection, doc.ID)
			if err != nil {
				continue
			}
			doc.Embedding = embedding
		}

		vector := documentToVector(doc)
		if len(vector.Embedding) != len(req.Embedding) {
			continue
		}

		// Calculate similarity score
		vectorScore := queryVector.CosineSimilarity(vector)

		// Apply hybrid weighting if specified
		finalScore := vectorScore
		if req.Options != nil && req.Options.HybridWeight != nil {
			hw := req.Options.HybridWeight
			metadataScore := 1.0 // Simple metadata match score
			finalScore = (hw.Vector * vectorScore) + (hw.Metadata * metadataScore)
		}

		results = append(results, &models.SearchResult{
			Vector: vector,
			Score:  finalScore,
		})
	}

	// Sort by score
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Limit results
	if req.TopK > 0 && len(results) > req.TopK {
		results = results[:req.TopK]
	}

	return results, nil
}

func (vsa *VectorStorageAdapter) AdvancedSearch(req *models.AdvancedSearchRequest, queryEmbedding []float64) ([]*models.SearchResult, error) {

	// Use shared search utility
	vectors := []*models.Vector{}
	collection, err := vsa.localStorage.GetCollection(vsa.collection)
	if err != nil {
		return nil, err
	}

	for _, doc := range collection.Documents {
		if doc.Embedding == nil {
			continue
		}

		// Load embedding if stored separately
		if len(doc.Embedding.Vector) == 0 && doc.Embedding.Path != "" {
			embedding, err := vsa.localStorage.loadEmbedding(vsa.collection, doc.ID)
			if err != nil {
				continue
			}
			doc.Embedding = embedding
		}

		vector := documentToVector(doc)
		vectors = append(vectors, vector)
	}

	// Convert req.Filters (map[string]models.FilterExpr) to []models.MetadataFilter
	var metadataFilters []models.MetadataFilter
	for key, expr := range req.Filters {

		// Assuming expr is of type models.FilterExpr (likely a struct or map), extract the operator as string.
		// If expr is a struct with a field "Operator", use expr.Operator.
		// If expr is a map, extract the operator string value (e.g., expr["operator"].(string)).
		// Here, let's assume expr is a struct with an Operator field.
		var operator string
		if op, ok := expr["operator"]; ok {
			operator, _ = op.(string)
		}
		metadataFilters = append(metadataFilters, models.MetadataFilter{
			Field:    key,
			Operator: operator,
		})
	}

	advancedReq := &models.SearchByEmbbedingRequest{
		Embedding: queryEmbedding,
		TopK:      req.TopK,
		Filters:   metadataFilters,
		Options:   req.Options,
	}

	searchResults := search.FilterAndScoreVectors(vectors, advancedReq)
	return searchResults, nil
}

// TemporalSearch implements the Storage interface.
// TODO: Replace the implementation with actual logic as needed.
func (v *VectorStorageAdapter) TemporalSearch(*models.TemporalSearchRequest, []float64) ([]*models.TemporalSearchResult, error) {
	// Implement the required logic or return nil/error as a stub
	return nil, nil
}

// Close closes the storage
func (vsa *VectorStorageAdapter) Close() error {
	return vsa.localStorage.Close()
}

// Helper functions

func documentToVector(doc *Document) *models.Vector {
	metadata := make(map[string]string)
	for k, v := range doc.Metadata {
		metadata[k] = fmt.Sprint(v)
	}

	vector := &models.Vector{
		ID:        doc.ID,
		Metadata:  metadata,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}

	if doc.Embedding != nil {
		vector.Embedding = doc.Embedding.Vector
	}

	return vector
}

func convertMetadataToInterface(metadata map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

func convertInterfaceToStringMap(metadata map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range metadata {
		result[k] = fmt.Sprint(v)
	}
	return result
}

func getEmbedderName(metadata map[string]string) string {
	if name, ok := metadata["embedder.name"]; ok {
		return name
	}
	return "unknown"
}

func extractTags(metadata map[string]string) []string {
	if tagsStr, ok := metadata["tags"]; ok {
		tags := []string{}
		current := ""
		for _, ch := range tagsStr {
			if ch == ',' {
				if current != "" {
					tags = append(tags, current)
					current = ""
				}
			} else {
				current += string(ch)
			}
		}
		if current != "" {
			tags = append(tags, current)
		}
		return tags
	}
	return []string{}
}
