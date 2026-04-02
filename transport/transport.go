// Package transport provides communication with the Claude Code CLI.
package transport

import (
	"context"
)

// Transport handles communication with the Claude Code CLI.
type Transport interface {
	// Start begins the transport connection.
	Start(ctx context.Context) error

	// Send sends a message to the CLI.
	Send(ctx context.Context, msg []byte) error

	// Receive returns a channel for receiving messages.
	Receive() <-chan []byte

	// Errors returns a channel for transport errors.
	Errors() <-chan error

	// Close terminates the transport.
	Close() error
}

// Option configures a transport.
type Option func(interface{})

// WithWorkingDir sets the working directory.
func WithWorkingDir(dir string) Option {
	return func(t interface{}) {
		if st, ok := t.(*SubprocessTransport); ok {
			st.workingDir = dir
		}
	}
}

// WithEnv sets environment variables.
func WithEnv(env map[string]string) Option {
	return func(t interface{}) {
		if st, ok := t.(*SubprocessTransport); ok {
			st.env = env
		}
	}
}

// WithArgs adds CLI arguments.
func WithArgs(args ...string) Option {
	return func(t interface{}) {
		if st, ok := t.(*SubprocessTransport); ok {
			st.args = append(st.args, args...)
		}
	}
}
