package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"ollama_go/internal/embedding"
	"ollama_go/internal/models"
	"ollama_go/internal/store"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/google/uuid"
)
func crawling() {
	// Initialize embedding service
	embService, err := embedding.NewService("llama3:latest")
	if err != nil {
		log.Fatal("Failed to initialize embedding service:", err)
	}

	// Initialize store
	docStore := store.NewDocumentStore()

	ctx := context.Background()
	
	// Thread-safe storage for documents
	var documentsMux sync.Mutex
	documents := make([]*models.PageContent, 0)

	// Thread-safe page counter
	var pageCountMux sync.Mutex
	pageCount := 0
	maxPages := 10 // Limit to 10 pages

	// Create collector with async enabled for concurrent crawling
	c := colly.NewCollector(
		colly.AllowedDomains("pkg.go.dev"),
		colly.MaxDepth(1),
		colly.Async(true), // Enable asynchronous mode for parallel crawling
	)
	maxPages = 10 // Limit to 10 pages

	// Limit parallelism - control how many requests run simultaneously
	c.Limit(&colly.LimitRule{
		DomainGlob:  "pkg.go.dev",
		Parallelism: 3,                    // 3 concurrent requests
		RandomDelay: 500 * time.Millisecond, // Delay between requests
	})

	// Extract and display page content
	c.OnHTML("html", func(e *colly.HTMLElement) {
		// Thread-safe page count check and increment
		pageCountMux.Lock()
		if pageCount >= maxPages {
			pageCountMux.Unlock()
			return
		}
		pageCount++
		currentPage := pageCount
		pageCountMux.Unlock()

		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Printf("PAGE #%d\n", currentPage)
		fmt.Println(strings.Repeat("=", 80))

		pageContent := &models.PageContent{
			URL:         e.Request.URL.String(),
			MainContent: make([]string, 0),
		}

		// Title
		title := e.ChildText("title")
		pageContent.Title = title
		fmt.Printf("📄 Title: %s\n", title)
		fmt.Printf("🔗 URL: %s\n", e.Request.URL.String())

		// Meta description
		e.DOM.Find("meta[name='description']").Each(func(_ int, s *goquery.Selection) {
			if desc, exists := s.Attr("content"); exists {
				pageContent.Description = desc
				fmt.Printf("📝 Description: %s\n", desc)
			}
		})

		// Extract MAIN content only (skip navigation and footer)
		fmt.Println("\n📖 Main Content:")
		fmt.Println(strings.Repeat("-", 80))

		var contentPrinted bool

		// Try to get main content area
		e.ForEach("main, article, .Documentation-content, .SearchResults", func(_ int, mainEl *colly.HTMLElement) {
			var textCount int
			mainEl.ForEach("h1, h2, h3, h4, p, pre, code, li", func(_ int, el *colly.HTMLElement) {
				text := strings.TrimSpace(el.Text)
				// Skip generic footer text
				if text != "" && textCount < 20 &&
					!strings.Contains(text, "Common problems companies solve") &&
					!strings.Contains(text, "Learn and network with Go") &&
					!strings.Contains(text, "Meet other local Go developers") &&
					len(text) > 20 { // Skip very short text
					fmt.Printf("\n• %s\n", text)
					pageContent.MainContent = append(pageContent.MainContent, text)
					textCount++
					contentPrinted = true
				}
			})
		})

		if !contentPrinted {
			fmt.Println("(No main content extracted)")
		}

		// Count links
		linkCount := 0
		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			linkCount++
		})
		pageContent.LinkCount = linkCount
		fmt.Printf("\n🔗 Links found: %d\n", linkCount)

		// Store the page content for later embedding generation (thread-safe)
		documentsMux.Lock()
		documents = append(documents, pageContent)
		documentsMux.Unlock()
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// Check page count in thread-safe manner
		pageCountMux.Lock()
		shouldSkip := pageCount >= maxPages
		pageCountMux.Unlock()
		
		if shouldSkip {
			return
		}

		link := e.Attr("href")
		absURL := e.Request.AbsoluteURL(link)

		// Skip version tabs and external links
		c.Visit(absURL)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("\n🔍 Crawling: %s\n", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("❌ Error: %v\n", err)
	})

	// Start crawling
	c.Visit("https://pkg.go.dev/std")
	
	// Wait for all async requests to complete
	c.Wait()
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("❌ Error: %v\n", err)
	})

	// Generate embeddings for all collected documents (parallelized)
	if len(documents) > 0 {
		fmt.Println("\n🔄 Generating embeddings for crawled content...")
		
		// Use worker pool for parallel embedding generation
		numWorkers := 3 // Number of parallel embedding workers
		docChan := make(chan struct {
			index   int
			content *models.PageContent
		}, len(documents))
		
		var wg sync.WaitGroup
		
		// Start workers
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				for job := range docChan {
					i := job.index
					pageContent := job.content
					
					// Combine title, description, and content for embedding
					combinedText := fmt.Sprintf("%s. %s. %s",
						pageContent.Title,
						pageContent.Description,
						strings.Join(pageContent.MainContent, " "))

					// Truncate if too long (optional - adjust as needed)
					if len(combinedText) > 2000 {
						combinedText = combinedText[:2000]
					}

					fmt.Printf("\n📊 [Worker %d] Generating embedding for page %d/%d: %s\n", 
						workerID, i+1, len(documents), pageContent.Title)

					// Generate embedding
					embedding, err := embService.GenerateEmbedding(ctx, combinedText)
					if err != nil {
						log.Printf("⚠️  Error generating embedding for %s: %v\n", pageContent.URL, err)
						continue
					}

					// Create document with embedding
					doc := &models.Document{
						ID:          uuid.New().String(),
						URL:         pageContent.URL,
						Title:       pageContent.Title,
						Description: pageContent.Description,
						Content:     strings.Join(pageContent.MainContent, "\n"),
						Embedding:   embedding,
						CreatedAt:   time.Now(),
					}

					// Save to store
					if err := docStore.SaveDocument(doc); err != nil {
						log.Printf("⚠️  Error saving document: %v\n", err)
						continue
					}

					fmt.Printf("✅ [Worker %d] Embedding generated (dim: %d) and document saved\n", 
						workerID, len(embedding))
				}
			}(w)
		}
		
		// Send jobs to workers
		for i, pageContent := range documents {
			docChan <- struct {
				index   int
				content *models.PageContent
			}{i, pageContent}
		}
		close(docChan)
		
		// Wait for all workers to complete
		wg.Wait()

		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Printf("✅ Generated and saved %d embeddings\n", len(documents))
		fmt.Println(strings.Repeat("=", 80))
	}
}
