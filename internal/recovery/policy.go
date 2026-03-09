package recovery

import "time"

type RetryPolicy struct {
	MaxAttempts int
	Backoff     time.Duration
}

func (p RetryPolicy) CanRetry(attempt int) bool {
	if attempt < 0 {
		return false
	}
	return attempt < p.MaxAttempts
}

func (p RetryPolicy) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return p.Backoff
	}
	return time.Duration(attempt+1) * p.Backoff
}

func IsStale(updatedAt time.Time, leaseDuration time.Duration, now time.Time) bool {
	return now.Sub(updatedAt) > leaseDuration
}
