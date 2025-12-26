# RAG in Golang with Langchain and Ollama

A complete RAG (Retrieval-Augmented Generation) system for Go documentation using Ollama and vector embeddings.

## What is RAG?

RAG enhances LLM responses by:
1. **Indexing** - Crawling and embedding documents into a vector database
2. **Retrieval** - Finding relevant documents using semantic search
3. **Augmentation** - Providing context to the LLM before generating a response

## Setup

### Prerequisites
- Go 1.21+
- [Ollama](https://ollama.ai/) installed and running
- `llama3:latest` model: `ollama pull llama3:latest`

## Usage

### Step 1: Index Documents (First Time)

```bash
go run . crawl
```

This crawls pkg.go.dev, generates embeddings, and saves to `data/documents.json`.

### Step 2: Ask Questions

```bash
go run .
```

Ask questions like:
- "What is the fmt package used for?"
- "How do I read files in Go?"
- "Explain Go channels"

## How It Works

```
User Query → Query Embedding → Vector Search → Top-K Docs → Augmented Prompt → LLM → Response
```

### Features

✅ **Concurrent Crawling** - 3 parallel workers  
✅ **Vector Embeddings** - Semantic search using Ollama  
✅ **In-Memory Vector DB** - Cosine similarity search  
✅ **Streaming Responses** - Real-time LLM output  
✅ **RAG Integration** - Context-aware answers  

## Configuration

Adjust in `online_crawl.go`:
```go
maxPages := 10           // Pages to crawl
Parallelism: 3,         // Concurrent crawlers
numWorkers := 3         // Embedding workers
```

Adjust RAG in `main.go`:
```go
ragService, err := internal.NewRAGService("llama3:latest", docStore, 3) // top-K = 3
```

