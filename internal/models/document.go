package models

import "time"

// Document represents a crawled web page with its embedding

type PageContent struct {
	// PageContent represents the extracted content from a web page
	LinkCount   int      `json:"link_count"`
	MainContent []string `json:"main_content"`
	Description string   `json:"description"`
	Title       string   `json:"title"`
	URL         string   `json:"url"`
}

type Document struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Embedding   []float32 `json:"embedding"`
	Content     string    `json:"content"`
	Description string    `json:"description"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
}