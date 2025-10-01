package ingestion

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/tahcohcat/same-same/internal/embedders"
	"github.com/tahcohcat/same-same/internal/models"
	"github.com/tahcohcat/same-same/internal/storage"
)

// Ingestor handles the ingestion pipeline
type Ingestor struct {
	source   Source
	embedder embedders.Embedder
	storage  storage.Storage
	config   *SourceConfig
	stats    *Stats
}

// Stats tracks ingestion statistics
type Stats struct {
	TotalRecords    int
	SuccessCount    int
	FailureCount    int
	SkippedCount    int
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	RecordsPerSec   float64
	FailureReasons  map[string]int
	Namespace       string
	StorageType     string
}

// NewIngestor creates a new ingestor
func NewIngestor(source Source, embedder embedders.Embedder, storage storage.Storage, config *SourceConfig) *Ingestor {
	storageType := "memory"
	// Try to determine storage type from the storage interface
	switch storage.(type) {
	default:
		storageType = "memory"
	}
	
	return &Ingestor{
		source:   source,
		embedder: embedder,
		storage:  storage,
		config:   config,
		stats: &Stats{
			FailureReasons: make(map[string]int),
			Namespace:      config.Namespace,
			StorageType:    storageType,
		},
	}
}

// Run executes the ingestion pipeline
func (ing *Ingestor) Run(ctx context.Context) (*Stats, error) {
	ing.stats.StartTime = time.Now()
	
	if err := ing.source.Open(ctx); err != nil {
		return nil, fmt.Errorf("failed to open source: %w", err)
	}
	defer ing.source.Close()
	
	if ing.config.Verbose {
		fmt.Printf("Starting ingestion from: %s\n", ing.source.Name())
	}
	
	batch := make([]*models.Vector, 0, ing.config.BatchSize)
	
	for {
		select {
		case <-ctx.Done():
			return ing.stats, ctx.Err()
		default:
		}
		
		record, err := ing.source.Next()
		if err == io.EOF {
			// Process remaining batch
			if len(batch) > 0 {
				ing.processBatch(batch)
			}
			break
		}
		
		if err != nil {
			ing.stats.FailureCount++
			ing.stats.FailureReasons["read_error"]++
			if ing.config.Verbose {
				fmt.Printf("Error reading record: %v\n", err)
			}
			continue
		}
		
		ing.stats.TotalRecords++
		
		// Skip empty text
		if record.Text == "" {
			ing.stats.SkippedCount++
			continue
		}
		
		// Generate embedding
		var embedding []float64
		
		// Check if this is an image record and embedder supports images
		if record.Metadata["type"] == "image" {
			if imgEmbedder, ok := ing.embedder.(interface {
				EmbedImage(string) ([]float64, error)
			}); ok {
				// Use image embedding
				embedding, err = imgEmbedder.EmbedImage(record.Text)
			} else {
				ing.stats.FailureCount++
				ing.stats.FailureReasons["embedder_not_multimodal"]++
				if ing.config.Verbose {
					fmt.Printf("Embedder does not support images, skipping: %s\n", record.Text)
				}
				continue
			}
		} else {
			// Use text embedding
			embedding, err = ing.embedder.Embed(record.Text)
		}
		if err != nil {
			ing.stats.FailureCount++
			ing.stats.FailureReasons["embed_error"]++
			if ing.config.Verbose {
				textPreview := record.Text
				if len(textPreview) > 50 {
					textPreview = textPreview[:50] + "..."
				}
				fmt.Printf("Error embedding text '%s': %v\n", textPreview, err)
			}
			continue
		}
		
		if ing.config.Verbose && ing.stats.TotalRecords <= 3 {
			fmt.Printf("Successfully embedded record %d with %d dimensions\n", ing.stats.TotalRecords, len(embedding))
		}
		
		// Create vector
		id := record.ID
		if id == "" {
			// Generate ID from text hash or use UUID
			id = fmt.Sprintf("vec_%d_%d", time.Now().UnixNano(), ing.stats.TotalRecords)
		}
		
		vector := &models.Vector{
			ID:        id,
			Embedding: embedding,
			Metadata:  record.Metadata,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		// Add to batch
		batch = append(batch, vector)
		
		// Process batch if full
		if len(batch) >= ing.config.BatchSize {
			ing.processBatch(batch)
			batch = make([]*models.Vector, 0, ing.config.BatchSize)
		}
		
		// Progress indicator
		if ing.config.Verbose && ing.stats.TotalRecords%100 == 0 {
			fmt.Printf("Processed %d records...\n", ing.stats.TotalRecords)
		}
	}
	
	ing.stats.EndTime = time.Now()
	ing.stats.Duration = ing.stats.EndTime.Sub(ing.stats.StartTime)
	
	if ing.stats.Duration.Seconds() > 0 {
		ing.stats.RecordsPerSec = float64(ing.stats.SuccessCount) / ing.stats.Duration.Seconds()
	}
	
	return ing.stats, nil
}

func (ing *Ingestor) processBatch(batch []*models.Vector) {
	if ing.config.DryRun {
		ing.stats.SuccessCount += len(batch)
		if ing.config.Verbose {
			fmt.Printf("[DRY RUN] Would store batch of %d vectors\n", len(batch))
		}
		return
	}
	
	for i, vector := range batch {
		if err := ing.storage.Store(vector); err != nil {
			ing.stats.FailureCount++
			ing.stats.FailureReasons["storage_error"]++
			if ing.config.Verbose {
				fmt.Printf("Error storing vector %d (ID: %s): %v\n", i, vector.ID, err)
			}
			continue
		}
		ing.stats.SuccessCount++
	}
}

// PrintStats prints ingestion statistics
func (s *Stats) Print() {
	fmt.Printf("\n=== Ingestion Complete ===\n")
	fmt.Printf("Total Records:    %d\n", s.TotalRecords)
	fmt.Printf("Successfully Ingested: %d\n", s.SuccessCount)
	fmt.Printf("Failed:           %d\n", s.FailureCount)
	fmt.Printf("Skipped:          %d\n", s.SkippedCount)
	fmt.Printf("Duration:         %v\n", s.Duration)
	fmt.Printf("Speed:            %.2f records/sec\n", s.RecordsPerSec)
	
	if len(s.FailureReasons) > 0 {
		fmt.Printf("\nFailure Breakdown:\n")
		for reason, count := range s.FailureReasons {
			fmt.Printf("  %s: %d\n", reason, count)
		}
	}
	
	fmt.Printf("\nStorage Details:\n")
	fmt.Printf("  Location:       %s\n", s.StorageType)
	if s.Namespace != "" {
		fmt.Printf("  Namespace:      %s\n", s.Namespace)
	}
	if s.StorageType == "memory" {
		fmt.Printf("  Note:           Data is in-memory only (will be lost on restart)\n")
		fmt.Printf("                  Use local file storage for persistence\n")
	}
	fmt.Printf("========================\n")
}
