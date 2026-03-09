package codex

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunnerSuccess(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "mock-codex.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho \"$SAGA_SANDBOX|$SAGA_NETWORK|$SAGA_MODEL\"\necho stderr 1>&2\n"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := NewRunner().Run(context.Background(), RunnerRequest{
		RunID:        "run-1",
		StageName:    "planner",
		CommandPath:  script,
		ArtifactRoot: filepath.Join(dir, "artifacts"),
		Sandbox:      SandboxWorkspaceWrite,
		Network:      false,
		Model:        "gpt-5-codex",
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

	stdoutBytes, err := os.ReadFile(result.StdoutPath)
	if err != nil {
		t.Fatalf("ReadFile(stdout) error = %v", err)
	}
	if !strings.Contains(string(stdoutBytes), "workspace-write|false|gpt-5-codex") {
		t.Fatalf("stdout = %q, want propagated runner env", stdoutBytes)
	}

	resultPath := filepath.Join(dir, "artifacts", "run-1", "planner", "result.json")
	resultBytes, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("ReadFile(result.json) error = %v", err)
	}
	var stored Result
	if err := json.Unmarshal(resultBytes, &stored); err != nil {
		t.Fatalf("Unmarshal(result.json) error = %v", err)
	}
	if stored.Model != "gpt-5-codex" {
		t.Fatalf("stored.Model = %q, want %q", stored.Model, "gpt-5-codex")
	}
}

func TestRunnerTimeout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "sleep.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 2\n"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := NewRunner().Run(context.Background(), RunnerRequest{
		RunID:        "run-2",
		StageName:    "tester",
		CommandPath:  script,
		ArtifactRoot: filepath.Join(dir, "artifacts"),
		Timeout:      50 * time.Millisecond,
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() error = %v, want context deadline exceeded", err)
	}
	if !result.TimedOut {
		t.Fatalf("TimedOut = %v, want true", result.TimedOut)
	}
}

func TestRunnerNonZeroExitPersistsArtifacts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fail.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho failed\necho boom 1>&2\nexit 42\n"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := NewRunner().Run(context.Background(), RunnerRequest{
		RunID:        "run-3",
		StageName:    "tester",
		CommandPath:  script,
		ArtifactRoot: filepath.Join(dir, "artifacts"),
	})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
	if result.ExitCode != 42 {
		t.Fatalf("exit code = %d, want 42", result.ExitCode)
	}

	resultPath := filepath.Join(dir, "artifacts", "run-3", "tester", "result.json")
	resultBytes, readErr := os.ReadFile(resultPath)
	if readErr != nil {
		t.Fatalf("ReadFile(result.json) error = %v", readErr)
	}
	var stored Result
	if err := json.Unmarshal(resultBytes, &stored); err != nil {
		t.Fatalf("Unmarshal(result.json) error = %v", err)
	}
	if stored.ExitCode != 42 {
		t.Fatalf("stored.ExitCode = %d, want 42", stored.ExitCode)
	}
}

func TestRunnerStartFailurePersistsArtifacts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	result, err := NewRunner().Run(context.Background(), RunnerRequest{
		RunID:        "run-4",
		StageName:    "tester",
		CommandPath:  filepath.Join(dir, "missing-command"),
		ArtifactRoot: filepath.Join(dir, "artifacts"),
	})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
	if result.ExitCode != -1 {
		t.Fatalf("exit code = %d, want -1", result.ExitCode)
	}

	for _, path := range []string{
		filepath.Join(dir, "artifacts", "run-4", "tester", "stdout.log"),
		filepath.Join(dir, "artifacts", "run-4", "tester", "stderr.log"),
		filepath.Join(dir, "artifacts", "run-4", "tester", "result.json"),
	} {
		if _, statErr := os.Stat(path); statErr != nil {
			t.Fatalf("Stat(%q) error = %v", path, statErr)
		}
	}
}
