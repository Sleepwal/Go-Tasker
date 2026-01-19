package scheduler

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testJob struct {
	id      string
	execErr error
}

func (j *testJob) ID() string {
	return j.id
}

func (j *testJob) Run(ctx context.Context) error {
	return j.execErr
}

type countingJob struct {
	id        string
	execErr   error
	execCount *int32
}

func (j *countingJob) ID() string {
	return j.id
}

func (j *countingJob) Run(ctx context.Context) error {
	if j.execCount != nil {
		atomic.AddInt32(j.execCount, 1)
	}
	return j.execErr
}

func TestRetry_Success(t *testing.T) {
	r := NewRetry()
	execCount := int32(0)
	j := &countingJob{id: "success", execCount: &execCount}

	err := r.Retry(context.Background(), j, func(ctx context.Context) error {
		return j.Run(ctx)
	}, DefaultRetryPolicy())

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if execCount != 1 {
		t.Errorf("expected 1 execution, got %d", execCount)
	}
}

func TestRetry_RetryOnFailure(t *testing.T) {
	r := NewRetry()
	execCount := int32(0)
	j := &countingJob{id: "retry", execCount: &execCount, execErr: errors.New("temporary error")}

	policy := DefaultRetryPolicy()
	policy.MaxRetries = 2

	err := r.Retry(context.Background(), j, func(ctx context.Context) error {
		return j.Run(ctx)
	}, policy)

	if err == nil {
		t.Error("expected error after retries exhausted")
	}
	if execCount != 3 {
		t.Errorf("expected 3 executions (1 initial + 2 retries), got %d", execCount)
	}
}

func TestRetry_RetrySucceeds(t *testing.T) {
	r := NewRetry()
	execCount := int32(0)
	failTimes := int32(1)

	j := &countingJob{id: "retry-success", execCount: &execCount, execErr: errors.New("temporary error")}

	policy := DefaultRetryPolicy()
	policy.MaxRetries = 3

	err := r.Retry(context.Background(), j, func(ctx context.Context) error {
		count := atomic.AddInt32(&execCount, 1)
		if count <= failTimes {
			return errors.New("retryable error")
		}
		return nil
	}, policy)

	if err != nil {
		t.Errorf("expected success after retries, got error: %v", err)
	}
	if execCount != 2 {
		t.Errorf("expected 2 executions (1 fail + 1 success), got %d", execCount)
	}
}

func TestRetry_ShouldRetryFilter(t *testing.T) {
	r := NewRetry()
	execCount := int32(0)
	retryableErr := errors.New("retryable error")
	fatalErr := errors.New("fatal error")

	j := &testJob{id: "filter", execErr: retryableErr}

	policy := DefaultRetryPolicy()
	policy.MaxRetries = 3
	policy.ShouldRetry = func(err error) bool {
		return err.Error() != "fatal error"
	}

	err := r.Retry(context.Background(), j, func(ctx context.Context) error {
		atomic.AddInt32(&execCount, 1)
		if atomic.LoadInt32(&execCount) == 2 {
			return fatalErr
		}
		return retryableErr
	}, policy)

	if err != fatalErr {
		t.Errorf("expected fatal error, got %v", err)
	}
	if execCount != 2 {
		t.Errorf("expected 2 executions, got %d", execCount)
	}
}

func TestRetry_ZeroMaxRetries(t *testing.T) {
	r := NewRetry()
	execCount := int32(0)
	j := &testJob{id: "zero-retries", execErr: errors.New("error")}

	policy := RetryPolicy{
		MaxRetries:   0,
		InitialDelay: 10 * time.Millisecond,
		Timeout:      5 * time.Second,
		ShouldRetry:  func(err error) bool { return false },
	}

	err := r.Retry(context.Background(), j, func(ctx context.Context) error {
		atomic.AddInt32(&execCount, 1)
		return j.Run(ctx)
	}, policy)

	if err == nil {
		t.Error("expected error on first execution")
	}
	if execCount != 1 {
		t.Errorf("expected 1 execution with MaxRetries=0, got %d", execCount)
	}
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	r := NewRetry()
	execCount := int32(0)
	j := &testJob{id: "backoff", execErr: errors.New("error")}

	policy := RetryPolicy{
		MaxRetries:   3,
		InitialDelay: 5 * time.Millisecond,
		Multiplier:   2.0,
		MaxDelay:     50 * time.Millisecond,
		Timeout:      5 * time.Second,
	}

	start := time.Now()
	err := r.Retry(context.Background(), j, func(ctx context.Context) error {
		atomic.AddInt32(&execCount, 1)
		return j.Run(ctx)
	}, policy)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected error after retries exhausted")
	}
	if execCount != 4 {
		t.Errorf("expected 4 executions, got %d", execCount)
	}

	minExpected := 5 + 10 + 20
	if elapsed < time.Duration(minExpected)*time.Millisecond-5*time.Millisecond {
		t.Errorf("expected backoff delay at least %dms, got %v", minExpected, elapsed)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	r := NewRetry()
	execCount := int32(0)
	j := &testJob{id: "cancel", execErr: errors.New("error")}

	policy := DefaultRetryPolicy()
	policy.MaxRetries = 100
	policy.InitialDelay = 1 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := r.Retry(ctx, j, func(ctx context.Context) error {
		atomic.AddInt32(&execCount, 1)
		return j.Run(ctx)
	}, policy)

	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
	if execCount > 2 {
		t.Errorf("expected few executions before cancel, got %d", execCount)
	}
}

