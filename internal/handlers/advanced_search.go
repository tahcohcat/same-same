// File: internal/handlers/advanced_search.go
package handlers

import (
	"encoding/json"
	"github/tahcohcat/same-same/internal/models"
	"net/http"
)

// AdvancedSearchResponse matches the API specification
type AdvancedSearchResponse struct {
	Results []AdvancedSearchResult `json:"results"`
	Total   int                    `json:"total"`
}

// AdvancedSearchResult represents a single search result with flattened metadata
type AdvancedSearchResult struct {
	ID       string                 `json:"id"`
	Text     string                 `json:"text,omitempty"`
	Author   string                 `json:"author,omitempty"`
	Year     interface{}            `json:"year,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"-"` // Additional metadata
}

// AdvancedSearch handles POST /api/v1/search with metadata filtering
func (vh *VectorHandler) AdvancedSearch(w http.ResponseWriter, r *http.Request) {
	var req models.AdvancedSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate embedding for the query text
	embedding, err := vh.embedder.Embed(req.Query)
	if err != nil {
		http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}

	// Perform advanced search with filters
	results, err := vh.storage.AdvancedSearch(&req, embedding)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Transform results to match API specification
	apiResults := make([]AdvancedSearchResult, len(results))
	for i, result := range results {
		apiResults[i] = AdvancedSearchResult{
			ID:    result.Vector.ID,
			Score: result.Score,
		}

		// Extract common metadata fields
		if text, ok := result.Vector.Metadata["text"]; ok {
			apiResults[i].Text = text
		}
		if author, ok := result.Vector.Metadata["author"]; ok {
			apiResults[i].Author = author
		}
		if year, ok := result.Vector.Metadata["year"]; ok {
			apiResults[i].Year = year
		}
		if tags, ok := result.Vector.Metadata["tags"]; ok {
			apiResults[i].Tags = parseTagsString(tags)
		}

		// Store all metadata for potential inline expansion
		metadata := make(map[string]interface{})
		for k, v := range result.Vector.Metadata {
			metadata[k] = v
		}
		apiResults[i].Metadata = metadata
	}

	response := AdvancedSearchResponse{
		Results: apiResults,
		Total:   len(apiResults),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// parseTagsString converts a comma-separated string to a slice
func parseTagsString(tags string) []string {
	if tags == "" {
		return []string{}
	}

	result := []string{}
	for _, tag := range splitAndTrim(tags, ",") {
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

// splitAndTrim splits a string and trims whitespace from each part
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimWhitespace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// splitString is a simple string split helper
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}

	result := []string{}
	current := ""

	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, current)
			current = ""
			i += len(sep) - 1
		} else {
			current += string(s[i])
		}
	}
	result = append(result, current)

	return result
}

// trimWhitespace removes leading and trailing whitespace
func trimWhitespace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
