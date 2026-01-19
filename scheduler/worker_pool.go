package scheduler

import (
	"context"
	"log"
	"sync"

	"github.com/SleepWalker/go-tasker/job"
)

type WorkerPool struct {
	workers int
	jobCh   chan job.Job
	wg      sync.WaitGroup
	retry   *Retry
}

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		jobCh:   make(chan job.Job),
		retry:   NewRetry(),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("worker panic: %v", r)
				}
				wp.wg.Done()
			}()
			for {
				select {
				case <-ctx.Done():
					return
				case j := <-wp.jobCh:
					if err := j.Run(ctx); err != nil {
						log.Printf("job %s failed: %v", j.ID(), err)
					}
				}
			}
		}()
	}
}

func (wp *WorkerPool) Submit(j job.Job) {
	wp.jobCh <- j
}

func (wp *WorkerPool) SubmitWithRetry(ctx context.Context, j job.Job, policy RetryPolicy) error {
	return wp.retry.Retry(ctx, j, func(ctx context.Context) error {
		errCh := make(chan error, 1)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case wp.jobCh <- j:
		}

		go func() {
			errCh <- j.Run(ctx)
		}()

		select {
		case err := <-errCh:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}, policy)
}

func (wp *WorkerPool) SubmitSync(ctx context.Context, j job.Job) error {
	errCh := make(chan error, 1)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case wp.jobCh <- j:
	}

	go func() {
		errCh <- j.Run(ctx)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.jobCh)
	wp.wg.Wait()
}
