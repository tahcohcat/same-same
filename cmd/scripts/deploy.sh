#!/bin/bash

# Deployment script for Same-Same Vector Database
set -e

echo "🚀 Deploying Same-Same Vector Database..."

# Check prerequisites
if [ ! -f .env ]; then
    echo "❌ .env file not found. Please copy .env.example to .env and configure it."
    exit 1
fi

if [ -z "$GEMINI_API_KEY" ]; then
    echo "❌ GEMINI_API_KEY not set. Please check your .env file."
    exit 1
fi

# Build and start services
echo "📦 Building Docker images..."
docker-compose build --no-cache

echo "🔄 Starting services..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 30

# Health check
echo "🔍 Performing health checks..."
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Same-Same API is healthy"
else
    echo "❌ Same-Same API health check failed"
    docker-compose logs same-same
    exit 1
fi

if curl -f http://localhost:80 > /dev/null 2>&1; then
    echo "✅ Nginx is serving content"
else
    echo "❌ Nginx health check failed"
    docker-compose logs nginx
    exit 1
fi

# Load sample data
echo "📊 Loading sample data..."
curl -X POST http://localhost:8080/api/v1/vectors/embed \
    -H "Content-Type: application/json" \
    -d '{"text": "The only way to do great work is to love what you do.", "author": "Steve Jobs"}' \
    > /dev/null 2>&1 || echo "⚠️ Sample data loading failed (may already exist)"

echo "🎉 Deployment completed successfully!"
echo ""
echo "📱 Access the application:"
echo "  - Web Interface: http://localhost"
echo "  - API Endpoint: http://localhost:8080"
echo "  - Health Check: http://localhost:8080/health"
echo ""
echo "📊 Optional monitoring (run with --profile monitoring):"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000 (admin/admin123)"
