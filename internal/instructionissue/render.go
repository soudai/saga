package instructionissue

import (
	"fmt"
	"regexp"
	"strings"
)

const markerPrefix = "<!-- sg:instruction-issue"

var markdownHeadingPattern = regexp.MustCompile(`^#{1,3}\s+\S`)
var markdownHeadingCapturePattern = regexp.MustCompile(`^#{1,3}\s+(.+)$`)

type RenderOptions struct {
	Repository string
	Title      string
}

type RenderedIssue struct {
	Title string
	Body  string
}

func ExtractTitle(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if matches := markdownHeadingCapturePattern.FindStringSubmatch(trimmed); len(matches) == 2 {
			return strings.TrimSpace(matches[1])
		}
		return trimmed
	}
	return ""
}

func Render(content string, opts RenderOptions) (RenderedIssue, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return RenderedIssue{}, fmt.Errorf("instruction issue content is required")
	}

	title := strings.TrimSpace(opts.Title)
	if title == "" {
		title = ExtractTitle(trimmed)
	}
	if title == "" {
		return RenderedIssue{}, fmt.Errorf("instruction issue title is required")
	}

	body := trimmed
	if !hasMarkdownHeading(trimmed) {
		body = fmt.Sprintf(`# %s

## Background / Goal

%s

## Scope

- Inspect the target repository and narrow the implementation scope to the actual code paths.

## Acceptance Criteria

- The requested change is implemented.
- Relevant tests or validation steps are updated.
`, title, trimmed)
	}

	if !strings.HasPrefix(body, markerPrefix) {
		body = fmt.Sprintf("%s\n\n%s", buildMarker(opts.Repository), body)
	}

	return RenderedIssue{
		Title: title,
		Body:  body,
	}, nil
}

func hasMarkdownHeading(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if markdownHeadingPattern.MatchString(strings.TrimSpace(line)) {
			return true
		}
	}
	return false
}

func buildMarker(repository string) string {
	if strings.TrimSpace(repository) == "" {
		return "<!-- sg:instruction-issue v1 -->"
	}
	return fmt.Sprintf("<!-- sg:instruction-issue v1 repo=%s -->", repository)
}
