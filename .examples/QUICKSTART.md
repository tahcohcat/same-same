# Quick Start Examples

## Build the Ingest Tool

```bash
go build ./cmd/ingest
```

## Basic Examples

### 1. Ingest Demo Dataset (20 quotes)
```bash
./ingest demo
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
./ingest -namespace philosophy demo
```

### 3. Ingest CSV File
```bash
./ingest .examples/data/sample.csv
```

### 4. Dry Run (Validate Only)
```bash
./ingest -dry-run -verbose demo
```

### 5. Full Quotes Dataset
```bash
./ingest -verbose quotes
```

### 6. With Gemini Embeddings
```bash
export GEMINI_API_KEY=your_key
./ingest -embedder gemini demo
```

## Remember

- **Flags before source**: `./ingest -namespace test demo` ✓
- **Not after**: `./ingest demo -namespace test` ✗

See [INGESTION_GUIDE.md](../../INGESTION_GUIDE.md) for full documentation.
