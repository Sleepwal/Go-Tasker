package main

import (
	"context"
	"fmt"

	"github.com/SleepWalker/go-tasker/scheduler"
)

func main() {
	s := scheduler.NewScheduler(5)
	s.Start()
	defer s.Stop()

	fmt.Println("Go Tasker - Task Scheduling Framework")
	fmt.Println("Scheduler started, press Ctrl+C to exit...")

	<-context.Background().Done()
}
