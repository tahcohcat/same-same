# Local File Storage System - Implementation Guide

## Overview

This local file storage implementation provides a robust, extensible, and schema-driven approach to persisting vector embeddings and multimodal data. It's designed to be idiomatic Go, support future use cases, and prioritize metadata retrieval performance.

## Architecture

### Directory Structure

```
data/
├── metadata.json              # Storage schema and collection metadata
├── collections/               # Document JSON files
│   ├── quotes/
│   │   ├── quote_001.json
│   │   ├── quote_002.json
│   │   └── ...
│   └── photos/
│       └── photo_001.json
├── embeddings/                # Separate embedding vectors
│   ├── quotes/
│   │   ├── quote_001.json
│   │   └── quote_002.json
│   └── photos/
│       └── photo_001.json
└── content/                   # Binary content files
    ├── quotes/
    ├── photos/
    │   └── photo_001/
    │       └── image.jpg
    └── audio/
        └── audio_001/
            └── track.mp3
```

### Key Features

* **Schema-Driven**: Collections have defined schemas with field types and indexes  
* **Multimodal Support**: Text, images, audio, video, and custom binary content  
* **Metadata Priority**: Fast metadata queries with indexing support  
* **Extensible**: Easy to add new content types and fields  
* **Relations**: Document relationships for complex data structures  
* **Separation of Concerns**: Large embeddings and content stored separately  
* **Version Control**: Document versioning built-in  
* **Export/Import**: Easy backup and migration  

## Usage Examples

### 1. Basic Setup

```go
import "github/tahcohcat/same-same/internal/storage/local"

// Create local storage
storage, err := local.NewLocalStorage("./data/storage")
if err != nil {
    log.Fatal(err)
}
defer storage.Close()
```

### 2. Create Collection with Schema

```go
schema := &local.CollectionSchema{
    Fields: map[string]local.FieldDefinition{
        "author": {
            Type:        "string",
            Description: "Author name",
            Indexed:     true,
        },
        "year": {
            Type:    "number",
            Indexed: true,
        },
        "tags": {
            Type: "array",
        },
    },
    Required: []string{"author"},
    VectorConfig: &local.VectorConfig{
        Dimension:    768,
        EmbedderType: "local",
        Metric:       "cosine",
    },
}

collection, err := storage.CreateCollection(
    "quotes",
    "Famous quotes collection",
    schema,
)
```

### 3. Store Documents

```go
doc := &local.Document{
    ID:   "quote_001",
    Type: local.TypeText,
    Metadata: map[string]interface{}{
        "author":   "Einstein",
        "year":     1930,
        "category": "science",
        "verified": true,
    },
    Content: &local.ContentData{
        Type: local.TypeText,
        Text: &local.TextContent{
            Raw:      "Time is relative...",
            Language: "en",
            Format:   "plain",
        },
    },
    Embedding: &local.EmbeddingData{
        Vector:    embedding, // []float64
        Dimension: len(embedding),
        Model:     "local-tfidf",
    },
    Tags: []string{"physics", "time"},
}

err = storage.StoreDocument("quotes", doc)
```

### 4. Query by Metadata

```go
results, err := storage.QueryByMetadata("quotes", map[string]interface{}{
    "author":   "Einstein",
    "verified": true,
})

for _, doc := range results {
    fmt.Printf("Found: %s\n", doc.ID)
}
```

### 5. Integration with Existing Storage

```go
// Create adapter for existing Vector interface
adapter, err := local.NewVectorStorageAdapter(
    "./data/storage",
    "vectors",
)
if err != nil {
    log.Fatal(err)
}

// Use like memory storage
vector := &models.Vector{
    ID:        "vec_001",
    Embedding: []float64{0.1, 0.2, 0.3},
    Metadata:  map[string]string{"type": "quote"},
}

err = adapter.Store(vector)
```

## Migration Tools

### Backup Memory to Local

```bash
go run ./cmd/migrate -mode backup \
  -target ./backup/$(date +%Y%m%d)
```

### Restore from Backup

```bash
go run ./cmd/migrate -mode restore \
  -source ./backup/20250930
```

### Export Collection

```bash
go run ./cmd/migrate -mode export \
  -source ./data/storage \
  -target ./exports \
  -collection quotes
```

## Server Integration

### Option 1: Replace Memory Storage

```go
// File: internal/server/server.go

func NewServer() *Server {
    // Replace memory storage with local storage
    adapter, err := local.NewVectorStorageAdapter(
        os.Getenv("STORAGE_PATH"),
        "vectors",
    )
    if err != nil {
        log.Fatal(err)
    }

    // ... rest of server setup
}
```

### Option 2: Hybrid Approach

```go
func NewServer() *Server {
    memStorage := memory.NewStorage()
    
    // Periodically sync to local storage
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        for range ticker.C {
            syncToLocal(memStorage)
        }
    }()
    
    // ... rest of server setup
}
```

## Performance Considerations

### Metadata Indexing

The schema supports field-level indexing for fast queries:

```go
schema := &local.CollectionSchema{
    Fields: map[string]local.FieldDefinition{
        "author": {Indexed: true},  // Fast author lookups
        "year":   {Indexed: true},  // Fast year range queries
    },
    Indexes: []local.Index{
        {
            Name:   "author_year",
            Fields: []string{"author", "year"},
        },
    },
}
```

### Embedding Storage

Large embeddings are stored separately to keep JSON files manageable:

- Documents < 1KB: Inline storage
- Embeddings: Separate files in `embeddings/` directory
- Binary content: Separate files in `content/` directory

### Caching Strategy

```go
// Implement LRU cache for frequently accessed documents
type CachedStorage struct {
    storage *local.LocalStorage
    cache   *lru.Cache
}
```

## Future Enhancements

### Planned Features

1. **Metadata Indexes**: B-tree indexes for fast range queries
2. **Compression**: Optional gzip compression for large documents
3. **Sharding**: Distribute collections across multiple directories
4. **Replication**: Built-in backup and replication
5. **Query Language**: SQL-like query syntax
6. **Transactions**: ACID compliance for batch operations
7. **Streaming**: Support for large file uploads
8. **Encryption**: At-rest encryption for sensitive data

### Multimodal Extensions

The schema is designed to easily support:

- **3D Models**: Point clouds, meshes, CAD files
- **Geospatial**: GeoJSON, shapefiles
- **Time Series**: Sensor data, metrics
- **Code**: Source code with syntax trees
- **Graphs**: Network data, knowledge graphs

## Testing

```bash
# Run storage tests
go test ./internal/storage/local -v

# Benchmark storage operations
go test ./internal/storage/local -bench=. -benchmem

# Test migration tools
go test ./cmd/migrate -v
```

## Troubleshooting

### Issue: Slow Metadata Queries

**Solution**: Enable indexing for frequently queried fields

### Issue: Large Storage Size

**Solution**: Use embedding separation and content references

### Issue: Concurrent Access

**Solution**: The storage uses RWMutex for thread safety

## Best Practices

1. **Schema First**: Define schemas before storing documents
2. **Index Strategy**: Index fields used in filters, not all fields
3. **Separation**: Store large content separately
4. **Versioning**: Use document versions for tracking changes
5. **Backup**: Regular exports to prevent data loss
6. **Relations**: Use relations instead of embedding for complex structures

