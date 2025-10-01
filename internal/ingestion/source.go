package ingestion

import (
	"context"
)

// Record represents a single data record to be ingested
type Record struct {
	ID       string
	Text     string
	Metadata map[string]string
}

// Source defines the interface for data sources
type Source interface {
	// Open prepares the source for reading
	Open(ctx context.Context) error
	
	// Next returns the next record or io.EOF when done
	Next() (*Record, error)
	
	// Close cleans up resources
	Close() error
	
	// Name returns a human-readable name for this source
	Name() string
}

// SourceConfig contains common configuration for all sources
type SourceConfig struct {
	// Namespace to use for ingested vectors
	Namespace string
	
	// BatchSize for bulk operations
	BatchSize int
	
	// DryRun if true, don't actually ingest
	DryRun bool
	
	// Verbose logging
	Verbose bool
}
