package claude

import (
	"errors"
	"fmt"
)

// Common sentinel errors.
var (
	// ErrClosed indicates the client or transport has been closed.
	ErrClosed = errors.New("claude: client closed")

	// ErrTimeout indicates an operation timed out.
	ErrTimeout = errors.New("claude: operation timed out")

	// ErrPermissionDenied indicates a permission check failed.
	ErrPermissionDenied = errors.New("claude: permission denied")
)

// SDKError is the base error type for SDK errors.
type SDKError struct {
	Message string
	Cause   error
}

func (e *SDKError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *SDKError) Unwrap() error {
	return e.Cause
}

// CLINotFoundError indicates the Claude Code CLI binary was not found.
type CLINotFoundError struct {
	SDKError
	Path       string
	SearchPath []string
}

func (e *CLINotFoundError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("claude code CLI not found at %s", e.Path)
	}
	return "claude code CLI not found in PATH or common locations"
}

// NewCLINotFoundError creates a new CLINotFoundError.
func NewCLINotFoundError(path string, searchPath []string) *CLINotFoundError {
	return &CLINotFoundError{
		SDKError:   SDKError{Message: "CLI not found"},
		Path:       path,
		SearchPath: searchPath,
	}
}

// CLIConnectionError indicates a connection or communication error with the CLI.
type CLIConnectionError struct {
	SDKError
}

// NewCLIConnectionError creates a new CLIConnectionError.
func NewCLIConnectionError(message string, cause error) *CLIConnectionError {
	return &CLIConnectionError{
		SDKError: SDKError{Message: message, Cause: cause},
	}
}

// ProcessError indicates the CLI process exited with an error.
type ProcessError struct {
	SDKError
	ExitCode int
	Stderr   string
}

func (e *ProcessError) Error() string {
	msg := e.Message
	if e.ExitCode != 0 {
		msg = fmt.Sprintf("%s (exit code: %d)", msg, e.ExitCode)
	}
	if e.Stderr != "" {
		msg = fmt.Sprintf("%s\nError output: %s", msg, e.Stderr)
	}
	return msg
}

// NewProcessError creates a new ProcessError.
func NewProcessError(message string, exitCode int, stderr string) *ProcessError {
	return &ProcessError{
		SDKError: SDKError{Message: message},
		ExitCode: exitCode,
		Stderr:   stderr,
	}
}

// JSONDecodeError indicates a JSON parsing error.
type JSONDecodeError struct {
	SDKError
	Line string
}

func (e *JSONDecodeError) Error() string {
	line := e.Line
	if len(line) > 100 {
		line = line[:100] + "..."
	}
	causeMsg := "unknown error"
	if e.Cause != nil {
		causeMsg = e.Cause.Error()
	}
	return fmt.Sprintf("failed to parse JSON: %s (line: %s)", causeMsg, line)
}

// NewJSONDecodeError creates a new JSONDecodeError.
func NewJSONDecodeError(line string, cause error) *JSONDecodeError {
	return &JSONDecodeError{
		SDKError: SDKError{Message: "JSON decode error", Cause: cause},
		Line:     line,
	}
}

// MessageParseError indicates a message could not be parsed.
type MessageParseError struct {
	SDKError
	RawData []byte
}

func (e *MessageParseError) Error() string {
	data := string(e.RawData)
	if len(data) > 200 {
		data = data[:200] + "..."
	}
	causeMsg := "unknown error"
	if e.Cause != nil {
		causeMsg = e.Cause.Error()
	}
	return fmt.Sprintf("failed to parse message: %s (data: %s)", causeMsg, data)
}

// NewMessageParseError creates a new MessageParseError.
func NewMessageParseError(data []byte, cause error) *MessageParseError {
	return &MessageParseError{
		SDKError: SDKError{Message: "message parse error", Cause: cause},
		RawData:  data,
	}
}

// PermissionDeniedError indicates a tool use was denied by permissions.
type PermissionDeniedError struct {
	SDKError
	ToolName string
	Reason   string
}

func (e *PermissionDeniedError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("permission denied for tool %s: %s", e.ToolName, e.Reason)
	}
	return fmt.Sprintf("permission denied for tool %s", e.ToolName)
}

// NewPermissionDeniedError creates a new PermissionDeniedError.
func NewPermissionDeniedError(toolName, reason string) *PermissionDeniedError {
	return &PermissionDeniedError{
		SDKError: SDKError{Message: "permission denied"},
		ToolName: toolName,
		Reason:   reason,
	}
}

// HookError indicates an error in hook execution.
type HookError struct {
	SDKError
	HookEvent string
	ToolName  string
}

func (e *HookError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("hook error for %s on tool %s: %v", e.HookEvent, e.ToolName, e.Cause)
	}
	return fmt.Sprintf("hook error for %s on tool %s", e.HookEvent, e.ToolName)
}

// NewHookError creates a new HookError.
func NewHookError(event, toolName string, cause error) *HookError {
	return &HookError{
		SDKError:  SDKError{Message: "hook error", Cause: cause},
		HookEvent: event,
		ToolName:  toolName,
	}
}

// IsNotFound returns true if the error indicates the CLI was not found.
func IsNotFound(err error) bool {
	var e *CLINotFoundError
	return errors.As(err, &e)
}

// IsConnectionError returns true if the error is a connection error.
func IsConnectionError(err error) bool {
	var e *CLIConnectionError
	return errors.As(err, &e)
}

// IsProcessError returns true if the error is a process error.
func IsProcessError(err error) bool {
	var e *ProcessError
	return errors.As(err, &e)
}

// IsPermissionDenied returns true if the error is a permission denied error.
func IsPermissionDenied(err error) bool {
	var e *PermissionDeniedError
	if errors.As(err, &e) {
		return true
	}
	return errors.Is(err, ErrPermissionDenied)
}
