package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/soudai/saga/internal/control"
	sagagithub "github.com/soudai/saga/internal/github"
	"github.com/soudai/saga/internal/store"
)

func TestIssueDraftRendersTemplate(t *testing.T) {
	t.Parallel()

	path := writeIssueSource(t, "Implement auth support")

	var stdout bytes.Buffer
	configPath := ""
	cmd := newIssueCommandWithDeps(strings.NewReader(""), &stdout, &configPath, issueCommandDeps{})
	cmd.SetArgs([]string{"draft", "soudai/saga", "--from-file", path})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	for _, want := range []string{
		"<!-- sg:instruction-issue v1 repo=soudai/saga -->",
		"# Implement auth support",
		"## Background / Goal",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout.String())
		}
	}
}

func TestIssueCreateCanEnqueueCreatedIssue(t *testing.T) {
	t.Parallel()

	path := writeIssueSource(t, "# Implement auth\n\nDetails")

	writer := &stubIssueWriter{
		issue: sagagithub.CreatedIssue{
			Number: 123,
			URL:    "https://github.com/soudai/saga/issues/123",
		},
	}

	var stdout bytes.Buffer
	configPath := ""
	cmd := newIssueCommandWithDeps(strings.NewReader(""), &stdout, &configPath, issueCommandDeps{
		writer: writer,
		enqueue: func(_ context.Context, configPath, repository string, issueNumber int64) (control.TaskResponse, error) {
			return control.TaskResponse{
				ID:          7,
				Repository:  repository,
				IssueNumber: issueNumber,
				State:       store.TaskStateQueued,
			}, nil
		},
	})
	cmd.SetArgs([]string{"create", "soudai/saga", "--from-file", path, "--label", "saga:ready", "--enqueue"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if writer.repository != "soudai/saga" {
		t.Fatalf("repository = %q, want %q", writer.repository, "soudai/saga")
	}
	if writer.request.Title != "Implement auth" {
		t.Fatalf("title = %q, want %q", writer.request.Title, "Implement auth")
	}
	if len(writer.request.Labels) != 1 || writer.request.Labels[0] != "saga:ready" {
		t.Fatalf("labels = %#v, want saga:ready", writer.request.Labels)
	}

	for _, want := range []string{
		"issue #123 https://github.com/soudai/saga/issues/123",
		"task id=7 repo=soudai/saga issue=123 state=queued",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout.String())
		}
	}
}

func TestIssueCommandRejectsInvalidRepository(t *testing.T) {
	t.Parallel()

	path := writeIssueSource(t, "Implement auth support")

	configPath := ""
	cmd := newIssueCommandWithDeps(strings.NewReader(""), io.Discard, &configPath, issueCommandDeps{})
	cmd.SetArgs([]string{"draft", "invalid", "--from-file", path})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "owner/repo") {
		t.Fatalf("Execute() error = %v, want owner/repo validation", err)
	}
}

func writeIssueSource(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "task.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

type stubIssueWriter struct {
	repository string
	request    sagagithub.CreateIssueRequest
	issue      sagagithub.CreatedIssue
	err        error
}

func (s *stubIssueWriter) Create(_ context.Context, repository string, req sagagithub.CreateIssueRequest) (sagagithub.CreatedIssue, error) {
	s.repository = repository
	s.request = req
	return s.issue, s.err
}
