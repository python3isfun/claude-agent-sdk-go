// Package hooks provides the hook system for intercepting agent behavior.
package hooks

import (
	"context"
	"time"

	"github.com/dayonghuang/claude-agent-sdk-go/permission"
)

// Event identifies the type of hook event.
type Event string

const (
	// EventPreToolUse fires before a tool is executed.
	EventPreToolUse Event = "PreToolUse"

	// EventPostToolUse fires after successful tool execution.
	EventPostToolUse Event = "PostToolUse"

	// EventPostToolUseFailure fires after tool execution fails.
	EventPostToolUseFailure Event = "PostToolUseFailure"

	// EventUserPromptSubmit fires when user input is submitted.
	EventUserPromptSubmit Event = "UserPromptSubmit"

	// EventPermissionRequest fires when permission is needed.
	EventPermissionRequest Event = "PermissionRequest"

	// EventSubagentStart fires when a subagent starts.
	EventSubagentStart Event = "SubagentStart"

	// EventSubagentStop fires when a subagent stops.
	EventSubagentStop Event = "SubagentStop"

	// EventStop fires when execution stops.
	EventStop Event = "Stop"

	// EventPreCompact fires before message compaction.
	EventPreCompact Event = "PreCompact"

	// EventNotification fires on system notifications.
	EventNotification Event = "Notification"
)

// Context provides context for hook execution.
type Context struct {
	// ToolUseID is the unique identifier for this tool use.
	ToolUseID string

	// AgentID identifies the agent (for subagent hooks).
	AgentID string

	// SessionID is the current session ID.
	SessionID string
}

// Input contains the input data for a hook.
type Input struct {
	// Event is the hook event type.
	Event Event

	// ToolName is the name of the tool (for tool-related events).
	ToolName string

	// ToolInput is the tool input (for PreToolUse).
	ToolInput map[string]interface{}

	// ToolOutput is the tool output (for PostToolUse).
	ToolOutput map[string]interface{}

	// Error is the error message (for PostToolUseFailure).
	Error string

	// Prompt is the user prompt (for UserPromptSubmit).
	Prompt string

	// Message is the notification message (for Notification).
	Message string

	// Title is the notification title (for Notification).
	Title string

	// Manual indicates manual compaction (for PreCompact).
	Manual bool

	// Description is the agent description (for SubagentStart).
	Description string

	// Suggestions are permission suggestions (for PermissionRequest).
	Suggestions []permission.Update
}

// Result contains the result of a hook execution.
type Result struct {
	// Continue indicates whether to continue execution.
	Continue bool

	// Decision is the permission decision (for PreToolUse).
	Decision permission.Decision

	// Reason explains the decision.
	Reason string

	// SystemMessage is a message to show to the user.
	SystemMessage string

	// UpdatedInput contains modified input (for PreToolUse).
	UpdatedInput map[string]interface{}

	// AdditionalContext is extra context to inject (for UserPromptSubmit).
	AdditionalContext string

	// StopReason explains why execution was stopped.
	StopReason string
}

// Allow creates a Result that allows the operation.
func Allow() Result {
	return Result{Continue: true, Decision: permission.Allow}
}

// AllowWithInput creates a Result that allows with modified input.
func AllowWithInput(input map[string]interface{}) Result {
	return Result{Continue: true, Decision: permission.Allow, UpdatedInput: input}
}

// Deny creates a Result that denies the operation.
func Deny(reason string) Result {
	return Result{Continue: true, Decision: permission.Deny, Reason: reason}
}

// DenyAndStop creates a Result that denies and stops execution.
func DenyAndStop(reason string) Result {
	return Result{Continue: false, Decision: permission.Deny, Reason: reason, StopReason: reason}
}

// Ask creates a Result that asks the user.
func Ask() Result {
	return Result{Continue: true, Decision: permission.Ask}
}

// WithMessage adds a system message to the result.
func (r Result) WithMessage(msg string) Result {
	r.SystemMessage = msg
	return r
}

// Handler is a function that handles hook events.
type Handler func(ctx context.Context, input Input, hctx Context) (Result, error)

// Matcher matches tools to hook handlers.
type Matcher struct {
	// Pattern is the tool name pattern (exact match, glob, or regex).
	// Use "*" to match all tools, or "" to match the event itself.
	Pattern string

	// Handlers are the hook handlers to execute.
	Handlers []Handler

	// Timeout is the maximum time for handler execution.
	// Defaults to 60 seconds if not set.
	Timeout time.Duration
}

// NewMatcher creates a new Matcher with the given pattern and handlers.
func NewMatcher(pattern string, handlers ...Handler) Matcher {
	return Matcher{
		Pattern:  pattern,
		Handlers: handlers,
		Timeout:  60 * time.Second,
	}
}

// AllToolsMatcher creates a Matcher that matches all tools.
func AllToolsMatcher(handlers ...Handler) Matcher {
	return NewMatcher("*", handlers...)
}

// ToolMatcher creates a Matcher for a specific tool.
func ToolMatcher(toolName string, handlers ...Handler) Matcher {
	return NewMatcher(toolName, handlers...)
}

// WithTimeout returns a copy of the Matcher with the specified timeout.
func (m Matcher) WithTimeout(d time.Duration) Matcher {
	m.Timeout = d
	return m
}
