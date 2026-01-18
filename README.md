# Go Tasker

Go Tasker is a lightweight, high-performance task scheduling framework in Go, supporting delayed jobs, cron jobs, retries, and graceful shutdown.

## Features

- Scheduled and delayed job execution
- Concurrent worker pool for parallel processing
- Retry mechanism for failed jobs
- Graceful shutdown support
- Simple and extensible API

## Installation

```bash
go get github.com/SleepWalker/go-tasker
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/SleepWalker/go-tasker/scheduler"
)

type PrintJob struct{}

func (p PrintJob) ID() string {
	return "print"
}

func (p PrintJob) Run(ctx context.Context) error {
	fmt.Println("hello tasker")
	return nil
}

func main() {
	s := scheduler.NewScheduler(5)
	s.Start()
	defer s.Stop()

	s.ScheduleAfter(PrintJob{}, 3*time.Second)
	time.Sleep(5 * time.Second)
}
```

## Documentation

For more information, see the [documentation](docs/).

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.
