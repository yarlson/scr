package capture

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// TTydServer manages the ttyd subprocess lifecycle.
type TTydServer struct {
	Command string       // the shell command to execute
	Port    int          // port number for ttyd to listen on
	cmd     *exec.Cmd    // the running ttyd process
	stderr  bytes.Buffer // to capture error output
}

// NewTTydServer creates a TTydServer instance without starting it.
func NewTTydServer(command string, port int) *TTydServer {
	return &TTydServer{
		Command: command,
		Port:    port,
		cmd:     nil,
	}
}

// Validate checks that the TTydServer configuration is valid.
func (s *TTydServer) Validate() error {
	if s.Command == "" {
		return fmt.Errorf("command must not be empty")
	}
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", s.Port)
	}
	return nil
}

// Start verifies ttyd binary exists, builds and starts the ttyd subprocess,
// and polls the health endpoint to verify readiness.
func (s *TTydServer) Start(ctx context.Context) error {
	// Validate configuration
	if err := s.Validate(); err != nil {
		return fmt.Errorf("invalid TTydServer configuration: %w", err)
	}

	// Verify ttyd binary exists in PATH
	ttydPath, err := exec.LookPath("ttyd")
	if err != nil {
		return fmt.Errorf("ttyd binary not found. Install ttyd and ensure it's in PATH. Visit: https://github.com/tsl0741/ttyd")
	}

	// Build ttyd command with options matching VHS configuration
	// These client options (-t) are passed to xterm.js for proper terminal emulation
	args := []string{
		"-p", strconv.Itoa(s.Port),
		"--interface", "127.0.0.1",
		"-t", "rendererType=canvas",
		"-t", "disableResizeOverlay=true",
		"-t", "enableSixel=true",
		"-t", "customGlyphs=true",
		"--writable",
		"bash", "--norc", "--noprofile", "-c", s.Command,
	}

	s.cmd = exec.CommandContext(ctx, ttydPath, args...)

	// Set environment for proper terminal emulation
	s.cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"PS1=> ",
	)

	// Attach stderr to capture error output
	s.cmd.Stderr = &s.stderr

	// Start process
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("start ttyd process: %w", err)
	}

	// Poll http://localhost:<port>/ for up to 5 seconds to verify readiness
	url := s.URL()
	deadline := time.Now().Add(5 * time.Second)
	client := &http.Client{Timeout: 1 * time.Second}

	for {
		select {
		case <-ctx.Done():
			_ = s.Stop() // Clean up process before returning
			return fmt.Errorf("context cancelled while waiting for ttyd to be ready: %w", ctx.Err())
		default:
		}

		if time.Now().After(deadline) {
			// Timeout occurred, kill the process
			if err := s.Stop(); err != nil {
				// Log that Stop failed and attempt direct kill as fallback
				_ = s.cmd.Process.Kill()
			}
			return fmt.Errorf("ttyd health check timeout after 5 seconds. stderr: %s", s.stderr.String())
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode != http.StatusNotFound {
			_ = resp.Body.Close()
			return nil
		}
		if resp != nil {
			_ = resp.Body.Close()
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// Stop gracefully terminates the ttyd process.
// If process is already finished, returns nil.
// Sends SIGTERM and waits up to 5 seconds, then SIGKILL if needed.
func (s *TTydServer) Stop() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}

	// Send SIGTERM for graceful shutdown
	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// Process might already be dead, check if it's still running
		if err == os.ErrProcessDone {
			return nil
		}
		// Try to kill it
		return s.cmd.Process.Kill()
	}

	// Wait up to 5 seconds for graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- s.cmd.Wait()
	}()

	select {
	case <-time.After(5 * time.Second):
		// Still running after timeout, send SIGKILL
		if err := s.cmd.Process.Kill(); err != nil && err != os.ErrProcessDone {
			return fmt.Errorf("kill ttyd process: %w", err)
		}
		// Wait for kill to complete
		<-done
		return nil
	case err := <-done:
		return err
	}
}

// URL returns the localhost address with the configured port.
func (s *TTydServer) URL() string {
	return fmt.Sprintf("http://localhost:%d", s.Port)
}
