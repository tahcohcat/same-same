package local

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	SchemaVersion     = "1.0.0"
	DefaultPermission = 0755
	MetadataFile      = "metadata.json"
	CollectionsDir    = "collections"
	EmbeddingsDir     = "embeddings"
	ContentDir        = "content"
)

// LocalStorage implements file-based persistent storage
type LocalStorage struct {
	basePath string
	schema   *StorageSchema
	mu       sync.RWMutex
	logger   *logrus.Logger
}

// NewLocalStorage creates a new local file storage
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	ls := &LocalStorage{
		basePath: basePath,
		logger:   logrus.New(),
	}

	// Create directory structure
	if err := ls.initializeDirectories(); err != nil {
		return nil, fmt.Errorf("failed to initialize directories: %w", err)
	}

	// Load or create schema
	if err := ls.loadOrCreateSchema(); err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return ls, nil
}

// initializeDirectories creates the required directory structure
func (ls *LocalStorage) initializeDirectories() error {
	dirs := []string{
		ls.basePath,
		filepath.Join(ls.basePath, CollectionsDir),
		filepath.Join(ls.basePath, EmbeddingsDir),
		filepath.Join(ls.basePath, ContentDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, DefaultPermission); err != nil {
			return err
		}
	}

	return nil
}

// loadOrCreateSchema loads existing schema or creates a new one
func (ls *LocalStorage) loadOrCreateSchema() error {
	metadataPath := filepath.Join(ls.basePath, MetadataFile)

	// Check if metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		// Create new schema
		ls.schema = &StorageSchema{
			Version:   SchemaVersion,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Metadata: StorageMetadata{
				Name:        "same-same-storage",
				Description: "Local file storage for Same-Same vector database",
				Tags:        []string{"vector", "embeddings", "multimodal"},
				Properties:  make(map[string]string),
			},
			Collections: make(map[string]*Collection),
		}

		return ls.saveSchema()
	}

	// Load existing schema
	file, err := os.Open(metadataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	ls.schema = &StorageSchema{}
	if err := json.NewDecoder(file).Decode(ls.schema); err != nil {
		return err
	}

	ls.logger.WithFields(logrus.Fields{
		"version":     ls.schema.Version,
		"collections": len(ls.schema.Collections),
	}).Info("loaded storage schema")

	return nil
}

// saveSchema persists the schema to disk
// saveSchema persists the schema to disk. Caller must hold the lock.
func (ls *LocalStorage) saveSchema() error {
	ls.schema.UpdatedAt = time.Now()

	metadataPath := filepath.Join(ls.basePath, MetadataFile)
	file, err := os.Create(metadataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ls.schema)
}

// CreateCollection creates a new collection
func (ls *LocalStorage) CreateCollection(name, description string, schema *CollectionSchema) (*Collection, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Check if collection already exists
	if _, exists := ls.schema.Collections[name]; exists {
		return nil, fmt.Errorf("collection %s already exists", name)
	}

	collection := &Collection{
		ID:          name,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Schema:      schema,
		Documents:   make(map[string]*Document),
		Stats: CollectionStats{
			DocumentCount: 0,
			TotalSize:     0,
			LastUpdated:   time.Now(),
		},
	}

	ls.schema.Collections[name] = collection

	// Create collection directory
	collectionPath := filepath.Join(ls.basePath, CollectionsDir, name)
	if err := os.MkdirAll(collectionPath, DefaultPermission); err != nil {
		return nil, err
	}

	// Already holding lock
	if err := ls.saveSchema(); err != nil {
		return nil, err
	}

	ls.logger.WithField("collection", name).Info("created collection")
	return collection, nil
}

// GetCollection retrieves a collection by name
func (ls *LocalStorage) GetCollection(name string) (*Collection, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	collection, exists := ls.schema.Collections[name]
	if !exists {
		return nil, fmt.Errorf("collection %s not found", name)
	}

	return collection, nil
}

// ListCollections returns all collections
func (ls *LocalStorage) ListCollections() []*Collection {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	collections := make([]*Collection, 0, len(ls.schema.Collections))
	for _, collection := range ls.schema.Collections {
		collections = append(collections, collection)
	}

	return collections
}

