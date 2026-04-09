package claude

import (
	"time"

	"github.com/python3isfun/claude-agent-sdk-go/hooks"
	"github.com/python3isfun/claude-agent-sdk-go/mcp"
	"github.com/python3isfun/claude-agent-sdk-go/permission"
)

// ClaudeAgentOptions configures the Claude agent.
type ClaudeAgentOptions struct {
	// SystemPrompt sets the system instruction for Claude (replaces default).
	SystemPrompt string

	// SystemPromptFile path to a file containing the system prompt (replaces default).
	SystemPromptFile string

	// AppendSystemPrompt appends to the default system prompt (preserves Claude Code capabilities).
	AppendSystemPrompt string

	// AppendSystemPromptFile path to a file to append to the default system prompt.
	AppendSystemPromptFile string

	// MaxTurns limits the number of conversation turns.
	MaxTurns int

	// AllowedTools is an allowlist of pre-approved tools.
	AllowedTools []string

	// DisallowedTools is a blocklist of denied tools.
	DisallowedTools []string

	// PermissionMode sets default permission behavior.
	PermissionMode permission.Mode

	// PermissionEvaluator provides custom permission evaluation.
	PermissionEvaluator permission.Evaluator

	// WorkingDir sets the working directory for file operations.
	WorkingDir string

	// CLIPath optionally specifies a custom CLI path.
	CLIPath string

	// MCPServers maps server names to MCP server instances.
	MCPServers map[string]mcp.Server

	// Hooks configures hook handlers by event type.
	Hooks map[hooks.Event][]hooks.Matcher

	// Model specifies the Claude model to use.
	Model string

	// FallbackModel specifies a fallback model if the primary is unavailable.
	FallbackModel string

	// Timeout sets the overall request timeout.
	Timeout time.Duration

	// StreamTimeout sets the timeout for stream inactivity.
	StreamTimeout time.Duration

	// Budget sets the maximum cost budget for the session.
	Budget float64

	// Resume specifies a session ID to resume.
	Resume string

	// OutputFormat specifies the expected output format (for structured outputs).
	OutputFormat interface{}

	// Sandbox enables sandboxed execution.
	Sandbox bool

	// SandboxConfig provides additional sandbox configuration.
	SandboxConfig map[string]interface{}

	// PluginDirs specifies directories containing plugins.
	PluginDirs []string

	// Agents defines subagent configurations.
	Agents []AgentConfig

	// ExtraArgs provides additional CLI arguments.
	ExtraArgs map[string]interface{}

	// Bare enables minimal mode: skips hooks, LSP, plugins, CLAUDE.md, auto-memory, etc.
	// Reduces subprocess startup latency.
	Bare bool

	// NoSessionPersistence disables session persistence to disk.
	NoSessionPersistence bool

	// JsonSchema sets a JSON Schema for structured output validation.
	// The CLI will constrain model output to match this schema.
	JsonSchema string
}

// AgentConfig defines a subagent configuration.
type AgentConfig struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tools       []string `json:"tools,omitempty"`
	Model       string   `json:"model,omitempty"`
}

// Option is a functional option for configuring queries.
type Option func(*ClaudeAgentOptions)

// WithSystemPrompt sets the system prompt.
func WithSystemPrompt(prompt string) Option {
	return func(o *ClaudeAgentOptions) {
		o.SystemPrompt = prompt
	}
}

// WithSystemPromptFile sets the system prompt from a file.
func WithSystemPromptFile(path string) Option {
	return func(o *ClaudeAgentOptions) {
		o.SystemPromptFile = path
	}
}

// WithAppendSystemPrompt appends to the default system prompt.
// This preserves Claude Code's built-in capabilities while adding custom instructions.
func WithAppendSystemPrompt(prompt string) Option {
	return func(o *ClaudeAgentOptions) {
		o.AppendSystemPrompt = prompt
	}
}

// WithAppendSystemPromptFile appends file contents to the default system prompt.
func WithAppendSystemPromptFile(path string) Option {
	return func(o *ClaudeAgentOptions) {
		o.AppendSystemPromptFile = path
	}
}

// WithMaxTurns sets the maximum number of turns.
func WithMaxTurns(n int) Option {
	return func(o *ClaudeAgentOptions) {
		o.MaxTurns = n
	}
}

// WithAllowedTools sets the list of allowed tools.
func WithAllowedTools(tools ...string) Option {
	return func(o *ClaudeAgentOptions) {
		o.AllowedTools = tools
	}
}

// WithDisallowedTools sets the list of disallowed tools.
func WithDisallowedTools(tools ...string) Option {
	return func(o *ClaudeAgentOptions) {
		o.DisallowedTools = tools
	}
}

// WithPermissionMode sets the permission mode.
func WithPermissionMode(mode permission.Mode) Option {
	return func(o *ClaudeAgentOptions) {
		o.PermissionMode = mode
	}
}

// WithPermissionEvaluator sets a custom permission evaluator.
func WithPermissionEvaluator(eval permission.Evaluator) Option {
	return func(o *ClaudeAgentOptions) {
		o.PermissionEvaluator = eval
	}
}

// WithWorkingDir sets the working directory.
func WithWorkingDir(dir string) Option {
	return func(o *ClaudeAgentOptions) {
		o.WorkingDir = dir
	}
}

// WithCLIPath sets a custom CLI path.
func WithCLIPath(path string) Option {
	return func(o *ClaudeAgentOptions) {
		o.CLIPath = path
	}
}

