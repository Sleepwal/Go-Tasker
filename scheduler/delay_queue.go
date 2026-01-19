package scheduler

import (
	"container/heap"
	"time"

	"github.com/SleepWalker/go-tasker/job"
)

// item 表示队列中的任务项，包含任务和执行时间
type item struct {
	job   job.Job
	runAt time.Time
}

// delayQueue 延迟队列，使用堆实现，按执行时间排序
type delayQueue []*item

func (dq delayQueue) Len() int {
	return len(dq)
}

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

func (dq *delayQueue) Init() {
	heap.Init(dq)
}
