# Getting Started

## Sample Quotes Demo

We include a **sample dataset** of public-domain quotes, so you can try out same-same immediately:

- **File path:** `.examples/data/quotes.txt` 

## Load the quotes into the index

```bash
while read line; do
  curl -s -X POST "http://localhost:8081/docs" \
    -H "Content-Type: application/json" \
    -d "{\"text\": \"$line\"}" > /dev/null
done < examples/data/quotes.txt
```

## Run a similarity search
```bash
curl -s "http://localhost:8081/query" \
  -H "Content-Type: application/json" \
  -d '{"text": "life purpose and happiness", "k": 3}' | jq
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