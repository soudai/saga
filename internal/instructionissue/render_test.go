package instructionissue

import (
	"strings"
	"testing"
)

func TestExtractTitlePrefersHeading(t *testing.T) {
	t.Parallel()

	got := ExtractTitle("\n# Implement auth\n\nDetails")
	if got != "Implement auth" {
		t.Fatalf("ExtractTitle() = %q, want %q", got, "Implement auth")
	}
}

func TestExtractTitleFallsBackToFirstNonEmptyLine(t *testing.T) {
	t.Parallel()

	got := ExtractTitle("\nImplement auth\n\nMore details")
	if got != "Implement auth" {
		t.Fatalf("ExtractTitle() = %q, want %q", got, "Implement auth")
	}
}

func TestRenderWrapsBriefInInstructionTemplate(t *testing.T) {
	t.Parallel()

	rendered, err := Render("Implement auth support", RenderOptions{Repository: "soudai/saga"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, want := range []string{
		"<!-- sg:instruction-issue v1 repo=soudai/saga -->",
		"# Implement auth support",
		"## Background / Goal",
		"## Scope",
		"## Acceptance Criteria",
	} {
		if !strings.Contains(rendered.Body, want) {
			t.Fatalf("rendered body missing %q:\n%s", want, rendered.Body)
		}
	}
	if rendered.Title != "Implement auth support" {
		t.Fatalf("title = %q, want %q", rendered.Title, "Implement auth support")
	}
}

func TestRenderPreservesExistingMarkdownHeading(t *testing.T) {
	t.Parallel()

	body := "# Task Instruction\n\n## Background\n\nDetails"
	rendered, err := Render(body, RenderOptions{Repository: "soudai/saga"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if strings.Count(rendered.Body, "# Task Instruction") != 1 {
		t.Fatalf("rendered body duplicated heading:\n%s", rendered.Body)
	}
	if !strings.HasPrefix(rendered.Body, "<!-- sg:instruction-issue v1 repo=soudai/saga -->") {
		t.Fatalf("rendered body missing marker:\n%s", rendered.Body)
	}
}
