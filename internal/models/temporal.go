package models

import (
	"fmt"
	"math"
	"time"
)

// TemporalDecayStrength represents the strength of temporal decay
type TemporalDecayStrength string

const (
	DecayStrong TemporalDecayStrength = "strong" // λ = 0.5 (rapid decay)
	DecayMedium TemporalDecayStrength = "medium" // λ = 0.1 (moderate decay)
	DecayWeak   TemporalDecayStrength = "weak"   // λ = 0.01 (slow decay)
	DecayNone   TemporalDecayStrength = "none"   // λ = 0 (no decay)
)

// TemporalSearchRequest extends search with temporal awareness
type TemporalSearchRequest struct {
	Query         string                `json:"query"`
	TopK          int                   `json:"top_k,omitempty"`
	Filters       map[string]FilterExpr `json:"filters,omitempty"`
	TemporalDecay TemporalDecayStrength `json:"temporal_decay,omitempty"` // strong, medium, weak, none
	ReferenceTime *time.Time            `json:"reference_time,omitempty"` // Defaults to now
	TimeField     string                `json:"time_field,omitempty"`     // Metadata field for timestamp
	Options       *SearchOptions        `json:"options,omitempty"`
}

// TemporalConfig holds temporal decay configuration
type TemporalConfig struct {
	Lambda        float64   // Decay rate
	ReferenceTime time.Time // Time to compute decay from
	TimeField     string    // Metadata field containing timestamp
}

func (tsr *TemporalSearchRequest) Validate() error {
	if tsr.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}
	if tsr.TopK <= 0 {
		tsr.TopK = 10
	}
	if tsr.TemporalDecay == "" {
		tsr.TemporalDecay = DecayNone
	}
	if tsr.TimeField == "" {
		tsr.TimeField = "created_at" // Default field
	}

	// Validate decay strength
	switch tsr.TemporalDecay {
	case DecayStrong, DecayMedium, DecayWeak, DecayNone:
		// Valid
	default:
		return fmt.Errorf("invalid temporal_decay value: %s (must be: strong, medium, weak, none)", tsr.TemporalDecay)
	}

	return nil
}

// GetTemporalConfig converts request to config
func (tsr *TemporalSearchRequest) GetTemporalConfig() *TemporalConfig {
	config := &TemporalConfig{
		Lambda:    tsr.GetLambda(),
		TimeField: tsr.TimeField,
	}

	if tsr.ReferenceTime != nil {
		config.ReferenceTime = *tsr.ReferenceTime
	} else {
		config.ReferenceTime = time.Now()
	}

	return config
}

// GetLambda returns the decay rate based on strength
func (tsr *TemporalSearchRequest) GetLambda() float64 {
	switch tsr.TemporalDecay {
	case DecayStrong:
		return 0.5 // Rapid decay: 60% score after 1 year
	case DecayMedium:
		return 0.1 // Moderate decay: 90% score after 1 year
	case DecayWeak:
		return 0.01 // Slow decay: 99% score after 1 year
	case DecayNone:
		return 0.0 // No decay
	default:
		return 0.0
	}
}

// TemporalScorer applies temporal decay to similarity scores
type TemporalScorer struct {
	config *TemporalConfig
}

// NewTemporalScorer creates a new temporal scorer
func NewTemporalScorer(config *TemporalConfig) *TemporalScorer {
	return &TemporalScorer{config: config}
}

// ApplyDecay applies temporal decay to a score
// Formula: score(q,d) = cos(q,d) × e^(-λ·Δt)
// where Δt is in years
func (ts *TemporalScorer) ApplyDecay(cosineSimilarity float64, documentTime time.Time) float64 {
	if ts.config.Lambda == 0 {
		return cosineSimilarity // No decay
	}

	// Calculate time difference in years
	deltaT := ts.config.ReferenceTime.Sub(documentTime).Hours() / (24 * 365.25)

	// Handle future dates (shouldn't decay)
	if deltaT < 0 {
		deltaT = 0
	}

	// Apply exponential decay
	decayFactor := math.Exp(-ts.config.Lambda * deltaT)

	return cosineSimilarity * decayFactor
}

// GetDecayFactor returns just the decay factor (for debugging)
func (ts *TemporalScorer) GetDecayFactor(documentTime time.Time) float64 {
	if ts.config.Lambda == 0 {
		return 1.0
	}

	deltaT := ts.config.ReferenceTime.Sub(documentTime).Hours() / (24 * 365.25)
	if deltaT < 0 {
		deltaT = 0
	}

	return math.Exp(-ts.config.Lambda * deltaT)
}

// TemporalSearchResult extends SearchResult with temporal info
type TemporalSearchResult struct {
	Vector       *Vector   `json:"vector"`
	Score        float64   `json:"score"`         // Final score with decay
	BaseScore    float64   `json:"base_score"`    // Original cosine similarity
	DecayFactor  float64   `json:"decay_factor"`  // Temporal decay applied
	DocumentTime time.Time `json:"document_time"` // Time used for decay
	Age          string    `json:"age,omitempty"` // Human-readable age
}

// CalculateAge returns a human-readable age string
func CalculateAge(t time.Time, reference time.Time) string {
	duration := reference.Sub(t)

	years := int(duration.Hours() / (24 * 365.25))
	if years > 0 {
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}

	months := int(duration.Hours() / (24 * 30.44))
	if months > 0 {
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	days := int(duration.Hours() / 24)
	if days > 0 {
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	hours := int(duration.Hours())
	if hours > 0 {
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	return "just now"
}
