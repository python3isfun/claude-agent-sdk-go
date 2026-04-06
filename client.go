package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/python3isfun/claude-agent-sdk-go/internal/cli"
	"github.com/python3isfun/claude-agent-sdk-go/message"
	"github.com/python3isfun/claude-agent-sdk-go/transport"
)

// Client provides an interactive session with Claude.
// Unlike Query(), Client maintains state across multiple exchanges
// and supports bidirectional communication.
type Client struct {
	options   *ClaudeAgentOptions
	transport *transport.SubprocessTransport
	msgChan   chan message.Message
	errChan   chan error

	sessionID string
	mu        sync.Mutex
	started   bool
	closed    bool
	cancel    context.CancelFunc
}

// NewClient creates a new interactive Client.
func NewClient(opts ...Option) (*Client, error) {
	cfg := DefaultOptions()
	applyOptions(cfg, opts...)

	// Find CLI
	cliPath, err := cli.FindCLI(cfg.CLIPath)
	if err != nil {
		return nil, NewCLINotFoundError(cfg.CLIPath, nil)
	}
	cfg.CLIPath = cliPath

	return &Client{
		options: cfg,
		msgChan: make(chan message.Message, 100),
		errChan: make(chan error, 1),
	}, nil
}

// Start initializes the client connection.
func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return nil
	}

	// Create cancellable context
	ctx, c.cancel = context.WithCancel(ctx)

	// Build transport options
	transportOpts := []transport.Option{}
	if c.options.WorkingDir != "" {
		transportOpts = append(transportOpts, transport.WithWorkingDir(c.options.WorkingDir))
	}

	// Build CLI arguments
	args := buildCLIArgs(c.options)
	transportOpts = append(transportOpts, transport.WithArgs(args...))

	// Create transport
	c.transport = transport.NewSubprocessTransport(c.options.CLIPath, transportOpts...)

	// Start transport
	if err := c.transport.Start(ctx); err != nil {
		return NewCLIConnectionError("failed to start CLI", err)
	}

	c.started = true

	// Start message reader
	go c.readMessages(ctx)

	return nil
}

// readMessages reads messages from the transport and forwards to the channel.
func (c *Client) readMessages(ctx context.Context) {
	defer close(c.msgChan)
	defer close(c.errChan)

	for {
		select {
		case <-ctx.Done():
			return

		case data, ok := <-c.transport.Receive():
			if !ok {
				return
			}

			msg, err := message.ParseMessage(data)
			if err != nil {
				continue
			}

			if msg == nil {
				continue
			}

			// Track session ID from assistant messages
			if assistant, ok := msg.(*message.AssistantMessage); ok {
				if assistant.SessionID != "" {
					c.mu.Lock()
					c.sessionID = assistant.SessionID
					c.mu.Unlock()
				}
			}

			select {
			case c.msgChan <- msg:
			case <-ctx.Done():
				return
			}

		case err := <-c.transport.Errors():
			if err != nil {
				select {
				case c.errChan <- err:
				default:
				}
				return
			}
		}
	}
}

// SendMessage sends a text message to Claude.
func (c *Client) SendMessage(ctx context.Context, prompt string) error {
	// Check if we need to start, with proper synchronization
	c.mu.Lock()
	needsStart := !c.started
	closed := c.closed
	c.mu.Unlock()

	if closed {
		return ErrClosed
	}

	if needsStart {
		if err := c.Start(ctx); err != nil {
			return err
		}
	}

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
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.transport.Send(ctx, msgBytes)
}

// ContentBlock represents a content block in a multimodal message.
// For text: {"type": "text", "text": "..."}
// For image: {"type": "image", "source": {"type": "base64", "media_type": "image/jpeg", "data": "..."}}
type ContentBlock = map[string]interface{}

