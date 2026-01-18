package scheduler

import "context"

type WorkerPool struct {
	workers int
	jobCh   chan Job
}

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		jobCh:   make(chan Job),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
}

func (wp *WorkerPool) Submit(job Job) {
}
