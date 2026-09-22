package main

import (
	"context"
	"errors"
	"fmt"
	"io"
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

	fmt.Println("Streaming response from Cortiqa:")
	stream, err := client.Chat.CreateStream(ctx, cortiqa.ChatCompletionRequest{
		Messages: []cortiqa.ChatMessage{
			{Role: "user", Content: "Write a short haiku about coding."},
		},
	})
	if err != nil {
		log.Fatalf("Failed to open stream: %v", err)
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatalf("Error reading stream: %v", err)
		}

		if len(chunk.Choices) > 0 {
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}

	fmt.Println("\n\nStream finished!")
}
