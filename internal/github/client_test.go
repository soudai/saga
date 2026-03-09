package github

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMatchesIssue(t *testing.T) {
	t.Parallel()

	issue := Issue{
		Number:    12,
		State:     "open",
		Labels:    []string{"saga:ready"},
		Assignees: []string{"saga-bot"},
		Comments:  []string{"/saga run"},
	}

	selector := Selector{
		Labels:    []string{"saga:ready"},
		Assignees: []string{"saga-bot"},
		Commands:  []string{"/saga run"},
	}

	if !MatchesIssue(issue, selector) {
		t.Fatal("MatchesIssue() = false, want true")
	}
}

func TestListOpenIssues(t *testing.T) {
	t.Parallel()

	client := NewClient("https://example.test", "soudai", "saga", &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			body := `[{"number":1,"state":"open","body":"issue","labels":[{"name":"saga:ready"}],"assignees":[{"login":"saga-bot"}]},{"number":2,"state":"open","body":"pr","pull_request":{}}]`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	})

	issues, err := client.ListOpenIssues(context.Background())
	if err != nil {
		t.Fatalf("ListOpenIssues() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("len(issues) = %d, want 1", len(issues))
	}
	if issues[0].Number != 1 {
		t.Fatalf("number = %d, want 1", issues[0].Number)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
