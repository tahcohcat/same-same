# Image Embedding with CLIP

Same-Same now supports multimodal embedding using OpenCLIP, enabling you to embed both images and text into the same vector space for semantic search across modalities.

## Overview

CLIP (Contrastive Language-Image Pre-training) is a neural network trained on image-text pairs that can:
- Embed images into vector representations
- Embed text into the same vector space
- Enable cross-modal search (find images with text queries, or vice versa)

## Installation

### Requirements

- Python 3.7+ (python3 or python command available)
- pip (Python package manager)

### Install Dependencies

```bash
pip install open_clip_torch pillow torch
```

For GPU support (optional, faster):
```bash
# NVIDIA GPU
pip install open_clip_torch pillow torch torchvision --index-url https://download.pytorch.org/whl/cu118

# Apple Silicon (M1/M2)
pip install open_clip_torch pillow torch torchvision
```

## Quick Start

### 1. Ingest Images

```bash
# Ingest all images in a directory
same-same ingest -e clip images:./photos

# Ingest with custom namespace
same-same ingest -e clip -n vacation images:./vacation_photos

# Non-recursive (don't scan subdirectories)
same-same ingest -e clip --recursive=false images:./photos
```

### 2. Search by Text

Once ingested, search for images using natural language:

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"text": "a sunset over the ocean", "limit": 5}'
```

## Usage Examples

### Example 1: Basic Image Ingestion

```bash
# Directory structure:
# photos/
#   ├── beach/
#   │   ├── sunset.jpg
#   │   └── waves.jpg
#   └── mountains/
#       └── peak.jpg

# Ingest all images recursively
same-same ingest -e clip images:./photos

# Result: 3 images ingested
```

### Example 2: Image List with Labels

Create `images.txt`:
```
beach/sunset.jpg,landscape
beach/waves.jpg,seascape
mountains/peak.jpg,landscape
city/downtown.jpg,urban
```

Ingest:
```bash
same-same ingest -e clip image-list:./images.txt
```

### Example 3: Custom CLIP Model

```bash
# Use larger, more accurate model
same-same ingest -e clip \
  --clip-model ViT-L-14 \
  --clip-pretrain laion2b_s34b_b79k \
  images:./photos
```

### Example 4: Multimodal Search

```bash
# Ingest both text and images with same embedder
same-same ingest -e clip -n documents demo  # Text documents
same-same ingest -e clip -n photos images:./photos  # Images

# Search across both
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"text": "nature and wildlife", "limit": 10}'
```

## CLIP Models

### Available Models

| Model | Dimensions | Speed | Quality | Use Case |
|-------|-----------|-------|---------|----------|
| `ViT-B-32` | 512 | Fast | Good | Default, general use |
| `ViT-B-16` | 512 | Medium | Better | More accurate |
| `ViT-L-14` | 768 | Slow | Best | Production, high quality |

### Pretrained Weights

| Weights | Training Data | Best For |
|---------|--------------|----------|
| `openai` | OpenAI dataset | Default, general images |
| `laion2b_s34b_b79k` | LAION-2B (2.3B images) | Best quality |
| `laion400m_e32` | LAION-400M | Good balance |

### Model Selection

```bash
# Default: ViT-B/32 with OpenAI weights
same-same ingest -e clip images:./photos

# High quality: ViT-L/14 with LAION weights
same-same ingest -e clip \
  --clip-model ViT-L-14 \
  --clip-pretrain laion2b_s34b_b79k \
  images:./photos

# Fast prototyping: ViT-B/32
same-same ingest -e clip \
  --clip-model ViT-B-32 \
  --clip-pretrain openai \
  images:./photos
```

## Image Sources

### Directory Source

Ingest all images in a directory:

```bash
# Recursive (default)
same-same ingest -e clip images:./photos

# Non-recursive
same-same ingest -e clip -r=false images:./photos
```

**Supported formats:**
- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- BMP (.bmp)
- WebP (.webp)

### Image List Source

Provide a list of image paths:

```bash
same-same ingest -e clip image-list:./images.txt
```

**Format:**
```
path/to/image1.jpg
path/to/image2.png,label
path/to/image3.jpg,category,subcategory
```

- Lines starting with `#` are comments
- Paths can be relative or absolute
- Optional: Add comma-separated metadata

