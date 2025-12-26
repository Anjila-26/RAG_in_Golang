package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"ollama_go/internal"
	"ollama_go/internal/store"
)

const logo = `
 ██████╗ ██╗     ██╗      █████╗ ███╗   ███╗ █████╗
██╔═══██╗██║     ██║     ██╔══██╗████╗ ████║██╔══██╗
██║   ██║██║     ██║     ███████║██╔████╔██║███████║
██║   ██║██║     ██║     ██╔══██║██║╚██╔╝██║██╔══██║
╚██████╔╝███████╗███████╗██║  ██║██║ ╚═╝ ██║██║  ██║
 ╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝  ╚═╝

 Local LLM CLI powered by Ollama
`

func cleanInput(str string) string {
	return strings.TrimSpace(str)
}

func main() {
	fmt.Println(logo)

	// Check if we should crawl
	if len(os.Args) > 1 && os.Args[1] == "crawl" {
		fmt.Println("Starting web crawler...")
		crawling()
		fmt.Println("\nCrawling completed!")
		return
	}

	// Initialize document store and load existing documents
	docStore := store.NewDocumentStore()
	if err := docStore.LoadFromDisk(); err != nil {
		log.Printf("Warning: Could not load documents: %v", err)
	}

	docs := docStore.GetAllDocuments()
	if len(docs) == 0 {
		fmt.Println("⚠️  No documents found! Please run 'go run . crawl' first to index documents.")
		return
	}

	fmt.Printf("✅ Loaded %d documents from index\n\n", len(docs))

	// Initialize RAG service
	ragService, err := internal.NewRAGService("llama3:latest", docStore, 3)
	if err != nil {
		log.Fatal("Error initializing RAG service:", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("🤖 RAG-powered Q&A ready! Ask questions about the indexed Go documentation.")
	fmt.Println("Type 'exit' to quit.\n")

	for {
		fmt.Printf("Prompt : ")
		scanner.Scan()

		text := cleanInput(scanner.Text())
		if err := scanner.Err(); err != nil {
			log.Println("Error reading input:", err)
			continue
		}

		// Exit condition
		if strings.ToLower(text) == "exit" {
			fmt.Println("Exiting CLI. Goodbye!")
			break
		}

		if text == "" {
			continue
		}

		start := time.Now()
		ctx := context.Background()

		fmt.Println("\n🔍 Searching for relevant context...")

		// Use RAG to generate response with retrieved context
		_, err := ragService.Query(ctx, text, func(chunk string) {
			fmt.Print(chunk)
		})
		if err != nil {
			log.Println("\n❌ Error generating response:", err)
			continue
		}

		elapsed := time.Since(start)

		fmt.Printf("\nExecution time: %s\n\n", elapsed)
	}

}
