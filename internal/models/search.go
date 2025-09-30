package models

import "fmt"

type SearchResult struct {
	Vector *Vector `json:"vector"`
	Score  float64 `json:"score"`
}

type SearchByEmbbedingRequest struct {
	Embedding []float64 `json:"embedding"`
	Limit     int       `json:"limit,omitempty"`

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
	if sr.Limit <= 0 {
		sr.Limit = 10
	}
	return nil
}

type SearchByTextRequest struct {
	Text      string `json:"text"`
	Limit     int    `json:"limit,omitempty"`
	Namespace string `json:"namespace,omitempty"`

	MetadataFilters []MetadataFilter `json:"metadata_filters,omitempty"`

	ReturnEmbedding bool `json:"return_embedding,omitempty"`
}

func (st *SearchByTextRequest) Validate() error {
	if len(st.Text) == 0 {
		return fmt.Errorf("text field cannot be empty")
	}
	if st.Limit <= 0 {
		st.Limit = 10
	}
	switch st.Namespace {
	case "", "quotes", "general":
		return nil
	default:
		return fmt.Errorf("invalid namespace: %s", st.Namespace)
	}
}
