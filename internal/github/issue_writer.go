package github

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type CreateIssueRequest struct {
	Title  string
	Body   string
	Labels []string
}

type CreatedIssue struct {
	Number int64
	URL    string
}

type IssueWriter interface {
	Create(ctx context.Context, repository string, req CreateIssueRequest) (CreatedIssue, error)
}

type commandRunner interface {
	CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error)
}

type execCommandRunner struct{}

func (execCommandRunner) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

type GHCLIIssueWriter struct {
	runner commandRunner
}

func NewGHCLIIssueWriter() GHCLIIssueWriter {
	return GHCLIIssueWriter{runner: execCommandRunner{}}
}

func (w GHCLIIssueWriter) Create(ctx context.Context, repository string, req CreateIssueRequest) (CreatedIssue, error) {
	args := []string{
		"issue",
		"create",
		"--repo", repository,
		"--title", req.Title,
		"--body", req.Body,
	}
	for _, label := range req.Labels {
		if strings.TrimSpace(label) == "" {
			continue
		}
		args = append(args, "--label", label)
	}

	output, err := w.runner.CombinedOutput(ctx, "gh", args...)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return CreatedIssue{}, fmt.Errorf("gh CLI not found; install GitHub CLI and run gh auth login")
		}

		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return CreatedIssue{}, fmt.Errorf("create issue via gh: %s", message)
	}

	issueURL, err := extractIssueURL(string(output))
	if err != nil {
		return CreatedIssue{}, err
	}
	number, err := parseIssueNumber(issueURL)
	if err != nil {
		return CreatedIssue{}, err
	}

	return CreatedIssue{
		Number: number,
		URL:    issueURL,
	}, nil
}

func extractIssueURL(output string) (string, error) {
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if _, err := url.ParseRequestURI(line); err == nil {
			return line, nil
		}
	}
	return "", fmt.Errorf("extract issue url: no URL found in gh output")
}

func parseIssueNumber(issueURL string) (int64, error) {
	parsed, err := url.Parse(issueURL)
	if err != nil {
		return 0, fmt.Errorf("parse issue url: %w", err)
	}

	number, err := strconv.ParseInt(path.Base(parsed.Path), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse issue number: %w", err)
	}
	return number, nil
}
