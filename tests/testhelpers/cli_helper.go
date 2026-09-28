package testhelpers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// RunCLI runs the forge CLI with given arguments and returns stdout, stderr, and exit error
func RunCLI(ctx context.Context, workdir string, env map[string]string, args []string, timeout time.Duration) (string, string, error) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmdPath, cmdArgs, err := resolveForgeCommand(args)
	if err != nil {
		return "", "", err
	}
	cmd := exec.CommandContext(ctx, cmdPath, cmdArgs...)
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

var (
	forgeCmdOnce sync.Once
	forgeCmdPath string
	forgeCmdErr  error
)

func resolveForgeCommand(args []string) (string, []string, error) {
	if path, err := exec.LookPath("forge"); err == nil {
		return path, args, nil
	}

	forgeCmdOnce.Do(func() {
		forgeDir := repoRoot()
		binName := "forge"
		if runtime.GOOS == "windows" {
			binName += ".exe"
		}
		tmpDir, err := os.MkdirTemp("", "forge-cli-")
		if err != nil {
			forgeCmdErr = err
			return
		}
		forgeCmdPath = filepath.Join(tmpDir, binName)

		buildCmd := exec.Command("go", "build", "-o", forgeCmdPath, "./cmd/forge")
		buildCmd.Dir = forgeDir
		buildCmd.Env = os.Environ()
		output, err := buildCmd.CombinedOutput()
		if err != nil {
			forgeCmdErr = fmt.Errorf("failed to build forge CLI: %w: %s", err, string(output))
		}
	})
	if forgeCmdErr != nil {
		return "", nil, forgeCmdErr
	}

	return forgeCmdPath, args, nil
}

func repoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}
