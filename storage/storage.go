package storage

import (
	"time"

	"github.com/SleepWalker/go-tasker/job"
)

type Storage interface {
	Save(j job.Job, runAt time.Time) error
	Delete(jobID string) error
	ListReady(now time.Time) ([]job.Job, error)
}