// StoreDocument stores a document in a collection
func (ls *LocalStorage) StoreDocument(collectionName string, doc *Document) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	collection, exists := ls.schema.Collections[collectionName]
	if !exists {
		return fmt.Errorf("collection %s not found", collectionName)
	}

	// Set document metadata
	now := time.Now()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}
	doc.UpdatedAt = now
	doc.CollectionID = collectionName
	doc.Version++

	// Store document in collection
	collection.Documents[doc.ID] = doc

	// Update collection stats
	collection.Stats.DocumentCount = len(collection.Documents)
	collection.Stats.LastUpdated = now
	collection.UpdatedAt = now

	// Save document to file
	if err := ls.saveDocument(collectionName, doc); err != nil {
		return err
	}

	// Save embeddings separately if present and large
	if doc.Embedding != nil && len(doc.Embedding.Vector) > 0 {
		if err := ls.saveEmbedding(collectionName, doc.ID, doc.Embedding); err != nil {
			return err
		}
		// Reference embedding file instead of storing inline
		doc.Embedding.Path = ls.getEmbeddingPath(collectionName, doc.ID)
		doc.Embedding.Vector = nil // Clear vector to save space
	}

	// Save content files separately for large content
	if doc.Content != nil {
		if err := ls.saveContent(collectionName, doc.ID, doc.Content); err != nil {
			return err
		}
	}

	// Already holding lock
	if err := ls.saveSchema(); err != nil {
		return err
	}

	ls.logger.WithFields(logrus.Fields{
		"collection": collectionName,
		"document":   doc.ID,
		"version":    doc.Version,
	}).Debug("stored document")

	return nil
}

