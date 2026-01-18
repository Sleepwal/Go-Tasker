package scheduler

import (
	"context"
	"testing"
	"time"
)

type workerTestJob struct {
	id string
}

func (tj *workerTestJob) ID() string {
	return tj.id
}

func (tj *workerTestJob) Run(ctx context.Context) error {
	return nil
}

func TestWorkerPool_Submit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := NewWorkerPool(2)
	pool.Start(ctx)
	defer pool.Stop()

	j := &workerTestJob{id: "test"}
	pool.Submit(j)

	time.Sleep(100 * time.Millisecond)
}

func TestWorkerPool_Stop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewWorkerPool(2)
	pool.Start(ctx)

	cancel()
	pool.Stop()
}
