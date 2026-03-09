package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/soudai/saga/internal/store"
)

func TestStoreStatus(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "saga.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		_ = s.Close()
	}()

	ctx := context.Background()
	task, err := s.CreateTask(ctx, "soudai/saga", 1, store.TaskStateQueued)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if _, err := s.CreateRun(ctx, task.ID, "running"); err != nil {
		t.Fatalf("CreateRun() error = %v", err)
	}

	if err := s.AcquireLease(ctx, "issue:1", "worker-1", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("AcquireLease() error = %v", err)
	}

	status, err := s.Status(ctx)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	if len(status.Tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(status.Tasks))
	}
	if status.ActiveRuns != 1 {
		t.Fatalf("active runs = %d, want 1", status.ActiveRuns)
	}
}
