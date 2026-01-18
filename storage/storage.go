package storage

import (
	"time"

	"github.com/SleepWalker/go-tasker/scheduler"
)

type Storage interface {
	Save(job scheduler.Job, runAt time.Time) error
	Delete(jobID string) error
	ListReady(now time.Time) ([]scheduler.Job, error)
}
