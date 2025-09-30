package models

import (
	"testing"
)

func TestFilterEvaluator_Eq(t *testing.T) {
	fe := NewFilterEvaluator()

	tests := []struct {
		name     string
		value    string
		exists   bool
		expr     FilterExpr
		expected bool
	}{
		{"exact match", "Einstein", true, FilterExpr{"eq": "Einstein"}, true},
		{"no match", "Einstein", true, FilterExpr{"eq": "Newton"}, false},
		{"field missing", "", false, FilterExpr{"eq": "Einstein"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fe.evaluateExpression(tt.value, tt.exists, tt.expr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFilterEvaluator_Neq(t *testing.T) {
	fe := NewFilterEvaluator()

	tests := []struct {
		name     string
		value    string
		exists   bool
		expr     FilterExpr
		expected bool
	}{
		{"not equal", "Einstein", true, FilterExpr{"neq": "Newton"}, true},
		{"equal fails", "Einstein", true, FilterExpr{"neq": "Einstein"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fe.evaluateExpression(tt.value, tt.exists, tt.expr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFilterEvaluator_Comparison(t *testing.T) {
	fe := NewFilterEvaluator()

	tests := []struct {
		name     string
		value    string
		exists   bool
		expr     FilterExpr
		expected bool
	}{
		{"gt true", "1950", true, FilterExpr{"gt": 1900}, true},
		{"gt false", "1850", true, FilterExpr{"gt": 1900}, false},
		{"gte true", "1900", true, FilterExpr{"gte": 1900}, true},
		{"lt true", "1850", true, FilterExpr{"lt": 1900}, true},
		{"lt false", "1950", true, FilterExpr{"lt": 1900}, false},
		{"lte true", "1900", true, FilterExpr{"lte": 1900}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fe.evaluateExpression(tt.value, tt.exists, tt.expr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFilterEvaluator_Between(t *testing.T) {
	fe := NewFilterEvaluator()

	tests := []struct {
		name     string
		value    string
		exists   bool
		expr     FilterExpr
		expected bool
	}{
		{"in range", "1925", true, FilterExpr{"between": []interface{}{1900, 1950}}, true},
		{"below range", "1850", true, FilterExpr{"between": []interface{}{1900, 1950}}, false},
		{"above range", "2000", true, FilterExpr{"between": []interface{}{1900, 1950}}, false},
		{"at lower bound", "1900", true, FilterExpr{"between": []interface{}{1900, 1950}}, true},
		{"at upper bound", "1950", true, FilterExpr{"between": []interface{}{1900, 1950}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fe.evaluateExpression(tt.value, tt.exists, tt.expr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFilterEvaluator_Contains(t *testing.T) {
	fe := NewFilterEvaluator()

	tests := []struct {
		name     string
		value    string
		exists   bool
		expr     FilterExpr
		expected bool
	}{
		{"contains substring", "science,physics", true, FilterExpr{"contains": "science"}, true},
		{"case insensitive", "SCIENCE", true, FilterExpr{"contains": "science"}, true},
		{"not contains", "mathematics", true, FilterExpr{"contains": "science"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fe.evaluateExpression(tt.value, tt.exists, tt.expr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFilterEvaluator_In(t *testing.T) {
	fe := NewFilterEvaluator()

	tests := []struct {
		name     string
		value    string
		exists   bool
		expr     FilterExpr
		expected bool
	}{
		{"in list", "Einstein", true, FilterExpr{"in": []interface{}{"Einstein", "Bohr", "Heisenberg"}}, true},
		{"not in list", "Newton", true, FilterExpr{"in": []interface{}{"Einstein", "Bohr"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fe.evaluateExpression(tt.value, tt.exists, tt.expr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFilterEvaluator_Exists(t *testing.T) {
	fe := NewFilterEvaluator()

	tests := []struct {
		name     string
		value    string
		exists   bool
		expr     FilterExpr
		expected bool
	}{
		{"field exists", "value", true, FilterExpr{"exists": true}, true},
		{"field missing", "", false, FilterExpr{"exists": false}, true},
		{"field exists but should not", "value", true, FilterExpr{"exists": false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fe.evaluateExpression(tt.value, tt.exists, tt.expr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFilterEvaluator_ComplexFilter(t *testing.T) {
	fe := NewFilterEvaluator()

	metadata := map[string]string{
		"author": "Einstein",
		"year":   "1925",
		"tags":   "physics,relativity,science",
	}

	filters := map[string]FilterExpr{
		"author": {"eq": "Einstein"},
		"year":   {"gte": 1900, "lte": 1950},
		"tags":   {"contains": "science"},
	}

	result := fe.Evaluate(metadata, filters)
	if !result {
		t.Error("expected complex filter to match")
	}
}
