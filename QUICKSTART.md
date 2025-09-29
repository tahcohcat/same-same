# 🚀 Getting Started 

This section will help you get up and running with the Same-Same Vector DB running on localhost:8080 usiung a Google Gemini API Key.

## Step 1: Start the Vector Database

### Set your API key
```bash
export GEMINI_API_KEY=your_google_gemini_api_key_here
```

### Start the service
```bash
go run ./cmd/same-same -addr :8080

# or with docker
docker run -d --name same-same -p 8080:8080 -e GEMINI_API_KEY=your_key same-same:latest
```

## Step 2: Launch the Demo Application

We include a **sample dataset** of public-domain quotes, so you can try out same-same immediately:

- **File path:** `.examples/data/quotes.txt` 

### Load the quotes into the index

```bash
 cat .examples/data/quotes_small.txt | tr -d '\r' | while IFS= read -r line; do   quote=$(printf '%s' "$line" | sed 's/ — .*//; s/\\/\\\\/g; s/"/\\"/g');   author=$(printf '%s' "$line" | sed 's/.* — //; s/\\/\\\\/g; s/"/\\"/g');\
    curl -s -X POST "http://localhost:8081/api/v1/vectors/embed" -H "Content-Type: application/json" -d "{\"text\":\"$quote\", \"author\":\"$author\"}"; done
```

### Run a similarity search
```bash
curl -s "http://localhost:8081/api/v1/search" \
  -H "Content-Type: application/json" \
  -d '{"text": "patience", "limit": 3, "namespace": "quotes"}'
```

### Example response:
```json
{
  "matches": [
    {
      "vector": {
        "id": "quote_1759143813",
        "metadata": {
          "author": "Aristotle",
          "text": "Wishing to be friends is quick work, but friendship is a slow ripening fruit.",
          "type": "quote"
        },
        "created_at": "0001-01-01T00:00:00Z",
        "updated_at": "2025-09-29T12:03:33.9520825+01:00"
      },
      "score": 0.5861640990760936
    },
    {
      "vector": {
        "id": "quote_1759143810",
        "metadata": {
          "author": "Plato",
          "text": "Opinion is the medium between knowledge and ignorance.",
          "type": "quote"
        },
        "created_at": "0001-01-01T00:00:00Z",
        "updated_at": "2025-09-29T12:03:30.8416675+01:00"
      },
      "score": 0.5502264895305855
    },
    {
      "vector": {
        "id": "quote_1759143809",
        "metadata": {
          "author": "Socrates",
          "text": "He who is not contented with what he has, would not be contented with what he would like to have.",
          "type": "quote"
        },
        "created_at": "0001-01-01T00:00:00Z",
        "updated_at": "2025-09-29T12:03:29.7642857+01:00"
      },
      "score": 0.5494706619549622
    }
  ]
}
```