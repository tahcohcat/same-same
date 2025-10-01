package ingestion

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// HuggingFaceSource reads from HuggingFace datasets
type HuggingFaceSource struct {
	dataset    string
	split      string
	subset     string
	tempFile   string
	scanner    *bufio.Scanner
	file       *os.File
	config     *SourceConfig
	textField  string
}

// NewHuggingFaceSource creates a source for HuggingFace datasets
// dataset format: "dataset_name" or "dataset_name:subset"
func NewHuggingFaceSource(dataset string, config *SourceConfig) *HuggingFaceSource {
	parts := strings.Split(dataset, ":")
	
	name := parts[0]
	subset := ""
	if len(parts) > 1 {
		subset = parts[1]
	}
	
	return &HuggingFaceSource{
		dataset:   name,
		subset:    subset,
		split:     "train", // Default split
		config:    config,
		textField: "text",  // Default text field
	}
}

// SetSplit sets which split to use (train, test, validation)
func (s *HuggingFaceSource) SetSplit(split string) {
	s.split = split
}

// SetTextField sets which field contains the text
func (s *HuggingFaceSource) SetTextField(field string) {
	s.textField = field
}

func (s *HuggingFaceSource) Open(ctx context.Context) error {
	// Check if Python is available
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err := exec.LookPath("python"); err != nil {
			return fmt.Errorf("python not found - required for HuggingFace datasets")
		}
	}
	
	// Create a temporary Python script to download and export the dataset
	script := s.generatePythonScript()
	
	tmpScript, err := os.CreateTemp("", "hf_download_*.py")
	if err != nil {
		return fmt.Errorf("failed to create temp script: %w", err)
	}
	defer os.Remove(tmpScript.Name())
	
	if _, err := tmpScript.WriteString(script); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}
	tmpScript.Close()
	
	// Create temp file for output
	tmpOutput, err := os.CreateTemp("", "hf_data_*.jsonl")
	if err != nil {
		return fmt.Errorf("failed to create temp output file: %w", err)
	}
	s.tempFile = tmpOutput.Name()
	tmpOutput.Close()
	
	if s.config.Verbose {
		fmt.Printf("Downloading HuggingFace dataset: %s\n", s.dataset)
	}
	
	// Execute Python script
	pythonCmd := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		pythonCmd = "python"
	}
	
	cmd := exec.CommandContext(ctx, pythonCmd, tmpScript.Name(), s.tempFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		os.Remove(s.tempFile)
		return fmt.Errorf("failed to download dataset: %w", err)
	}
	
	// Open the downloaded file
	file, err := os.Open(s.tempFile)
	if err != nil {
		return fmt.Errorf("failed to open downloaded data: %w", err)
	}
	
	s.file = file
	s.scanner = bufio.NewScanner(file)
	
	// Increase buffer size for large JSON lines
	buf := make([]byte, 0, 64*1024)
	s.scanner.Buffer(buf, 1024*1024)
	
	return nil
}

func (s *HuggingFaceSource) generatePythonScript() string {
	datasetArg := fmt.Sprintf("'%s'", s.dataset)
	if s.subset != "" {
		datasetArg = fmt.Sprintf("'%s', '%s'", s.dataset, s.subset)
	}
	
	return fmt.Sprintf(`#!/usr/bin/env python3
import sys
import json
from datasets import load_dataset

output_file = sys.argv[1]

try:
    dataset = load_dataset(%s, split='%s')
    
    with open(output_file, 'w', encoding='utf-8') as f:
        for item in dataset:
            f.write(json.dumps(item) + '\n')
    
    print(f"Successfully downloaded {len(dataset)} records")
except Exception as e:
    print(f"Error: {e}", file=sys.stderr)
    sys.exit(1)
`, datasetArg, s.split)
}

func (s *HuggingFaceSource) Next() (*Record, error) {
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	
	line := s.scanner.Bytes()
	if len(line) == 0 {
		return s.Next()
	}
	
	var data map[string]interface{}
	if err := json.Unmarshal(line, &data); err != nil {
		if s.config.Verbose {
			fmt.Printf("skipping invalid JSON: %v\n", err)
		}
		return s.Next()
	}
	
	// Extract text field
	text, ok := data[s.textField].(string)
	if !ok {
		if s.config.Verbose {
			fmt.Printf("skipping record without '%s' field\n", s.textField)
		}
		return s.Next()
	}
	
	// Build metadata from other fields
	metadata := make(map[string]string)
	for key, value := range data {
		if key == s.textField {
			continue
		}
		
		switch v := value.(type) {
		case string:
			metadata[key] = v
		case float64, int, int64, bool:
			metadata[key] = fmt.Sprintf("%v", v)
		}
	}
	
	metadata["source"] = "huggingface"
	metadata["dataset"] = s.dataset
	
	if s.config.Namespace != "" {
		metadata["namespace"] = s.config.Namespace
	}
	
	return &Record{
		Text:     text,
		Metadata: metadata,
	}, nil
}

func (s *HuggingFaceSource) Close() error {
	if s.file != nil {
		s.file.Close()
	}
	if s.tempFile != "" {
		os.Remove(s.tempFile)
	}
	return nil
}

func (s *HuggingFaceSource) Name() string {
	if s.subset != "" {
		return fmt.Sprintf("hf:%s:%s", s.dataset, s.subset)
	}
	return fmt.Sprintf("hf:%s", s.dataset)
}

// HuggingFaceAPISource uses the HuggingFace API without Python (simpler but limited)
type HuggingFaceAPISource struct {
	dataset string
	apiKey  string
	config  *SourceConfig
}

// NewHuggingFaceAPISource creates a lightweight API-based source
func NewHuggingFaceAPISource(dataset, apiKey string, config *SourceConfig) *HuggingFaceAPISource {
	return &HuggingFaceAPISource{
		dataset: dataset,
		apiKey:  apiKey,
		config:  config,
	}
}

func (s *HuggingFaceAPISource) Open(ctx context.Context) error {
	// This is a simpler implementation using direct API calls
	// For now, return error suggesting to use full implementation
	return fmt.Errorf("HuggingFace API source not yet implemented - use hf: prefix with Python installed")
}

func (s *HuggingFaceAPISource) Next() (*Record, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *HuggingFaceAPISource) Close() error {
	return nil
}

func (s *HuggingFaceAPISource) Name() string {
	return fmt.Sprintf("hf-api:%s", s.dataset)
}
