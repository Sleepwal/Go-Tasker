package storage

import (
	"log"
	"sync"
	"time"

	"github.com/SleepWalker/go-tasker/job"
)

type MemoryStorage struct {
	jobs map[string]*storageItem
	mu   sync.RWMutex
}

type storageItem struct {
	job   job.Job
	runAt time.Time
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		jobs: make(map[string]*storageItem),
	}
}

func (m *MemoryStorage) Save(j job.Job, runAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[j.ID()] = &storageItem{job: j, runAt: runAt}
	log.Printf("job %s saved to run at %v", j.ID(), runAt)
	return nil
}

func (m *MemoryStorage) Delete(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.jobs, jobID)
	log.Printf("job %s deleted", jobID)
	return nil
}

func (m *MemoryStorage) ListReady(now time.Time) ([]job.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var ready []job.Job
	for _, item := range m.jobs {
		if item.runAt.Before(now) {
			ready = append(ready, item.job)
		}
	}
	return ready, nil
}
