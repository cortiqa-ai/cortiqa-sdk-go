# Cortiqa Go SDK API Reference

Comprehensive reference for package `github.com/cortiqa-ai/cortiqa-sdk-go`.

---

## Table of Contents

- [Client Initialization](#client-initialization)
- [Chat & Messages Services](#chat--messages-services)
  - [`Create`](#create)
  - [`CreateStream`](#createstream)
- [Models Service](#models-service)
  - [`List`](#list)
- [Types & Structs](#types--structs)
- [Error Handling](#error-handling)

---

## Client Initialization

```go
import "github.com/cortiqa-ai/cortiqa-sdk-go"

// 1. From environment variable CORTIQA_API_KEY
client := cortiqa.NewClient("")

// 2. Explicit API key
client := cortiqa.NewClient("sk-cortiqa-your-api-key")

// 3. With functional options
client := cortiqa.NewClient("sk-cortiqa-your-api-key",
    cortiqa.WithBaseURL("https://api.cortiqa.co"),
    cortiqa.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
)
```

---

## Chat & Messages Services

Both `client.Chat` and `client.Messages` offer the exact same methods:

### `Create`

Sends a synchronous chat completion request.

```go
ctx := context.Background()

resp, err := client.Chat.Create(ctx, &cortiqa.ChatCompletionRequest{
    Model: "falin-01",
    Messages: []cortiqa.ChatMessage{
        {Role: cortiqa.RoleUser, Content: "Hello Cortiqa!"},
    },
    Temperature: 0.7,
    MaxTokens:   500,
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Choices[0].Message.Content)
```

### `CreateStream`

Streams response chunks over Server-Sent Events (SSE).

```go
stream, err := client.Chat.CreateStream(ctx, &cortiqa.ChatCompletionRequest{
    Model: "falin-01",
    Messages: []cortiqa.ChatMessage{
        {Role: cortiqa.RoleUser, Content: "Write a haiku."},
    },
})
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for {
    chunk, err := stream.Recv()
    if err == io.EOF {
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

## Models Service

```go
models, err := client.Models.List(ctx)
if err != nil {
    log.Fatal(err)
}

for _, m := range models.Data {
    fmt.Printf("Model: %s - %s\n", m.ID, m.Description)
}
```

---

## Error Handling

Errors returned by the SDK can be checked against `cortiqa.APIError`:

```go
resp, err := client.Chat.Create(ctx, req)
if err != nil {
    var apiErr *cortiqa.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error [%d]: %s (type: %s)\n", 
            apiErr.StatusCode, apiErr.Message, apiErr.Type)
    } else {
        fmt.Printf("Network or client error: %v\n", err)
    }
}
```
