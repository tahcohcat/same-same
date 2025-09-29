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
  -d '{"text": "patience", "limit": 1, "namespace": "quotes"}'
```

### Example response:
```json
{
  "matches": [
    {
      "text": "The happiness of your life depends upon the quality of your thoughts. — Marcus Aurelius",
      "score": 0.88
    },
    {
      "text": "Happiness depends upon ourselves. — Aristotle",
      "score": 0.84
    },
    {
      "text": "The soul becomes dyed with the color of its thoughts. — Marcus Aurelius",
      "score": 0.80
    }
  ]
}
```