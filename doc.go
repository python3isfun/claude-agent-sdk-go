// Package claude provides a Go SDK for the Claude Agent, enabling programmatic
// interaction with Claude Code.
//
// The SDK provides two main interfaces:
//
//   - Query(): A simple function for one-shot queries that returns streaming messages
//   - Client: An interactive client for multi-turn conversations with session management
//
// # Quick Start
//
// Simple query:
//
//	ctx := context.Background()
//	msgChan, errChan := claude.Query(ctx, "Hello Claude!",
//	    claude.WithSystemPrompt("You are helpful"),
//	)
//
//	for msg := range msgChan {
//	    if assistant, ok := msg.(*message.AssistantMessage); ok {
//	        for _, block := range assistant.Content() {
//	            if text, ok := block.(*message.TextBlock); ok {
//	                fmt.Print(text.Text())
//	            }
//	        }
//	    }
//	}
//
//	if err := <-errChan; err != nil {
//	    log.Fatal(err)
//	}
//
// Interactive session:
//
//	client, err := claude.NewClient(
//	    claude.WithSystemPrompt("You are a coding assistant"),
//	    claude.WithWorkingDir("/path/to/project"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	client.SendMessage(ctx, "What files are here?")
//	for msg := range client.Receive() {
//	    // process messages
//	}
//
// # Custom MCP Tools
//
// You can register custom tools that Claude can invoke:
//
//	server := mcp.NewServer("my-tools", "1.0.0")
//	server.RegisterTool(mcp.Tool{
//	    Name:        "add",
//	    Description: "Add two numbers",
//	    Handler: func(ctx context.Context, input map[string]interface{}) (mcp.ToolResult, error) {
//	        a := input["a"].(float64)
//	        b := input["b"].(float64)
//	        return mcp.TextResult(fmt.Sprintf("%.2f", a+b)), nil
//	    },
//	})
//
//	claude.Query(ctx, "What is 2+2?", claude.WithMCPServer("math", server))
//
// # Hooks
//
// Hooks allow you to intercept and control agent behavior:
//
//	bashGuard := func(ctx context.Context, hctx hooks.HookContext) (hooks.HookResult, error) {
//	    cmd := hctx.Input["command"].(string)
//	    if strings.Contains(cmd, "rm -rf") {
//	        return hooks.Deny("Dangerous command blocked"), nil
//	    }
//	    return hooks.Allow(), nil
//	}
//
//	claude.Query(ctx, prompt, claude.WithPreToolUseHook("Bash", bashGuard))
package claude
