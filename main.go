package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	ollama "github.com/prathyushnallamothu/ollamago"
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

	if len(os.Args) < 2 {
		fmt.Println("Usage: ollama-go \"your prompt here\"")
		os.Exit(1)
	}

	prompt := os.Args[1]

	client := ollama.NewClient(
		ollama.WithTimeout(5*time.Minute),
	)

	resp, err := client.Generate(
		context.Background(),
		ollama.GenerateRequest{
			Model:  "llama3:latest",
			Prompt: prompt,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n🤖 Response:\n")
	fmt.Println(resp.Response)
}
