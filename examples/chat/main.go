package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cortiqa-ai/cortiqa-sdk-go"
)

func main() {
	apiKey := os.Getenv("CORTIQA_API_KEY")
	if apiKey == "" {
		apiKey = "sk-cortiqa-your-key"
	}

	client := cortiqa.NewClient(apiKey)
	ctx := context.Background()

	fmt.Println("Sending prompt to Cortiqa Falin-01...")
	resp, err := client.Chat.Create(ctx, cortiqa.ChatCompletionRequest{
		Model: "falin-01",
		Messages: []cortiqa.ChatMessage{
			{Role: "system", Content: "You are an assistant built by Cortiqa."},
			{Role: "user", Content: "Explain Goroutines in 2 short sentences."},
		},
	})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("\nResponse:")
	fmt.Println(resp.Content())
	fmt.Printf("\nTokens used: %d\n", resp.Usage.TotalTokens)
}
