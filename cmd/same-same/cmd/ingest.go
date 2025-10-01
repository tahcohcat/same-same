package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tahcohcat/same-same/internal/embedders"
	"github.com/tahcohcat/same-same/internal/embedders/clip"
	"github.com/tahcohcat/same-same/internal/embedders/quotes/gemini"
	"github.com/tahcohcat/same-same/internal/embedders/quotes/huggingface"
	"github.com/tahcohcat/same-same/internal/embedders/quotes/local/tfidf"
	"github.com/tahcohcat/same-same/internal/ingestion"
	"github.com/tahcohcat/same-same/internal/storage/memory"
)

var (
	// Ingest-specific flags
	textCol      string
	idCol        string
	metaCol      string
	sample       int
	split        string
	maxTokens    int
	benchmark    bool
	batchSize    int
	embedderType string
	timeout      time.Duration
	output       string
	recursive    bool
	clipModel    string
	clipPretrain string
)

func init() {
	rootCmd.AddCommand(ingestCmd)

	// Ingest flags
	ingestCmd.Flags().StringVar(&textCol, "text-col", "text", "Name of the text column (CSV)")
	ingestCmd.Flags().StringVar(&idCol, "id-col", "id", "Name of the ID column (optional)")
	ingestCmd.Flags().StringVar(&metaCol, "meta-col", "", "Name of the metadata column (optional)")
	ingestCmd.Flags().IntVar(&sample, "sample", 0, "Sample N rows (0 = all)")
	ingestCmd.Flags().StringVar(&split, "split", "train", "Dataset split (HuggingFace only)")
	ingestCmd.Flags().IntVar(&maxTokens, "max-tokens", 512, "Max tokens per document")
	ingestCmd.Flags().BoolVar(&benchmark, "benchmark", false, "Run in benchmark mode")
	ingestCmd.Flags().IntVar(&batchSize, "batch-size", 100, "Batch size for bulk operations")
	ingestCmd.Flags().StringVarP(&embedderType, "embedder", "e", "", "Embedder type (local, gemini, huggingface)")
	ingestCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Minute, "Timeout for ingestion")
	ingestCmd.Flags().StringVarP(&output, "output", "o", "", "Output file for exported vectors")
}

var ingestCmd = &cobra.Command{
	Use:   "ingest <source>",
	Short: "Ingest data into same-same",
	Long: `Ingest data from various sources into the same-same vector database.

Sources:
  demo, quotes, quotes-small    Built-in datasets from .examples/data
  hf:<dataset>                  HuggingFace dataset (e.g., hf:imdb, hf:squad:v2)
  file.csv                      CSV file
  file.jsonl                    JSONL file (each line is a JSON object)
  file.json                     Same as JSONL
  images:<directory>            Directory of images (requires -e clip)
  image-list:<file.txt>         Text file with image paths (requires -e clip)

The ingestion pipeline:
  1. Reads records from the source
  2. Generates embeddings using the selected embedder
  3. Stores vectors in the database`,
	Example: `  # Ingest built-in demo dataset
  same-same ingest demo

  # Ingest with custom namespace
  same-same ingest -n quotes demo

  # Ingest HuggingFace dataset
  same-same ingest hf:imdb --split train --sample 1000

  # Ingest from CSV file
  same-same ingest mydata.csv --text-col content

  # Ingest from JSONL file
  same-same ingest data.jsonl -v

  # Dry run to validate data
  same-same ingest --dry-run -v data.jsonl

  # Use specific embedder
  same-same ingest -e gemini demo
  
  # Ingest images with CLIP
  same-same ingest -e clip images:./photos
  
  # Ingest images from list
  same-same ingest -e clip image-list:images.txt`,
	Args: cobra.ExactArgs(1),
	Run:  runIngest,
}

func runIngest(cmd *cobra.Command, args []string) {
	source := args[0]

	// Create config
	config := &ingestion.SourceConfig{
		Namespace: namespace,
		BatchSize: batchSize,
		DryRun:    dryRun,
		Verbose:   verbose,
	}

	// Create source
	src, err := createSource(source, config)
	if err != nil {
		log.Fatalf("Failed to create source: %v", err)
	}

	// Create embedder
	embedder, err := createEmbedder(embedderType)
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
	}

	// Create storage
	storage := memory.NewStorage()

	// Create ingestor
	ingestor := ingestion.NewIngestor(src, embedder, storage, config)

	// Run ingestion
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("Starting ingestion from: %s\n", src.Name())
	if dryRun {
		fmt.Println("DRY RUN MODE - no data will be stored")
	}
	if benchmark {
		fmt.Println("⚡ Benchmark mode enabled")
	}

	stats, err := ingestor.Run(ctx)
	if err != nil {
		log.Fatalf("Ingestion failed: %v", err)
	}

	// Print statistics
	stats.Print()

	// Export if requested
	if output != "" && !dryRun {
		if err := exportVectors(storage, output); err != nil {
			log.Fatalf("Failed to export vectors: %v", err)
		}
		fmt.Printf("Vectors exported to: %s\n", output)
	}
}

func createSource(sourceArg string, config *ingestion.SourceConfig) (ingestion.Source, error) {
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

	case "clip":
		// Check if using Python-based CLIP or simple Go-based
		if os.Getenv("CLIP_USE_PYTHON") == "true" {
			embedder := clip.NewCLIPEmbedder(clipModel, clipPretrain)
			if verbose {
				fmt.Printf("Using Python CLIP model: %s with pretrained: %s\n", clipModel, clipPretrain)
			}
			return embedder, nil
		} else {
			// Use simple Go-based embedder (no Python required!)
			embedder := clip.NewSimpleCLIPEmbedder()
			if verbose {
				fmt.Printf("Using Simple CLIP embedder (pure Go, no Python required)\n")
			}
			return embedder, nil
		}

	default:
		return nil, fmt.Errorf("unknown embedder type: %s (supported: local, gemini, huggingface, clip)", embedderType)
	}
}

func exportVectors(storage *memory.Storage, filename string) error {
	// Placeholder - implement based on your export needs
	fmt.Printf("Export functionality not yet implemented\n")
	return nil
}
