package control_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/soudai/saga/internal/control"
	"github.com/soudai/saga/internal/store"
	"github.com/soudai/saga/internal/store/sqlite"
)

func TestStatusHandler(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "saga.db")
	sqliteStore, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		_ = sqliteStore.Close()
	}()

	ctx := context.Background()
	task, err := sqliteStore.CreateTask(ctx, "soudai/saga", 22, store.TaskStateQueued)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if _, err := sqliteStore.CreateRun(ctx, task.ID, "running"); err != nil {
		t.Fatalf("CreateRun() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()

	control.NewServer(sqliteStore).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", rec.Code)
	}

	var payload control.StatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if len(payload.Tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(payload.Tasks))
	}
	if payload.ActiveRuns != 1 {
		t.Fatalf("active runs = %d, want 1", payload.ActiveRuns)
	}
}
