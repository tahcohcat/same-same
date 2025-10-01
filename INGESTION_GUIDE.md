# Ingestion Guide

The `ingest` command provides a flexible way to load data from multiple sources into the same-same vector database.

## Quick Start

```bash
# Build the same-same CLI tool
go build ./cmd/same-same

# Ingest built-in demo dataset
./same-same ingest demo

# Ingest with namespace (short flags)
./same-same ingest -n quotes demo

# Dry run to validate data
./same-same ingest --dry-run -v data.jsonl
```

## Supported Sources

### 1. Built-in Datasets

Load pre-packaged datasets from the `.examples/data` directory.

**Available datasets:**
- `demo` or `quotes-small` - 20 philosophical quotes
- `quotes` - Full quotes dataset

**Usage:**
```bash
./ingest demo
./ingest quotes-small
./ingest quotes
```

**Format:** Text files with format: `"Quote text — Author"`

### 2. HuggingFace Datasets

Load any public dataset from HuggingFace.

**Requirements:**
- Python 3 installed
- `datasets` library: `pip install datasets`

**Usage:**
```bash
# Basic dataset
same-same ingest hf:imdb

# Dataset with subset
same-same ingest hf:squad:v2

# Specify split
same-same ingest --split test hf:imdb

# Use different embedder
same-same ingest -e gemini hf:imdb
```

**How it works:**
1. Downloads dataset using Python's `datasets` library
2. Exports to temporary JSONL file
3. Processes and embeds each record
4. Stores in vector database

### 3. CSV Files

Load data from CSV files.

**Usage:**
```bash
# Ingest CSV (text column named "text")
same-same ingest data.csv

# Specify custom text column
same-same ingest --text-col content data.csv

# With namespace
same-same ingest -n products --text-col description products.csv
```

**CSV Format:**
```csv
text,author,year
"The only way to do great work is to love what you do.",Steve Jobs,2005
"Innovation distinguishes between a leader and a follower.",Steve Jobs,1998
```

**Features:**
- First row must contain headers
- `-text-col` flag specifies which column contains the text to embed
- All other columns become metadata

### 4. JSONL (JSON Lines) Files

Load data from JSONL/NDJSON files.

**Usage:**
```bash
./ingest data.jsonl
./ingest data.ndjson
```

**JSONL Format:**
```jsonl
{"text": "First quote", "author": "Author 1", "year": 2020}
{"text": "Second quote", "author": "Author 2", "category": "wisdom"}
```

**Features:**
- Each line must be a valid JSON object
- Automatically detects text field (tries: `text`, `content`, `body`, `message`, `quote`)
- All other fields become metadata
- Flexible schema - each record can have different fields

## Command Flags

### Core Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-namespace` | string | `default` | Namespace for ingested vectors |
| `-batch-size` | int | `100` | Number of records to process in each batch |
| `-dry-run` | bool | `false` | Validate data without storing |
| `-verbose` | bool | `false` | Enable detailed logging |
| `-timeout` | duration | `30m` | Maximum time for ingestion |

### Embedder Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-embedder` | string | `local` | Embedder type: `local`, `gemini`, `huggingface` |

**Environment variables:**
- `EMBEDDER_TYPE` - Default embedder (overridden by `-embedder` flag)
- `GEMINI_API_KEY` - Required for Gemini embedder
- `HUGGINGFACE_API_KEY` - Required for HuggingFace embedder

### Source-Specific Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-text-col` | string | `text` | CSV: column name containing text |
| `-split` | string | `train` | HuggingFace: dataset split to use |

### Advanced Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-output` | string | `` | Export vectors to file after ingestion |

## Examples

### Example 1: Ingest Demo Data with Gemini Embeddings

```bash
export GEMINI_API_KEY=your_key_here
same-same ingest -e gemini -n philosophy demo
```

### Example 2: Ingest Custom CSV

```bash
# products.csv format:
# name,description,price,category
# "Product 1","Great product...",29.99,electronics

same-same ingest --text-col description -n products products.csv
```

### Example 3: Ingest HuggingFace Dataset

```bash
# Install dependencies
pip install datasets

# Ingest IMDB reviews
same-same ingest -n reviews --batch-size 500 hf:imdb
```

