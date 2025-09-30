package models

import (
	"fmt"
	"reflect"
	"strings"
)

// FilterExpr represents a filter expression with operators
type FilterExpr map[string]interface{}

// AdvancedSearchRequest extends SearchByEmbeddingRequest with filters
type AdvancedSearchRequest struct {
	Query   string                `json:"query,omitempty"`
	TopK    int                   `json:"top_k,omitempty"`
	Filters map[string]FilterExpr `json:"filters,omitempty"`
	Options *SearchOptions        `json:"options,omitempty"`
}

// SearchOptions for hybrid search weighting
type SearchOptions struct {
	HybridWeight *HybridWeight `json:"hybrid_weight,omitempty"`
}

// HybridWeight controls vector vs metadata scoring
type HybridWeight struct {
	Vector   float64 `json:"vector"`
	Metadata float64 `json:"metadata"`
}

func (asr *AdvancedSearchRequest) Validate() error {
	if asr.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}
	if asr.TopK <= 0 {
		asr.TopK = 10
	}
	
	// Validate hybrid weights if provided
	if asr.Options != nil && asr.Options.HybridWeight != nil {
		hw := asr.Options.HybridWeight
		if hw.Vector < 0 || hw.Vector > 1 || hw.Metadata < 0 || hw.Metadata > 1 {
			return fmt.Errorf("hybrid weights must be between 0 and 1")
		}
		if hw.Vector+hw.Metadata != 1.0 {
			return fmt.Errorf("hybrid weights must sum to 1.0")
		}
	}
	
	return nil
}

// FilterEvaluator handles filter evaluation logic
type FilterEvaluator struct{}

// NewFilterEvaluator creates a new filter evaluator
func NewFilterEvaluator() *FilterEvaluator {
	return &FilterEvaluator{}
}

// Evaluate checks if metadata matches all filters
func (fe *FilterEvaluator) Evaluate(metadata map[string]string, filters map[string]FilterExpr) bool {
	if len(filters) == 0 {
		return true // No filters means match all
	}

	for field, expr := range filters {
		value, exists := metadata[field]
		
		if !fe.evaluateExpression(value, exists, expr) {
			return false
		}
	}
	
	return true
}

// evaluateExpression evaluates a single filter expression
func (fe *FilterEvaluator) evaluateExpression(value string, exists bool, expr FilterExpr) bool {
	for op, expectedVal := range expr {
		switch op {
		case "eq":
			if !exists || value != fmt.Sprint(expectedVal) {
				return false
			}
		case "neq":
			if !exists || value == fmt.Sprint(expectedVal) {
				return false
			}
		case "lt":
			if !exists || !fe.compareLess(value, expectedVal, false) {
				return false
			}
		case "lte":
			if !exists || !fe.compareLess(value, expectedVal, true) {
				return false
			}
		case "gt":
			if !exists || !fe.compareGreater(value, expectedVal, false) {
				return false
			}
		case "gte":
			if !exists || !fe.compareGreater(value, expectedVal, true) {
				return false
			}
		case "between":
			if !exists || !fe.compareBetween(value, expectedVal) {
				return false
			}
		case "contains":
			if !exists || !strings.Contains(strings.ToLower(value), strings.ToLower(fmt.Sprint(expectedVal))) {
				return false
			}
		case "in":
			if !exists || !fe.compareIn(value, expectedVal) {
				return false
			}
		case "exists":
			expectedExists, ok := expectedVal.(bool)
			if !ok {
				return false
			}
			if exists != expectedExists {
				return false
			}
		default:
			return false // Unknown operator
		}
	}
	
	return true
}

// compareLess handles less than comparisons
func (fe *FilterEvaluator) compareLess(value string, expected interface{}, orEqual bool) bool {
	valFloat, valErr := fe.toFloat64(value)
	expFloat, expErr := fe.toFloat64(expected)
	
	if valErr == nil && expErr == nil {
		if orEqual {
			return valFloat <= expFloat
		}
		return valFloat < expFloat
	}
	
	// String comparison fallback
	if orEqual {
		return value <= fmt.Sprint(expected)
	}
	return value < fmt.Sprint(expected)
}

// compareGreater handles greater than comparisons
func (fe *FilterEvaluator) compareGreater(value string, expected interface{}, orEqual bool) bool {
	valFloat, valErr := fe.toFloat64(value)
	expFloat, expErr := fe.toFloat64(expected)
	
	if valErr == nil && expErr == nil {
		if orEqual {
			return valFloat >= expFloat
		}
		return valFloat > expFloat
	}
	
	// String comparison fallback
	if orEqual {
		return value >= fmt.Sprint(expected)
	}
	return value > fmt.Sprint(expected)
}

// compareBetween handles range comparisons
func (fe *FilterEvaluator) compareBetween(value string, expected interface{}) bool {
	rangeSlice, ok := expected.([]interface{})
	if !ok || len(rangeSlice) != 2 {
		return false
	}
	
	valFloat, err := fe.toFloat64(value)
	if err != nil {
		return false
	}
	
	minFloat, err1 := fe.toFloat64(rangeSlice[0])
	maxFloat, err2 := fe.toFloat64(rangeSlice[1])
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	return valFloat >= minFloat && valFloat <= maxFloat
}

// compareIn handles "in" operator (value in list)
func (fe *FilterEvaluator) compareIn(value string, expected interface{}) bool {
	list, ok := expected.([]interface{})
	if !ok {
		return false
	}
	
	for _, item := range list {
		if value == fmt.Sprint(item) {
			return true
		}
	}
	
	return false
}

// toFloat64 converts interface{} to float64
func (fe *FilterEvaluator) toFloat64(val interface{}) (float64, error) {
	v := reflect.ValueOf(val)
	
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.String:
		var f float64
		_, err := fmt.Sscanf(v.String(), "%f", &f)
		return f, err
	default:
		return 0, fmt.Errorf("cannot convert to float64")
	}
}
