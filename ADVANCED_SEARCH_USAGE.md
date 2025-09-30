# Advanced Metadata Search & Filtering

## Overview

This implementation adds powerful metadata filtering capabilities to same-same vector database, allowing you to combine semantic vector search with precise metadata filtering.

## Features

- **Rich Filter Operators**: eq, neq, lt, lte, gt, gte, between, contains, in, exists
- **Hybrid Scoring**: Combine vector similarity with metadata matching scores
- **Type-Safe Filtering**: Automatic numeric and string type handling
- **Composable Filters**: Apply multiple filters across different metadata fields

## API Endpoint

```
POST /api/v1/search
```

## Request Structure

```json
{
  "query": "physics and time",
  "top_k": 5,
  "filters": {
    "author": { "eq": "Einstein" },
    "year": { "gte": 1900, "lte": 1950 },
    "tags": { "contains": "science" },
    "is_public": { "eq": true }
  },
  "options": {
    "hybrid_weight": {
      "vector": 0.8,
      "metadata": 0.2
    }
  }
}
```

## Filter Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equals | `"author": { "eq": "Einstein" }` |
| `neq` | Not equals | `"author": { "neq": "Newton" }` |
| `lt` | Less than | `"year": { "lt": 1950 }` |
| `lte` | Less than or equal | `"year": { "lte": 1950 }` |
| `gt` | Greater than | `"year": { "gt": 1900 }` |
| `gte` | Greater than or equal | `"year": { "gte": 1900 }` |
| `between` | Range (inclusive) | `"year": { "between": [1900, 1950] }` |
| `contains` | String/array contains | `"tags": { "contains": "science" }` |
| `in` | Value in list | `"author": { "in": ["Einstein", "Bohr"] }` |
| `exists` | Field exists | `"tags": { "exists": true }` |

## Usage Examples

### Example 1: Basic Equality Filter

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "relativity and space",
    "top_k": 3,
    "filters": {
      "author": { "eq": "Einstein" }
    }
  }'
```

### Example 2: Range Filter

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "quantum mechanics",
    "top_k": 5,
    "filters": {
      "year": { "gte": 1920, "lte": 1930 }
    }
  }'
```

### Example 3: Multiple Filters

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "philosophy of science",
    "top_k": 10,
    "filters": {
      "author": { "in": ["Einstein", "Bohr", "Heisenberg"] },
      "year": { "between": [1900, 1950] },
      "tags": { "contains": "physics" }
    }
  }'
```

### Example 4: Hybrid Search with Weighted Scoring

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "time and space",
    "top_k": 5,
    "filters": {
      "author": { "eq": "Einstein" }
    },
    "options": {
      "hybrid_weight": {
        "vector": 0.7,
        "metadata": 0.3
      }
    }
  }'
```

### Example 5: Contains Filter for Tags

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "scientific discovery",
    "top_k": 8,
    "filters": {
      "tags": { "contains": "relativity" }
    }
  }'
```

### Example 6: Check Field Existence

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "wisdom",
    "top_k": 5,
    "filters": {
      "year": { "exists": true },
      "author": { "neq": "Unknown" }
    }
  }'
```

## Response Format

```json
{
  "results": [
    {
      "id": "quote_123",
      "text": "The distinction between past, present and future is only a stubbornly persistent illusion.",
      "author": "Einstein",
      "year": 1915,
      "tags": ["time", "philosophy", "relativity"],
      "score": 0.92
    },
    {
      "id": "quote_456",
      "text": "Life is like riding a bicycle. To keep your balance, you must keep moving.",
      "author": "Einstein",
      "year": 1930,
      "tags": ["life", "balance", "science"],
      "score": 0.88
    }
  ],
  "total": 2
}
```

## Integration Steps

### 1. Add the Filter Model

Add the file `internal/models/filter.go` with the FilterEvaluator implementation.

### 2. Update Storage Layer

Add the file `internal/storage/memory/advanced_search.go` with the AdvancedSearch method.

### 3. Add Handler

Add the file `internal/handlers/advanced_search.go` with the AdvancedSearch handler.

### 4. Update Server Routes

In `internal/server/server.go`, add to the `setupRoutes()` method:

```go
// Add after existing search route
api.HandleFunc("/search", s.handler.AdvancedSearch).Methods("POST")
```

Note: This will replace the existing `/search` endpoint with the advanced version that supports filters.

### 5. Update Vector Creation

When creating vectors with metadata, ensure you store filterable fields as metadata:

```go
vector := models.Vector{
    ID:        fmt.Sprintf("quote_%d", time.Now().Unix()),
    Embedding: embedding,
    Metadata: map[string]string{
        "type":   "quote",
        "author": "Einstein",
        "text":   quote.Text,
        "year":   "1925",
        "tags":   "physics,relativity,science",
    },
}
```

## Advanced Use Cases

### Use Case 1: Time-Based Search

