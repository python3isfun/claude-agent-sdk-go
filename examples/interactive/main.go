// Example: Interactive multi-turn session using the Claude Agent SDK
package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/python3isfun/claude-agent-sdk-go"
	"github.com/python3isfun/claude-agent-sdk-go/message"
)

func main() {
	ctx := context.Background()

	// Create an interactive client
	client, err := claude.NewClient(
		claude.WithSystemPrompt("You are a helpful coding assistant."),
		claude.WithWorkingDir("."),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Start the session
	if err := client.Start(ctx); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}

	// First turn
	fmt.Println("=== Turn 1 ===")
	if err := client.SendMessage(ctx, "What programming language is this SDK written in?"); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	messages, err := client.ReceiveUntilResult(ctx)
	if err != nil {
		log.Fatalf("Error receiving messages: %v", err)
	}
	printMessages(messages)

	// Second turn (continues the conversation)
	fmt.Println("\n=== Turn 2 ===")
	if err := client.SendMessage(ctx, "What are its main features?"); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	messages, err = client.ReceiveUntilResult(ctx)
	if err != nil {
		log.Fatalf("Error receiving messages: %v", err)
	}
	printMessages(messages)

	fmt.Printf("\nSession ID: %s\n", client.SessionID())
}

func printMessages(messages []message.Message) {
	for _, msg := range messages {
		switch m := msg.(type) {
		case *message.AssistantMessage:
			fmt.Print(m.TextContent())
		case *message.ResultMessage:
			fmt.Printf("\n[Tokens: %d]\n", m.TotalTokens())
		}
	}
}
