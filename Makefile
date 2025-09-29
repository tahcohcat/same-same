# Enhanced Makefile for Same-Same Vector Database
.PHONY: help build run test clean docker-build docker-run docker-stop logs dev-setup lint format benchmark security-scan

# Variables
APP_NAME := same-same
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse HEAD)
DOCKER_IMAGE := ghcr.io/tahcohcat/$(APP_NAME)
DOCKER_COMPOSE := docker-compose
GO_VERSION := $(shell go version | cut -d' ' -f3)
PLATFORMS := linux/amd64,linux/arm64,darwin/amd64,darwin/arm64,windows/amd64

# Build flags
LDFLAGS := -X main.version=$(VERSION) \
		   -X main.buildTime=$(BUILD_TIME) \
		   -X main.gitCommit=$(GIT_COMMIT) \
		   -w -s

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
PURPLE := \033[0;35m
CYAN := \033[0;36m
WHITE := \033[0;37m
NC := \033[0m # No Color

## help: Show this help message
help:
	@echo "$(CYAN)Same-Same Vector Database - Available Commands$(NC)"
	@echo ""
	@echo "$(GREEN)Development Commands:$(NC)"
	@echo "  $(BLUE)build$(NC)          - Build the application binary"
	@echo "  $(BLUE)run$(NC)            - Run the application locally"
	@echo "  $(BLUE)dev$(NC)            - Run in development mode with auto-reload"
	@echo "  $(BLUE)test$(NC)           - Run all tests"
	@echo "  $(BLUE)test-coverage$(NC)  - Run tests with coverage report"
	@echo "  $(BLUE)benchmark$(NC)      - Run performance benchmarks"
	@echo "  $(BLUE)lint$(NC)           - Run code linters"
	@echo "  $(BLUE)format$(NC)         - Format code and organize imports"
	@echo "  $(BLUE)security-scan$(NC)  - Run security vulnerability scan"
	@echo ""
	@echo "$(GREEN)Docker Commands:$(NC)"
	@echo "  $(BLUE)docker-build$(NC)   - Build Docker image"
	@echo "  $(BLUE)docker-run$(NC)     - Start services with Docker Compose"
	@echo "  $(BLUE)docker-stop$(NC)    - Stop Docker services"
	@echo "  $(BLUE)docker-logs$(NC)    - Show Docker logs"
	@echo "  $(BLUE)docker-clean$(NC)   - Clean up Docker resources"
	@echo ""
	@echo "$(GREEN)Deployment Commands:$(NC)"
	@echo "  $(BLUE)deploy$(NC)         - Deploy to production"
	@echo "  $(BLUE)monitoring-up$(NC)  - Start with monitoring stack"
	@echo "  $(BLUE)monitoring-down$(NC) - Stop monitoring stack"
	@echo ""
	@echo "$(GREEN)Utility Commands:$(NC)"
	@echo "  $(BLUE)clean$(NC)          - Clean build artifacts"
	@echo "  $(BLUE)deps$(NC)           - Download and tidy dependencies"
	@echo "  $(BLUE)dev-setup$(NC)      - Set up development environment"
	@echo "  $(BLUE)load-test$(NC)      - Run load tests"
	@echo "  $(BLUE)health-check$(NC)   - Check if service is healthy"
	@echo "  $(BLUE)backup$(NC)         - Backup vector database"
	@echo "  $(BLUE)release$(NC)        - Create a new release"

## build: Build the application binary
build:
	@echo "$(GREEN)Building $(APP_NAME)...$(NC)"
	@mkdir -p bin
	go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME) ./cmd/same-same
	@echo "$(GREEN)✅ Build complete: bin/$(APP_NAME)$(NC)"

## build-all: Build for all platforms
build-all:
	@echo "$(GREEN)Building for all platforms...$(NC)"
	@mkdir -p dist
	@for platform in $(subst $(comma), ,$(PLATFORMS)); do \
		os=$(echo $platform | cut -d'/' -f1); \
		arch=$(echo $platform | cut -d'/' -f2); \
		ext=""; \
		if [ "$os" = "windows" ]; then ext=".exe"; fi; \
		echo "$(BLUE)Building for $os/$arch...$(NC)"; \
		GOOS=$os GOARCH=$arch go build -ldflags="$(LDFLAGS)" \
			-o dist/$(APP_NAME)-$os-$arch$ext ./cmd/same-same; \
	done
	@echo "$(GREEN)✅ Multi-platform build complete$(NC)"

## run: Run the application locally
run:
	@echo "$(GREEN)Starting $(APP_NAME) on :8080...$(NC)"
	@if [ -z "$GEMINI_API_KEY" ]; then \
		echo "$(YELLOW)⚠️  GEMINI_API_KEY not set. Loading from .env if available$(NC)"; \
	fi
	go run -ldflags="$(LDFLAGS)" ./cmd/same-same -addr :8080

## dev: Run in development mode with auto-reload
dev:
	@echo "$(GREEN)Starting development mode...$(NC)"
	@if ! command -v air >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing air for hot reload...$(NC)"; \
		go install github.com/cosmtrek/air@latest; \
	fi
	air

## test: Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v -race ./...
	@echo "$(GREEN)✅ All tests passed$(NC)"

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@mkdir -p coverage
	go test -v -race -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@coverage=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}' | sed 's/%//'); \
	echo "$(GREEN)✅ Coverage: $coverage%$(NC)"; \
	if [ $(echo "$coverage < 70" | bc -l 2>/dev/null || echo "0") = "1" ]; then \
		echo "$(YELLOW)⚠️  Coverage below 70%$(NC)"; \
	fi
	@echo "$(BLUE)Coverage report: coverage/coverage.html$(NC)"

## benchmark: Run performance benchmarks
benchmark:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@mkdir -p benchmarks
	go test -bench=. -benchmem -cpuprofile=benchmarks/cpu.prof -memprofile=benchmarks/mem.prof ./... | tee benchmarks/results.txt
	@echo "$(GREEN)✅ Benchmarks complete$(NC)"
	@echo "$(BLUE)Results: benchmarks/results.txt$(NC)"

## lint: Run code linters
lint:
	@echo "$(GREEN)Running linters...$(NC)"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing golangci-lint...$(NC)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest; \
	fi
	golangci-lint run --timeout=10m
	@echo "$(GREEN)✅ Linting complete$(NC)"

## format: Format code and organize imports
format:
	@echo "$(GREEN)Formatting code...$(NC)"
	gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "$(YELLOW)Installing goimports...$(NC)"; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		goimports -w .; \
	fi
	go mod tidy
	@echo "$(GREEN)✅ Code formatted$(NC)"

## security-scan: Run security vulnerability scan
security-scan:
	@echo "$(GREEN)Running security scan...$(NC)"
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing gosec...$(NC)"; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	gosec -fmt json -out gosec-report.json -severity medium ./...
	@if ! command -v nancy >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing nancy...$(NC)"; \
		go install github.com/sonatypeoss/nancy@latest; \
	fi
	go list -json -deps ./... | nancy sleuth
	@echo "$(GREEN)✅ Security scan complete$(NC)"

## deps: Download and tidy dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	go mod download
	go mod tidy
	go mod verify
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

## clean: Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -rf bin/
	rm -rf dist/
	rm -rf coverage/
	rm -rf benchmarks/
	rm -f coverage.out coverage.html
	rm -f gosec-report.json
	go clean -cache -testcache -modcache
	@echo "$(GREEN)✅ Cleanup complete$(NC)"

## dev-setup: Set up development environment
dev-setup:
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.example .env 2>/dev/null || touch .env; \
		echo "$(YELLOW)⚠️  Please edit .env file with your API keys$(NC)"; \
	fi
	@if [ ! -f .golangci.yml ]; then \
		echo "$(BLUE)Creating .golangci.yml...$(NC)"; \
		curl -s https://raw.githubusercontent.com/golangci/golangci-lint/master/.golangci.reference.yml > .golangci.yml; \
	fi
	go mod download
	@echo "$(GREEN)✅ Development environment ready$(NC)"
	@echo "$(BLUE)Next steps:$(NC)"
	@echo "  1. Edit .env file with your API keys"
	@echo "  2. Run 'make run' to start the server"
	@echo "  3. Run 'make test' to verify everything works"

## docker-build: Build Docker image
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .
	@echo "$(GREEN)✅ Docker image built: $(DOCKER_IMAGE):$(VERSION)$(NC)"

## docker-build-multi: Build multi-architecture Docker image
docker-build-multi:
	@echo "$(GREEN)Building multi-architecture Docker image...$(NC)"
	docker buildx create --name multiarch --use --bootstrap 2>/dev/null || true
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t $(DOCKER_IMAGE):$(VERSION) \
		-t $(DOCKER_IMAGE):latest \
		--push .
	@echo "$(GREEN)✅ Multi-arch Docker image built and pushed$(NC)"

## docker-run: Start services with Docker Compose
docker-run:
	@echo "$(GREEN)Starting Docker services...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)✅ Services started$(NC)"
	@echo "$(BLUE)API: http://localhost:8080$(NC)"
	@echo "$(BLUE)Web: http://localhost$(NC)"

## docker-stop: Stop Docker services
docker-stop:
	@echo "$(GREEN)Stopping Docker services...$(NC)"
	$(DOCKER_COMPOSE) down
	@echo "$(GREEN)✅ Services stopped$(NC)"

## docker-logs: Show Docker logs
docker-logs:
	$(DOCKER_COMPOSE) logs -f $(APP_NAME)

## docker-clean: Clean up Docker resources
docker-clean:
	@echo "$(GREEN)Cleaning Docker resources...$(NC)"
	$(DOCKER_COMPOSE) down -v --remove-orphans
	docker image prune -f
	docker volume prune -f
	@echo "$(GREEN)✅ Docker cleanup complete$(NC)"

## monitoring-up: Start with monitoring stack
monitoring-up:
	@echo "$(GREEN)Starting monitoring stack...$(NC)"
	$(DOCKER_COMPOSE) --profile monitoring up -d
	@echo "$(GREEN)✅ Monitoring stack started$(NC)"
	@echo "$(BLUE)Prometheus: http://localhost:9090$(NC)"
	@echo "$(BLUE)Grafana: http://localhost:3000 (admin/admin123)$(NC)"

## monitoring-down: Stop monitoring stack
monitoring-down:
	@echo "$(GREEN)Stopping monitoring stack...$(NC)"
	$(DOCKER_COMPOSE) --profile monitoring down
	@echo "$(GREEN)✅ Monitoring stack stopped$(NC)"

## deploy: Deploy to production
deploy:
	@echo "$(GREEN)Deploying to production...$(NC)"
	@chmod +x cmd/scripts/deploy.sh
	./cmd/scripts/deploy.sh

## health-check: Check if service is healthy
health-check:
	@echo "$(GREEN)Checking service health...$(NC)"
	@if curl -sf http://localhost:8080/health > /dev/null; then \
		echo "$(GREEN)✅ Service is healthy$(NC)"; \
	else \
		echo "$(RED)❌ Service is unhealthy$(NC)"; \
		exit 1; \
	fi

## load-test: Run load tests
load-test:
	@echo "$(GREEN)Running load tests...$(NC)"
	@if ! command -v ab >/dev/null 2>&1; then \
		echo "$(RED)❌ Apache Bench (ab) not found. Install apache2-utils$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Testing health endpoint...$(NC)"
	ab -n 1000 -c 10 http://localhost:8080/health
	@echo "$(BLUE)Testing vector count endpoint...$(NC)"
	ab -n 100 -c 5 http://localhost:8080/api/v1/vectors/count
	@echo "$(GREEN)✅ Load tests complete$(NC)"

## backup: Backup vector database
backup:
	@echo "$(GREEN)Creating backup...$(NC)"
	@chmod +x cmd/scripts/backup.sh
	./cmd/scripts/backup.sh

## release: Create a new release
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)❌ VERSION not set. Usage: make release VERSION=v1.0.0$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating release $(VERSION)...$(NC)"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	$(MAKE) build-all
	@echo "$(GREEN)✅ Release $(VERSION) created$(NC)"

## install: Install the application
install: build
	@echo "$(GREEN)Installing $(APP_NAME)...$(NC)"
	cp bin/$(APP_NAME) /usr/local/bin/
	@echo "$(GREEN)✅ $(APP_NAME) installed to /usr/local/bin/$(NC)"

## uninstall: Uninstall the application
uninstall:
	@echo "$(GREEN)Uninstalling $(APP_NAME)...$(NC)"
	rm -f /usr/local/bin/$(APP_NAME)
	@echo "$(GREEN)✅ $(APP_NAME) uninstalled$(NC)"

## info: Show project information
info:
	@echo "$(CYAN)Same-Same Vector Database$(NC)"
	@echo "$(BLUE)Version:$(NC) $(VERSION)"
	@echo "$(BLUE)Go Version:$(NC) $(GO_VERSION)"
	@echo "$(BLUE)Git Commit:$(NC) $(GIT_COMMIT)"
	@echo "$(BLUE)Build Time:$(NC) $(BUILD_TIME)"
	@echo "$(BLUE)Docker Image:$(NC) $(DOCKER_IMAGE)"