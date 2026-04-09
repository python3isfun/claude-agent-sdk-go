package claude

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/python3isfun/claude-agent-sdk-go/permission"
)

// Validation patterns for CLI argument safety.
var (
	// modelPattern allows alphanumeric, dots, dashes, colons, slashes, underscores.
	// e.g. "claude-sonnet-4-20250514", "claude-opus-4-0-20250514"
	modelPattern = regexp.MustCompile(`^[a-zA-Z0-9._:/-]+$`)

	// resumePattern allows alphanumeric, dashes, underscores (session IDs).
	resumePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// toolNamePattern allows alphanumeric, dashes, underscores, dots, colons, slashes, wildcards.
	// e.g. "Bash", "mcp__server__tool", "Read", "computer:*"
	toolNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_\-.*:/]+$`)
)

// validPermissionModes is the set of known permission modes.
var validPermissionModes = map[permission.Mode]bool{
	permission.ModeDefault:           true,
	permission.ModeAcceptEdits:       true,
	permission.ModePlan:              true,
	permission.ModeBypassPermissions: true,
	permission.ModeDontAsk:           true,
}

// ValidationError represents an input validation failure.
type ValidationError struct {
	Field   string
	Value   string
	Reason  string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
}

// containsFlagPrefix checks if a string looks like a CLI flag.
func containsFlagPrefix(s string) bool {
	return strings.HasPrefix(s, "-")
}

// validateModel validates a model name.
func validateModel(model string) error {
	if containsFlagPrefix(model) {
		return &ValidationError{Field: "model", Value: model, Reason: "must not start with '-'"}
	}
	if !modelPattern.MatchString(model) {
		return &ValidationError{Field: "model", Value: model, Reason: "contains invalid characters; allowed: alphanumeric, '.', '-', '_', ':', '/'"}
	}
	if len(model) > 256 {
		return &ValidationError{Field: "model", Value: model, Reason: "exceeds maximum length of 256"}
	}
	return nil
}

// validatePermissionMode validates a permission mode against known values.
func validatePermissionMode(mode permission.Mode) error {
	if !validPermissionModes[mode] {
		return &ValidationError{Field: "permission_mode", Value: string(mode), Reason: "unknown permission mode"}
	}
	return nil
}

// validateResume validates a session ID for resume.
func validateResume(sessionID string) error {
	if containsFlagPrefix(sessionID) {
		return &ValidationError{Field: "resume", Value: sessionID, Reason: "must not start with '-'"}
	}
	if !resumePattern.MatchString(sessionID) {
		return &ValidationError{Field: "resume", Value: sessionID, Reason: "contains invalid characters; allowed: alphanumeric, '-', '_'"}
	}
	if len(sessionID) > 256 {
		return &ValidationError{Field: "resume", Value: sessionID, Reason: "exceeds maximum length of 256"}
	}
	return nil
}

// validateFilePath validates a file path argument to prevent path traversal.
func validateFilePath(field, path string) error {
	if containsFlagPrefix(path) {
		return &ValidationError{Field: field, Value: path, Reason: "must not start with '-'"}
	}

	// Clean and check for path traversal
	cleaned := filepath.Clean(path)

	// Reject paths with ".." components
	for _, part := range strings.Split(cleaned, string(filepath.Separator)) {
		if part == ".." {
			return &ValidationError{Field: field, Value: path, Reason: "must not contain '..' path traversal"}
		}
	}

	// Reject paths to common sensitive system files/directories
	sensitivePrefix := []string{
		"/etc/shadow", "/etc/passwd", "/etc/sudoers",
		"/proc/", "/sys/",
	}
	lower := strings.ToLower(cleaned)
	for _, prefix := range sensitivePrefix {
		if lower == prefix || strings.HasPrefix(lower, prefix) {
			return &ValidationError{Field: field, Value: path, Reason: "points to a sensitive system path"}
		}
	}

	return nil
}

// validateToolName validates a tool name.
func validateToolName(tool string) error {
	if containsFlagPrefix(tool) {
		return &ValidationError{Field: "tool", Value: tool, Reason: "must not start with '-'"}
	}
	if !toolNamePattern.MatchString(tool) {
		return &ValidationError{Field: "tool", Value: tool, Reason: "contains invalid characters"}
	}
	if len(tool) > 256 {
		return &ValidationError{Field: "tool", Value: tool, Reason: "exceeds maximum length of 256"}
	}
	return nil
}

// validateStringArg validates a generic string argument to ensure it is not a flag injection.
func validateStringArg(field, value string) error {
	if containsFlagPrefix(value) {
		return &ValidationError{Field: field, Value: value, Reason: "must not start with '-'"}
	}
	return nil
}

// validateCLIArgs validates all CLI arguments before building the command.
func validateCLIArgs(cfg *ClaudeAgentOptions) error {
	if cfg.SystemPrompt != "" {
		if err := validateStringArg("system_prompt", cfg.SystemPrompt); err != nil {
			return err
		}
	}

	if cfg.SystemPromptFile != "" {
		if err := validateFilePath("system_prompt_file", cfg.SystemPromptFile); err != nil {
			return err
		}
	}

	if cfg.AppendSystemPrompt != "" {
		if err := validateStringArg("append_system_prompt", cfg.AppendSystemPrompt); err != nil {
			return err
		}
	}

	if cfg.AppendSystemPromptFile != "" {
		if err := validateFilePath("append_system_prompt_file", cfg.AppendSystemPromptFile); err != nil {
			return err
		}
	}

	if cfg.PermissionMode != "" {
		if err := validatePermissionMode(cfg.PermissionMode); err != nil {
			return err
		}
	}

	if cfg.Model != "" {
		if err := validateModel(cfg.Model); err != nil {
			return err
		}
	}

	if cfg.Resume != "" {
		if err := validateResume(cfg.Resume); err != nil {
			return err
		}
	}

	for _, tool := range cfg.AllowedTools {
		if err := validateToolName(tool); err != nil {
			return fmt.Errorf("allowed tool %q: %w", tool, err)
		}
	}

	for _, tool := range cfg.DisallowedTools {
		if err := validateToolName(tool); err != nil {
			return fmt.Errorf("disallowed tool %q: %w", tool, err)
		}
	}

	return nil
}
