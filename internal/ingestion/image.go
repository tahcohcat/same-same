package ingestion

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ImageSource reads images from a directory
type ImageSource struct {
	directory string
	files     []string
	index     int
	config    *SourceConfig
	recursive bool
}

// NewImageSource creates a source for image files
func NewImageSource(directory string, config *SourceConfig) (*ImageSource, error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", directory)
	}

	return &ImageSource{
		directory: directory,
		config:    config,
		recursive: true,
		index:     0,
	}, nil
}

// SetRecursive sets whether to scan subdirectories
func (s *ImageSource) SetRecursive(recursive bool) {
	s.recursive = recursive
}

func (s *ImageSource) Open(ctx context.Context) error {
	var files []string

	// Supported image extensions
	imageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
	}

	// Walk directory
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !s.recursive && path != s.directory {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if imageExts[ext] {
			files = append(files, path)
		}

		return nil
	}

	if err := filepath.Walk(s.directory, walkFn); err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	s.files = files

	if len(s.files) == 0 {
		return fmt.Errorf("no images found in directory: %s", s.directory)
	}

	if s.config.Verbose {
		fmt.Printf("Found %d images in %s\n", len(s.files), s.directory)
	}

	return nil
}

func (s *ImageSource) Next() (*Record, error) {
	if s.index >= len(s.files) {
		return nil, io.EOF
	}

	path := s.files[s.index]
	s.index++

	// Get relative path for ID and metadata
	relPath, _ := filepath.Rel(s.directory, path)
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Create record with image path
	// The embedder will handle reading the image
	record := &Record{
		ID:   fmt.Sprintf("img_%s", strings.ReplaceAll(nameWithoutExt, " ", "_")),
		Text: path, // Store path in Text field - embedder will treat it as image path
		Metadata: map[string]string{
			"type":      "image",
			"filename":  filename,
			"path":      relPath,
			"extension": ext,
		},
	}

	if s.config.Namespace != "" {
		record.Metadata["namespace"] = s.config.Namespace
	}

	return record, nil
}

func (s *ImageSource) Close() error {
	return nil
}

func (s *ImageSource) Name() string {
	return fmt.Sprintf("image:%s", filepath.Base(s.directory))
}

// ImageListSource reads images from a list file (CSV or text)
type ImageListSource struct {
	listFile string
	baseDir  string
	scanner  *bufio.Scanner
	file     *os.File
	config   *SourceConfig
}

// NewImageListSource creates a source from an image list file
// Format: path,label (CSV) or just path (text file)
func NewImageListSource(listFile string, config *SourceConfig) (*ImageListSource, error) {
	if _, err := os.Stat(listFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("list file does not exist: %s", listFile)
	}

	baseDir := filepath.Dir(listFile)

	return &ImageListSource{
		listFile: listFile,
		baseDir:  baseDir,
		config:   config,
	}, nil
}

// SetBaseDir sets the base directory for relative paths
func (s *ImageListSource) SetBaseDir(baseDir string) {
	s.baseDir = baseDir
}

func (s *ImageListSource) Open(ctx context.Context) error {
	file, err := os.Open(s.listFile)
	if err != nil {
		return fmt.Errorf("failed to open list file: %w", err)
	}

	s.file = file
	s.scanner = bufio.NewScanner(file)

	return nil
}

func (s *ImageListSource) Next() (*Record, error) {
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	line := strings.TrimSpace(s.scanner.Text())
	if line == "" || strings.HasPrefix(line, "#") {
		return s.Next() // Skip empty lines and comments
	}

	// Parse line: "path" or "path,label"
	parts := strings.Split(line, ",")
	imagePath := strings.TrimSpace(parts[0])

	// Make path absolute if relative
	if !filepath.IsAbs(imagePath) {
		imagePath = filepath.Join(s.baseDir, imagePath)
	}

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		if s.config.Verbose {
			fmt.Printf("skipping missing file: %s\n", imagePath)
		}
		return s.Next()
	}

	filename := filepath.Base(imagePath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	metadata := map[string]string{
		"type":      "image",
		"filename":  filename,
		"path":      imagePath,
		"extension": ext,
	}

	// Add label if present
	if len(parts) > 1 {
		label := strings.TrimSpace(parts[1])
		metadata["label"] = label
	}

	if s.config.Namespace != "" {
		metadata["namespace"] = s.config.Namespace
	}

	record := &Record{
		ID:       fmt.Sprintf("img_%s", strings.ReplaceAll(nameWithoutExt, " ", "_")),
		Text:     imagePath,
		Metadata: metadata,
	}

	return record, nil
}

func (s *ImageListSource) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

func (s *ImageListSource) Name() string {
	return fmt.Sprintf("image-list:%s", filepath.Base(s.listFile))
}
