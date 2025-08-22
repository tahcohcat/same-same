# Same-Same Vector Database Microservice

A lightweight RESTful microservice for storing and searching vectors using cosine similarity.

## Features

- In-memory vector storage
- RESTful API for CRUD operations
- Vector similarity search using cosine similarity
- Metadata filtering
- JSON API responses

## API Endpoints

### Vectors
- `POST /api/v1/vectors` - Create a new vector
- `GET /api/v1/vectors` - List all vectors
- `GET /api/v1/vectors/{id}` - Get a specific vector
- `PUT /api/v1/vectors/{id}` - Update a vector
- `DELETE /api/v1/vectors/{id}` - Delete a vector
- `POST /api/v1/vectors/search` - Search vectors by similarity

### Health
- `GET /health` - Health check endpoint

## Usage

### Start the server
```bash
go run ./cmd/same-same -addr :8080
```

### Create a vector
```bash
curl -X POST http://localhost:8080/api/v1/vectors \
  -H "Content-Type: application/json" \
  -d '{
    "id": "doc1",
    "embedding": [0.1, 0.2, 0.3, 0.4],
    "metadata": {"type": "document", "category": "tech"}
  }'
```

### Search vectors
```bash
curl -X POST http://localhost:8080/api/v1/vectors/search \
  -H "Content-Type: application/json" \
  -d '{
    "embedding": [0.1, 0.2, 0.3, 0.4],
    "limit": 5,
    "metadata": {"type": "document"}
  }'
```

### Get all vectors
```bash
curl http://localhost:8080/api/v1/vectors
```

## Development

Build the project:
```bash
go build ./cmd/same-same
```

Run tests:
```bash
go test ./...
```