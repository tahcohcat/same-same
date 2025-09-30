package local

import (
	"fmt"
	"time"
)

// ExampleUsage demonstrates how to use the local storage system
func ExampleUsage() error {
	// 1. Create local storage
	storage, err := NewLocalStorage("./data/storage")
	if err != nil {
		return err
	}
	defer storage.Close()

	// 2. Create a collection with schema
	schema := &CollectionSchema{
		Fields: map[string]FieldDefinition{
			"author": {
				Type:        "string",
				Description: "Author of the quote",
				Indexed:     true,
			},
			"year": {
				Type:        "number",
				Description: "Year of publication",
				Indexed:     true,
			},
			"category": {
				Type:        "string",
				Description: "Category of content",
				Indexed:     true,
				Enum:        []string{"science", "philosophy", "literature"},
			},
			"verified": {
				Type:        "boolean",
				Description: "Whether the quote is verified",
				Indexed:     true,
			},
		},
		Required: []string{"author", "year"},
		Indexes: []Index{
			{Name: "author_year", Fields: []string{"author", "year"}, Unique: false},
		},
		VectorConfig: &VectorConfig{
			Dimension:    768,
			EmbedderType: "local",
			Metric:       "cosine",
		},
	}

	_, err = storage.CreateCollection("quotes", "Famous quotes collection", schema)
	if err != nil {
		return err
	}

	// 3. Store a text document with embedding
	doc1 := &Document{
		ID:   "quote_001",
		Type: TypeText,
		Metadata: map[string]interface{}{
			"author":   "Einstein",
			"year":     1930,
			"category": "science",
			"verified": true,
			"tags":     []string{"time", "relativity", "philosophy"},
		},
		Content: &ContentData{
			Type: TypeText,
			Text: &TextContent{
				Raw:      "Time is relative; its only worth depends upon what we do as it is passing.",
				Language: "en",
				Format:   "plain",
			},
		},
		Embedding: &EmbeddingData{
			Vector:    []float64{0.1, 0.2, 0.3}, // Simplified example
			Dimension: 3,
			Model:     "local-tfidf",
			CreatedAt: time.Now(),
		},
		Tags: []string{"physics", "time", "philosophy"},
	}

	if err := storage.StoreDocument("quotes", doc1); err != nil {
		return err
	}

	// 4. Store an image document (multimodal)
	doc2 := &Document{
		ID:   "photo_001",
		Type: TypeImage,
		Metadata: map[string]interface{}{
			"photographer": "Ansel Adams",
			"year":         1942,
			"location":     "Yosemite",
			"verified":     true,
		},
		Content: &ContentData{
			Type: TypeImage,
			Image: &ImageContent{
				Format:     "jpg",
				Width:      1920,
				Height:     1080,
				Size:       2048000,
				Path:       "./content/photos/photo_001.jpg",
				ColorSpace: "RGB",
			},
		},
		Embedding: &EmbeddingData{
			Vector:    []float64{0.4, 0.5, 0.6}, // Image embedding
			Dimension: 3,
			Model:     "clip-vit-b32",
			CreatedAt: time.Now(),
		},
		Tags: []string{"landscape", "nature", "mountains"},
	}

	if err := storage.StoreDocument("photos", doc2); err != nil {
		// Collection doesn't exist, create it first
		_, _ = storage.CreateCollection("photos", "Photo collection", nil)
		if err := storage.StoreDocument("photos", doc2); err != nil {
			return err
		}
	}

	// 5. Query by metadata
	results, err := storage.QueryByMetadata("quotes", map[string]interface{}{
		"author":   "Einstein",
		"verified": true,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Found %d documents matching query\n", len(results))

	// 6. Get specific document
	doc, err := storage.GetDocument("quotes", "quote_001")
	if err != nil {
		return err
	}

	fmt.Printf("Retrieved document: %s\n", doc.ID)

	// 7. Export collection
	if err := storage.Export("quotes", "./backup/quotes.json"); err != nil {
		return err
	}

	// 8. Get statistics
	stats := storage.GetStats()
	fmt.Printf("Storage stats: %+v\n", stats)

	return nil
}

// ExampleMultimodalStorage demonstrates storing different content types
func ExampleMultimodalStorage() error {
	storage, err := NewLocalStorage("./data/multimodal")
	if err != nil {
		return err
	}
	defer storage.Close()

	// Create multimodal collection
	_, err = storage.CreateCollection("multimedia", "Multimodal content", nil)
	if err != nil {
		return err
	}

	// Store audio document
	audioDoc := &Document{
		ID:   "audio_001",
		Type: TypeAudio,
		Metadata: map[string]interface{}{
			"title":  "Symphony No. 5",
			"artist": "Beethoven",
			"genre":  "classical",
		},
		Content: &ContentData{
			Type: TypeAudio,
			Audio: &AudioContent{
				Format:     "mp3",
				Duration:   432.5,
				SampleRate: 44100,
				Bitrate:    320,
				Size:       10485760,
				Path:       "./content/audio/symphony_5.mp3",
			},
		},
		Embedding: &EmbeddingData{
			Vector:    []float64{0.7, 0.8, 0.9},
			Dimension: 3,
			Model:     "music-embedder",
		},
	}

	if err := storage.StoreDocument("multimedia", audioDoc); err != nil {
		return err
	}

	// Store video document
	videoDoc := &Document{
		ID:   "video_001",
		Type: TypeVideo,
		Metadata: map[string]interface{}{
			"title":    "Nature Documentary",
			"director": "David Attenborough",
			"duration": 3600,
		},
		Content: &ContentData{
			Type: TypeVideo,
			Video: &VideoContent{
				Format:    "mp4",
				Duration:  3600.0,
				Width:     3840,
				Height:    2160,
				FrameRate: 60.0,
				Size:      5368709120,
				Path:      "./content/video/nature_doc.mp4",
				Thumbnail: "./content/video/thumbnails/nature_doc_thumb.jpg",
			},
		},
		Embedding: &EmbeddingData{
			Vector:    []float64{0.15, 0.25, 0.35},
			Dimension: 3,
			Model:     "video-embedder",
		},
	}

	if err := storage.StoreDocument("multimedia", videoDoc); err != nil {
		return err
	}

	// Store document with relations
	relatedDoc := &Document{
		ID:   "article_001",
		Type: TypeDocument,
		Metadata: map[string]interface{}{
			"title":  "The Art of Nature Photography",
			"author": "Jane Smith",
		},
		Relations: []Relation{
			{Type: "references", DocumentID: "photo_001"},
			{Type: "references", DocumentID: "video_001"},
		},
	}

	if err := storage.StoreDocument("multimedia", relatedDoc); err != nil {
		return err
	}

	return nil
}
