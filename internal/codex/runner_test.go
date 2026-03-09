package codex

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunnerSuccess(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "mock-codex.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho stdout\necho stderr 1>&2\n"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := NewRunner().Run(context.Background(), RunnerRequest{
		RunID:        "run-1",
		StageName:    "planner",
		CommandPath:  script,
		ArtifactRoot: filepath.Join(dir, "artifacts"),
		Sandbox:      SandboxWorkspaceWrite,
		Network:      false,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", result.ExitCode)
	}
	if _, err := os.Stat(result.StdoutPath); err != nil {
		t.Fatalf("stdout artifact error = %v", err)
	}
	if _, err := os.Stat(result.StderrPath); err != nil {
		t.Fatalf("stderr artifact error = %v", err)
	}
}

func TestRunnerTimeout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "sleep.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 2\n"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := NewRunner().Run(context.Background(), RunnerRequest{
		RunID:        "run-2",
		StageName:    "tester",
		CommandPath:  script,
		ArtifactRoot: filepath.Join(dir, "artifacts"),
		Timeout:      50 * time.Millisecond,
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() error = %v, want context deadline exceeded", err)
	}
}
