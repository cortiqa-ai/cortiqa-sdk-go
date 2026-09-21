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

	fmt.Println("Fetching available Cortiqa AI Models...")
	models, err := client.Models.List(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch models: %v", err)
	}

	fmt.Printf("\nFound %d models:\n", len(models))
	fmt.Println("------------------------------------------------------------")
	for _, m := range models {
		tier := "Pro / Enterprise"
		if m.Free {
			tier = "Free Tier"
		}
		fmt.Printf("ID:          %s (%s)\n", m.ID, m.Name)
		fmt.Printf("Provider:    %s\n", m.Provider)
		fmt.Printf("Tier:        %s\n", tier)
		fmt.Printf("Description: %s\n", m.Description)
		fmt.Println("------------------------------------------------------------")
	}
}
