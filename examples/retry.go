package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/SleepWalker/go-tasker/scheduler"
)

type unreliableService struct {
	callCount int
	maxFails  int
}

func (s *unreliableService) ID() string {
	return "unreliable-service"
}

func (s *unreliableService) Run(ctx context.Context) error {
	s.callCount++
	if s.callCount <= s.maxFails {
		return fmt.Errorf("service temporarily unavailable (attempt %d)", s.callCount)
	}
	log.Printf("service succeeded on attempt %d", s.callCount)
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool := scheduler.NewWorkerPool(3)
	pool.Start(ctx)
	defer pool.Stop()

	policy := scheduler.RetryPolicy{
		MaxRetries:   5,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		ShouldRetry: func(err error) bool {
			return err.Error() != "permanent failure"
		},
		Timeout: 10 * time.Second,
		OnFailure: func(j scheduler.Job, err error) {
			log.Printf("job %s permanently failed: %v", j.ID(), err)
		},
	}

	service := &unreliableService{maxFails: 2}

	err := pool.SubmitWithRetry(ctx, service, policy)
	if err != nil {
		log.Printf("final result: %v", err)
	} else {
		log.Println("service recovered successfully!")
	}
}
