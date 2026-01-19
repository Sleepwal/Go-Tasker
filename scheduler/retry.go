package scheduler

import (
	"context"
	"math"
	"time"

	"github.com/SleepWalker/go-tasker/job"
)

type Retry struct{}

func NewRetry() *Retry {
	return &Retry{}
}

func (r *Retry) Retry(ctx context.Context, j job.Job, fn func(context.Context) error, policy RetryPolicy) error {
	policy = policy.withDefaults()

	var lastErr error
	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		timeoutCtx, cancel := context.WithTimeout(ctx, policy.Timeout)
		err := fn(timeoutCtx)
		cancel()

		if err == nil {
			return nil
		}

		lastErr = err

		if !policy.ShouldRetry(err) {
			return err
		}

		if attempt < policy.MaxRetries {
			delay := time.Duration(float64(policy.InitialDelay) * math.Pow(policy.Multiplier, float64(attempt)))
			if delay > policy.MaxDelay {
				delay = policy.MaxDelay
			}

			select {
			case <-ctx.Done():
				if policy.OnFailure != nil {
					policy.OnFailure(j, ctx.Err())
				}
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	if policy.OnFailure != nil {
		policy.OnFailure(j, lastErr)
	}

	return lastErr
}
