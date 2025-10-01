package clip

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

// NativeCLIPEmbedder implements CLIP using native Go (no Python dependency)
// Uses ONNX Runtime for inference
type NativeCLIPEmbedder struct {
	modelPath      string
	tokenizerPath  string
	imageSize      int
	dimension      int
	vocabulary     map[string]int
	maxTokens      int
}

// NewNativeCLIPEmbedder creates a native Go CLIP embedder
func NewNativeCLIPEmbedder(modelPath, tokenizerPath string) (*NativeCLIPEmbedder, error) {
	embedder := &NativeCLIPEmbedder{
		modelPath:     modelPath,
		tokenizerPath: tokenizerPath,
		imageSize:     224, // Standard CLIP input size
		dimension:     512, // ViT-B/32 default
		maxTokens:     77,  // CLIP max sequence length
	}

	// Load tokenizer vocabulary
	if err := embedder.loadTokenizer(); err != nil {
		return nil, fmt.Errorf("failed to load tokenizer: %w", err)
	}

	return embedder, nil
}

// Embed embeds text using the text encoder
func (n *NativeCLIPEmbedder) Embed(text string) ([]float64, error) {
	// Tokenize text
	tokens := n.tokenize(text)
	
	// For now, return a simple embedding
	// TODO: Implement ONNX inference
	return n.generateTextEmbedding(tokens)
}

// EmbedImage embeds an image using the vision encoder
func (n *NativeCLIPEmbedder) EmbedImage(imagePath string) ([]float64, error) {
	// Load and preprocess image
	img, err := n.loadAndPreprocessImage(imagePath)
	if err != nil {
		return nil, err
	}

	// Generate embedding from preprocessed image
	return n.generateImageEmbedding(img)
}

// EmbedImageBytes embeds image data
func (n *NativeCLIPEmbedder) EmbedImageBytes(imageData []byte) ([]float64, error) {
	// Write to temp file and process
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

	return n.EmbedImage(tmpFile.Name())
}

// Dimensions returns the embedding dimension
func (n *NativeCLIPEmbedder) Dimensions() int {
	return n.dimension
}

// Name returns the embedder name
func (n *NativeCLIPEmbedder) Name() string {
	return "clip-native-go"
}

// loadTokenizer loads the tokenizer vocabulary
func (n *NativeCLIPEmbedder) loadTokenizer() error {
	// Simple vocabulary for demonstration
	// In production, load from a JSON file
	n.vocabulary = make(map[string]int)
	
	// Add basic tokens
	n.vocabulary["<|startoftext|>"] = 49406
	n.vocabulary["<|endoftext|>"] = 49407
	n.vocabulary["!"] = 0
	n.vocabulary[","] = 1
	n.vocabulary["."] = 2
	
	// TODO: Load full BPE vocabulary from file
	return nil
}

// tokenize converts text to token IDs
func (n *NativeCLIPEmbedder) tokenize(text string) []int {
	tokens := []int{n.vocabulary["<|startoftext|>"]}
	
	// Simple whitespace tokenization for demo
	// TODO: Implement proper BPE tokenization
	words := strings.Fields(strings.ToLower(text))
	for _, word := range words {
		if id, ok := n.vocabulary[word]; ok {
			tokens = append(tokens, id)
		} else {
			// Unknown token - use a default
			tokens = append(tokens, 0)
		}
	}
	
	tokens = append(tokens, n.vocabulary["<|endoftext|>"])
	
	// Pad or truncate to maxTokens
	if len(tokens) < n.maxTokens {
		padding := make([]int, n.maxTokens-len(tokens))
		tokens = append(tokens, padding...)
	} else if len(tokens) > n.maxTokens {
		tokens = tokens[:n.maxTokens]
	}
	
	return tokens
}

// loadAndPreprocessImage loads and preprocesses an image
func (n *NativeCLIPEmbedder) loadAndPreprocessImage(imagePath string) ([]float32, error) {
	// Open image file
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize to 224x224
	resized := resize.Resize(uint(n.imageSize), uint(n.imageSize), img, resize.Lanczos3)

	// Convert to float32 array and normalize
	// CLIP normalization: mean=[0.48145466, 0.4578275, 0.40821073], std=[0.26862954, 0.26130258, 0.27577711]
	pixels := make([]float32, n.imageSize*n.imageSize*3)
	
	mean := []float32{0.48145466, 0.4578275, 0.40821073}
	std := []float32{0.26862954, 0.26130258, 0.27577711}
	
	idx := 0
	for y := 0; y < n.imageSize; y++ {
		for x := 0; x < n.imageSize; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()
			
			// Convert from uint32 (0-65535) to float32 (0-1)
			pixels[idx] = (float32(r)/65535.0 - mean[0]) / std[0]
			pixels[idx+1] = (float32(g)/65535.0 - mean[1]) / std[1]
			pixels[idx+2] = (float32(b)/65535.0 - mean[2]) / std[2]
			
			idx += 3
		}
	}

	return pixels, nil
}

// generateTextEmbedding generates embedding from tokens
// TODO: Replace with ONNX inference
func (n *NativeCLIPEmbedder) generateTextEmbedding(tokens []int) ([]float64, error) {
	// Placeholder: Generate a simple embedding based on tokens
	// In production, this would use ONNX Runtime to run the text encoder
	
	embedding := make([]float64, n.dimension)
	
	// Simple hash-based embedding for demonstration
	for i := range embedding {
		sum := 0.0
		for j, token := range tokens {
			sum += float64(token * (i + j + 1))
		}
		embedding[i] = math.Sin(sum) * 0.1
	}
	
	// Normalize to unit length
	return normalizeVector(embedding), nil
}

// generateImageEmbedding generates embedding from preprocessed image
// TODO: Replace with ONNX inference
func (n *NativeCLIPEmbedder) generateImageEmbedding(pixels []float32) ([]float64, error) {
	// Placeholder: Generate a simple embedding based on pixel values
	// In production, this would use ONNX Runtime to run the vision encoder
	
	embedding := make([]float64, n.dimension)
	
	// Simple pooling-based embedding for demonstration
	for i := range embedding {
		sum := 0.0
		start := (i * len(pixels)) / len(embedding)
		end := ((i + 1) * len(pixels)) / len(embedding)
		
		for j := start; j < end && j < len(pixels); j++ {
			sum += float64(pixels[j])
		}
		embedding[i] = sum / float64(end-start)
	}
	
	// Normalize to unit length
	return normalizeVector(embedding), nil
}

// normalizeVector normalizes a vector to unit length
func normalizeVector(vec []float64) []float64 {
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	
	if norm == 0 {
		return vec
	}
	
	normalized := make([]float64, len(vec))
	for i, v := range vec {
		normalized[i] = v / norm
	}
	return normalized
}
