package embedding

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/ollama"
)

// Service handles embedding generation
type Service struct {
	llm *ollama.LLM
}

// NewService creates a new embedding service
func NewService(modelName string) (*Service, error) {
	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Ollama LLM: %w", err)
	}
	return &Service{
		llm: llm,
	}, nil
}

// GenerateEmbedding creates an embedding for a single text
func (s *Service) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}
	embs, err := s.llm.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}
	if len(embs) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embs[0], nil
}

// GenerateBatchEmbeddings creates embeddings for multiple texts
func (s *Service) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}
	embs, err := s.llm.CreateEmbedding(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings: %w", err)
	}
	log.Printf("Generated %d embeddings (dimension: %d)\n", len(embs), len(embs[0]))
	return embs, nil
}