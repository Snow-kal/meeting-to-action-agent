package syncer

import (
	"context"
	"errors"
	"time"
)

type HTTPStatusError struct {
	StatusCode int
	Message    string
}

func (e *HTTPStatusError) Error() string {
	return e.Message
}

type RetryConfig struct {
	MaxAttempts int
	BaseBackoff time.Duration
}

func (c RetryConfig) normalized() RetryConfig {
	rc := c
	if rc.MaxAttempts <= 0 {
		rc.MaxAttempts = 1
	}
	if rc.BaseBackoff <= 0 {
		rc.BaseBackoff = 200 * time.Millisecond
	}
	return rc
}

func doWithRetry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	cfg = cfg.normalized()
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt == cfg.MaxAttempts || !isRetriable(err) {
				return err
			}

			wait := backoff(cfg.BaseBackoff, attempt)
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			continue
		}
		return nil
	}

	return lastErr
}

func isRetriable(err error) bool {
	var httpErr *HTTPStatusError
	if errors.As(err, &httpErr) {
		if httpErr.StatusCode == 429 || httpErr.StatusCode >= 500 {
			return true
		}
		return false
	}
	return true
}

func backoff(base time.Duration, attempt int) time.Duration {
	// 200ms, 400ms, 800ms...
	return base * time.Duration(1<<(attempt-1))
}
