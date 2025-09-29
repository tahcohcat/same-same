package memory

import (
	"github/tahcohcat/same-same/internal/models"

	"testing"
)

func TestMatchesMetadata(t *testing.T) {
	tests := []struct {
		vectorMeta map[string]string
		queryMeta  map[string]string
		want       bool
	}{
		{map[string]string{"a": "1", "b": "2"}, map[string]string{"a": "1"}, true},
		{map[string]string{"a": "1", "b": "2"}, map[string]string{"b": "2"}, true},
		{map[string]string{"a": "1", "b": "2"}, map[string]string{"a": "2"}, false},
		{map[string]string{"a": "1", "b": "2"}, map[string]string{"c": "3"}, false},
		{map[string]string{"a": "1"}, map[string]string{}, true},
	}
	for i, tt := range tests {
		got := matchesMetadata(tt.vectorMeta, tt.queryMeta)
		if got != tt.want {
			t.Errorf("test %d: MatchesMetadata(%v, %v) = %v, want %v", i, tt.vectorMeta, tt.queryMeta, got, tt.want)
		}
	}
}

func TestSearchBasic(t *testing.T) {
	store := NewStorage()

	vec1 := &models.Vector{ID: "v1", Embedding: []float64{1, 0, 0}}
	vec2 := &models.Vector{ID: "v2", Embedding: []float64{0, 1, 0}}
	vec3 := &models.Vector{ID: "v3", Embedding: []float64{0, 0, 1}}

	_ = store.Store(vec1)
	_ = store.Store(vec2)
	_ = store.Store(vec3)

	req := &models.SearchByEmbbedingRequest{
		Embedding: []float64{1, 0, 0},
		Limit:     2,
	}
	results, err := store.Search(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if results[0].Vector.ID != "v1" {
		t.Errorf("expected first result to be v1, got %s", results[0].Vector.ID)
	}
}

func TestSearch_EmbeddingLengthMismatch(t *testing.T) {
	store := NewStorage()
	vec := &models.Vector{ID: "v1", Embedding: []float64{1, 2, 3}}
	_ = store.Store(vec)

	req := &models.SearchByEmbbedingRequest{
		Embedding: []float64{1, 2}, // length mismatch
	}
	results, err := store.Search(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