// WithMCPServer adds an MCP server.
func WithMCPServer(name string, server mcp.Server) Option {
	return func(o *ClaudeAgentOptions) {
		if o.MCPServers == nil {
			o.MCPServers = make(map[string]mcp.Server)
		}
		o.MCPServers[name] = server
	}
}

// WithModel sets the Claude model.
func WithModel(model string) Option {
	return func(o *ClaudeAgentOptions) {
		o.Model = model
	}
}

// WithFallbackModel sets a fallback model.
func WithFallbackModel(model string) Option {
	return func(o *ClaudeAgentOptions) {
		o.FallbackModel = model
	}
}

// WithTimeout sets the overall request timeout.
func WithTimeout(d time.Duration) Option {
	return func(o *ClaudeAgentOptions) {
		o.Timeout = d
	}
}

// WithStreamTimeout sets the stream inactivity timeout.
func WithStreamTimeout(d time.Duration) Option {
	return func(o *ClaudeAgentOptions) {
		o.StreamTimeout = d
	}
}

// WithBudget sets the maximum cost budget.
func WithBudget(budget float64) Option {
	return func(o *ClaudeAgentOptions) {
		o.Budget = budget
	}
}

// WithResume resumes an existing session.
func WithResume(sessionID string) Option {
	return func(o *ClaudeAgentOptions) {
		o.Resume = sessionID
	}
}

// WithOutputFormat sets the expected output format for structured outputs.
func WithOutputFormat(format interface{}) Option {
	return func(o *ClaudeAgentOptions) {
		o.OutputFormat = format
	}
}

// WithSandbox enables sandboxed execution.
func WithSandbox(enabled bool) Option {
	return func(o *ClaudeAgentOptions) {
		o.Sandbox = enabled
	}
}

// WithSandboxConfig sets sandbox configuration.
func WithSandboxConfig(config map[string]interface{}) Option {
	return func(o *ClaudeAgentOptions) {
		o.SandboxConfig = config
	}
}

// WithPluginDir adds a plugin directory.
func WithPluginDir(dir string) Option {
	return func(o *ClaudeAgentOptions) {
		o.PluginDirs = append(o.PluginDirs, dir)
	}
}

// WithAgent adds a subagent configuration.
func WithAgent(config AgentConfig) Option {
	return func(o *ClaudeAgentOptions) {
		o.Agents = append(o.Agents, config)
	}
}

// WithHook adds a hook for the specified event.
func WithHook(event hooks.Event, matcher hooks.Matcher) Option {
	return func(o *ClaudeAgentOptions) {
		if o.Hooks == nil {
			o.Hooks = make(map[hooks.Event][]hooks.Matcher)
		}
		o.Hooks[event] = append(o.Hooks[event], matcher)
	}
}

// WithPreToolUseHook adds a PreToolUse hook for the specified tool pattern.
func WithPreToolUseHook(pattern string, handler hooks.Handler) Option {
	return func(o *ClaudeAgentOptions) {
		if o.Hooks == nil {
			o.Hooks = make(map[hooks.Event][]hooks.Matcher)
		}
		o.Hooks[hooks.EventPreToolUse] = append(o.Hooks[hooks.EventPreToolUse], hooks.Matcher{
			Pattern:  pattern,
			Handlers: []hooks.Handler{handler},
		})
	}
}

// WithPostToolUseHook adds a PostToolUse hook for the specified tool pattern.
func WithPostToolUseHook(pattern string, handler hooks.Handler) Option {
	return func(o *ClaudeAgentOptions) {
		if o.Hooks == nil {
			o.Hooks = make(map[hooks.Event][]hooks.Matcher)
		}
		o.Hooks[hooks.EventPostToolUse] = append(o.Hooks[hooks.EventPostToolUse], hooks.Matcher{
			Pattern:  pattern,
			Handlers: []hooks.Handler{handler},
		})
	}
}

// WithBare enables bare/minimal mode (skips hooks, LSP, plugins, etc.).
func WithBare(enabled bool) Option {
	return func(o *ClaudeAgentOptions) {
		o.Bare = enabled
	}
}

// WithNoSessionPersistence disables session persistence to disk.
func WithNoSessionPersistence(enabled bool) Option {
	return func(o *ClaudeAgentOptions) {
		o.NoSessionPersistence = enabled
	}
}

// WithJsonSchema sets a JSON Schema for structured output validation.
// The CLI constrains model output to match this schema, improving consistency
// and reducing output verbosity.
func WithJsonSchema(schema string) Option {
	return func(o *ClaudeAgentOptions) {
		o.JsonSchema = schema
	}
}

// WithExtraArg sets an extra CLI argument.
func WithExtraArg(key string, value interface{}) Option {
	return func(o *ClaudeAgentOptions) {
		if o.ExtraArgs == nil {
			o.ExtraArgs = make(map[string]interface{})
		}
		o.ExtraArgs[key] = value
	}
}

// applyOptions applies the given options to a base configuration.
func applyOptions(base *ClaudeAgentOptions, opts ...Option) *ClaudeAgentOptions {
	for _, opt := range opts {
		opt(base)
	}
	return base
}

// DefaultOptions returns the default options.
func DefaultOptions() *ClaudeAgentOptions {
	return &ClaudeAgentOptions{
		PermissionMode: permission.ModeDefault,
		Timeout:        10 * time.Minute,
		StreamTimeout:  60 * time.Second,
	}
}
