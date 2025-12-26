package main

import (
	"context"
	"fmt"
	"log"
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

func main() {
	fmt.Println(logo)

	llm, err := ollama.New(ollama.WithModel("llama3:latest"))
	if err != nil {
		log.Fatal(err)
	}

	query := "What is the capital city of France?"

	start := time.Now()
	ctx := context.Background()

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response : \n\n")

	elapsed := time.Since(start)

	fmt.Print(completion)
	fmt.Printf("\n\nExectution time: %s\n\n", elapsed)
}
