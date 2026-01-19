package scheduler

import (
	"time"

	"github.com/SleepWalker/go-tasker/job"
)

const (
	DefaultMaxRetries   = 3
	DefaultInitialDelay = 1 * time.Second
	DefaultMaxDelay     = 1 * time.Minute
	DefaultMultiplier   = 2.0
	DefaultTimeout      = 30 * time.Second
)

type RetryPolicy struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	ShouldRetry  func(err error) bool
	Timeout      time.Duration
	OnFailure    func(j job.Job, err error)
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:   DefaultMaxRetries,
		InitialDelay: DefaultInitialDelay,
		MaxDelay:     DefaultMaxDelay,
		Multiplier:   DefaultMultiplier,
		ShouldRetry:  func(err error) bool { return true },
		Timeout:      DefaultTimeout,
		OnFailure:    nil,
	}
}

func (p RetryPolicy) withDefaults() RetryPolicy {
	if p.MaxRetries == 0 && p.InitialDelay == 0 && p.MaxDelay == 0 && p.Multiplier == 0 && p.Timeout == 0 {
		p.MaxRetries = DefaultMaxRetries
		p.InitialDelay = DefaultInitialDelay
		p.MaxDelay = DefaultMaxDelay
		p.Multiplier = DefaultMultiplier
		p.Timeout = DefaultTimeout
	} else {
		if p.MaxRetries == 0 {
			p.MaxRetries = DefaultMaxRetries
		}
		if p.InitialDelay == 0 {
			p.InitialDelay = DefaultInitialDelay
		}
		if p.MaxDelay == 0 {
			p.MaxDelay = DefaultMaxDelay
		}
		if p.Multiplier == 0 {
			p.Multiplier = DefaultMultiplier
		}
		if p.Timeout == 0 {
			p.Timeout = DefaultTimeout
		}
	}
	if p.ShouldRetry == nil {
		p.ShouldRetry = func(err error) bool { return true }
	}
	return p
}
