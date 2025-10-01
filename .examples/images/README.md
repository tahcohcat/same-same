# Example Images for CLIP Ingestion

This directory contains sample images for testing CLIP-based image embedding and ingestion.

## Setup

To use CLIP embeddings, you need to install the required Python dependencies:

```bash
pip install open_clip_torch pillow torch
```

## Usage

### Ingest all images in this directory

```bash
same-same ingest -e clip images:.examples/images
```

### Ingest with custom CLIP model

```bash
# Use ViT-L/14 with LAION weights (higher quality, slower)
same-same ingest -e clip --clip-model ViT-L-14 --clip-pretrain laion2b_s34b_b79k images:.examples/images

# Use ViT-B/32 with OpenAI weights (default, faster)
same-same ingest -e clip --clip-model ViT-B-32 --clip-pretrain openai images:.examples/images
```

### Ingest from image list

Create a file `images.txt`:
```
cat.jpg,animal
dog.jpg,animal
car.jpg,vehicle
```

Then ingest:
```bash
same-same ingest -e clip image-list:.examples/images/images.txt
```

## Multimodal Search

Once images are ingested with CLIP, you can search for them using text:

```bash
# Search for images matching the text "a cute cat"
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"text": "a cute cat", "limit": 5}'
```

This works because CLIP embeds both images and text into the same vector space!

## Supported Formats

- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- BMP (.bmp)
- WebP (.webp)

## CLIP Models

Available models:
- `ViT-B-32` (default) - 512 dimensions, fast
- `ViT-B-16` - 512 dimensions, more accurate
- `ViT-L-14` - 768 dimensions, best quality but slower

Pretrained weights:
- `openai` (default) - Original OpenAI weights
- `laion2b_s34b_b79k` - Trained on LAION-2B dataset
- `laion400m_e32` - Trained on LAION-400M

## Example: Search Similar Images

1. Ingest images:
```bash
same-same ingest -e clip -n photos images:./my_photos
```

2. Start server (if not running):
```bash
same-same serve
```

3. Search by text:
```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"text": "sunset over mountains", "namespace": "photos", "limit": 3}'
```

4. The results will include images that match the semantic meaning of your query!
