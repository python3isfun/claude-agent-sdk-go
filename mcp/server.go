// Package mcp provides MCP (Model Context Protocol) server integration.
package mcp

import (
	"context"
)

// Server represents an MCP server.
type Server interface {
	// Name returns the server name.
	Name() string

	// Version returns the server version.
	Version() string

	// Tools returns the list of available tools.
	Tools() []Tool

	// Execute runs a tool with the given input.
	Execute(ctx context.Context, toolName string, input map[string]interface{}) (ToolResult, error)

	// Start starts the server (for external servers).
	Start(ctx context.Context) error

	// Close stops the server and releases resources.
	Close() error
}

// Tool represents an MCP tool.
type Tool struct {
	// Name is the tool identifier.
	Name string `json:"name"`

	// Description explains what the tool does.
	Description string `json:"description"`

	// InputSchema is the JSON schema for tool input.
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`

	// Handler is the function that executes the tool (for in-process servers).
	Handler ToolHandler `json:"-"`
}

// ToolHandler is a function that handles tool execution.
type ToolHandler func(ctx context.Context, input map[string]interface{}) (ToolResult, error)

// ToolResult represents the result of tool execution.
type ToolResult struct {
	// Content contains the result content blocks.
	Content []ContentBlock `json:"content"`

	// IsError indicates if the result represents an error.
	IsError bool `json:"isError,omitempty"`
}

// ContentBlock represents a content block in tool results.
type ContentBlock struct {
	// Type is the content type (e.g., "text", "image").
	Type string `json:"type"`

	// Text is the text content (for type "text").
	Text string `json:"text,omitempty"`

	// Data is base64-encoded data (for type "image").
	Data string `json:"data,omitempty"`

	// MimeType is the MIME type (for binary content).
	MimeType string `json:"mimeType,omitempty"`
}

// TextResult creates a ToolResult with text content.
func TextResult(text string) ToolResult {
	return ToolResult{
		Content: []ContentBlock{{Type: "text", Text: text}},
	}
}

// ErrorResult creates a ToolResult representing an error.
func ErrorResult(err error) ToolResult {
	return ToolResult{
		Content: []ContentBlock{{Type: "text", Text: err.Error()}},
		IsError: true,
	}
}

// ServerConfig represents configuration for an MCP server.
type ServerConfig struct {
	// Type is the server type: "stdio", "sse", "http", "sdk"
	Type string `json:"type"`

	// Command is the command to run (for stdio servers).
	Command string `json:"command,omitempty"`

	// Args are command arguments (for stdio servers).
	Args []string `json:"args,omitempty"`

	// Env are environment variables (for stdio servers).
	Env map[string]string `json:"env,omitempty"`

	// URL is the server URL (for SSE/HTTP servers).
	URL string `json:"url,omitempty"`

	// Headers are HTTP headers (for SSE/HTTP servers).
	Headers map[string]string `json:"headers,omitempty"`

	// Instance is the in-process server instance (for SDK servers).
	Instance Server `json:"-"`
}

// StdioServerConfig creates a config for a stdio-based MCP server.
func StdioServerConfig(command string, args ...string) ServerConfig {
	return ServerConfig{
		Type:    "stdio",
		Command: command,
		Args:    args,
	}
}

// SSEServerConfig creates a config for an SSE-based MCP server.
func SSEServerConfig(url string) ServerConfig {
	return ServerConfig{
		Type: "sse",
		URL:  url,
	}
}

// HTTPServerConfig creates a config for an HTTP-based MCP server.
func HTTPServerConfig(url string) ServerConfig {
	return ServerConfig{
		Type: "http",
		URL:  url,
	}
}

// SDKServerConfig creates a config for an in-process SDK server.
func SDKServerConfig(server Server) ServerConfig {
	return ServerConfig{
		Type:     "sdk",
		Instance: server,
	}
}
