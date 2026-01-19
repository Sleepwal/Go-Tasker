package scheduler

import (
	"context"
	"log"
	"sync"

	"github.com/SleepWalker/go-tasker/job"
)

// WorkerPool 管理 goroutine 池，用于并发执行任务
type WorkerPool struct {
	workers int
	jobCh   chan job.Job
	wg      sync.WaitGroup
	retry   *Retry
}

// NewWorkerPool 创建指定数量的 worker 池
func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		jobCh:   make(chan job.Job),
		retry:   NewRetry(),
	}
}

// Start 启动所有 worker goroutine
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

// Submit 提交任务到 worker 池（无重试）
func (wp *WorkerPool) Submit(j job.Job) {
	wp.jobCh <- j
}

// SubmitWithRetry 提交任务并执行重试策略
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

// SubmitSync 同步提交任务并等待执行结果
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

// Stop 优雅关闭 worker 池
func (wp *WorkerPool) Stop() {
	close(wp.jobCh)
	wp.wg.Wait()
}