## Metadata

Images are automatically tagged with metadata:

```json
{
  "type": "image",
  "filename": "sunset.jpg",
  "path": "beach/sunset.jpg",
  "extension": ".jpg",
  "namespace": "photos"
}
```

From image-list:
```json
{
  "type": "image",
  "filename": "sunset.jpg",
  "path": "/full/path/sunset.jpg",
  "extension": ".jpg",
  "label": "landscape",
  "namespace": "photos"
}
```

## Performance Tips

### 1. Use GPU

Set device to CUDA for faster embedding (requires NVIDIA GPU):

```python
# Modify clip embedder to use GPU
# Currently defaults to CPU
```

### 2. Batch Processing

Process multiple images at once:

```bash
same-same ingest -e clip --batch-size 32 images:./large_collection
```

### 3. Model Selection

- **Prototyping**: Use `ViT-B-32` with `openai` weights
- **Production**: Use `ViT-L-14` with `laion2b_s34b_b79k` weights

### 4. Reduce Dimensions (Future)

For very large datasets, consider dimensionality reduction post-processing.

## Use Cases

### 1. Image Search by Description

```bash
# Ingest product images
same-same ingest -e clip -n products images:./product_photos

# Search for specific products
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"text": "red dress with floral pattern", "namespace": "products", "limit": 10}'
```

### 2. Content Moderation

```bash
# Ingest user-uploaded images
same-same ingest -e clip -n user_content images:./uploads

# Find potentially problematic content
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"text": "inappropriate content keywords", "namespace": "user_content"}'
```

### 3. Photo Organization

```bash
# Ingest personal photo library
same-same ingest -e clip -n my_photos images:~/Pictures

# Find specific memories
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"text": "birthday party with cake", "namespace": "my_photos", "limit": 20}'
```

### 4. Duplicate Detection

```bash
# Ingest image collection
same-same ingest -e clip images:./collection

# Find similar images
curl -X POST http://localhost:8080/api/v1/vectors/search \
  -H "Content-Type: application/json" \
  -d '{"embedding": [/* vector from image */], "limit": 5}'
```

## Troubleshooting

### Python Not Found

```bash
# Error: python not found
# Solution: Install Python 3
# Windows: Download from python.org
# Mac: brew install python3
# Linux: sudo apt install python3
```

### Missing Dependencies

```bash
# Error: No module named 'open_clip'
# Solution: Install dependencies
pip install open_clip_torch pillow torch
```

### Out of Memory

```bash
# Error: CUDA out of memory / RuntimeError
# Solution 1: Use CPU
# Solution 2: Use smaller model (ViT-B-32 instead of ViT-L-14)
# Solution 3: Reduce batch size
same-same ingest -e clip --batch-size 8 images:./photos
```

### Slow Performance

```bash
# Solution 1: Use GPU if available
# Solution 2: Use smaller model
same-same ingest -e clip --clip-model ViT-B-32 images:./photos

# Solution 3: Increase batch size (if you have enough memory)
same-same ingest -e clip --batch-size 64 images:./photos
```

## Advanced: Programmatic Usage

```go
package main

import (
    "github.com/tahcohcat/same-same/internal/embedders/clip"
)

func main() {
    // Create CLIP embedder
    embedder := clip.NewCLIPEmbedder("ViT-B-32", "openai")
    
    // Embed text
    textEmb, err := embedder.Embed("a photo of a cat")
    if err != nil {
        panic(err)
    }
    
    // Embed image
    imageEmb, err := embedder.EmbedImage("./cat.jpg")
    if err != nil {
        panic(err)
    }
    
    // Both embeddings are in the same 512-dimensional space
    // You can compute similarity between them
}
```

## Future Enhancements

Planned features:
- [ ] GPU device selection via flag
- [ ] Batch processing optimization
- [ ] Image preprocessing options
- [ ] Multi-image embedding
- [ ] Image-to-image search API endpoint
- [ ] Support for more CLIP variants (EVA-CLIP, SigLIP)

## References

- [OpenCLIP GitHub](https://github.com/mlfoundations/open_clip)
- [CLIP Paper](https://arxiv.org/abs/2103.00020)
- [LAION Dataset](https://laion.ai/)
