// Example: Simple one-shot query using the Claude Agent SDK
package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/dayonghuang/claude-agent-sdk-go"
	"github.com/dayonghuang/claude-agent-sdk-go/message"
)

func main() {
	ctx := context.Background()

	// Simple query with streaming response
	msgChan, errChan := claude.Query(ctx, "What is 2 + 2? Reply briefly.",
		claude.WithSystemPrompt("You are a helpful assistant. Keep responses concise."),
		claude.WithMaxTurns(1),
	)

	fmt.Println("Response:")
	fmt.Println("=========")

	// Process streaming messages
	for msg := range msgChan {
		switch m := msg.(type) {
		case *message.AssistantMessage:
			// Print text content as it arrives
			fmt.Print(m.TextContent())

		case *message.ResultMessage:
			// Print final stats
			fmt.Printf("\n\n[Done] Tokens: %d, Cost: $%.4f\n",
				m.TotalTokens(), m.Cost)
		}
	}

	// Check for errors
	if err := <-errChan; err != nil {
		log.Fatalf("Query failed: %v", err)
	}
}
