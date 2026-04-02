// Package cli provides utilities for finding and managing the Claude Code CLI.
package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// DefaultCLIName is the default CLI binary name.
const DefaultCLIName = "claude"

// FindCLI searches for the Claude Code CLI binary.
// It searches in the following order:
// 1. Custom path if provided
// 2. PATH environment variable
// 3. Common installation locations
func FindCLI(customPath string) (string, error) {
	// 1. Custom path
	if customPath != "" {
		if _, err := os.Stat(customPath); err == nil {
			return customPath, nil
		}
		return "", &NotFoundError{Path: customPath}
	}

	// 2. PATH lookup
	if path, err := exec.LookPath(DefaultCLIName); err == nil {
		return path, nil
	}

	// 3. Common locations
	commonPaths := getCommonPaths()
	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", &NotFoundError{SearchPaths: commonPaths}
}

// getCommonPaths returns common CLI installation paths for the current OS.
func getCommonPaths() []string {
	home, _ := os.UserHomeDir()

	var paths []string

	switch runtime.GOOS {
	case "darwin":
		paths = []string{
			"/usr/local/bin/claude",
			"/opt/homebrew/bin/claude",
			filepath.Join(home, ".local/bin/claude"),
			filepath.Join(home, ".claude/bin/claude"),
			"/Applications/Claude.app/Contents/Resources/claude",
		}
	case "linux":
		paths = []string{
			"/usr/local/bin/claude",
			"/usr/bin/claude",
			filepath.Join(home, ".local/bin/claude"),
			filepath.Join(home, ".claude/bin/claude"),
		}
	case "windows":
		appData := os.Getenv("LOCALAPPDATA")
		paths = []string{
			filepath.Join(appData, "Programs", "Claude", "claude.exe"),
			filepath.Join(home, ".claude", "bin", "claude.exe"),
			"C:\\Program Files\\Claude\\claude.exe",
		}
	}

	// Also check npm global installations
	npmPaths := []string{
		filepath.Join(home, ".npm-global/bin/claude"),
		"/usr/local/lib/node_modules/@anthropic-ai/claude-code/bin/claude",
	}

	return append(paths, npmPaths...)
}

// NotFoundError indicates the CLI was not found.
type NotFoundError struct {
	Path        string
	SearchPaths []string
}

func (e *NotFoundError) Error() string {
	if e.Path != "" {
		return "claude code CLI not found at: " + e.Path
	}
	return "claude code CLI not found in PATH or common locations"
}

// CheckVersion verifies the CLI version meets minimum requirements.
func CheckVersion(cliPath string, minVersion string) error {
	// For now, just check if the CLI exists and is executable
	_, err := exec.LookPath(cliPath)
	return err
}
