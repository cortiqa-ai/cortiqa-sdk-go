# Helpers & Advanced Usage in Cortiqa Go SDK

This guide covers streaming patterns, HTTP timeouts, context cancellation, and concurrent requests.

---

## 1. Context & Timeouts

Always pass a `context.Context` to control cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()

resp, err := client.Chat.Create(ctx, req)
```

---

## 2. Channel-based Streaming Helper

You can convert the stream to a Go channel of strings:

```go
func StreamToChan(ctx context.Context, stream *cortiqa.Stream) <-chan string {
    out := make(chan string)
    go func() {
        defer close(out)
        for {
            chunk, err := stream.Recv()
            if err != nil {
                return
            }
            if len(chunk.Choices) > 0 {
                text := chunk.Choices[0].Delta.Content
                if text != "" {
                    select {
                    case <-ctx.Done():
                        return
                    case out <- text:
                    }
                }
            }
        }
    }()
    return out
}
```

---

## 3. Custom HTTP Transport / Proxy

```go
proxyURL, _ := url.Parse("http://corp-proxy:8080")
customTransport := &http.Transport{
    Proxy: http.ProxyURL(proxyURL),
    MaxIdleConns: 50,
}

client := cortiqa.NewClient("sk-cortiqa-...",
    cortiqa.WithHTTPClient(&http.Client{
        Transport: customTransport,
        Timeout: 60 * time.Second,
    }),
)
```
