package control

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/soudai/saga/internal/store"
	"github.com/soudai/saga/internal/store/sqlite"
)

func TestClientEnqueue(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "sg.db")
	sqliteStore, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		_ = sqliteStore.Close()
	}()

	server := httptest.NewServer(NewServer(sqliteStore).Handler())
	defer server.Close()

	client, err := newTestClient(server)
	if err != nil {
		t.Fatalf("newTestClient() error = %v", err)
	}

	task, err := client.Enqueue(context.Background(), "soudai/saga", 123)
	if err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	if task.Repository != "soudai/saga" {
		t.Fatalf("repository = %q, want %q", task.Repository, "soudai/saga")
	}
	if task.IssueNumber != 123 {
		t.Fatalf("issue number = %d, want 123", task.IssueNumber)
	}
	if task.State != store.TaskStateQueued {
		t.Fatalf("state = %q, want %q", task.State, store.TaskStateQueued)
	}
}

func TestClientEnqueueError(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "sg.db")
	sqliteStore, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		_ = sqliteStore.Close()
	}()

	server := httptest.NewServer(NewServer(sqliteStore).Handler())
	defer server.Close()

	client, err := newTestClient(server)
	if err != nil {
		t.Fatalf("newTestClient() error = %v", err)
	}

	_, err = client.Enqueue(context.Background(), "", 123)
	if err == nil {
		t.Fatal("Enqueue() error = nil, want error")
	}
	if got := err.Error(); got != "enqueue request failed: repository is required" {
		t.Fatalf("error = %q, want %q", got, "enqueue request failed: repository is required")
	}
}

func newTestClient(server *httptest.Server) (*Client, error) {
	baseURL, err := url.Parse(server.URL)
	if err != nil {
		return nil, err
	}

	return &Client{
		http: &http.Client{
			Transport: rewriteHostTransport{
				baseURL:   baseURL,
				transport: server.Client().Transport,
			},
		},
	}, nil
}

type rewriteHostTransport struct {
	baseURL   *url.URL
	transport http.RoundTripper
}

func (t rewriteHostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	urlCopy := *clone.URL
	urlCopy.Scheme = t.baseURL.Scheme
	urlCopy.Host = t.baseURL.Host
	clone.URL = &urlCopy
	clone.Host = t.baseURL.Host
	return t.transport.RoundTrip(clone)
}
