package control_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/soudai/saga/internal/control"
	"github.com/soudai/saga/internal/store"
	"github.com/soudai/saga/internal/store/sqlite"
)

func TestTaskActions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		action     string
		wantState  store.TaskState
		statusCode int
	}{
		{name: "cancel", action: "cancel", wantState: store.TaskStateCancelled, statusCode: http.StatusNoContent},
		{name: "retry", action: "retry", wantState: store.TaskStateQueued, statusCode: http.StatusNoContent},
		{name: "resume", action: "resume", wantState: store.TaskStateRunning, statusCode: http.StatusNoContent},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
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
			task, err := sqliteStore.CreateTask(ctx, "soudai/saga", 1, store.TaskStateQueued)
			if err != nil {
				t.Fatalf("CreateTask() error = %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/tasks/"+int64ToString(task.ID)+"/"+tt.action, nil)
			rec := httptest.NewRecorder()

			control.NewServer(sqliteStore).Handler().ServeHTTP(rec, req)

			if rec.Code != tt.statusCode {
				t.Fatalf("status code = %d, want %d", rec.Code, tt.statusCode)
			}

			tasks, err := sqliteStore.ListTasks(ctx)
			if err != nil {
				t.Fatalf("ListTasks() error = %v", err)
			}
			if tasks[0].State != tt.wantState {
				t.Fatalf("task state = %q, want %q", tasks[0].State, tt.wantState)
			}
		})
	}
}

func TestTaskActionErrors(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "saga.db")
	sqliteStore, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		_ = sqliteStore.Close()
	}()

	handler := control.NewServer(sqliteStore).Handler()

	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
	}{
		{name: "unknown action", method: http.MethodPost, path: "/tasks/1/pause", statusCode: http.StatusNotFound},
		{name: "non post", method: http.MethodGet, path: "/tasks/1/cancel", statusCode: http.StatusMethodNotAllowed},
		{name: "unknown task", method: http.MethodPost, path: "/tasks/999/cancel", statusCode: http.StatusNotFound},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.statusCode {
				t.Fatalf("status code = %d, want %d", rec.Code, tt.statusCode)
			}
		})
	}
}

func int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}
