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
}

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		jobCh:   make(chan job.Job),
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

func (wp *WorkerPool) Stop() {
	close(wp.jobCh)
	wp.wg.Wait()
}
