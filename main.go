package main

import (
	"context"
	"fmt"
	"log"

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

	ctx := context.Background()

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response : \n")
	fmt.Print(completion)
}
