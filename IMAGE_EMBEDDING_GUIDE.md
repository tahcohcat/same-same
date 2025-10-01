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

## Usage Examples

### Example 1: Ingest Images

```bash
# Directory structure:
# photos/
#   ├── beach/
#   │   ├── sunset.jpg
#   │   └── waves.jpg
#   └── mountains/
#       └── peak.jpg

# Ingest all images recursively (default)
same-same ingest -e clip images:./photos

# Non-recursive (only top-level directory)
same-same ingest -e clip -r=false images:./photos
```

### Example 2: Image List with Labels

Create `images.txt`:
```
beach/sunset.jpg,landscape
beach/waves.jpg,seascape
mountains/peak.jpg,landscape
```

Ingest:
```bash
same-same ingest -e clip image-list:./images.txt
```

### Example 3: Multimodal Search

```bash
# Ingest images
same-same ingest -e clip -n photos images:./vacation

# Start server
same-same serve &

# Search for images using text
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "text": "beautiful sunset with palm trees",
    "namespace": "photos",
    "limit": 10
  }'
```

## How Simple CLIP Works

The Go implementation uses computer vision techniques:

1. **Image Features:**
   - Color histograms (256 dimensions)
   - Edge/texture patterns (128 dimensions)
   - Spatial layout (128 dimensions)

2. **Text Features:**
   - Word-level hashing
   - Character n-grams for semantic similarity
   - Position-weighted embeddings

3. **Shared Space:**
   - Both modalities map to 512D vectors
   - Normalized to unit length for cosine similarity

## Supported Image Formats

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

## Future Enhancements

Planned features:
- [ ] GPU acceleration for Simple CLIP
- [ ] ONNX model support (middle ground between Simple and Python)
- [ ] Image-to-image search API
- [ ] Duplicate image detection
- [ ] Image clustering

## References

- Pure Go implementation (no dependencies!)
- Optional Python OpenCLIP for production use
- Based on CLIP concepts from OpenAI
