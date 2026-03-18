package syncer

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoWithRetryRetriable(t *testing.T) {
	attempts := 0
	err := doWithRetry(context.Background(), RetryConfig{
		MaxAttempts: 3,
		BaseBackoff: 1 * time.Millisecond,
	}, func() error {
		attempts++
		if attempts < 3 {
			return &HTTPStatusError{StatusCode: 500, Message: "server error"}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("want 3 attempts, got %d", attempts)
	}
}

func TestDoWithRetryNonRetriable(t *testing.T) {
	attempts := 0
	err := doWithRetry(context.Background(), RetryConfig{
		MaxAttempts: 3,
		BaseBackoff: 1 * time.Millisecond,
	}, func() error {
		attempts++
		return &HTTPStatusError{StatusCode: 400, Message: "bad request"}
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if attempts != 1 {
		t.Fatalf("want 1 attempt, got %d", attempts)
	}
}

func TestDoWithRetryContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := doWithRetry(ctx, RetryConfig{
		MaxAttempts: 3,
		BaseBackoff: 1 * time.Second,
	}, func() error {
		return errors.New("network error")
	})
	if err == nil {
		t.Fatalf("expected context error")
	}
}
