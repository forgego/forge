package helpers

import (
	"context"
	"os"
	"os/exec"
	"time"
)

// RunCLI runs the forge CLI with given arguments and returns stdout, stderr, and exit error
func RunCLI(ctx context.Context, workdir string, env map[string]string, args []string, timeout time.Duration) (string, string, error) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "forge", args...)
	cmd.Dir = workdir

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	// Capture output
	output, err := cmd.CombinedOutput()
	return string(output), "", err
}
