package analytics

import "time"

// RetryPolicy controls exponential retry timing for failed analytics outbox rows.
type RetryPolicy struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

// DefaultRetryPolicy returns the retry schedule used by the analytics processor.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		BaseDelay: 5 * time.Second,
		MaxDelay:  5 * time.Minute,
	}
}

// RetryDelay returns the backoff delay for a given attempt number.
func RetryDelay(attempt int, policy RetryPolicy) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	// backfill any missing policy values with defaults
	if policy.BaseDelay <= 0 {
		policy.BaseDelay = DefaultRetryPolicy().BaseDelay
	}
	if policy.MaxDelay <= 0 {
		policy.MaxDelay = DefaultRetryPolicy().MaxDelay
	}

	delay := policy.BaseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= policy.MaxDelay {
			return policy.MaxDelay
		}
	}
	if delay > policy.MaxDelay {
		return policy.MaxDelay
	}
	return delay
}

// NextRetryAt returns the absolute retry time for a failed attempt.
func NextRetryAt(attempt int, now time.Time, policy RetryPolicy) time.Time {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return now.UTC().Add(RetryDelay(attempt, policy))
}