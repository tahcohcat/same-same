package models

import "fmt"

type SearchResult struct {
	Vector *Vector `json:"vector"`
	Score  float64 `json:"score"`
}

type SearchByEmbbedingRequest struct {
	Embedding []float64 `json:"embedding"`
	TopK      int       `json:"top_K,omitempty"`

	Options *SearchOptions `json:"options,omitempty"`

	Filters []MetadataFilter `json:"filters,omitempty"`
}

// MetadataFilter supports advanced filtering
type MetadataFilter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // =, in, not_in, >=, <=, >, <
	Value    interface{} `json:"value"`
}

func (sr *SearchByEmbbedingRequest) Validate() error {
	if len(sr.Embedding) == 0 {
		return fmt.Errorf("embedding cannot be empty")
	}
	if sr.TopK <= 0 {
		sr.TopK = 10
	}
	return nil
}

type SearchByTextRequest struct {
	Text      string `json:"text"`
	TopK      int    `json:"top_K,omitempty"`
	Namespace string `json:"namespace,omitempty"`

	MetadataFilters []MetadataFilter `json:"metadata_filters,omitempty"`

	ReturnEmbedding bool `json:"return_embedding,omitempty"`
}

func (st *SearchByTextRequest) Validate() error {
	if len(st.Text) == 0 {
		return fmt.Errorf("text field cannot be empty")
	}
	if st.TopK <= 0 {
		st.TopK = 10
	}
	switch st.Namespace {
	case "", "quotes", "general":
		return nil
	default:
		return fmt.Errorf("invalid namespace: %s", st.Namespace)
	}
}
