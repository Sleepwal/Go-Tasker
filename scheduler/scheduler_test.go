package scheduler

import (
	"context"
	"testing"
	"time"
)

type schedulerTestJob struct {
	id string
}

func (tj *schedulerTestJob) ID() string {
	return tj.id
}

func (tj *schedulerTestJob) Run(ctx context.Context) error {
	return nil
}

func TestScheduler_ScheduleAfter(t *testing.T) {
	s := NewScheduler(2)
	s.Start()
	defer s.Stop()

	j := &schedulerTestJob{id: "test"}
	s.ScheduleAfter(j, 100*time.Millisecond)

	time.Sleep(200 * time.Millisecond)
}

func TestScheduler_Schedule(t *testing.T) {
	s := NewScheduler(2)
	s.Start()
	defer s.Stop()

	j := &schedulerTestJob{id: "test"}
	runAt := time.Now().Add(100 * time.Millisecond)
	s.Schedule(j, runAt)

	time.Sleep(200 * time.Millisecond)
}

func TestScheduler_Stop(t *testing.T) {
	s := NewScheduler(2)
	s.Start()
	s.Stop()
}