// saveDocument saves a document to its JSON file
func (ls *LocalStorage) saveDocument(collectionName string, doc *Document) error {
	docPath := ls.getDocumentPath(collectionName, doc.ID)

	if err := os.MkdirAll(filepath.Dir(docPath), DefaultPermission); err != nil {
		return err
	}

	file, err := os.Create(docPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}

// saveEmbedding saves embedding vector to a separate binary file
func (ls *LocalStorage) saveEmbedding(collectionName, docID string, embedding *EmbeddingData) error {
	embPath := ls.getEmbeddingPath(collectionName, docID)

	if err := os.MkdirAll(filepath.Dir(embPath), DefaultPermission); err != nil {
		return err
	}

	file, err := os.Create(embPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(embedding)
}

// saveContent saves large content to separate files
func (ls *LocalStorage) saveContent(collectionName, docID string, content *ContentData) error {
	// For binary content (images, audio, video), save to content directory
	if content.Image != nil && content.Image.Path == "" {
		// Placeholder for actual image saving logic
		content.Image.Path = ls.getContentPath(collectionName, docID, "image")
	}
	if content.Audio != nil && content.Audio.Path == "" {
		content.Audio.Path = ls.getContentPath(collectionName, docID, "audio")
	}
	if content.Video != nil && content.Video.Path == "" {
		content.Video.Path = ls.getContentPath(collectionName, docID, "video")
	}
	if content.Binary != nil && content.Binary.Path == "" {
		content.Binary.Path = ls.getContentPath(collectionName, docID, "binary")
	}

	return nil
}

// GetDocument retrieves a document by ID
func (ls *LocalStorage) GetDocument(collectionName, docID string) (*Document, error) {
	ls.mu.RLock()
	collection, exists := ls.schema.Collections[collectionName]
	ls.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}

	// Try to get from memory first
	if doc, exists := collection.Documents[docID]; exists {
		// Load embedding if it was stored separately
		if doc.Embedding != nil && doc.Embedding.Path != "" {
			embedding, err := ls.loadEmbedding(collectionName, docID)
			if err == nil {
				doc.Embedding = embedding
			}
		}
		return doc, nil
	}

	// Load from file
	return ls.loadDocument(collectionName, docID)
}

// loadDocument loads a document from its JSON file
func (ls *LocalStorage) loadDocument(collectionName, docID string) (*Document, error) {
	docPath := ls.getDocumentPath(collectionName, docID)

	file, err := os.Open(docPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var doc Document
	if err := json.NewDecoder(file).Decode(&doc); err != nil {
		return nil, err
	}

	// Load embedding if stored separately
	if doc.Embedding != nil && doc.Embedding.Path != "" {
		embedding, err := ls.loadEmbedding(collectionName, docID)
		if err == nil {
			doc.Embedding = embedding
		}
	}

	return &doc, nil
}

// loadEmbedding loads embedding from separate file
func (ls *LocalStorage) loadEmbedding(collectionName, docID string) (*EmbeddingData, error) {
	embPath := ls.getEmbeddingPath(collectionName, docID)

	file, err := os.Open(embPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var embedding EmbeddingData
	if err := json.NewDecoder(file).Decode(&embedding); err != nil {
		return nil, err
	}

	return &embedding, nil
}

// Path helpers
func (ls *LocalStorage) getDocumentPath(collectionName, docID string) string {
	return filepath.Join(ls.basePath, CollectionsDir, collectionName, fmt.Sprintf("%s.json", docID))
}

func (ls *LocalStorage) getEmbeddingPath(collectionName, docID string) string {
	return filepath.Join(ls.basePath, EmbeddingsDir, collectionName, fmt.Sprintf("%s.json", docID))
}

func (ls *LocalStorage) getContentPath(collectionName, docID, contentType string) string {
	return filepath.Join(ls.basePath, ContentDir, collectionName, docID, contentType)
}

// DeleteDocument deletes a document
func (ls *LocalStorage) DeleteDocument(collectionName, docID string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	collection, exists := ls.schema.Collections[collectionName]
	if !exists {
		return fmt.Errorf("collection %s not found", collectionName)
	}

	delete(collection.Documents, docID)

	// Delete document file
	docPath := ls.getDocumentPath(collectionName, docID)
	os.Remove(docPath)

	// Delete embedding file
	embPath := ls.getEmbeddingPath(collectionName, docID)
	os.Remove(embPath)

	// Update stats
	collection.Stats.DocumentCount = len(collection.Documents)
	collection.Stats.LastUpdated = time.Now()

	// Already holding lock
	return ls.saveSchema()
}

// QueryByMetadata queries documents by metadata filters
func (ls *LocalStorage) QueryByMetadata(collectionName string, filters map[string]interface{}) ([]*Document, error) {
	ls.mu.RLock()
	collection, exists := ls.schema.Collections[collectionName]
	ls.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}

	results := make([]*Document, 0)

	for _, doc := range collection.Documents {
		if ls.matchesFilters(doc.Metadata, filters) {
			results = append(results, doc)
		}
	}

	return results, nil
}

// matchesFilters checks if document metadata matches filters
func (ls *LocalStorage) matchesFilters(metadata map[string]interface{}, filters map[string]interface{}) bool {
	for key, value := range filters {
		metaValue, exists := metadata[key]
		if !exists || metaValue != value {
			return false
		}
	}
	return true
}

// Export exports collection to a file
func (ls *LocalStorage) Export(collectionName, outputPath string) error {
	ls.mu.RLock()
	collection, exists := ls.schema.Collections[collectionName]
	ls.mu.RUnlock()

	if !exists {
		return fmt.Errorf("collection %s not found", collectionName)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(collection)
}

// Import imports collection from a file
func (ls *LocalStorage) Import(collectionName, inputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var collection Collection
	if err := json.NewDecoder(file).Decode(&collection); err != nil {
		return err
	}

	ls.mu.Lock()
	ls.schema.Collections[collectionName] = &collection
	// Already holding lock
	err = ls.saveSchema()
	ls.mu.Unlock()
	return err
}

// Close closes the storage
func (ls *LocalStorage) Close() error {
	return ls.saveSchema()
}

// GetStats returns storage statistics
func (ls *LocalStorage) GetStats() map[string]interface{} {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	totalDocs := 0
	for _, collection := range ls.schema.Collections {
		totalDocs += collection.Stats.DocumentCount
	}

	return map[string]interface{}{
		"version":         ls.schema.Version,
		"collections":     len(ls.schema.Collections),
		"total_documents": totalDocs,
		"created_at":      ls.schema.CreatedAt,
		"updated_at":      ls.schema.UpdatedAt,
	}
}
