package store

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"ollama_go/internal/models"
)

// DocumentStore handles storage of documents with embeddings
type DocumentStore struct {
	mu        sync.RWMutex
	documents map[string]*models.Document
	filePath  string
}

// NewDocumentStore creates a new document store
func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		documents: make(map[string]*models.Document),
		filePath:  "data/documents.json",
	}
}

// SaveDocument saves a document with its embedding
func (ds *DocumentStore) SaveDocument(doc *models.Document) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.documents[doc.ID] = doc

	// Persist to disk
	return ds.persistToDisk()
}

// GetDocument retrieves a document by ID
func (ds *DocumentStore) GetDocument(id string) (*models.Document, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	doc, exists := ds.documents[id]
	if !exists {
		return nil, fmt.Errorf("document not found: %s", id)
	}

	return doc, nil
}

// GetAllDocuments returns all stored documents
func (ds *DocumentStore) GetAllDocuments() []*models.Document {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	docs := make([]*models.Document, 0, len(ds.documents))
	for _, doc := range ds.documents {
		docs = append(docs, doc)
	}

	return docs
}

// LoadFromDisk loads documents from disk
func (ds *DocumentStore) LoadFromDisk() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(ds.filePath); os.IsNotExist(err) {
		log.Println("No existing documents file found, starting fresh")
		return nil
	}

	data, err := os.ReadFile(ds.filePath)
	if err != nil {
		return fmt.Errorf("failed to read documents file: %w", err)
	}

	var docs []*models.Document
	if err := json.Unmarshal(data, &docs); err != nil {
		return fmt.Errorf("failed to unmarshal documents: %w", err)
	}

	ds.documents = make(map[string]*models.Document)
	for _, doc := range docs {
		ds.documents[doc.ID] = doc
	}

	log.Printf("Loaded %d documents from disk\n", len(docs))
	return nil
}

// persistToDisk saves all documents to disk
func (ds *DocumentStore) persistToDisk() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(ds.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Convert map to slice
	docs := make([]*models.Document, 0, len(ds.documents))
	for _, doc := range ds.documents {
		docs = append(docs, doc)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal documents: %w", err)
	}

	// Write to file
	if err := os.WriteFile(ds.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write documents file: %w", err)
	}

	return nil
}

// SearchBySimilarity finds documents similar to the query embedding
// This is a simple implementation - for production, use a vector database
func (ds *DocumentStore) SearchBySimilarity(queryEmbedding []float32, topK int) []*models.Document {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	type scoredDoc struct {
		doc   *models.Document
		score float32
	}

	scores := make([]scoredDoc, 0, len(ds.documents))

	for _, doc := range ds.documents {
		similarity := cosineSimilarity(queryEmbedding, doc.Embedding)
		scores = append(scores, scoredDoc{doc: doc, score: similarity})
	}

	// Sort by similarity (descending)
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// Return top K results
	if topK > len(scores) {
		topK = len(scores)
	}

	results := make([]*models.Document, topK)
	for i := 0; i < topK; i++ {
		results[i] = scores[i].doc
	}

	return results
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

// sqrt is a simple square root implementation for float32
func sqrt(x float32) float32 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
