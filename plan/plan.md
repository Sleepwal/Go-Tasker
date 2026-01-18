🚀 Go 并发任务调度器（开源级设计）

> 项目目标：
> 一个 **高性能、可扩展、易集成** 的 Go 任务调度框架
> 支持：**定时 / 延迟 / 并发执行 / 重试 / 优雅关闭**

------

## 一、核心能力拆分（非常重要）

我们把系统拆成 **5 个核心模块**：

```
Scheduler（调度）
   ├── Job（任务定义）
   ├── DelayQueue（延迟队列）
   ├── WorkerPool（执行器）
   ├── Retry（失败重试）
   └── Storage（任务持久化）
```

👉 这样设计 **开源可读性高、易扩展、易 PR**

------

## 二、项目目录结构（推荐直接用）

```text
go-tasker/
├── cmd/
│   └── tasker/
│       └── main.go
├── scheduler/
│   ├── scheduler.go
│   ├── job.go
│   ├── worker_pool.go
│   ├── delay_queue.go
│   └── retry.go
├── storage/
│   ├── storage.go
│   └── memory.go
├── examples/
│   └── basic.go
├── README.md
├── LICENSE
└── go.mod
```

------

## 三、核心接口设计（⭐ 开源成败关键）

### 1️⃣ Job 定义

```go
// scheduler/job.go
package scheduler

import "context"

type Job interface {
	ID() string
	Run(ctx context.Context) error
}
```

✅ 用 interface，而不是 func
✅ 用户可以自定义 Job（非常开源友好）

------

### 2️⃣ Scheduler 核心接口

```go
// scheduler/scheduler.go
package scheduler

import "time"

type Scheduler interface {
	Start()
	Stop()
	Schedule(job Job, runAt time.Time)
	ScheduleAfter(job Job, delay time.Duration)
}
```

------

### 3️⃣ Storage（为 Redis 留好接口）

```go
// storage/storage.go
package storage

import (
	"time"
	"your_project/scheduler"
)

type Storage interface {
	Save(job scheduler.Job, runAt time.Time) error
	Delete(jobID string) error
	ListReady(now time.Time) ([]scheduler.Job, error)
}
```

------

## 四、Delay Queue（技术亮点之一）

用 `heap` 实现 **时间优先队列**

```go
// scheduler/delay_queue.go
package scheduler

import (
	"container/heap"
	"time"
)

type item struct {
	job   Job
	runAt time.Time
}

type delayQueue []*item

func (dq delayQueue) Len() int { return len(dq) }
func (dq delayQueue) Less(i, j int) bool {
	return dq[i].runAt.Before(dq[j].runAt)
}
func (dq delayQueue) Swap(i, j int) {
	dq[i], dq[j] = dq[j], dq[i]
}

func (dq *delayQueue) Push(x interface{}) {
	*dq = append(*dq, x.(*item))
}

func (dq *delayQueue) Pop() interface{} {
	old := *dq
	n := len(old)
	it := old[n-1]
	*dq = old[:n-1]
	return it
}
```

------

## 五、Worker Pool（并发执行核心）

```go
// scheduler/worker_pool.go
package scheduler

import "context"

type WorkerPool struct {
	workers int
	jobCh   chan Job
}

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		jobCh:   make(chan Job),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job := <-wp.jobCh:
					_ = job.Run(ctx)
				}
			}
		}()
	}
}

func (wp *WorkerPool) Submit(job Job) {
	wp.jobCh <- job
}
```

------

## 六、Scheduler 主逻辑（调度核心）

```go
// scheduler/scheduler.go
package scheduler

import (
	"context"
	"time"
)

type schedulerImpl struct {
	ctx    context.Context
	cancel context.CancelFunc
	pool   *WorkerPool
	queue  delayQueue
}

func NewScheduler(workers int) Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &schedulerImpl{
		ctx:    ctx,
		cancel: cancel,
		pool:   NewWorkerPool(workers),
		queue:  delayQueue{},
	}
}

func (s *schedulerImpl) Start() {
	s.pool.Start(s.ctx)
	go s.loop()
}

func (s *schedulerImpl) loop() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			// 检查到期任务（伪代码）
			_ = now
		}
	}
}

func (s *schedulerImpl) Stop() {
	s.cancel()
}
```

------

## 七、示例（examples/basic.go）

```go
package main

import (
	"context"
	"fmt"
	"time"
)

type PrintJob struct{}

func (p PrintJob) ID() string { return "print" }
func (p PrintJob) Run(ctx context.Context) error {
	fmt.Println("hello tasker")
	return nil
}

func main() {
	s := scheduler.NewScheduler(5)
	s.Start()

	s.ScheduleAfter(PrintJob{}, 3*time.Second)
	time.Sleep(5 * time.Second)
	s.Stop()
}
```

------

## 八、README 一句话模板（送你）

> **go-tasker** is a lightweight, high-performance task scheduling framework in Go,
> supporting delayed jobs, cron jobs, retries, and graceful shutdown.

------

## 九、后续升级路线（决定 Star 数）

1️⃣ v0.1：内存调度 + worker pool
2️⃣ v0.2：失败重试 + timeout
3️⃣ v0.3：cron 表达式
4️⃣ v0.4：Redis storage
5️⃣ v1.0：gRPC / HTTP 管理 API

## 