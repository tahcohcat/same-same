package embedders

// ImageEmbedder can embed images into vectors
type ImageEmbedder interface {
	EmbedImage(imagePath string) ([]float64, error)
	EmbedImageBytes(imageData []byte) ([]float64, error)
	Name() string
}

// MultiModalEmbedder can embed both text and images into the same vector space
type MultiModalEmbedder interface {
	Embedder
	ImageEmbedder
	
	// Dimensions returns the embedding dimension
	Dimensions() int
}
