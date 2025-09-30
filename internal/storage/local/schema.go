package local

import (
	"time"
)

// StorageSchema represents the top-level storage structure
type StorageSchema struct {
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    StorageMetadata        `json:"metadata"`
	Collections map[string]*Collection `json:"collections"`
}

// StorageMetadata contains storage-level information
type StorageMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// Collection represents a logical grouping of documents
type Collection struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Schema      *CollectionSchema    `json:"schema,omitempty"`
	Documents   map[string]*Document `json:"documents"`
	Stats       CollectionStats      `json:"stats"`
}

// CollectionSchema defines the structure and constraints for a collection
type CollectionSchema struct {
	Fields       map[string]FieldDefinition `json:"fields"`
	Required     []string                   `json:"required,omitempty"`
	Indexes      []Index                    `json:"indexes,omitempty"`
	VectorConfig *VectorConfig              `json:"vector_config,omitempty"`
}

// FieldDefinition describes a metadata field
type FieldDefinition struct {
	Type        string   `json:"type"` // string, number, boolean, array, object
	Description string   `json:"description,omitempty"`
	Indexed     bool     `json:"indexed"`
	Unique      bool     `json:"unique,omitempty"`
	Enum        []string `json:"enum,omitempty"` // Allowed values
}

// Index represents a metadata index for fast retrieval
type Index struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
	Unique bool     `json:"unique"`
}

// VectorConfig defines vector embedding configuration
type VectorConfig struct {
	Dimension    int    `json:"dimension"`
	EmbedderType string `json:"embedder_type"` // local, gemini, huggingface
	Metric       string `json:"metric"`        // cosine, euclidean, dot
}

// Document represents a single stored item (multimodal support)
type Document struct {
	ID           string                 `json:"id"`
	CollectionID string                 `json:"collection_id"`
	Type         DocumentType           `json:"type"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Version      int                    `json:"version"`
	Metadata     map[string]interface{} `json:"metadata"`
	Content      *ContentData           `json:"content,omitempty"`
	Embedding    *EmbeddingData         `json:"embedding,omitempty"`
	Relations    []Relation             `json:"relations,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
}

// DocumentType represents the type of content
type DocumentType string

const (
	TypeText     DocumentType = "text"
	TypeImage    DocumentType = "image"
	TypeAudio    DocumentType = "audio"
	TypeVideo    DocumentType = "video"
	TypeDocument DocumentType = "document"
	TypeCustom   DocumentType = "custom"
)

// ContentData holds the actual content (supports multimodal)
type ContentData struct {
	Type       DocumentType           `json:"type"`
	Text       *TextContent           `json:"text,omitempty"`
	Image      *ImageContent          `json:"image,omitempty"`
	Audio      *AudioContent          `json:"audio,omitempty"`
	Video      *VideoContent          `json:"video,omitempty"`
	Binary     *BinaryContent         `json:"binary,omitempty"`
	References []ContentReference     `json:"references,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// TextContent represents text data
type TextContent struct {
	Raw      string `json:"raw"`
	Language string `json:"language,omitempty"`
	Format   string `json:"format,omitempty"` // plain, markdown, html
}

// ImageContent represents image data
type ImageContent struct {
	Format     string `json:"format"` // jpg, png, webp
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Size       int64  `json:"size"`
	Path       string `json:"path"` // Relative path to file
	Thumbnail  string `json:"thumbnail,omitempty"`
	ColorSpace string `json:"color_space,omitempty"`
}

// AudioContent represents audio data
type AudioContent struct {
	Format     string  `json:"format"` // mp3, wav, flac
	Duration   float64 `json:"duration"`
	SampleRate int     `json:"sample_rate"`
	Bitrate    int     `json:"bitrate"`
	Size       int64   `json:"size"`
	Path       string  `json:"path"`
}

// VideoContent represents video data
type VideoContent struct {
	Format    string  `json:"format"` // mp4, webm, avi
	Duration  float64 `json:"duration"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	FrameRate float64 `json:"frame_rate"`
	Size      int64   `json:"size"`
	Path      string  `json:"path"`
	Thumbnail string  `json:"thumbnail,omitempty"`
}

// BinaryContent represents arbitrary binary data
type BinaryContent struct {
	Format      string `json:"format"`
	Size        int64  `json:"size"`
	Path        string `json:"path"`
	Checksum    string `json:"checksum"`
	Compression string `json:"compression,omitempty"`
}

// ContentReference points to external content
type ContentReference struct {
	Type string `json:"type"` // url, file, s3, etc.
	URI  string `json:"uri"`
}

// EmbeddingData represents vector embedding information
type EmbeddingData struct {
	Vector    []float64         `json:"vector,omitempty"`
	Dimension int               `json:"dimension"`
	Model     string            `json:"model"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Path      string            `json:"path,omitempty"` // Path to separate embedding file
}

// Relation represents a relationship between documents
type Relation struct {
	Type       string                 `json:"type"` // parent, child, related, reference
	DocumentID string                 `json:"document_id"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// CollectionStats contains collection statistics
type CollectionStats struct {
	DocumentCount int       `json:"document_count"`
	TotalSize     int64     `json:"total_size"`
	LastUpdated   time.Time `json:"last_updated"`
}
