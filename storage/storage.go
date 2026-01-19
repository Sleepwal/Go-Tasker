package storage

import (
	"time"

	"github.com/SleepWalker/go-tasker/job"
)

// Storage 任务存储接口，用于持久化任务调度信息
type Storage interface {
	// Save 保存任务及执行时间
	Save(j job.Job, runAt time.Time) error
	// Delete 删除任务
	Delete(jobID string) error
	// ListReady 列出所有到执行时间的任务
	ListReady(now time.Time) ([]job.Job, error)
}
