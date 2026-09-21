# Tool Calling (Function Calling) with Cortiqa Go SDK

The Cortiqa Go SDK supports structured tool calling for Falin models.

---

## Defining a Tool

```go
package main

import "github.com/cortiqa-ai/cortiqa-sdk-go"

var weatherTool = cortiqa.Tool{
    Type: "function",
    Function: cortiqa.FunctionDefinition{
        Name:        "get_weather",
        Description: "Get temperature for a city",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "city": map[string]interface{}{
                    "type":        "string",
                    "description": "City name, e.g. London, Mumbai, Tokyo",
                },
            },
            "required": []string{"city"},
        },
    },
}
```

---

## Tool Execution Loop

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/cortiqa-ai/cortiqa-sdk-go"
)

type WeatherArgs struct {
    City string `json:"city"`
}

func main() {
    client := cortiqa.NewClient("")
    ctx := context.Background()

    messages := []cortiqa.ChatMessage{
        {Role: cortiqa.RoleUser, Content: "What's the weather like in Mumbai?"},
    }

    resp, err := client.Chat.Create(ctx, &cortiqa.ChatCompletionRequest{
        Model:    "falin-01",
        Messages: messages,
        Tools:    []cortiqa.Tool{weatherTool},
    })
    if err != nil {
        log.Fatal(err)
    }

    choice := resp.Choices[0]
    if len(choice.Message.ToolCalls) > 0 {
        toolCall := choice.Message.ToolCalls[0]
        fmt.Printf("Model invoked function: %s\n", toolCall.Function.Name)

        var args WeatherArgs
        _ = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

        // Mock result
        result := fmt.Sprintf(`{"city": "%s", "temperature": "30C", "condition": "Sunny"}`, args.City)

        // Append assistant tool call and tool response
        messages = append(messages, choice.Message)
        messages = append(messages, cortiqa.ChatMessage{
            Role:       cortiqa.RoleTool,
            ToolCallID: toolCall.ID,
            Content:    result,
        })

        // Request final response
        finalResp, err := client.Chat.Create(ctx, &cortiqa.ChatCompletionRequest{
            Model:    "falin-01",
            Messages: messages,
        })
        if err != nil {
            log.Fatal(err)
        }

        fmt.Println("Final Answer:\n", finalResp.Choices[0].Message.Content)
    }
}
```
