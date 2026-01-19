package scheduler

import (
	"time"

	"github.com/SleepWalker/go-tasker/job"
)

const (
	// DefaultMaxRetries 默认最大重试次数
	DefaultMaxRetries = 3
	// DefaultInitialDelay 默认初始延迟时间
	DefaultInitialDelay = 1 * time.Second
	// DefaultMaxDelay 默认最大延迟时间
	DefaultMaxDelay = 1 * time.Minute
	// DefaultMultiplier 默认延迟倍数（指数退避）
	DefaultMultiplier = 2.0
	// DefaultTimeout 默认单次执行超时时间
	DefaultTimeout = 30 * time.Second
)

type RetryPolicy struct {
	// MaxRetries 最大重试次数，默认 3 次
	// 设置为 0 表示不重试
	MaxRetries int
	// InitialDelay 首次重试的延迟时间，默认 1 秒
	// 后续重试延迟会按 Multiplier 递增
	InitialDelay time.Duration
	// MaxDelay 单次重试的最大延迟时间，默认 1 分钟
	// 防止延迟时间过长
	MaxDelay time.Duration
	// Multiplier 延迟时间的乘数，默认 2.0（指数退避）
	// 每次重试延迟 = InitialDelay × Multiplier^retryCount
	Multiplier float64
	// ShouldRetry 自定义重试条件函数
	// 返回 true 继续重试，false 立即停止
	ShouldRetry func(err error) bool
	// Timeout 单次执行的超时时间，默认 30 秒
	// 为 0 时使用默认值
	Timeout time.Duration
	// OnFailure 所有重试耗尽后的回调函数
	// 参数为失败的任务和最终错误
	OnFailure func(j job.Job, err error)
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:   DefaultMaxRetries,
		InitialDelay: DefaultInitialDelay,
		MaxDelay:     DefaultMaxDelay,
		Multiplier:   DefaultMultiplier,
		ShouldRetry:  func(err error) bool { return true },
		Timeout:      DefaultTimeout,
		OnFailure:    nil,
	}
}

func (p RetryPolicy) withDefaults() RetryPolicy {
	if p.MaxRetries == 0 && p.InitialDelay == 0 && p.MaxDelay == 0 && p.Multiplier == 0 && p.Timeout == 0 {
		p.MaxRetries = DefaultMaxRetries
		p.InitialDelay = DefaultInitialDelay
		p.MaxDelay = DefaultMaxDelay
		p.Multiplier = DefaultMultiplier
		p.Timeout = DefaultTimeout
	} else {
		if p.MaxRetries == 0 {
			p.MaxRetries = DefaultMaxRetries
		}
		if p.InitialDelay == 0 {
			p.InitialDelay = DefaultInitialDelay
		}
		if p.MaxDelay == 0 {
			p.MaxDelay = DefaultMaxDelay
		}
		if p.Multiplier == 0 {
			p.Multiplier = DefaultMultiplier
		}
		if p.Timeout == 0 {
			p.Timeout = DefaultTimeout
		}
	}
	if p.ShouldRetry == nil {
		p.ShouldRetry = func(err error) bool { return true }
	}
	return p
}
