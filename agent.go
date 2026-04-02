package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/dayonghuang/claude-agent-sdk-go/internal/cli"
	"github.com/dayonghuang/claude-agent-sdk-go/message"
	"github.com/dayonghuang/claude-agent-sdk-go/transport"
)

// Query sends a prompt to Claude and returns streaming messages.
// This is the main entry point for simple, one-shot queries.
//
// Example:
//
//	ctx := context.Background()
//	msgChan, errChan := claude.Query(ctx, "Hello!",
//	    claude.WithSystemPrompt("You are helpful"),
//	)
//
//	for msg := range msgChan {
//	    switch m := msg.(type) {
//	    case *message.AssistantMessage:
//	        fmt.Print(m.TextContent())
//	    case *message.ResultMessage:
//	        fmt.Printf("\nDone! Tokens: %d\n", m.TotalTokens())
//	    }
//	}
//
//	if err := <-errChan; err != nil {
//	    log.Fatal(err)
//	}
func Query(ctx context.Context, prompt string, opts ...Option) (<-chan message.Message, <-chan error) {
	msgChan := make(chan message.Message, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(msgChan)
		defer close(errChan)

		if err := runQuery(ctx, prompt, msgChan, opts...); err != nil {
			errChan <- err
		}
	}()

	return msgChan, errChan
}

// runQuery executes the query and streams messages to the channel.
func runQuery(ctx context.Context, prompt string, msgChan chan<- message.Message, opts ...Option) error {
	// Apply options
	cfg := DefaultOptions()
	applyOptions(cfg, opts...)

	// Apply timeout if configured
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	// Find CLI
	cliPath, err := cli.FindCLI(cfg.CLIPath)
	if err != nil {
		return NewCLINotFoundError(cfg.CLIPath, nil)
	}

	// Build transport options
	transportOpts := []transport.Option{}
	if cfg.WorkingDir != "" {
		transportOpts = append(transportOpts, transport.WithWorkingDir(cfg.WorkingDir))
	}

	// Build CLI arguments
	args := buildCLIArgs(cfg)
	transportOpts = append(transportOpts, transport.WithArgs(args...))

	// Create transport
	t := transport.NewSubprocessTransport(cliPath, transportOpts...)

	// Start transport
	if err := t.Start(ctx); err != nil {
		return NewCLIConnectionError("failed to start CLI", err)
	}
	defer t.Close()

	// Send initial prompt
	userMsg := map[string]interface{}{
		"type": "user",
		"message": map[string]interface{}{
			"role": "user",
			"content": []map[string]interface{}{
				{"type": "text", "text": prompt},
			},
		},
	}

	msgBytes, err := json.Marshal(userMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal user message: %w", err)
	}

	if err := t.Send(ctx, msgBytes); err != nil {
		return NewCLIConnectionError("failed to send prompt", err)
	}

	// Read responses
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case data, ok := <-t.Receive():
			if !ok {
				return nil // Channel closed, done
			}

			msg, err := message.ParseMessage(data)
			if err != nil {
				// Log but continue on parse errors for forward compatibility
				continue
			}

			if msg == nil {
				continue
			}

			// Send message to channel
			select {
			case msgChan <- msg:
			case <-ctx.Done():
				return ctx.Err()
			}

			// Check for result message (end of stream)
			if _, ok := msg.(*message.ResultMessage); ok {
				return nil
			}

		case err := <-t.Errors():
			if err != nil {
				return err
			}
		}
	}
}

// buildCLIArgs builds the CLI arguments from options.
func buildCLIArgs(cfg *ClaudeAgentOptions) []string {
	var args []string

	// System prompt
	if cfg.SystemPrompt != "" {
		args = append(args, "--system-prompt", cfg.SystemPrompt)
	} else if cfg.SystemPromptFile != "" {
		args = append(args, "--system-prompt-file", cfg.SystemPromptFile)
	}

	// Max turns
	if cfg.MaxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", cfg.MaxTurns))
	}

	// Permission mode
	if cfg.PermissionMode != "" {
		args = append(args, "--permission-mode", string(cfg.PermissionMode))
	}

	// Allowed tools
	if len(cfg.AllowedTools) > 0 {
		for _, tool := range cfg.AllowedTools {
			args = append(args, "--allowedTools", tool)
		}
	}

	// Disallowed tools
	if len(cfg.DisallowedTools) > 0 {
		for _, tool := range cfg.DisallowedTools {
			args = append(args, "--disallowedTools", tool)
		}
	}

	// Model
	if cfg.Model != "" {
		args = append(args, "--model", cfg.Model)
	}

	// MCP servers
	if len(cfg.MCPServers) > 0 {
		mcpConfig := make(map[string]interface{})
		for name, server := range cfg.MCPServers {
			// For in-process servers, we'll handle them separately
			// For now, just include external server configs
			mcpConfig[name] = map[string]interface{}{
				"type": "sdk",
				"name": server.Name(),
			}
		}
		if len(mcpConfig) > 0 {
			configBytes, _ := json.Marshal(map[string]interface{}{"mcpServers": mcpConfig})
			args = append(args, "--mcp-config", string(configBytes))
		}
	}

	// Resume session
	if cfg.Resume != "" {
		args = append(args, "--resume", cfg.Resume)
	}

	return args
}

// Agent represents a Claude agent instance.
type Agent struct {
	options   *ClaudeAgentOptions
	transport transport.Transport
	mu        sync.Mutex
	closed    bool
}

// NewAgent creates a new Agent with the given options.
func NewAgent(opts ...Option) (*Agent, error) {
	cfg := DefaultOptions()
	applyOptions(cfg, opts...)

	// Find CLI
	cliPath, err := cli.FindCLI(cfg.CLIPath)
	if err != nil {
		return nil, NewCLINotFoundError(cfg.CLIPath, nil)
	}
	cfg.CLIPath = cliPath

	return &Agent{
		options: cfg,
	}, nil
}

// Query sends a prompt and returns streaming messages.
func (a *Agent) Query(ctx context.Context, prompt string, opts ...Option) (<-chan message.Message, <-chan error) {
	// Merge agent options with per-query options
	mergedOpts := []Option{}

	// Start with agent's base options converted to Options
	if a.options.SystemPrompt != "" {
		mergedOpts = append(mergedOpts, WithSystemPrompt(a.options.SystemPrompt))
	}
	if a.options.WorkingDir != "" {
		mergedOpts = append(mergedOpts, WithWorkingDir(a.options.WorkingDir))
	}
	if a.options.CLIPath != "" {
		mergedOpts = append(mergedOpts, WithCLIPath(a.options.CLIPath))
	}

	// Add per-query options (will override)
	mergedOpts = append(mergedOpts, opts...)

	return Query(ctx, prompt, mergedOpts...)
}

// Close releases agent resources.
func (a *Agent) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return nil
	}
	a.closed = true

	if a.transport != nil {
		return a.transport.Close()
	}
	return nil
}
