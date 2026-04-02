# Claude Agent SDK for Go

A Go SDK for the Claude Agent, enabling programmatic interaction with Claude Code.

## Installation

```bash
go get github.com/dayonghuang/claude-agent-sdk-go
```

## Prerequisites

- Go 1.21 or later
- Claude Code CLI installed and accessible in PATH

## Quick Start

### Simple Query

```go
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

    msgChan, errChan := claude.Query(ctx, "Hello Claude!",
        claude.WithSystemPrompt("You are helpful"),
    )

    for msg := range msgChan {
        if assistant, ok := msg.(*message.AssistantMessage); ok {
            fmt.Print(assistant.TextContent())
        }
    }

    if err := <-errChan; err != nil {
        log.Fatal(err)
    }
}
```

### Interactive Session

```go
client, err := claude.NewClient(
    claude.WithSystemPrompt("You are a coding assistant"),
    claude.WithWorkingDir("/path/to/project"),
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

client.Start(ctx)
client.SendMessage(ctx, "What files are here?")

for msg := range client.Receive() {
    // process messages
}
```

### Custom MCP Tools

```go
server := mcp.NewBuilder("math-tools", "1.0.0").
    ToolFunc("add", "Add two numbers", func(ctx context.Context, input map[string]interface{}) (mcp.ToolResult, error) {
        a := input["a"].(float64)
        b := input["b"].(float64)
        return mcp.TextResult(fmt.Sprintf("%.2f", a+b)), nil
    }).
    Build()

claude.Query(ctx, "What is 2+2?", claude.WithMCPServer("math", server))
```

### Hooks for Permission Control

```go
bashGuard := func(ctx context.Context, input hooks.Input, hctx hooks.Context) (hooks.Result, error) {
    cmd := input.ToolInput["command"].(string)
    if strings.Contains(cmd, "rm -rf") {
        return hooks.Deny("Dangerous command blocked"), nil
    }
    return hooks.Allow(), nil
}

claude.Query(ctx, prompt, claude.WithPreToolUseHook("Bash", bashGuard))
```

## Features

- **Streaming responses** via Go channels
- **Interactive sessions** with multi-turn conversations
- **Custom MCP tools** with in-process servers
- **Hooks system** for intercepting tool use
- **Permission control** with multiple modes
- **Context-based cancellation** and timeouts

## Package Structure

```
github.com/dayonghuang/claude-agent-sdk-go/
├── agent.go              # Query() entry point
├── client.go             # Interactive client
├── options.go            # Configuration options
├── errors.go             # Error types
├── message/              # Message types
├── transport/            # CLI communication
├── mcp/                  # MCP server integration
├── hooks/                # Hook system
└── permission/           # Permission management
```

## Options

| Option | Description |
|--------|-------------|
| `WithSystemPrompt(s)` | Set system prompt |
| `WithMaxTurns(n)` | Limit conversation turns |
| `WithAllowedTools(...)` | Allowlist tools |
| `WithPermissionMode(m)` | Set permission mode |
| `WithWorkingDir(d)` | Set working directory |
| `WithMCPServer(n, s)` | Add MCP server |
| `WithPreToolUseHook(p, h)` | Add pre-tool hook |
| `WithTimeout(d)` | Set request timeout |
| `WithModel(m)` | Set Claude model |

## Permission Modes

- `ModeDefault` - Requires approval for sensitive operations
- `ModeAcceptEdits` - Auto-approves file edits
- `ModePlan` - Requires explicit approval for all changes
- `ModeBypassPermissions` - Skips all permission checks
- `ModeDontAsk` - Suppresses prompts, uses defaults

## License

This SDK is governed by Anthropic's Commercial Terms of Service.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
