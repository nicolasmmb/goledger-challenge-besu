package blockchain

import (
	"context"
	"fmt"
	"time"
)

func Retry(ctx context.Context, attempts int, baseDelay time.Duration, fn func() error) error {
	if attempts <= 0 {
		attempts = 1
	}
	delay := baseDelay
	var lastErr error

	for i := 0; i < attempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if i == attempts-1 {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay *= 2
	}

	return fmt.Errorf("retry exhausted: %w", lastErr)
}
