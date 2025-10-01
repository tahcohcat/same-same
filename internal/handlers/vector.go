package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/tahcohcat/same-same/internal/embedders"
	"github.com/tahcohcat/same-same/internal/embedders/quotes/local/tfidf"
	"github.com/tahcohcat/same-same/internal/models"
	"github.com/tahcohcat/same-same/internal/storage"
)

type VectorHandler struct {
	storage  storage.Storage
	embedder embedders.Embedder
}

func NewVectorHandler(storage storage.Storage, embedder embedders.Embedder) *VectorHandler {
	return &VectorHandler{
		storage:  storage,
		embedder: embedder,
	}
}

func (vh *VectorHandler) CreateVector(w http.ResponseWriter, r *http.Request) {
	var vector models.Vector
	if err := json.NewDecoder(r.Body).Decode(&vector); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := vector.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := vh.storage.Store(&vector); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vector)
}

func (vh *VectorHandler) EmbedVector(w http.ResponseWriter, r *http.Request) {
	var quote models.Quote
	if err := json.NewDecoder(r.Body).Decode(&quote); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Generate embedding for the quote text
	fullText := quote.Text + " - " + quote.Author

	var embedding []float64
	var err error

	// Generate embedding
	embedding, err = vh.embedder.Embed(fullText)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate embedding: %v", err), http.StatusInternalServerError)
		return
	}

	// Verify non-zero embedding
	hasNonZero := false
	for _, val := range embedding {
		if val != 0 {
			hasNonZero = true
			break
		}
	}

	if !hasNonZero {
		logrus.Warn("Generated zero-only embedding, vocabulary may need bootstrapping")
	}

	// Create vector with generated ID and metadata
	vector := models.Vector{
		ID:        fmt.Sprintf("quote_%d", time.Now().Unix()),
		Embedding: embedding,
		Metadata: map[string]string{
			"type":          "quote",
			"author":        quote.Author,
			"text":          quote.Text,
			"embedder.name": vh.embedder.Name(),
		},
		CreatedAt: time.Now(), // Set creation time
		UpdatedAt: time.Now(), // Set update time
	}

	if err := vh.storage.Store(&vector); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vector)
}

func (vh *VectorHandler) GetVector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	vector, err := vh.storage.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vector)
}

func (vh *VectorHandler) UpdateVector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var vector models.Vector
	if err := json.NewDecoder(r.Body).Decode(&vector); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	vector.ID = id

	if err := vh.storage.Store(&vector); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vector)
}

func (vh *VectorHandler) DeleteVector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := vh.storage.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (vh *VectorHandler) ListVectors(w http.ResponseWriter, r *http.Request) {
	vectors, err := vh.storage.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vectors)
}

func (vh *VectorHandler) ListVectorMetadata(w http.ResponseWriter, r *http.Request) {
	vectors, err := vh.storage.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	meta := make([]map[string]interface{}, len(vectors))
	for i, vector := range vectors {
		meta[i] = map[string]interface{}{
			"id":         vector.ID,
			"length":     len(vector.Embedding),
			"metadata":   vector.Metadata,
			"created_at": vector.CreatedAt,
			"updated_at": vector.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}

func (vh *VectorHandler) SearchVectors(w http.ResponseWriter, r *http.Request) {
	var req models.SearchByEmbbedingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := vh.storage.Search(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}

func (vh *VectorHandler) SearchByText(w http.ResponseWriter, r *http.Request) {

	var req models.SearchByTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.TopK == 0 {
		req.TopK = 5
	}

	// 1. Embed the text
	embedding, err := vh.embedder.Embed(req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Run similarity search
	results, err := vh.storage.Search(&models.SearchByEmbbedingRequest{
		Embedding: embedding,
		TopK:      req.TopK,
		Filters:   req.MetadataFilters,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !req.ReturnEmbedding {
		for _, res := range results {
			res.Vector.Embedding = nil
		}
	}
	w.Header().Set("Content-Type", "application/json")

	// 3. Return matches
	json.NewEncoder(w).Encode(map[string]interface{}{
		"matches": results,
	})
}

func (vh *VectorHandler) CountVectors(w http.ResponseWriter, r *http.Request) {
	count := vh.storage.Count()

	response := map[string]int{
		"count": count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (vh *VectorHandler) GetEmbedderStats(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]interface{})

	stats["type"] = vh.embedder.Name()

	if vh.embedder.Name() == "local.tfidf" {
		stats["type"] = "local.tfidf"

		if tfidfEmbedder, ok := vh.embedder.(*tfidf.TFIDFEmbedder); ok {
			stats["vocabulary_size"] = tfidfEmbedder.GetVocabularySize()
			stats["document_count"] = tfidfEmbedder.GetDocumentCount()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
