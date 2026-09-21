# Cortiqa AI Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/cortiqa-ai/cortiqa-sdk-go.svg)](https://pkg.go.dev/github.com/cortiqa-ai/cortiqa-sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/cortiqa-ai/cortiqa-sdk-go)](https://goreportcard.com/report/github.com/cortiqa-ai/cortiqa-sdk-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Official Go client library for **Cortiqa AI** and **Falin Foundation Models**. Built with pure Go standard library, zero external dependencies, robust retry logic, and native streaming support.

---

## ⚡ Installation

```bash
go get github.com/cortiqa-ai/cortiqa-sdk-go
```

---

## 🚀 Quickstart

Set your API key:

```bash
export CORTIQA_API_KEY="sk-cortiqa-your-api-key"
```

Then create your first completion:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cortiqa-ai/cortiqa-sdk-go"
)

func main() {
	client := cortiqa.NewClient("") // Reads from CORTIQA_API_KEY env var
	ctx := context.Background()

	resp, err := client.Chat.Create(ctx, cortiqa.ChatCompletionRequest{
		Model: "falin-01",
		Messages: []cortiqa.ChatMessage{
			{Role: "system", Content: "You are a concise AI assistant by Cortiqa."},
			{Role: "user", Content: "Explain Goroutines in 2 sentences."},
		},
	})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println(resp.Content())
	fmt.Printf("Total tokens: %d\n", resp.Usage.TotalTokens)
}
```

---

## 🌊 Real-Time Token Streaming

```go
stream, err := client.Chat.CreateStream(ctx, cortiqa.ChatCompletionRequest{
	Model: "falin-01",
	Messages: []cortiqa.ChatMessage{
		{Role: "user", Content: "Write a poem about Bangalore tech culture."},
	},
})
if err != nil {
	log.Fatal(err)
}
defer stream.Close()

for {
	chunk, err := stream.Recv()
	if errors.Is(err, io.EOF) {
		break
	}
	if err != nil {
		log.Fatal(err)
	}

	if len(chunk.Choices) > 0 {
		fmt.Print(chunk.Choices[0].Delta.Content)
	}
}
```

---

## 🧠 Available Cortiqa Models

| Model ID | Provider | Description | Tier |
|---|---|---|---|
| `falin-01` | Cortiqa | Flagship foundation model for fast reasoning & conversational AI | Free / Starter |
| `falin-pro` | Cortiqa | Advanced model for complex coding, mathematical proofs, and system design | Pro |
| `falin-vision` | Cortiqa | Multimodal vision, OCR, document perception, and image reasoning | Free |
| `falin-ultra` | Cortiqa | Maximum reasoning capability for deep research and enterprise automation | Enterprise |

---

## ⚙️ Custom Configuration

```go
client := cortiqa.NewClient(
	"sk-cortiqa-...",
	cortiqa.WithBaseURL("https://api.cortiqa.co"),
	cortiqa.WithTimeout(30 * time.Second),
	cortiqa.WithMaxRetries(3),
)
```

---

## 🛡️ Error Handling

```go
resp, err := client.Chat.Create(ctx, req)
if err != nil {
	var authErr *cortiqa.AuthenticationError
	var rateErr *cortiqa.RateLimitError

	switch {
	case errors.As(err, &authErr):
		fmt.Println("Invalid API Key! Please verify your sk-cortiqa credentials.")
	case errors.As(err, &rateErr):
		fmt.Println("Rate limit exceeded. Please back off.")
	default:
		fmt.Printf("API error: %v\n", err)
	}
}
```

---

## 📄 License

MIT © [Cortiqa AI](https://cortiqa.co)
