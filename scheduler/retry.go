package scheduler

import (
	"context"
	"math"
	"time"

	"github.com/SleepWalker/go-tasker/job"
)

// Retry 重试处理器，负责执行带重试的任务
type Retry struct{}

func NewRetry() *Retry {
	return &Retry{}
}

// Retry 执行重试逻辑，包含指数退避和超时控制
//
// 参数：
//
//	ctx - 上下文，用于控制超时和取消
//	j - 需要执行的任务
//	fn - 任务执行函数
//	policy - 重试策略配置
//
// 返回值：
//
//	所有重试均失败返回最后错误；成功返回 nil
//
// 重试逻辑：
//  1. 首次执行后若失败，检查是否应该重试
//  2. 计算延迟时间（初始延迟 × 乘数^重试次数）
//  3. 等待延迟后进行下一次尝试
//  4. 达到最大重试次数后调用 OnFailure 回调
func (r *Retry) Retry(ctx context.Context, j job.Job, fn func(context.Context) error, policy RetryPolicy) error {
	policy = policy.withDefaults()

	var lastErr error
	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		timeoutCtx, cancel := context.WithTimeout(ctx, policy.Timeout)
		err := fn(timeoutCtx)
		cancel()

		if err == nil {
			return nil
		}

		lastErr = err

		if !policy.ShouldRetry(err) {
			return err
		}

		if attempt < policy.MaxRetries {
			delay := time.Duration(float64(policy.InitialDelay) * math.Pow(policy.Multiplier, float64(attempt)))
			if delay > policy.MaxDelay {
				delay = policy.MaxDelay
			}

			select {
			case <-ctx.Done():
				if policy.OnFailure != nil {
					policy.OnFailure(j, ctx.Err())
				}
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	if policy.OnFailure != nil {
		policy.OnFailure(j, lastErr)
	}

	return lastErr
}
