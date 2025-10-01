#!/bin/bash

# CLIP Image Embedding Quick Start
# This script demonstrates how to use CLIP for image embedding

echo "=== CLIP Image Embedding Quick Start ==="
echo ""

# Check if Python is installed
if ! command -v python3 &> /dev/null && ! command -v python &> /dev/null; then
    echo "❌ Python not found. Please install Python 3.7+"
    exit 1
fi

echo "✓ Python found"

# Check if dependencies are installed
echo ""
echo "Checking dependencies..."
python3 -c "import open_clip" 2>/dev/null || python -c "import open_clip" 2>/dev/null
if [ $? -ne 0 ]; then
    echo "❌ OpenCLIP not installed"
    echo ""
    echo "Installing dependencies..."
    pip install open_clip_torch pillow torch
else
    echo "✓ OpenCLIP installed"
fi

# Build same-same if needed
if [ ! -f "./same-same" ]; then
    echo ""
    echo "Building same-same..."
    go build ./cmd/same-same
fi

echo ""
echo "=== Ready to use CLIP! ==="
echo ""
echo "Example commands:"
echo ""
echo "1. Ingest images from a directory:"
echo "   same-same ingest -e clip images:./photos"
echo ""
echo "2. Ingest with custom namespace:"
echo "   same-same ingest -e clip -n vacation images:./vacation_photos"
echo ""
echo "3. Use high-quality model:"
echo "   same-same ingest -e clip --clip-model ViT-L-14 --clip-pretrain laion2b_s34b_b79k images:./photos"
echo ""
echo "4. Search images by text (after ingesting and starting server):"
echo "   curl -X POST http://localhost:8080/api/v1/search \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"text\": \"sunset over ocean\", \"limit\": 5}'"
echo ""
