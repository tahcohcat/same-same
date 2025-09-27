# Makefile
.PHONY: build run test clean docker-build docker-run docker-stop logs

# Variables
APP_NAME=same-same
DOCKER_IMAGE=same-same:latest
DOCKER_COMPOSE=docker-compose

# Development commands
build:
	go build -o bin/$(APP_NAME) ./cmd/same-same

run:
	go run ./cmd/same-same -addr :8080

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	$(DOCKER_COMPOSE) up -d

docker-stop:
	$(DOCKER_COMPOSE) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f same-same

docker-restart:
	$(DOCKER_COMPOSE) restart same-same

# Monitoring
monitoring-up:
	$(DOCKER_COMPOSE) --profile monitoring up -d

monitoring-down:
	$(DOCKER_COMPOSE) --profile monitoring down

# Health checks
health-check:
	curl -f http://localhost:8080/health || exit 1

api-test:
	curl -X GET http://localhost:8080/api/v1/vectors/count

# Development helpers
dev-setup:
	cp .env.example .env
	@echo "Please edit .env file with your API keys"

lint:
	golangci-lint run

format:
	gofmt -s -w .
	go mod tidy

# Load testing
load-test:
	ab -n 100 -c 10 http://localhost:8080/api/v1/vectors/count

# Database operations
load-sample-data:
	curl -X POST http://localhost:8080/api/v1/vectors/embed \
		-H "Content-Type: application/json" \
		-d '{"text": "The only way to do great work is to love what you do.", "author": "Steve Jobs"}'
