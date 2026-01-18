package scheduler

import (
	"container/heap"
	"context"
	"log"
	"sync"
	"time"

	"github.com/SleepWalker/go-tasker/job"
	"github.com/SleepWalker/go-tasker/storage"
)

type Scheduler interface {
	Start()
	Stop()
	Schedule(j job.Job, runAt time.Time)
	ScheduleAfter(j job.Job, delay time.Duration)
}

type schedulerImpl struct {
	ctx     context.Context
	cancel  context.CancelFunc
	pool    *WorkerPool
	queue   delayQueue
	storage storage.Storage
	mu      sync.RWMutex
}

func NewScheduler(workers int) Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	q := delayQueue{}
	q.Init()
	return &schedulerImpl{
		ctx:     ctx,
		cancel:  cancel,
		pool:    NewWorkerPool(workers),
		queue:   q,
		storage: storage.NewMemoryStorage(),
	}
}

func (s *schedulerImpl) Start() {
	log.Printf("scheduler starting with %d workers", s.pool.workers)
	s.pool.Start(s.ctx)
	go s.loop()
	log.Println("scheduler started")
}

func (s *schedulerImpl) Stop() {
	log.Println("scheduler stopping...")
	s.cancel()
	s.pool.Stop()
	log.Println("scheduler stopped")
}

func (s *schedulerImpl) Schedule(j job.Job, runAt time.Time) {
	s.storage.Save(j, runAt)
	s.mu.Lock()
	heap.Push(&s.queue, &item{job: j, runAt: runAt})
	s.mu.Unlock()
	log.Printf("job %s scheduled for %v", j.ID(), runAt)
}

func (s *schedulerImpl) ScheduleAfter(j job.Job, delay time.Duration) {
	runAt := time.Now().Add(delay)
	s.Schedule(j, runAt)
}

func (s *schedulerImpl) loop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case now := <-ticker.C:
			ready, err := s.storage.ListReady(now)
			if err != nil {
				log.Printf("failed to list ready jobs: %v", err)
				continue
			}

			for _, j := range ready {
				s.mu.Lock()
				if s.queue.Len() > 0 && s.queue[0].job.ID() == j.ID() {
					heap.Pop(&s.queue)
				}
				s.mu.Unlock()

				s.pool.Submit(j)
			}
		}
	}
}
