// Example: Custom MCP tools using the Claude Agent SDK
package main

import (
	"context"
	"fmt"
	"log"
	"math"

	claude "github.com/dayonghuang/claude-agent-sdk-go"
	"github.com/dayonghuang/claude-agent-sdk-go/mcp"
	"github.com/dayonghuang/claude-agent-sdk-go/message"
)

func main() {
	ctx := context.Background()

	// Create an in-process MCP server with custom tools
	mathServer := mcp.NewBuilder("math-tools", "1.0.0").
		ToolFunc("add", "Add two numbers", addHandler).
		ToolFunc("multiply", "Multiply two numbers", multiplyHandler).
		ToolFunc("sqrt", "Calculate square root", sqrtHandler).
		Build()

	// Query with custom tools
	msgChan, errChan := claude.Query(ctx,
		"Calculate: (3 + 5) * 2, then find its square root",
		claude.WithMCPServer("math", mathServer),
		claude.WithSystemPrompt("Use the math tools to perform calculations step by step."),
	)

	fmt.Println("Response:")
	fmt.Println("=========")

	for msg := range msgChan {
		switch m := msg.(type) {
		case *message.AssistantMessage:
			fmt.Print(m.TextContent())

			// Show tool uses
			for _, tool := range m.ToolUses() {
				fmt.Printf("\n[Tool: %s, Input: %v]\n", tool.Name(), tool.Input())
			}

		case *message.ResultMessage:
			fmt.Printf("\n\n[Done] Tokens: %d\n", m.TotalTokens())
		}
	}

	if err := <-errChan; err != nil {
		log.Fatalf("Query failed: %v", err)
	}
}

func addHandler(ctx context.Context, input map[string]interface{}) (mcp.ToolResult, error) {
	a, _ := input["a"].(float64)
	b, _ := input["b"].(float64)
	result := a + b
	return mcp.TextResult(fmt.Sprintf("%.2f + %.2f = %.2f", a, b, result)), nil
}

func multiplyHandler(ctx context.Context, input map[string]interface{}) (mcp.ToolResult, error) {
	a, _ := input["a"].(float64)
	b, _ := input["b"].(float64)
	result := a * b
	return mcp.TextResult(fmt.Sprintf("%.2f * %.2f = %.2f", a, b, result)), nil
}

func sqrtHandler(ctx context.Context, input map[string]interface{}) (mcp.ToolResult, error) {
	n, _ := input["n"].(float64)
	if n < 0 {
		return mcp.ErrorResult(fmt.Errorf("cannot calculate square root of negative number")), nil
	}
	result := math.Sqrt(n)
	return mcp.TextResult(fmt.Sprintf("sqrt(%.2f) = %.4f", n, result)), nil
}
