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
		_, err := llms.GenerateFromSinglePrompt(
			ctx,
			llm,
			text,
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				// Buffer for clean word streaming
				var buffer strings.Builder
				for _, b := range chunk {
					buffer.WriteByte(b)
					if b == ' ' || b == '\n' || b == '.' || b == ',' {
						fmt.Print(buffer.String())
						buffer.Reset()
					}
				}
				// Print remaining buffer if any
				if buffer.Len() > 0 {
					fmt.Print(buffer.String())
				}
				return nil
			}),
		)
		if err != nil {
			log.Println("Error generating response:", err)
			continue
		}

		elapsed := time.Since(start)

		fmt.Printf("\nExecution time: %s\n\n", elapsed)
	}
}
