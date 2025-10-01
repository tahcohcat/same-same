package ingestion

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// BuiltinSource reads from built-in datasets in .examples/data
type BuiltinSource struct {
	dataset string
	file    *os.File
	scanner *bufio.Scanner
	config  *SourceConfig
}

// NewBuiltinSource creates a source for built-in datasets
// dataset can be: "demo", "quotes", "quotes-small"
func NewBuiltinSource(dataset string, config *SourceConfig) *BuiltinSource {
	return &BuiltinSource{
		dataset: dataset,
		config:  config,
	}
}

func (s *BuiltinSource) Open(ctx context.Context) error {
	// Map dataset names to files
	fileMap := map[string]string{
		"demo":         ".examples/data/quotes_small.txt",
		"quotes":       ".examples/data/quotes.txt",
		"quotes-small": ".examples/data/quotes_small.txt",
	}
	
	filePath, ok := fileMap[s.dataset]
	if !ok {
		return fmt.Errorf("unknown builtin dataset: %s (available: demo, quotes, quotes-small)", s.dataset)
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open dataset file: %w", err)
	}
	
	s.file = file
	s.scanner = bufio.NewScanner(file)
	
	return nil
}

func (s *BuiltinSource) Next() (*Record, error) {
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	
	line := strings.TrimSpace(s.scanner.Text())
	if line == "" {
		return s.Next() // Skip empty lines
	}
	
	// Parse format: "Quote text — Author"
	parts := strings.Split(line, " — ")
	if len(parts) != 2 {
		// Skip malformed lines
		if s.config.Verbose {
			fmt.Printf("skipping malformed line: %s\n", line)
		}
		return s.Next()
	}
	
	text := strings.TrimSpace(parts[0])
	author := strings.TrimSpace(parts[1])
	
	record := &Record{
		Text: text,
		Metadata: map[string]string{
			"author": author,
			"type":   "quote",
		},
	}
	
	if s.config.Namespace != "" {
		record.Metadata["namespace"] = s.config.Namespace
	}
	
	return record, nil
}

func (s *BuiltinSource) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

func (s *BuiltinSource) Name() string {
	return fmt.Sprintf("builtin:%s", s.dataset)
}
