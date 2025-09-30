package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tahcohcat/same-same/internal/embedders/quotes/huggingface"
)

func main() {
	apiKey := os.Getenv("HUGGINGFACE_API_KEY")
	if apiKey == "" {
		log.Fatal("HUGGINGFACE_API_KEY environment variable is required")
	}

	embedder := huggingface.NewHuggingFaceEmbedder(apiKey)

	quote := "The only way to do great work is to love what you do. - Steve Jobs"

	embedding, err := embedder.Embed(quote)
	if err != nil {
		log.Fatalf("Failed to embed quote: %v", err)
	}

	fmt.Printf("Quote: %s\n", quote)
	fmt.Printf("Embedding dimension: %d\n", len(embedding))
	fmt.Printf("First 5 values: %v\n", embedding[:5])
}
