package ingestion

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileSource reads from CSV or JSONL files
type FileSource struct {
	path     string
	fileType string
	file     *os.File
	
	// CSV specific
	csvReader *csv.Reader
	headers   []string
	textCol   string
	
	// JSONL specific
	scanner *bufio.Scanner
	
	config *SourceConfig
}

// NewFileSource creates a source for CSV or JSONL files
func NewFileSource(path string, config *SourceConfig) (*FileSource, error) {
	ext := strings.ToLower(filepath.Ext(path))
	
	var fileType string
	switch ext {
	case ".csv":
		fileType = "csv"
	case ".jsonl", ".ndjson":
		fileType = "jsonl"
	case ".json":
		// Could be JSONL or regular JSON array
		fileType = "jsonl"
	default:
		return nil, fmt.Errorf("unsupported file type: %s (supported: .csv, .jsonl, .json)", ext)
	}
	
	return &FileSource{
		path:     path,
		fileType: fileType,
		config:   config,
		textCol:  "text", // Default text column name
	}, nil
}

// SetTextColumn sets which column contains the text (for CSV)
func (s *FileSource) SetTextColumn(col string) {
	s.textCol = col
}

func (s *FileSource) Open(ctx context.Context) error {
	file, err := os.Open(s.path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	
	s.file = file
	
	switch s.fileType {
	case "csv":
		s.csvReader = csv.NewReader(file)
		
		// Read headers
		headers, err := s.csvReader.Read()
		if err != nil {
			return fmt.Errorf("failed to read CSV headers: %w", err)
		}
		s.headers = headers
		
	case "jsonl":
		s.scanner = bufio.NewScanner(file)
		// Increase buffer size for large JSON lines
		buf := make([]byte, 0, 64*1024)
		s.scanner.Buffer(buf, 1024*1024)
	}
	
	return nil
}

func (s *FileSource) Next() (*Record, error) {
	switch s.fileType {
	case "csv":
		return s.nextCSV()
	case "jsonl":
		return s.nextJSONL()
	default:
		return nil, fmt.Errorf("unknown file type: %s", s.fileType)
	}
}

func (s *FileSource) nextCSV() (*Record, error) {
	row, err := s.csvReader.Read()
	if err != nil {
		return nil, err
	}
	
	// Find text column index
	textIdx := -1
	for i, header := range s.headers {
		if header == s.textCol {
			textIdx = i
			break
		}
	}
	
	if textIdx == -1 {
		return nil, fmt.Errorf("text column '%s' not found in CSV headers: %v", s.textCol, s.headers)
	}
	
	if textIdx >= len(row) {
		return nil, fmt.Errorf("text column index %d out of range for row with %d columns", textIdx, len(row))
	}
	
	text := row[textIdx]
	
	// Build metadata from other columns
	metadata := make(map[string]string)
	for i, value := range row {
		if i != textIdx && i < len(s.headers) {
			metadata[s.headers[i]] = value
		}
	}
	
	if s.config.Namespace != "" {
		metadata["namespace"] = s.config.Namespace
	}
	
	return &Record{
		Text:     text,
		Metadata: metadata,
	}, nil
}

func (s *FileSource) nextJSONL() (*Record, error) {
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	
	line := s.scanner.Bytes()
	if len(line) == 0 {
		return s.Next() // Skip empty lines
	}
	
	var data map[string]interface{}
	if err := json.Unmarshal(line, &data); err != nil {
		if s.config.Verbose {
			fmt.Printf("skipping invalid JSON line: %v\n", err)
		}
		return s.Next()
	}
	
	// Extract text field
	text, ok := data["text"].(string)
	if !ok {
		// Try alternative field names
		for _, field := range []string{"content", "body", "message", "quote"} {
			if t, ok := data[field].(string); ok {
				text = t
				break
			}
		}
	}
	
	if text == "" {
		if s.config.Verbose {
			fmt.Printf("skipping record without text field\n")
		}
		return s.Next()
	}
	
	// Build metadata from other fields
	metadata := make(map[string]string)
	for key, value := range data {
		if key == "text" || key == "content" || key == "body" || key == "message" {
			continue
		}
		
		// Convert value to string
		switch v := value.(type) {
		case string:
			metadata[key] = v
		case float64, int, int64, bool:
			metadata[key] = fmt.Sprintf("%v", v)
		}
	}
	
	if s.config.Namespace != "" {
		metadata["namespace"] = s.config.Namespace
	}
	
	return &Record{
		Text:     text,
		Metadata: metadata,
	}, nil
}

func (s *FileSource) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

func (s *FileSource) Name() string {
	return fmt.Sprintf("file:%s", filepath.Base(s.path))
}
