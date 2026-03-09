package github

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestListOpenIssues(t *testing.T) {
	t.Parallel()

	pages := map[string]string{
		"1": `[{"number":1,"state":"open","body":"issue","labels":[{"name":"saga:ready"}],"assignees":[{"login":"saga-bot"}]},{"number":2,"state":"open","body":"pr","pull_request":{}}]`,
		"2": `[{"number":3,"state":"open","body":"issue-2","labels":[{"name":"saga:ready"}],"assignees":[{"login":"saga-bot"}]}]`,
		"3": `[]`,
	}
	client := NewClient("https://example.test", "soudai", "saga", &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			query, err := url.ParseQuery(req.URL.RawQuery)
			if err != nil {
				t.Fatalf("ParseQuery() error = %v", err)
			}
			if query.Get("per_page") != "100" {
				t.Fatalf("per_page = %q, want 100", query.Get("per_page"))
			}
			body := pages[query.Get("page")]
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
	if len(issues) != 2 {
		t.Fatalf("len(issues) = %d, want 2", len(issues))
	}
	if issues[0].Number != 1 {
		t.Fatalf("number = %d, want 1", issues[0].Number)
	}
	if issues[1].Number != 3 {
		t.Fatalf("number = %d, want 3", issues[1].Number)
	}
}

func TestNewClientSetsDefaultTimeout(t *testing.T) {
	t.Parallel()

	client := NewClient("https://example.test", "soudai", "saga", nil)
	if client.HTTP == nil {
		t.Fatal("HTTP client = nil")
	}
	if client.HTTP.Timeout != 30*time.Second {
		t.Fatalf("HTTP timeout = %s, want 30s", client.HTTP.Timeout)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
