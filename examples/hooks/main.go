// Example: Using hooks for permission control
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	claude "github.com/python3isfun/claude-agent-sdk-go"
	"github.com/python3isfun/claude-agent-sdk-go/hooks"
	"github.com/python3isfun/claude-agent-sdk-go/message"
)

func main() {
	ctx := context.Background()

	// Define a hook that guards bash commands
	bashGuard := func(ctx context.Context, input hooks.Input, hctx hooks.Context) (hooks.Result, error) {
		command, _ := input.ToolInput["command"].(string)

		// List of dangerous patterns to block
		dangerous := []string{
			"rm -rf",
			"sudo",
			"chmod 777",
			"dd if=/dev",
			"> /dev/",
			"mkfs",
		}

		for _, pattern := range dangerous {
			if strings.Contains(command, pattern) {
				fmt.Printf("[HOOK] Blocked dangerous command: %s\n", command)
				return hooks.Deny(fmt.Sprintf("Blocked: command contains '%s'", pattern)), nil
			}
		}

		// Log allowed commands
		fmt.Printf("[HOOK] Allowed command: %s\n", command)
		return hooks.Allow(), nil
	}

	// Define a hook that logs all tool uses
	logAllTools := func(ctx context.Context, input hooks.Input, hctx hooks.Context) (hooks.Result, error) {
		fmt.Printf("[LOG] Tool: %s, ID: %s\n", input.ToolName, hctx.ToolUseID)
		return hooks.Allow(), nil
	}

	// Query with hooks
	msgChan, errChan := claude.Query(ctx,
		"List the files in the current directory",
		claude.WithPreToolUseHook("Bash", bashGuard),
		claude.WithPreToolUseHook("*", logAllTools),
		claude.WithSystemPrompt("You are a helpful assistant with access to bash."),
	)

	fmt.Println("\nResponse:")
	fmt.Println("=========")

	for msg := range msgChan {
		switch m := msg.(type) {
		case *message.AssistantMessage:
			fmt.Print(m.TextContent())
		case *message.ResultMessage:
			fmt.Printf("\n[Done]\n")
		}
	}

	if err := <-errChan; err != nil {
		log.Fatalf("Query failed: %v", err)
	}
}
