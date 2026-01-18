package scheduler

import "time"

type Scheduler interface {
	Start()
	Stop()
	Schedule(job Job, runAt time.Time)
	ScheduleAfter(job Job, delay time.Duration)
}
