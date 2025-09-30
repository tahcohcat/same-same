package tfidf

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/tahcohcat/same-same/internal/embedders"
)

// TFIDFEmbedder implements a local TF-IDF based embedder
// This provides a simple, zero-dependency embedding solution
type TFIDFEmbedder struct {
	vocabulary  map[string]int // word -> index mapping
	idf         []float64      // inverse document frequency for each term
	mu          sync.RWMutex
	documents   []string // corpus for IDF calculation
	minDf       int      // minimum document frequency
	maxDf       float64  // maximum document frequency ratio
	maxFeatures int      // maximum vocabulary size
}

// NewTFIDFEmbedder creates a new TF-IDF embedder
func NewTFIDFEmbedder() embedders.Embedder {
	return &TFIDFEmbedder{
		vocabulary:  make(map[string]int),
		documents:   make([]string, 0),
		minDf:       1,
		maxDf:       0.95,
		maxFeatures: 5000,
	}
}

// NewTFIDFEmbedderWithConfig creates a configured TF-IDF embedder
func NewTFIDFEmbedderWithConfig(minDf int, maxDf float64, maxFeatures int) embedders.Embedder {
	return &TFIDFEmbedder{
		vocabulary:  make(map[string]int),
		documents:   make([]string, 0),
		minDf:       minDf,
		maxDf:       maxDf,
		maxFeatures: maxFeatures,
	}
}

// preprocessText cleans and tokenizes text
func (t *TFIDFEmbedder) preprocessText(text string) []string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove punctuation and special characters, keep only letters and spaces
	reg := regexp.MustCompile(`[^a-z\s]+`)
	text = reg.ReplaceAllString(text, " ")

	// Split into words and filter empty strings
	words := strings.Fields(text)

	// Remove common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "can": true, "this": true,
		"that": true, "these": true, "those": true, "i": true, "you": true,
		"he": true, "she": true, "it": true, "we": true, "they": true,
		"me": true, "him": true, "her": true, "us": true, "them": true,
		"my": true, "your": true, "his": true, "its": true, "our": true,
		"their": true, "am": true, "so": true, "as": true,
	}

	filtered := make([]string, 0, len(words))
	for _, word := range words {
		if len(word) > 2 && !stopWords[word] { // Keep words longer than 2 chars
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// buildVocabulary creates vocabulary from the document corpus
func (t *TFIDFEmbedder) buildVocabulary() {
	// Count document frequency for each term
	termDocFreq := make(map[string]int)

	for _, doc := range t.documents {
		words := t.preprocessText(doc)
		seen := make(map[string]bool)

		for _, word := range words {
			if !seen[word] {
				termDocFreq[word]++
				seen[word] = true
			}
		}
	}

	// Filter terms by document frequency
	numDocs := len(t.documents)
	validTerms := make([]string, 0)

	for term, df := range termDocFreq {
		if df >= t.minDf && float64(df)/float64(numDocs) <= t.maxDf {
			validTerms = append(validTerms, term)
		}
	}

	// Sort terms by frequency (descending) and take top features
	sort.Slice(validTerms, func(i, j int) bool {
		return termDocFreq[validTerms[i]] > termDocFreq[validTerms[j]]
	})

	if len(validTerms) > t.maxFeatures {
		validTerms = validTerms[:t.maxFeatures]
	}

	// Build vocabulary mapping
	t.vocabulary = make(map[string]int)
	for i, term := range validTerms {
		t.vocabulary[term] = i
	}

	// Calculate IDF values
	t.idf = make([]float64, len(t.vocabulary))
	for term, idx := range t.vocabulary {
		df := termDocFreq[term]
		t.idf[idx] = math.Log(float64(numDocs)/float64(df)) + 1.0
	}
}

// AddDocument adds a document to the corpus for vocabulary building
func (t *TFIDFEmbedder) AddDocument(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.documents = append(t.documents, text)

	// Rebuild vocabulary if we have enough documents
	if len(t.documents)%100 == 0 || len(t.vocabulary) == 0 {
		t.buildVocabulary()
	}
}

// AddDocuments adds multiple documents at once
func (t *TFIDFEmbedder) AddDocuments(texts []string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.documents = append(t.documents, texts...)
	t.buildVocabulary()
}

// Embed converts text to TF-IDF vector
func (t *TFIDFEmbedder) Embed(text string) ([]float64, error) {
	t.mu.Lock() // Use write lock for potential vocabulary building
	defer t.mu.Unlock()

	// Bootstrap vocabulary if empty
	if len(t.vocabulary) == 0 {
		// Build initial vocabulary with common terms and current text
		bootstrapDocs := []string{
			text, // Current text
			// Common meaningful words to bootstrap vocabulary
			"life work time people world way things make know think feel see",
			"love good great true real best better never always friend",
			"success failure happiness wisdom knowledge learning education",
			"truth justice freedom equality peace war change progress",
		}
		t.documents = append(t.documents, bootstrapDocs...)
		t.buildVocabulary()
	} else {
		// Add document to corpus for future vocabulary updates
		t.documents = append(t.documents, text)

		// Rebuild vocabulary periodically
		if len(t.documents)%50 == 0 {
			t.buildVocabulary()
		}
	}

	words := t.preprocessText(text)

	// Count term frequencies
	tf := make(map[string]float64)
	for _, word := range words {
		tf[word]++
	}

	// Normalize term frequencies
	maxTf := 0.0
	for _, freq := range tf {
		if freq > maxTf {
			maxTf = freq
		}
	}

	if maxTf > 0 {
		for word := range tf {
			tf[word] = tf[word] / maxTf
		}
	}

	// Create TF-IDF vector
	embedding := make([]float64, len(t.vocabulary))

	for word, freq := range tf {
		if idx, exists := t.vocabulary[word]; exists {
			embedding[idx] = freq * t.idf[idx]
		}
	}

	// L2 normalize the vector
	norm := 0.0
	for _, val := range embedding {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	} else {
		// If still zero, create a minimal non-zero embedding
		// This ensures we never return all zeros
		for i := range embedding {
			embedding[i] = 1.0 / math.Sqrt(float64(len(embedding)))
		}
	}

	return embedding, nil
}

// GetVocabularySize returns the current vocabulary size
func (t *TFIDFEmbedder) GetVocabularySize() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.vocabulary)
}

// GetDocumentCount returns the number of documents in corpus
func (t *TFIDFEmbedder) GetDocumentCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.documents)
}

func (t *TFIDFEmbedder) Name() string {
	return "local.tfidf"
}
