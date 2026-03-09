package recovery

import (
	"testing"
	"time"
)

func TestRetryPolicy(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{MaxAttempts: 3, Backoff: time.Second}
	if !policy.CanRetry(2) {
		t.Fatal("CanRetry(2) = false, want true")
	}
	if policy.CanRetry(-1) {
		t.Fatal("CanRetry(-1) = true, want false")
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
