// Package permission provides permission management for the Claude Agent SDK.
package permission

import "context"

// Mode defines the default permission behavior.
type Mode string

const (
	// ModeDefault requires approval for sensitive operations.
	ModeDefault Mode = "default"

	// ModeAcceptEdits automatically approves file edits.
	ModeAcceptEdits Mode = "acceptEdits"

	// ModePlan requires explicit approval for all changes.
	ModePlan Mode = "plan"

	// ModeBypassPermissions skips all permission checks.
	ModeBypassPermissions Mode = "bypassPermissions"

	// ModeDontAsk suppresses permission prompts and uses defaults.
	ModeDontAsk Mode = "dontAsk"
)

// Decision represents the result of a permission evaluation.
type Decision int

const (
	// Allow permits the tool use.
	Allow Decision = iota

	// Deny blocks the tool use.
	Deny

	// Ask prompts the user for a decision.
	Ask
)

// Result contains the full result of a permission evaluation.
type Result struct {
	// Decision is the permission decision.
	Decision Decision

	// Reason explains why the decision was made.
	Reason string

	// UpdatedInput contains modified input if the hook changed it.
	UpdatedInput map[string]interface{}

	// Updates contains permission rule updates to apply.
	Updates []Update

	// Interrupt indicates whether to stop execution.
	Interrupt bool
}

// ResultAllow creates an Allow result.
func ResultAllow() Result {
	return Result{Decision: Allow}
}

// ResultAllowWithInput creates an Allow result with modified input.
func ResultAllowWithInput(input map[string]interface{}) Result {
	return Result{Decision: Allow, UpdatedInput: input}
}

// ResultDeny creates a Deny result.
func ResultDeny(reason string) Result {
	return Result{Decision: Deny, Reason: reason}
}

// ResultDenyAndInterrupt creates a Deny result that also interrupts execution.
func ResultDenyAndInterrupt(reason string) Result {
	return Result{Decision: Deny, Reason: reason, Interrupt: true}
}

// ResultAsk creates an Ask result.
func ResultAsk() Result {
	return Result{Decision: Ask}
}

// Update represents a permission rule update.
type Update struct {
	// Operation is the type of update: "add_rule", "set_mode", "add_directory"
	Operation string `json:"operation"`

	// Target is where to apply: "user", "project", "local", "session"
	Target string `json:"target"`

	// Value is the update value (depends on operation).
	Value interface{} `json:"value"`
}

// Context provides context for permission evaluation.
type Context struct {
	// ToolUseID is the unique identifier for this tool use.
	ToolUseID string

	// AgentID identifies the agent making the request (for subagents).
	AgentID string

	// Suggestions are permission suggestions from the CLI.
	Suggestions []Update
}

// Evaluator evaluates tool permissions.
type Evaluator interface {
	// Evaluate returns a permission decision for a tool use.
	Evaluate(ctx context.Context, toolName string, input map[string]interface{}, pctx Context) (Result, error)
}

// EvaluatorFunc is a function that implements Evaluator.
type EvaluatorFunc func(ctx context.Context, toolName string, input map[string]interface{}, pctx Context) (Result, error)

// Evaluate implements Evaluator.
func (f EvaluatorFunc) Evaluate(ctx context.Context, toolName string, input map[string]interface{}, pctx Context) (Result, error) {
	return f(ctx, toolName, input, pctx)
}

// AllowAll returns an evaluator that allows all tool uses.
func AllowAll() Evaluator {
	return EvaluatorFunc(func(ctx context.Context, toolName string, input map[string]interface{}, pctx Context) (Result, error) {
		return ResultAllow(), nil
	})
}

// DenyAll returns an evaluator that denies all tool uses.
func DenyAll() Evaluator {
	return EvaluatorFunc(func(ctx context.Context, toolName string, input map[string]interface{}, pctx Context) (Result, error) {
		return ResultDeny("all tool uses denied"), nil
	})
}

// AllowTools returns an evaluator that allows only the specified tools.
func AllowTools(tools ...string) Evaluator {
	allowed := make(map[string]bool)
	for _, t := range tools {
		allowed[t] = true
	}
	return EvaluatorFunc(func(ctx context.Context, toolName string, input map[string]interface{}, pctx Context) (Result, error) {
		if allowed[toolName] {
			return ResultAllow(), nil
		}
		return ResultDeny("tool not in allowed list"), nil
	})
}

// DenyTools returns an evaluator that denies the specified tools.
func DenyTools(tools ...string) Evaluator {
	denied := make(map[string]bool)
	for _, t := range tools {
		denied[t] = true
	}
	return EvaluatorFunc(func(ctx context.Context, toolName string, input map[string]interface{}, pctx Context) (Result, error) {
		if denied[toolName] {
			return ResultDeny("tool in denied list"), nil
		}
		return ResultAllow(), nil
	})
}
