package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github/tahcohcat/same-same/internal/models"
	"github/tahcohcat/same-same/internal/storage/memory"
)

type VectorHandler struct {
	storage *memory.Storage
}

func NewVectorHandler(storage *memory.Storage) *VectorHandler {
	return &VectorHandler{storage: storage}
}

func (vh *VectorHandler) CreateVector(w http.ResponseWriter, r *http.Request) {
	var vector models.Vector
	if err := json.NewDecoder(r.Body).Decode(&vector); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
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

func (vh *VectorHandler) SearchVectors(w http.ResponseWriter, r *http.Request) {
	var req models.SearchRequest
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
