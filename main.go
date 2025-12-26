package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

var logo = `
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

	scanner := bufio.NewScanner(os.Stdin)

	// Initialize Ollama LLM
	llm, err := ollama.New(ollama.WithModel("llama3:latest"))
	if err != nil {
		log.Fatal("Error initializing Ollama LLM:", err)
	}

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

		// Generate completion
		completion, err := llms.GenerateFromSinglePrompt(
			ctx,
			llm,
			text,
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				fmt.Printf("%s", chunk)
				return nil
			}),
		)
		if err != nil {
			log.Println("Error generating response:", err)
			continue
		}

		elapsed := time.Since(start)

		fmt.Printf("\nResponse :\n\n%s\n", completion)
		fmt.Printf("\nExecution time: %s\n\n", elapsed)
	}
}