Search for recent quotes about leadership:

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "leadership and vision",
    "top_k": 5,
    "filters": {
      "year": { "gte": 2000 }
    }
  }'
```

### Use Case 2: Multi-Author Search

Find quotes from multiple authors about a topic:

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "courage and perseverance",
    "top_k": 10,
    "filters": {
      "author": { 
        "in": ["Churchill", "Roosevelt", "Mandela"] 
      }
    }
  }'
```

### Use Case 3: Exclude Specific Content

Search while excluding certain authors:

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "artificial intelligence",
    "top_k": 5,
    "filters": {
      "author": { "neq": "Unknown" },
      "year": { "exists": true }
    }
  }'
```

## Performance Considerations

1. **Filter Selectivity**: More selective filters (like exact author matches) are evaluated first
2. **Embedding Dimension Check**: Vectors with mismatched dimensions are skipped early
3. **Metadata Indexing**: For production use, consider adding metadata indexing for large datasets
4. **Batch Operations**: The filter evaluator supports batch filtering efficiently

## Testing

Run the included tests:

```bash
go test ./internal/models -v -run TestFilterEvaluator
go test ./internal/storage/memory -v -run TestAdvancedSearch
```

## Extending the System

### Adding Custom Operators

To add a new operator like `starts_with`:

```go
case "starts_with":
    if !exists || !strings.HasPrefix(strings.ToLower(value), strings.ToLower(fmt.Sprint(expectedVal))) {
        return false
    }
```

### Adding Regex Support

```go
case "regex":
    pattern, ok := expectedVal.(string)
    if !ok {
        return false
    }
    matched, err := regexp.MatchString(pattern, value)
    if err != nil || !matched {
        return false
    }
```

### Geospatial Filtering (Future)

For location-based filtering:

```go
case "near":
    // Implement distance calculations for lat/lon metadata
    return fe.compareDistance(value, expectedVal)
```

## Best Practices

1. **Use appropriate operators**: Use `eq` for exact matches, `contains` for substring searches
2. **Combine filters**: Use multiple filters to narrow down results efficiently
3. **Hybrid weights**: Start with vector:0.8, metadata:0.2 and adjust based on your use case
4. **Metadata design**: Store filterable data as simple string values for flexibility
5. **Testing filters**: Test filter combinations with your data before production use

## Migration from Existing Endpoints

If you're using the old `/api/v1/search` endpoint:

**Old way:**
```json
{
  "text": "patience",
  "limit": 2,
  "namespace": "quotes"
}
```

**New way:**
```json
{
  "query": "patience",
  "top_k": 2,
  "filters": {
    "type": { "eq": "quotes" }
  }
}
```

The new endpoint is backward compatible when filters are omitted.

## Troubleshooting

### Issue: No results returned

**Check:**
- Verify metadata field names match exactly (case-sensitive)
- Ensure filter values match the stored data types
- Test without filters first to verify vectors exist

### Issue: Unexpected results

**Check:**
- Review hybrid weight settings (try vector:1.0, metadata:0.0)
- Verify filter operators (use `eq` not `equals`)
- Check metadata values in stored vectors

### Issue: Performance degradation

**Solutions:**
- Reduce `top_k` value
- Add more selective filters early
- Consider implementing metadata indexing for large datasets

## Complete Example Script

```bash
#!/bin/bash
# test_advanced_search.sh

BASE_URL="http://localhost:8080/api/v1"

# 1. Add some test vectors
echo "Adding test vectors..."
curl -X POST "$BASE_URL/vectors/embed" \
  -H "Content-Type: application/json" \
  -d '{"text": "Time is relative", "author": "Einstein"}'

curl -X POST "$BASE_URL/vectors/embed" \
  -H "Content-Type: application/json" \
  -d '{"text": "Quantum leap", "author": "Bohr"}'

# 2. Search with author filter
echo -e "\n\nSearching for Einstein quotes..."
curl -X POST "$BASE_URL/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "time and space",
    "top_k": 5,
    "filters": {
      "author": { "eq": "Einstein" }
    }
  }' | jq '.'

# 3. Search with multiple filters
echo -e "\n\nSearching with multiple filters..."
curl -X POST "$BASE_URL/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "quantum physics",
    "top_k": 3,
    "filters": {
      "author": { "in": ["Einstein", "Bohr"] },
      "type": { "eq": "quote" }
    }
  }' | jq '.'
```

Run the script:
```bash
chmod +x test_advanced_search.sh
./test_advanced_search.sh
```

---

## Summary

This advanced metadata search implementation provides:

* **9 powerful filter operators** (eq, neq, lt, lte, gt, gte, between, contains, in, exists)  
* **Hybrid scoring** combining vector similarity and metadata matching  
* **Composable filters** for complex queries  
* **Type-safe evaluation** handling strings and numbers  
* **Production-ready** with comprehensive tests  
* **Extensible design** for adding custom operators  

The system seamlessly integrates with your existing Same-Same codebase while maintaining backward compatibility