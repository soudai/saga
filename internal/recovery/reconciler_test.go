package recovery

import (
	"testing"

	"github.com/soudai/saga/internal/store"
)

func TestReconcile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		snapshot Snapshot
		want     Decision
	}{
		{
			name: "cancelled",
			snapshot: Snapshot{
				Task: store.Task{State: store.TaskStateCancelled},
			},
			want: DecisionCancel,
		},
		{
			name: "queued",
			snapshot: Snapshot{
				Task: store.Task{State: store.TaskStateQueued},
			},
			want: DecisionRetry,
		},
		{
			name: "running with pr",
			snapshot: Snapshot{
				Task:   store.Task{State: store.TaskStateRunning},
				PROpen: true,
			},
			want: DecisionKeepRunning,
		},
		{
			name: "running with branch",
			snapshot: Snapshot{
				Task:         store.Task{State: store.TaskStateRunning},
				BranchExists: true,
			},
			want: DecisionKeepRunning,
		},
		{
			name: "running without pr or branch",
			snapshot: Snapshot{
				Task: store.Task{State: store.TaskStateRunning},
			},
			want: DecisionRetry,
		},
		{
			name: "completed",
			snapshot: Snapshot{
				Task: store.Task{State: store.TaskStateCompleted},
			},
			want: DecisionComplete,
		},
		{
			name: "failed",
			snapshot: Snapshot{
				Task: store.Task{State: store.TaskStateFailed},
			},
			want: DecisionRetry,
		},
		{
			name: "unknown with pr",
			snapshot: Snapshot{
				Task:   store.Task{State: store.TaskState("mystery")},
				PROpen: true,
			},
			want: DecisionComplete,
		},
		{
			name: "unknown without pr",
			snapshot: Snapshot{
				Task: store.Task{State: store.TaskState("mystery")},
			},
			want: DecisionRetry,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Reconcile(tt.snapshot); got != tt.want {
				t.Fatalf("Reconcile() = %s, want %s", got, tt.want)
			}
		})
	}
}
