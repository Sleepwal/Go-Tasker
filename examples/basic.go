package main

import (
	"context"
	"fmt"
	"time"

	"github.com/SleepWalker/go-tasker/scheduler"
)

type PrintJob struct {
	name string
}

func (p PrintJob) ID() string {
	return "print-" + p.name
}

func (p PrintJob) Run(ctx context.Context) error {
	fmt.Printf("hello tasker: %s\n", p.name)
	return nil
}

func main() {
	s := scheduler.NewScheduler(5)
	s.Start()
	defer s.Stop()

	s.ScheduleAfter(PrintJob{name: "job1"}, 3*time.Second)
	s.ScheduleAfter(PrintJob{name: "job2"}, 2*time.Second)
	s.ScheduleAfter(PrintJob{name: "job3"}, 1*time.Second)

	time.Sleep(5 * time.Second)
}
