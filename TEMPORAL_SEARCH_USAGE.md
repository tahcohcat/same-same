# Temporal-Aware Vector Retrieval

## Overview

Temporal search applies time-based decay to similarity scores, giving more weight to recent documents. This is useful for:
- News and current events
- Time-sensitive recommendations
- Evolving knowledge bases
- Trend analysis

## Decay Formula

```
score(q,d) = cos(q,d) × e^(-λ·Δt)
```

Where:
- `cos(q,d)` = cosine similarity between query and document
- `λ` = decay rate (lambda)
- `Δt` = time difference in years

## Decay Strengths

| Strength | Lambda | 1 Year | 2 Years | 5 Years |
|----------|--------|--------|---------|---------|
| **strong** | 0.5 | 61% | 37% | 8% |
| **medium** | 0.1 | 90% | 82% | 61% |
| **weak** | 0.01 | 99% | 98% | 95% |
| **none** | 0.0 | 100% | 100% | 100% |

## API Endpoint

```
POST /api/v1/search/temporal
```

## Request Examples

### Example 1: Strong Decay (Recent Content Priority)

```bash
curl -X POST http://localhost:8080/api/v1/search/temporal \
  -H "Content-Type: application/json" \
  -d '{
    "query": "artificial intelligence trends",
    "top_k": 5,
    "temporal_decay": "strong"
  }'
```

### Example 2: Medium Decay with Filters

```bash
curl -X POST http://localhost:8080/api/v1/search/temporal \
  -H "Content-Type: application/json" \
  -d '{
    "query": "quantum computing",
    "top_k": 10,
    "temporal_decay": "medium",
    "filters": {
      "author": { "in": ["Einstein", "Feynman"] }
    }
  }'
```

### Example 3: Custom Reference Time

```bash
curl -X POST http://localhost:8080/api/v1/search/temporal \
  -H "Content-Type: application/json" \
  -d '{
    "query": "historical events",
    "top_k": 5,
    "temporal_decay": "weak",
    "reference_time": "2020-01-01T00:00:00Z",
    "time_field": "publication_date"
  }'
```

### Example 4: No Decay (Standard Search)

```bash
curl -X POST http://localhost:8080/api/v1/search/temporal \
  -H "Content-Type: application/json" \
  -d '{
    "query": "philosophy quotes",
    "top_k": 5,
    "temporal_decay": "none"
  }'
```

## Response Format

```json
{
  "results": [
    {
      "vector": {
        "id": "quote_123",
        "metadata": {
          "text": "AI is transforming our world",
          "author": "Modern Thinker",
          "type": "quote"
        },
        "created_at": "2024-01-15T10:00:00Z",
        "updated_at": "2024-01-15T10:00:00Z"
      },
      "score": 0.85,
      "base_score": 0.92,
      "decay_factor": 0.92,
      "document_time": "2024-01-15T10:00:00Z",
      "age": "8 months ago"
    }
  ],
  "total": 5,
  "query": "artificial intelligence trends",
  "decay": "medium",
  "timestamp": "2025-09-30T12:00:00Z"
}
```

## Field Explanations

- **score**: Final score with temporal decay applied
- **base_score**: Original cosine similarity (0-1)
- **decay_factor**: Temporal decay multiplier (0-1)
- **document_time**: Timestamp used for decay calculation
- **age**: Human-readable age ("2 years ago", "3 months ago")

## Use Cases

### 1. News Search (Strong Decay)
```json
{
  "query": "breaking news technology",
  "temporal_decay": "strong"
}
```
Prioritizes recent news heavily.

### 2. Research Papers (Medium Decay)
```json
{
  "query": "machine learning papers",
  "temporal_decay": "medium",
  "filters": {
    "category": { "eq": "research" }
  }
}
```
Balances relevance with recency.

### 3. Historical Quotes (Weak/None Decay)
```json
{
  "query": "philosophy wisdom",
  "temporal_decay": "weak"
}
```
Historical content remains relevant.
