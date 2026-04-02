package mcp

import (
	"context"
	"fmt"
	"sync"
)

// InProcessServer is an MCP server that runs in-process.
type InProcessServer struct {
	name    string
	version string
	tools   map[string]Tool
	mu      sync.RWMutex
}

// NewServer creates a new in-process MCP server.
func NewServer(name, version string) *InProcessServer {
	return &InProcessServer{
		name:    name,
		version: version,
		tools:   make(map[string]Tool),
	}
}

// Name returns the server name.
func (s *InProcessServer) Name() string {
	return s.name
}

// Version returns the server version.
func (s *InProcessServer) Version() string {
	return s.version
}

// RegisterTool registers a tool with the server.
func (s *InProcessServer) RegisterTool(tool Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
}

// RegisterToolFunc registers a tool using a typed function.
func (s *InProcessServer) RegisterToolFunc(name, description string, handler ToolHandler) {
	s.RegisterTool(Tool{
		Name:        name,
		Description: description,
		Handler:     handler,
	})
}

// Tools returns all registered tools.
func (s *InProcessServer) Tools() []Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Execute runs a tool with the given input.
func (s *InProcessServer) Execute(ctx context.Context, toolName string, input map[string]interface{}) (ToolResult, error) {
	s.mu.RLock()
	tool, ok := s.tools[toolName]
	s.mu.RUnlock()

	if !ok {
		return ToolResult{}, fmt.Errorf("tool not found: %s", toolName)
	}

	if tool.Handler == nil {
		return ToolResult{}, fmt.Errorf("tool has no handler: %s", toolName)
	}

	return tool.Handler(ctx, input)
}

// Start is a no-op for in-process servers.
func (s *InProcessServer) Start(ctx context.Context) error {
	return nil
}

// Close is a no-op for in-process servers.
func (s *InProcessServer) Close() error {
	return nil
}

// Builder provides a fluent interface for building an MCP server.
type Builder struct {
	server *InProcessServer
}

// NewBuilder creates a new server builder.
func NewBuilder(name, version string) *Builder {
	return &Builder{
		server: NewServer(name, version),
	}
}

// Tool adds a tool to the server.
func (b *Builder) Tool(tool Tool) *Builder {
	b.server.RegisterTool(tool)
	return b
}

// ToolFunc adds a tool using a handler function.
func (b *Builder) ToolFunc(name, description string, handler ToolHandler) *Builder {
	b.server.RegisterToolFunc(name, description, handler)
	return b
}

// Build returns the built server.
func (b *Builder) Build() *InProcessServer {
	return b.server
}

// SimpleHandler creates a ToolHandler from a simple function.
func SimpleHandler(fn func(input map[string]interface{}) (string, error)) ToolHandler {
	return func(ctx context.Context, input map[string]interface{}) (ToolResult, error) {
		result, err := fn(input)
		if err != nil {
			return ErrorResult(err), nil
		}
		return TextResult(result), nil
	}
}

// ContextHandler creates a ToolHandler that receives context.
func ContextHandler(fn func(ctx context.Context, input map[string]interface{}) (string, error)) ToolHandler {
	return func(ctx context.Context, input map[string]interface{}) (ToolResult, error) {
		result, err := fn(ctx, input)
		if err != nil {
			return ErrorResult(err), nil
		}
		return TextResult(result), nil
	}
}
