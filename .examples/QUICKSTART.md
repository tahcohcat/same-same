# Quick Start Examples

## Build the CLI Tool

```bash
go build ./cmd/same-same
```

## Basic Examples

### 1. Ingest Demo Dataset (20 quotes)
```bash
same-same ingest demo
```

Output:
```
=== Ingestion Complete ===
Total Records:    20
Successfully Ingested: 20
Duration:         511µs
Speed:            39138 records/sec

Storage Details:
  Location:       memory
  Namespace:      default
  Note:           Data is in-memory only
```

### 2. Ingest with Custom Namespace
```bash
same-same ingest -n philosophy demo
```

### 3. Ingest CSV File
```bash
same-same ingest .examples/data/sample.csv
```

### 4. Dry Run (Validate Only)
```bash
same-same ingest --dry-run -v demo
```

### 5. Full Quotes Dataset
```bash
same-same ingest -v quotes
```

### 6. With Gemini Embeddings
```bash
export GEMINI_API_KEY=your_key
same-same ingest -e gemini demo
```

## CLI Commands

- **Ingest data**: `same-same ingest <source> [flags]`
- **Start server**: `same-same serve [flags]`
- **Help**: `same-same --help` or `same-same ingest --help`

See [INGESTION_GUIDE.md](../../INGESTION_GUIDE.md) for full documentation.
