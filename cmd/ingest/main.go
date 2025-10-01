package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tahcohcat/same-same/internal/embedders"
	"github.com/tahcohcat/same-same/internal/embedders/quotes/gemini"
	"github.com/tahcohcat/same-same/internal/embedders/quotes/huggingface"
	"github.com/tahcohcat/same-same/internal/embedders/quotes/local/tfidf"
	"github.com/tahcohcat/same-same/internal/ingestion"
	"github.com/tahcohcat/same-same/internal/storage/memory"
)

func main() {
	// Flags
	var (
		namespace    = flag.String("namespace", "default", "Namespace for ingested vectors")
		batchSize    = flag.Int("batch-size", 100, "Batch size for bulk operations")
		dryRun       = flag.Bool("dry-run", false, "Don't actually ingest, just validate")
		verbose      = flag.Bool("verbose", false, "Verbose logging")
		embedderType = flag.String("embedder", "", "Embedder type (local, gemini, huggingface) - defaults to env EMBEDDER_TYPE or 'local'")
		textCol      = flag.String("text-col", "text", "Column name for text (CSV only)")
		split        = flag.String("split", "train", "Dataset split (HuggingFace only)")
		timeout      = flag.Duration("timeout", 30*time.Minute, "Timeout for ingestion")
		output = flag.String("output", "", "Output file for exported vectors (optional)")
	)
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %s [flags] <source>

Ingest data from various sources into the same-same vector database.

Sources:
  demo, quotes, quotes-small    Built-in datasets from .examples/data
  hf:<dataset>                  HuggingFace dataset (e.g., hf:imdb, hf:squad:v2)
  file.csv                      CSV file (requires -text-col flag)
  file.jsonl                    JSONL file (each line is a JSON object with "text" field)
  file.json                     Same as JSONL

Examples:
  # Ingest built-in demo dataset
  %s demo

  # Ingest with custom namespace
  %s -namespace quotes demo

  # Ingest HuggingFace dataset
  %s hf:imdb

  # Ingest from CSV file
  %s -text-col content data.csv

  # Ingest from JSONL file
  %s data.jsonl

  # Dry run to validate data
  %s -dry-run -verbose data.jsonl

  # Use specific embedder
  %s -embedder gemini demo

Flags:
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
	
	flag.Parse()
	
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	
	sourceArg := flag.Arg(0)
	
	// Setup logging
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	
	// Create config
	config := &ingestion.SourceConfig{
		Namespace: *namespace,
		BatchSize: *batchSize,
		DryRun:    *dryRun,
		Verbose:   *verbose,
	}
	
	// Create source
	source, err := createSource(sourceArg, config, *textCol, *split)
	if err != nil {
		log.Fatalf("Failed to create source: %v", err)
	}
	
	// Create embedder
	embedder, err := createEmbedder(*embedderType)
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
	}
	
	// Create storage
	storage := memory.NewStorage()
	
	// Create ingestor
	ingestor := ingestion.NewIngestor(source, embedder, storage, config)
	
	// Run ingestion
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	
	fmt.Printf("Starting ingestion from: %s\n", source.Name())
	if *dryRun {
		fmt.Println("DRY RUN MODE - no data will be stored")
	}
	
	stats, err := ingestor.Run(ctx)
	if err != nil {
		log.Fatalf("Ingestion failed: %v", err)
	}
	
	// Print statistics
	stats.Print()
	
	// Export if requested
	if *output != "" && !*dryRun {
		if err := exportVectors(storage, *output); err != nil {
			log.Fatalf("Failed to export vectors: %v", err)
		}
		fmt.Printf("Vectors exported to: %s\n", *output)
	}
}

func createSource(sourceArg string, config *ingestion.SourceConfig, textCol, split string) (ingestion.Source, error) {
	// Check for HuggingFace dataset
	if strings.HasPrefix(sourceArg, "hf:") {
		dataset := strings.TrimPrefix(sourceArg, "hf:")
		source := ingestion.NewHuggingFaceSource(dataset, config)
		source.SetSplit(split)
		return source, nil
	}
	
	// Check for built-in datasets
	builtinDatasets := map[string]bool{
		"demo":         true,
		"quotes":       true,
		"quotes-small": true,
	}
	
	if builtinDatasets[sourceArg] {
		return ingestion.NewBuiltinSource(sourceArg, config), nil
	}
	
	// Check if it's a file
	if _, err := os.Stat(sourceArg); err == nil {
		source, err := ingestion.NewFileSource(sourceArg, config)
		if err != nil {
			return nil, err
		}
		
		// Set text column for CSV files
		if strings.HasSuffix(strings.ToLower(sourceArg), ".csv") {
			source.SetTextColumn(textCol)
		}
		
		return source, nil
	}
	
	return nil, fmt.Errorf("unknown source: %s", sourceArg)
}

func createEmbedder(embedderType string) (embedders.Embedder, error) {
	// Use environment variable if not specified
	if embedderType == "" {
		embedderType = os.Getenv("EMBEDDER_TYPE")
		if embedderType == "" {
			embedderType = "local"
		}
	}
	
	switch strings.ToLower(embedderType) {
	case "local":
		return tfidf.NewTFIDFEmbedder(), nil
		
	case "gemini":
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
		}
		return gemini.NewGeminiEmbedder(apiKey), nil
		
	case "huggingface", "hf":
		apiKey := os.Getenv("HUGGINGFACE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("HUGGINGFACE_API_KEY environment variable not set")
		}
		return huggingface.NewHuggingFaceEmbedder(apiKey), nil
		
	default:
		return nil, fmt.Errorf("unknown embedder type: %s (supported: local, gemini, huggingface)", embedderType)
	}
}

func exportVectors(storage *memory.Storage, filename string) error {
	// This is a placeholder - implement based on your export needs
	fmt.Printf("Export functionality not yet implemented\n")
	return nil
}