// SendMessageWithContent sends a multimodal message to Claude with multiple content blocks.
// This supports text and images in the same message.
//
// Example usage:
//
//	content := []ContentBlock{
//	    {"type": "text", "text": "What's in this image?"},
//	    {"type": "image", "source": map[string]interface{}{
//	        "type": "base64",
//	        "media_type": "image/jpeg",
//	        "data": base64EncodedImage,
//	    }},
//	}
//	client.SendMessageWithContent(ctx, content)
func (c *Client) SendMessageWithContent(ctx context.Context, content []ContentBlock) error {
	// Check if we need to start, with proper synchronization
	c.mu.Lock()
	needsStart := !c.started
	closed := c.closed
	c.mu.Unlock()

	if closed {
		return ErrClosed
	}

	if needsStart {
		if err := c.Start(ctx); err != nil {
			return err
		}
	}

	userMsg := map[string]interface{}{
		"type": "user",
		"message": map[string]interface{}{
			"role":    "user",
			"content": content,
		},
	}

	msgBytes, err := json.Marshal(userMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.transport.Send(ctx, msgBytes)
}

// Receive returns the channel for receiving messages.
// Messages are streamed as they arrive from Claude.
func (c *Client) Receive() <-chan message.Message {
	return c.msgChan
}

// ReceiveUntilResult receives messages until a ResultMessage is received.
// Returns all messages received including the ResultMessage.
func (c *Client) ReceiveUntilResult(ctx context.Context) ([]message.Message, error) {
	var messages []message.Message

	for {
		select {
		case <-ctx.Done():
			return messages, ctx.Err()

		case msg, ok := <-c.msgChan:
			if !ok {
				return messages, nil
			}

			messages = append(messages, msg)

			if _, isResult := msg.(*message.ResultMessage); isResult {
				return messages, nil
			}

		case err := <-c.errChan:
			return messages, err
		}
	}
}

// Errors returns the channel for errors.
func (c *Client) Errors() <-chan error {
	return c.errChan
}

// SessionID returns the current session ID.
func (c *Client) SessionID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionID
}

// Interrupt sends an interrupt signal to stop the current operation.
func (c *Client) Interrupt(ctx context.Context) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClosed
	}
	if c.transport == nil {
		c.mu.Unlock()
		return fmt.Errorf("client not started")
	}
	c.mu.Unlock()

	interruptMsg := map[string]interface{}{
		"type": "interrupt",
	}

	msgBytes, err := json.Marshal(interruptMsg)
	if err != nil {
		return err
	}

	return c.transport.Send(ctx, msgBytes)
}

// Fork creates a new client that branches from the current session.
func (c *Client) Fork(ctx context.Context) (*Client, error) {
	c.mu.Lock()
	sessionID := c.sessionID
	c.mu.Unlock()

	if sessionID == "" {
		return nil, fmt.Errorf("cannot fork: no active session")
	}

	// Create new client with resume option
	opts := []Option{
		WithResume(sessionID),
		WithWorkingDir(c.options.WorkingDir),
		WithCLIPath(c.options.CLIPath),
	}
	if c.options.SystemPrompt != "" {
		opts = append(opts, WithSystemPrompt(c.options.SystemPrompt))
	}
	if c.options.AppendSystemPrompt != "" {
		opts = append(opts, WithAppendSystemPrompt(c.options.AppendSystemPrompt))
	}
	newClient, err := NewClient(opts...)
	if err != nil {
		return nil, err
	}

	if err := newClient.Start(ctx); err != nil {
		return nil, err
	}

	return newClient, nil
}

// Close terminates the client connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	if c.cancel != nil {
		c.cancel()
	}

	if c.transport != nil {
		return c.transport.Close()
	}

	return nil
}

// SetPermissionMode changes the permission mode during the session.
func (c *Client) SetPermissionMode(ctx context.Context, mode string) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClosed
	}
	if c.transport == nil {
		c.mu.Unlock()
		return fmt.Errorf("client not started")
	}
	c.mu.Unlock()

	controlMsg := map[string]interface{}{
		"type": "control",
		"control": map[string]interface{}{
			"type": "set_permission_mode",
			"mode": mode,
		},
	}

	msgBytes, err := json.Marshal(controlMsg)
	if err != nil {
		return err
	}

	return c.transport.Send(ctx, msgBytes)
}

// SetModel changes the model during the session.
func (c *Client) SetModel(ctx context.Context, model string) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClosed
	}
	if c.transport == nil {
		c.mu.Unlock()
		return fmt.Errorf("client not started")
	}
	c.mu.Unlock()

	controlMsg := map[string]interface{}{
		"type": "control",
		"control": map[string]interface{}{
			"type":  "set_model",
			"model": model,
		},
	}

	msgBytes, err := json.Marshal(controlMsg)
	if err != nil {
		return err
	}

	return c.transport.Send(ctx, msgBytes)
}
