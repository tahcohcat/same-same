package clip

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/tahcohcat/same-same/internal/embedders"
)

// CLIPEmbedder implements multimodal embedding using OpenCLIP
type CLIPEmbedder struct {
	model      string
	pretrained string
	device     string
	pythonPath string
	dimension  int
}

// EmbeddingResponse represents the response from the Python CLIP service
type EmbeddingResponse struct {
	Embedding  []float64 `json:"embedding"`
	Dimensions int       `json:"dimensions"`
	Error      string    `json:"error,omitempty"`
}

// NewCLIPEmbedder creates a new CLIP embedder
// model: e.g., "ViT-B-32", "ViT-L-14"
// pretrained: e.g., "openai", "laion2b_s34b_b79k"
func NewCLIPEmbedder(model, pretrained string) *CLIPEmbedder {
	if model == "" {
		model = "ViT-B-32"
	}
	if pretrained == "" {
		pretrained = "openai"
	}

	return &CLIPEmbedder{
		model:      model,
		pretrained: pretrained,
		device:     "cpu", // Default to CPU, use "cuda" for GPU
		dimension:  512,   // ViT-B/32 default, will be updated on first use
	}
}

// SetDevice sets the device for inference (cpu, cuda, mps)
func (c *CLIPEmbedder) SetDevice(device string) {
	c.device = device
}

// Embed embeds text using CLIP
func (c *CLIPEmbedder) Embed(text string) ([]float64, error) {
	return c.embedText(text)
}

// EmbedImage embeds an image file using CLIP
func (c *CLIPEmbedder) EmbedImage(imagePath string) ([]float64, error) {
	return c.embedImage(imagePath, false)
}

// EmbedImageBytes embeds image data using CLIP
func (c *CLIPEmbedder) EmbedImageBytes(imageData []byte) ([]float64, error) {
	// Write to temp file
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

	return c.embedImage(tmpFile.Name(), false)
}

// Dimensions returns the embedding dimension
func (c *CLIPEmbedder) Dimensions() int {
	return c.dimension
}

// Name returns the embedder name
func (c *CLIPEmbedder) Name() string {
	return fmt.Sprintf("clip-%s-%s", c.model, c.pretrained)
}

func (c *CLIPEmbedder) embedText(text string) ([]float64, error) {
	script := c.generatePythonScript()
	return c.runPythonScript(script, "text", text)
}

func (c *CLIPEmbedder) embedImage(path string, isBytes bool) ([]float64, error) {
	script := c.generatePythonScript()
	return c.runPythonScript(script, "image", path)
}

func (c *CLIPEmbedder) runPythonScript(script, mode, input string) ([]float64, error) {
	// Check if Python is available
	pythonCmd := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err := exec.LookPath("python"); err != nil {
			return nil, fmt.Errorf("python not found - required for CLIP embeddings")
		}
		pythonCmd = "python"
	}

	// Create temporary script file
	tmpScript, err := os.CreateTemp("", "clip_embed_*.py")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp script: %w", err)
	}
	defer os.Remove(tmpScript.Name())

	if _, err := tmpScript.WriteString(script); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}
	tmpScript.Close()

	// Execute Python script
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, pythonCmd, tmpScript.Name(), mode, input, c.model, c.pretrained, c.device)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("python script failed: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON response
	var response EmbeddingResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w\nOutput: %s", err, string(output))
	}

	if response.Error != "" {
		return nil, fmt.Errorf("embedding error: %s", response.Error)
	}

	// Update dimension if needed
	if response.Dimensions > 0 && c.dimension != response.Dimensions {
		c.dimension = response.Dimensions
	}

	return response.Embedding, nil
}

func (c *CLIPEmbedder) generatePythonScript() string {
	return `#!/usr/bin/env python3
import sys
import json
import torch
from PIL import Image
import open_clip

def main():
    if len(sys.argv) < 6:
        print(json.dumps({"error": "Usage: script.py <mode> <input> <model> <pretrained> <device>"}))
        sys.exit(1)
    
    mode = sys.argv[1]  # 'text' or 'image'
    input_data = sys.argv[2]
    model_name = sys.argv[3]
    pretrained = sys.argv[4]
    device = sys.argv[5]
    
    try:
        # Load model
        model, _, preprocess = open_clip.create_model_and_transforms(
            model_name, 
            pretrained=pretrained,
            device=device
        )
        model.eval()
        
        with torch.no_grad():
            if mode == 'text':
                # Tokenize and embed text
                tokenizer = open_clip.get_tokenizer(model_name)
                text_tokens = tokenizer([input_data]).to(device)
                text_features = model.encode_text(text_tokens)
                text_features = text_features / text_features.norm(dim=-1, keepdim=True)
                embedding = text_features[0].cpu().numpy().tolist()
                
            elif mode == 'image':
                # Load and embed image
                image = Image.open(input_data).convert('RGB')
                image_input = preprocess(image).unsqueeze(0).to(device)
                image_features = model.encode_image(image_input)
                image_features = image_features / image_features.norm(dim=-1, keepdim=True)
                embedding = image_features[0].cpu().numpy().tolist()
            else:
                print(json.dumps({"error": f"Unknown mode: {mode}"}))
                sys.exit(1)
        
        result = {
            "embedding": embedding,
            "dimensions": len(embedding)
        }
        print(json.dumps(result))
        
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)

if __name__ == "__main__":
    main()
`
}

// EmbedBatch embeds multiple texts in a batch (more efficient)
func (c *CLIPEmbedder) EmbedBatch(texts []string) ([][]float64, error) {
	// TODO: Implement batch processing
	results := make([][]float64, len(texts))
	for i, text := range texts {
		embedding, err := c.Embed(text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text %d: %w", i, err)
		}
		results[i] = embedding
	}
	return results, nil
}

// EmbedImageBatch embeds multiple images in a batch
func (c *CLIPEmbedder) EmbedImageBatch(imagePaths []string) ([][]float64, error) {
	// TODO: Implement batch processing
	results := make([][]float64, len(imagePaths))
	for i, path := range imagePaths {
		embedding, err := c.EmbedImage(path)
		if err != nil {
			return nil, fmt.Errorf("failed to embed image %d: %w", i, err)
		}
		results[i] = embedding
	}
	return results, nil
}

// Ensure CLIPEmbedder implements the interfaces
var _ embedders.Embedder = (*CLIPEmbedder)(nil)
var _ embedders.ImageEmbedder = (*CLIPEmbedder)(nil)
var _ embedders.MultiModalEmbedder = (*CLIPEmbedder)(nil)
