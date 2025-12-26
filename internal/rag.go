package internal

import (
	"context"
	"fmt"
	"strings"

	"ollama_go/internal/embedding"
	"ollama_go/internal/models"
	"ollama_go/internal/store"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// RAGService handles retrieval-augmented generation
type RAGService struct {
	llm        *ollama.LLM
	embService *embedding.Service
	docStore   *store.DocumentStore
	topK       int
}

// NewRAGService creates a new RAG service
func NewRAGService(modelName string, docStore *store.DocumentStore, topK int) (*RAGService, error) {
	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM: %w", err)
	}

	embService, err := embedding.NewService(modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize embedding service: %w", err)
	}

	return &RAGService{
		llm:        llm,
		embService: embService,
		docStore:   docStore,
		topK:       topK,
	}, nil
}

// Query performs RAG: retrieves relevant documents and generates response
func (r *RAGService) Query(ctx context.Context, query string, streamFunc func(string)) (string, error) {
	// Generate embedding for the query
	queryEmbedding, err := r.embService.GenerateEmbedding(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Retrieve similar documents
	similarDocs := r.docStore.SearchBySimilarity(queryEmbedding, r.topK)

	if len(similarDocs) == 0 {
		return "", fmt.Errorf("no relevant documents found")
	}

	// Build context from retrieved documents
	contextParts := make([]string, 0, len(similarDocs))
	for i, doc := range similarDocs {
		contextParts = append(contextParts, fmt.Sprintf(
			"[Document %d]\nTitle: %s\nURL: %s\nContent: %s\n",
			i+1, doc.Title, doc.URL, doc.Content,
		))
	}
	context1 := strings.Join(contextParts, "\n---\n\n")

	// Create augmented prompt
	augmentedPrompt := fmt.Sprintf(`Based on the following context, answer the question.

Context:
%s

---

Question: %s

Answer:`, context1, query)

	// Generate response with streaming
	var responseBuilder strings.Builder
	response, err := llms.GenerateFromSinglePrompt(
		ctx,
		r.llm,
		augmentedPrompt,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			text := string(chunk)
			responseBuilder.WriteString(text)
			if streamFunc != nil {
				streamFunc(text)
			}
			return nil
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	if response == "" {
		return responseBuilder.String(), nil
	}

	return response, nil
}

// GetRetrievedDocuments returns the documents that would be retrieved for a query
func (r *RAGService) GetRetrievedDocuments(ctx context.Context, query string) ([]*models.Document, error) {
	queryEmbedding, err := r.embService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	return r.docStore.SearchBySimilarity(queryEmbedding, r.topK), nil
}
