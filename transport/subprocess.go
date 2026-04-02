package transport

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// SubprocessTransport manages communication with the Claude Code CLI via subprocess.
type SubprocessTransport struct {
	cliPath    string
	workingDir string
	args       []string
	env        map[string]string

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	receiveCh chan []byte
	errorCh   chan error

	mu          sync.Mutex
	started     bool
	closed      bool
	cancel      context.CancelFunc
	closeOnce   sync.Once
	readersDone chan struct{} // signals when all reader goroutines are done
}

// NewSubprocessTransport creates a new subprocess transport.
func NewSubprocessTransport(cliPath string, opts ...Option) *SubprocessTransport {
	t := &SubprocessTransport{
		cliPath:     cliPath,
		receiveCh:   make(chan []byte, 100),
		errorCh:     make(chan error, 1),
		readersDone: make(chan struct{}),
		args: []string{
			"--input-format", "stream-json",
			"--output-format", "stream-json",
		},
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Start begins the subprocess and communication channels.
func (t *SubprocessTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.started {
		return nil
	}

	// Create cancellable context
	ctx, t.cancel = context.WithCancel(ctx)

	// Build command
	t.cmd = exec.CommandContext(ctx, t.cliPath, t.args...)

	if t.workingDir != "" {
		t.cmd.Dir = t.workingDir
	}

	// Set environment
	t.cmd.Env = os.Environ()
	for k, v := range t.env {
		t.cmd.Env = append(t.cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Setup pipes
	var err error
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	t.stderr, err = t.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start process
	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start CLI: %w", err)
	}

	t.started = true

	// Start reader goroutines with coordination
	var wg sync.WaitGroup
	wg.Add(2) // readOutput and readErrors

	go func() {
		t.readOutput()
		wg.Done()
	}()

	go func() {
		t.readErrors()
		wg.Done()
	}()

	// Coordinator goroutine: waits for readers, then closes errorCh
	go func() {
		wg.Wait()
		t.closeOnce.Do(func() {
			close(t.errorCh)
		})
		close(t.readersDone)
	}()

	go t.waitForExit()

	return nil
}

// readOutput reads JSON messages from stdout.
func (t *SubprocessTransport) readOutput() {
	scanner := bufio.NewScanner(t.stdout)

	// Increase buffer for large messages (1MB)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, len(buf))

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Make a copy since scanner reuses buffer
		data := make([]byte, len(line))
		copy(data, line)

		select {
		case t.receiveCh <- data:
		default:
			// Channel full, drop message (shouldn't happen with buffered channel)
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case t.errorCh <- fmt.Errorf("stdout read error: %w", err):
		default:
		}
	}

	close(t.receiveCh)
}

// readErrors reads stderr output.
func (t *SubprocessTransport) readErrors() {
	data, err := io.ReadAll(t.stderr)
	if err != nil {
		return
	}

	if len(data) > 0 {
		select {
		case t.errorCh <- &TransportError{
			Message: "CLI stderr output",
			Stderr:  string(data),
		}:
		default:
		}
	}
}

// waitForExit waits for the process to exit.
func (t *SubprocessTransport) waitForExit() {
	err := t.cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Try to send error, but don't block if channel is full
			select {
			case t.errorCh <- &TransportError{
				Message:  "CLI exited with error",
				ExitCode: exitErr.ExitCode(),
			}:
			default:
			}
		}
	}
	// Note: errorCh is closed by the coordinator goroutine after all readers are done
}

// Send sends a message to the CLI's stdin.
func (t *SubprocessTransport) Send(ctx context.Context, msg []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport closed")
	}

	if !t.started {
		return fmt.Errorf("transport not started")
	}

	// Write message followed by newline
	if _, err := t.stdin.Write(msg); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}
	if _, err := t.stdin.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// Receive returns the channel for receiving messages.
func (t *SubprocessTransport) Receive() <-chan []byte {
	return t.receiveCh
}

// Errors returns the channel for errors.
func (t *SubprocessTransport) Errors() <-chan error {
	return t.errorCh
}

// Close terminates the subprocess.
func (t *SubprocessTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}
	t.closed = true

	// Cancel context
	if t.cancel != nil {
		t.cancel()
	}

	// Close stdin to signal EOF
	if t.stdin != nil {
		t.stdin.Close()
	}

	// Give process time to exit gracefully
	if t.cmd != nil && t.cmd.Process != nil {
		done := make(chan struct{})
		go func() {
			t.cmd.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Process exited
		case <-time.After(5 * time.Second):
			// Force kill
			t.cmd.Process.Kill()
		}
	}

	return nil
}

// TransportError represents a transport-level error.
type TransportError struct {
	Message  string
	ExitCode int
	Stderr   string
}

func (e *TransportError) Error() string {
	msg := e.Message
	if e.ExitCode != 0 {
		msg = fmt.Sprintf("%s (exit code: %d)", msg, e.ExitCode)
	}
	if e.Stderr != "" {
		stderr := e.Stderr
		if len(stderr) > 500 {
			stderr = stderr[:500] + "..."
		}
		msg = fmt.Sprintf("%s\nStderr: %s", msg, stderr)
	}
	return msg
}