func TestRetry_CustomTimeout(t *testing.T) {
	r := NewRetry()
	execCount := int32(0)
	j := &testJob{id: "timeout", execErr: errors.New("error")}

	policy := DefaultRetryPolicy()
	policy.MaxRetries = 2
	policy.InitialDelay = 100 * time.Millisecond
	policy.Timeout = 50 * time.Millisecond

	ctx := context.Background()
	err := r.Retry(ctx, j, func(ctx context.Context) error {
		atomic.AddInt32(&execCount, 1)
		return j.Run(ctx)
	}, policy)

	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestRetry_OnFailureCallback(t *testing.T) {
	r := NewRetry()
	callbackCalled := int32(0)
	j := &testJob{id: "callback", execErr: errors.New("final error")}

	policy := DefaultRetryPolicy()
	policy.MaxRetries = 1
	policy.OnFailure = func(job Job, err error) {
		atomic.AddInt32(&callbackCalled, 1)
	}

	err := r.Retry(context.Background(), j, func(ctx context.Context) error {
		return j.Run(ctx)
	}, policy)

	if err == nil {
		t.Error("expected error")
	}
	if atomic.LoadInt32(&callbackCalled) != 1 {
		t.Errorf("expected OnFailure callback to be called once, got %d", callbackCalled)
	}
}

func TestRetry_DefaultTimeout(t *testing.T) {
	policy := DefaultRetryPolicy()

	if policy.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", policy.Timeout)
	}
}

func TestRetryPolicy_WithDefaults(t *testing.T) {
	policy := RetryPolicy{}
	policy = policy.withDefaults()

	if policy.MaxRetries != DefaultMaxRetries {
		t.Errorf("expected MaxRetries %d, got %d", DefaultMaxRetries, policy.MaxRetries)
	}
	if policy.InitialDelay != DefaultInitialDelay {
		t.Errorf("expected InitialDelay %v, got %v", DefaultInitialDelay, policy.InitialDelay)
	}
	if policy.MaxDelay != DefaultMaxDelay {
		t.Errorf("expected MaxDelay %v, got %v", DefaultMaxDelay, policy.MaxDelay)
	}
	if policy.Multiplier != DefaultMultiplier {
		t.Errorf("expected Multiplier %f, got %f", DefaultMultiplier, policy.Multiplier)
	}
	if policy.ShouldRetry == nil {
		t.Error("expected ShouldRetry to be set")
	}
	if policy.Timeout != DefaultTimeout {
		t.Errorf("expected Timeout %v, got %v", DefaultTimeout, policy.Timeout)
	}
}

func TestRetry_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	jobs := 10
	retries := 2
	successCount := int32(0)

	var start sync.WaitGroup
	start.Add(1)

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start.Wait()
			policy := DefaultRetryPolicy()
			policy.MaxRetries = retries
			policy.InitialDelay = 10 * time.Millisecond

			r := NewRetry()
			j := &testJob{id: "concurrent"}
			r.Retry(context.Background(), j, func(ctx context.Context) error {
				if atomic.AddInt32(&successCount, 1) <= int32(jobs) {
					return errors.New("keep trying")
				}
				return nil
			}, policy)
		}()
	}

	start.Done()
	wg.Wait()

	if atomic.LoadInt32(&successCount) < int32(jobs) {
		t.Errorf("expected at least %d total executions, got %d", jobs, successCount)
	}
}
