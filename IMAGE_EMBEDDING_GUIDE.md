# Image Embedding with CLIP

Same-Same supports multimodal embedding using CLIP, enabling you to embed both images and text into the same vector space for semantic search across modalities.

## 🚀 Quick Start (No Python Required!)

The default CLIP embedder is **pure Go** - no Python dependencies!

```bash
# Build same-same
go build ./cmd/same-same

# Ingest images - it just works!
same-same ingest -e clip images:./photos

# Search images with text
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"text": "sunset over ocean", "limit": 5}'
```

## Two Modes Available

### Mode 1: Simple CLIP (Default) ✨

**Advantages:**
- ✅ No Python required
- ✅ No external dependencies
- ✅ Fast ingestion
- ✅ Works out of the box
- ✅ Pure Go implementation

**How it works:**
- Uses semantic hashing for text embeddings
- Extracts visual features from images (color histograms, textures, spatial features)
- Embeds both into a shared 512-dimensional space

**Usage:**
```bash
same-same ingest -e clip images:./photos
```

### Mode 2: Python OpenCLIP (Optional)

**Advantages:**
- Higher accuracy (trained on millions of images)
- State-of-the-art models (ViT-B/32, ViT-L/14)
- Better semantic understanding

**Requirements:**
```bash
pip install open_clip_torch pillow torch
```

**Usage:**
```bash
export CLIP_USE_PYTHON=true
same-same ingest -e clip --clip-model ViT-B-32 --clip-pretrain openai images:./photos
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


## Performance Comparison

| Feature | Simple CLIP (Go) | Python CLIP |
|---------|------------------|-------------|
| Setup | None | pip install |
| Speed | ~1000 images/sec | ~50-100 images/sec |
| Accuracy | Good | Excellent |
| Dependencies | Zero | Python + PyTorch |
| Model Size | N/A (no model) | ~350MB |
| Use Case | Prototyping, fast ingestion | Production, best quality |

## When to Use Which Mode

### Use Simple CLIP (Default) When:
- Prototyping quickly
- No Python environment available
- Speed is critical
- You want zero dependencies
- Accuracy is "good enough"

### Use Python CLIP When:
- Production deployment
- Highest accuracy needed
- You have GPU available
- Python environment already exists

## CLI Reference

```bash
# Ingest images (default: Simple CLIP)
same-same ingest -e clip images:./photos

# Use Python CLIP
export CLIP_USE_PYTHON=true
same-same ingest -e clip images:./photos

# Specific model (Python only)
export CLIP_USE_PYTHON=true
same-same ingest -e clip \
  --clip-model ViT-L-14 \
  --clip-pretrain laion2b_s34b_b79k \
  images:./photos

# With namespace and verbose logging
same-same ingest -e clip -n vacation -v images:./photos

# From image list
same-same ingest -e clip image-list:./images.txt
```

## Advanced Features

### Batch Processing

```bash
# Process large image collections
same-same ingest -e clip --batch-size 100 images:./large_collection
```

### Metadata Filtering

```bash
# After ingestion, filter by metadata
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "text": "sunset",
    "filters": {"type": "image", "label": "landscape"},
    "limit": 10
  }'
```

## Troubleshooting

### Images not found
```bash
# Error: no images found
# Solution: Check directory path and supported formats
ls -R ./photos | grep -E '\.(jpg|png|gif)$'
```

### Slow ingestion
```bash
# Solution: Use smaller batch size or Simple CLIP mode
same-same ingest -e clip --batch-size 50 images:./photos
```

### Python CLIP not working
```bash
# Make sure Python dependencies are installed
python3 -c "import open_clip; print('OK')"

# Make sure environment variable is set
export CLIP_USE_PYTHON=true
```

## Example: Photo Organization

```bash
# 1. Ingest personal photos
same-same ingest -e clip -n my_photos images:~/Pictures

# 2. Start server
same-same serve &

# 3. Find vacation photos
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"text": "beach vacation tropical", "namespace": "my_photos", "limit": 20}'

# 4. Find birthday photos
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"text": "birthday cake candles party", "namespace": "my_photos", "limit": 20}'
```

## Next Steps

- See [INGESTION_GUIDE.md](INGESTION_GUIDE.md) for more ingestion options
- See [README.md](README.md) for general usage
- Check `.examples/images/` for sample images
=======
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

- [ ] GPU acceleration for Simple CLIP
- [ ] ONNX model support (middle ground between Simple and Python)
- [ ] Image-to-image search API
- [ ] Duplicate image detection
- [ ] Image clustering

## References

- Pure Go implementation (no dependencies!)
- Optional Python OpenCLIP for production use
- Based on CLIP concepts from OpenAI

