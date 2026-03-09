package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"version"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}

	if got := strings.TrimSpace(stdout.String()); got == "" {
		t.Fatal("expected version output")
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunDoctor(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"doctor"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}

	if !strings.Contains(stdout.String(), "config validation") {
		t.Fatalf("stdout = %q, want doctor output", stdout.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"unknown"}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run() code = %d, want 1", code)
	}

	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr = %q, want unknown command error", stderr.String())
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunEnqueueInvalidIssue(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"enqueue", "soudai/saga", "not-a-number"}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run() code = %d, want 1", code)
	}

	if !strings.Contains(stderr.String(), "invalid issue number") {
		t.Fatalf("stderr = %q, want invalid issue number error", stderr.String())
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}
