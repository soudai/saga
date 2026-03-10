package github

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestGHCLIIssueWriterCreate(t *testing.T) {
	t.Parallel()

	runner := &stubCommandRunner{
		output: []byte("https://github.com/soudai/saga/issues/123\n"),
	}
	writer := GHCLIIssueWriter{runner: runner}

	issue, err := writer.Create(context.Background(), "soudai/saga", CreateIssueRequest{
		Title:  "Implement auth",
		Body:   "# Implement auth",
		Labels: []string{"saga:ready", "backend"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if issue.Number != 123 {
		t.Fatalf("number = %d, want 123", issue.Number)
	}
	if issue.URL != "https://github.com/soudai/saga/issues/123" {
		t.Fatalf("url = %q, want issue URL", issue.URL)
	}

	wantArgs := []string{
		"issue", "create",
		"--repo", "soudai/saga",
		"--title", "Implement auth",
		"--body", "# Implement auth",
		"--label", "saga:ready",
		"--label", "backend",
	}
	if !reflect.DeepEqual(runner.args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", runner.args, wantArgs)
	}
}

func TestGHCLIIssueWriterCreateReturnsHelpfulError(t *testing.T) {
	t.Parallel()

	runner := &stubCommandRunner{
		output: []byte("authentication failed"),
		err:    errors.New("exit status 1"),
	}
	writer := GHCLIIssueWriter{runner: runner}

	_, err := writer.Create(context.Background(), "soudai/saga", CreateIssueRequest{
		Title: "Implement auth",
		Body:  "# Implement auth",
	})
	if err == nil || err.Error() != "create issue via gh: authentication failed" {
		t.Fatalf("Create() error = %v, want gh output in error", err)
	}
}

type stubCommandRunner struct {
	args   []string
	output []byte
	err    error
}

func (s *stubCommandRunner) CombinedOutput(_ context.Context, _ string, args ...string) ([]byte, error) {
	s.args = append([]string(nil), args...)
	return s.output, s.err
}
