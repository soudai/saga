package recovery

import (
	"testing"
	"time"

	"github.com/soudai/saga/internal/store"
)

func TestRetryPolicy(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{MaxAttempts: 3, Backoff: time.Second}
	if !policy.CanRetry(2) {
		t.Fatal("CanRetry(2) = false, want true")
	}
	if policy.CanRetry(3) {
		t.Fatal("CanRetry(3) = true, want false")
	}
	if delay := policy.NextDelay(2); delay != 3*time.Second {
		t.Fatalf("delay = %s, want 3s", delay)
	}
}

func TestIsStale(t *testing.T) {
	t.Parallel()

	now := time.Now()
	if !IsStale(now.Add(-2*time.Minute), time.Minute, now) {
		t.Fatal("IsStale() = false, want true")
	}
}

func TestReconcile(t *testing.T) {
	t.Parallel()

	decision := Reconcile(Snapshot{
		Task: store.Task{
			State: store.TaskStateRunning,
		},
		PROpen: false,
	})
	if decision != DecisionRetry {
		t.Fatalf("decision = %s, want %s", decision, DecisionRetry)
	}
}
