package installer

import (
	"context"
	"os/exec"
)

// CommandRunner abstracts command execution for testing.
type CommandRunner interface {
	// Run executes a command and returns combined output.
	Run(ctx context.Context, name string, args ...string) ([]byte, error)

	// LookPath searches for a binary in PATH.
	LookPath(name string) (string, error)
}

// OSCommandRunner executes real OS commands.
type OSCommandRunner struct{}

// Run executes a command via exec.CommandContext and returns combined output.
func (r *OSCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// LookPath searches for a binary in PATH.
func (r *OSCommandRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