### Example 4: Validate JSONL Before Ingesting

```bash
# Check data quality first
same-same ingest --dry-run -v data.jsonl

# If validation passes, ingest for real
same-same ingest -n mydata data.jsonl
```

### Example 5: Large Dataset with Timeout

```bash
same-same ingest --timeout 2h --batch-size 1000 hf:wikipedia
```

## Output and Statistics

After ingestion completes, you'll see statistics:

```
=== Ingestion Complete ===
Total Records:    1000
Successfully Ingested: 995
Failed:           5
Skipped:          0
Duration:         2m15s
Speed:            7.36 records/sec

Failure Breakdown:
  embed_error: 3
  storage_error: 2
========================
```

**Metrics:**
- **Total Records** - Number of records read from source
- **Successfully Ingested** - Vectors successfully stored
- **Failed** - Records that couldn't be processed
- **Skipped** - Empty or invalid records
- **Duration** - Total ingestion time
- **Speed** - Records processed per second
- **Failure Breakdown** - Categories of failures

## Error Handling

Common errors and solutions:

### "python not found"
```bash
# Install Python 3
# Windows: Download from python.org
# Linux: sudo apt install python3
# Mac: brew install python3
```

### "GEMINI_API_KEY environment variable not set"
```bash
export GEMINI_API_KEY=your_key_here
# or
./ingest -embedder local demo  # Use local embedder instead
```

### "text column 'X' not found in CSV headers"
```bash
# Check your CSV headers
head -1 your_file.csv

# Specify correct column
./ingest -text-col your_column_name your_file.csv
```

### "failed to download dataset"
```bash
# Install HuggingFace datasets
pip install datasets

# Check dataset name at https://huggingface.co/datasets
```

## Programmatic Usage

You can also use the ingestion library programmatically:

```go
package main

import (
    "context"
    "github.com/tahcohcat/same-same/internal/ingestion"
    "github.com/tahcohcat/same-same/internal/embedders/quotes/local/tfidf"
    "github.com/tahcohcat/same-same/internal/storage/memory"
)

func main() {
    config := &ingestion.SourceConfig{
        Namespace: "mydata",
        BatchSize: 100,
        Verbose:   true,
    }
    
    source := ingestion.NewBuiltinSource("demo", config)
    embedder := tfidf.NewTFIDFEmbedder()
    storage := memory.NewStorage()
    
    ingestor := ingestion.NewIngestor(source, embedder, storage, config)
    
    stats, err := ingestor.Run(context.Background())
    if err != nil {
        panic(err)
    }
    
    stats.Print()
}
```

## Custom Sources

To implement a custom data source, implement the `ingestion.Source` interface:

```go
type Source interface {
    Open(ctx context.Context) error
    Next() (*Record, error)  // Returns io.EOF when done
    Close() error
    Name() string
}

type Record struct {
    ID       string
    Text     string
    Metadata map[string]string
}
```

## Performance Tips

1. **Batch Size**: Larger batches are faster but use more memory
   - Small datasets: 100-500
   - Large datasets: 1000-5000

2. **Embedder Choice**:
   - **Local TF-IDF**: Fastest, no API calls, good for prototyping
   - **Gemini**: High quality, requires API key, rate limits apply
   - **HuggingFace**: Very high quality, slower, rate limits apply

3. **Parallel Processing**: For multiple files, run multiple ingest commands in parallel

4. **Dry Run First**: Always test with `-dry-run -verbose` on a sample before full ingestion

## Next Steps

After ingestion:

1. **Verify data**: Use the API to count and list vectors
   ```bash
   curl http://localhost:8080/api/v1/vectors/count
   ```

2. **Test search**: Try similarity search
   ```bash
   curl -X POST http://localhost:8080/api/v1/search \
     -H "Content-Type: application/json" \
     -d '{"text": "your query", "limit": 5}'
   ```

3. **Persist data**: Use local file storage for persistence (see [LOCAL_FILE_STORAGE.md](LOCAL_FILE_STORAGE.md))

## Troubleshooting

Enable verbose mode for detailed logging:

```bash
same-same ingest -v your_source
```

For issues, check:
- File format and structure
- API keys and environment variables
- Python and dependencies (for HuggingFace)
- Network connectivity (for API-based embedders)
