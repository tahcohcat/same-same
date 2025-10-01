package clip

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

// SimpleCLIPEmbedder is a lightweight CLIP-inspired embedder
// Uses simple hashing and image features instead of deep learning models
// No Python or ONNX dependencies - pure Go!
type SimpleCLIPEmbedder struct {
	imageSize int
	dimension int
}

// NewSimpleCLIPEmbedder creates a simple CLIP-like embedder
func NewSimpleCLIPEmbedder() *SimpleCLIPEmbedder {
	return &SimpleCLIPEmbedder{
		imageSize: 224,
		dimension: 512,
	}
}

// Embed embeds text using semantic hashing
func (s *SimpleCLIPEmbedder) Embed(text string) ([]float64, error) {
	return s.embedText(text), nil
}

// EmbedImage embeds an image using visual features
func (s *SimpleCLIPEmbedder) EmbedImage(imagePath string) ([]float64, error) {
	img, err := s.loadImage(imagePath)
	if err != nil {
		return nil, err
	}
	return s.embedImage(img), nil
}

// EmbedImageBytes embeds image data
func (s *SimpleCLIPEmbedder) EmbedImageBytes(imageData []byte) ([]float64, error) {
	tmpFile, err := os.CreateTemp("", "clip_image_*.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(imageData); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	return s.EmbedImage(tmpFile.Name())
}

// Dimensions returns the embedding dimension
func (s *SimpleCLIPEmbedder) Dimensions() int {
	return s.dimension
}

// Name returns the embedder name
func (s *SimpleCLIPEmbedder) Name() string {
	return "clip-simple-go"
}

// embedText creates a semantic embedding from text
func (s *SimpleCLIPEmbedder) embedText(text string) []float64 {
	embedding := make([]float64, s.dimension)
	
	// Normalize text
	text = strings.ToLower(strings.TrimSpace(text))
	words := strings.Fields(text)
	
	// Create embedding using multiple hash functions
	for i := 0; i < s.dimension; i++ {
		value := 0.0
		
		// Word-level features
		for j, word := range words {
			// Hash word with position
			hash := hashString(word, i, j)
			value += math.Sin(float64(hash)) * (1.0 / float64(len(words)))
		}
		
		// Character n-grams (for semantic similarity)
		for j := 0; j < len(text)-2; j++ {
			trigram := text[j : j+3]
			hash := hashString(trigram, i, 0)
			value += math.Cos(float64(hash)) * 0.1
		}
		
		embedding[i] = value
	}
	
	return normalizeVector(embedding)
}

// embedImage creates a visual embedding from image
func (s *SimpleCLIPEmbedder) embedImage(img image.Image) []float64 {
	embedding := make([]float64, s.dimension)
	
	// Resize image to standard size
	resized := resize.Resize(uint(s.imageSize), uint(s.imageSize), img, resize.Lanczos3)
	
	// Extract visual features
	// 1. Color histogram features (first 256 dims)
	colorHist := s.extractColorHistogram(resized)
	copy(embedding[:256], colorHist)
	
	// 2. Texture features (next 128 dims)
	textureFeatures := s.extractTextureFeatures(resized)
	copy(embedding[256:384], textureFeatures)
	
	// 3. Spatial features (remaining dims)
	spatialFeatures := s.extractSpatialFeatures(resized)
	copy(embedding[384:], spatialFeatures)
	
	return normalizeVector(embedding)
}

// loadImage loads an image from file
func (s *SimpleCLIPEmbedder) loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// extractColorHistogram extracts color distribution features
func (s *SimpleCLIPEmbedder) extractColorHistogram(img image.Image) []float64 {
	bounds := img.Bounds()
	histogram := make([]float64, 256)
	
	// Sample pixels and build histogram
	sampleSize := 16 // Sample every 16th pixel
	count := 0
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleSize {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleSize {
			r, g, b, _ := img.At(x, y).RGBA()
			
			// Convert to grayscale and bin
			gray := (r + g + b) / 3
			bin := int(gray >> 8) // 0-255
			if bin < 256 {
				histogram[bin]++
				count++
			}
		}
	}
	
	// Normalize
	if count > 0 {
		for i := range histogram {
			histogram[i] /= float64(count)
		}
	}
	
	return histogram
}

// extractTextureFeatures extracts texture and edge features
func (s *SimpleCLIPEmbedder) extractTextureFeatures(img image.Image) []float64 {
	bounds := img.Bounds()
	features := make([]float64, 128)
	
	// Simple edge detection in multiple directions
	sampleSize := 8
	
	for y := bounds.Min.Y; y < bounds.Max.Y-sampleSize; y += sampleSize {
		for x := bounds.Min.X; x < bounds.Max.X-sampleSize; x += sampleSize {
			// Get 2x2 patch
			r1, g1, b1, _ := img.At(x, y).RGBA()
			r2, g2, b2, _ := img.At(x+sampleSize, y).RGBA()
			r3, g3, b3, _ := img.At(x, y+sampleSize).RGBA()
			_, _, _, _ = img.At(x+sampleSize, y+sampleSize).RGBA()
			
			// Horizontal edge
			hEdge := math.Abs(float64(r2-r1) + float64(g2-g1) + float64(b2-b1))
			// Vertical edge
			vEdge := math.Abs(float64(r3-r1) + float64(g3-g1) + float64(b3-b1))
			
			// Bin edges
			hBin := int(hEdge/65536*64) % 64
			vBin := int(vEdge/65536*64) % 64
			
			features[hBin]++
			features[64+vBin]++
		}
	}
	
	// Normalize
	sum := 0.0
	for _, v := range features {
		sum += v
	}
	if sum > 0 {
		for i := range features {
			features[i] /= sum
		}
	}
	
	return features
}

// extractSpatialFeatures extracts spatial/layout features
func (s *SimpleCLIPEmbedder) extractSpatialFeatures(img image.Image) []float64 {
	bounds := img.Bounds()
	features := make([]float64, 128)
	
	// Divide image into grid and extract features from each cell
	gridSize := 8
	cellWidth := bounds.Dx() / gridSize
	cellHeight := bounds.Dy() / gridSize
	
	idx := 0
	for gy := 0; gy < gridSize && idx < len(features); gy++ {
		for gx := 0; gx < gridSize && idx < len(features); gx++ {
			// Sample center of each grid cell
			cx := bounds.Min.X + gx*cellWidth + cellWidth/2
			cy := bounds.Min.Y + gy*cellHeight + cellHeight/2
			
			if cx < bounds.Max.X && cy < bounds.Max.Y {
				r, g, b, _ := img.At(cx, cy).RGBA()
				
				// Average RGB as spatial feature
				avg := float64(r+g+b) / (3.0 * 65535.0)
				features[idx] = avg
				
				// Brightness as secondary feature
				if idx+1 < len(features) {
					brightness := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
					features[idx+1] = brightness / 65535.0
				}
				
				idx += 2
			}
		}
	}
	
	return features
}

// hashString creates a hash from string with seed
func hashString(s string, seed1, seed2 int) uint32 {
	h := md5.New()
	h.Write([]byte(s))
	h.Write([]byte{byte(seed1), byte(seed2)})
	sum := h.Sum(nil)
	return binary.BigEndian.Uint32(sum[:4])
}
