package grpc

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	JitterFactor  float64
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxRetries:    5,
	InitialDelay:  100 * time.Millisecond,
	MaxDelay:      30 * time.Second,
	BackoffFactor: 2.0,
	JitterFactor:  0.2,
}

// RetryableError 可重试的错误
type RetryableError struct {
	Err      error
	Retry    bool
	RetryNow bool
}

// Error 实现 error 接口
func (e *RetryableError) Error() string {
	return e.Err.Error()
}

// Unwrap 实现 errors.Unwrap 接口
func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NewRetryableError 创建可重试错误
func NewRetryableError(err error, retry bool) *RetryableError {
	return &RetryableError{
		Err:   err,
		Retry: retry,
	}
}

// NewRetryNowError 创建立即重试错误
func NewRetryNowError(err error) *RetryableError {
	return &RetryableError{
		Err:      err,
		Retry:    true,
		RetryNow: true,
	}
}

// Retry 重试函数
func Retry(ctx context.Context, fn func() error, config RetryConfig) error {
	var lastErr error
	for i := 0; i <= config.MaxRetries; i++ {
		// 执行函数
		err := fn()
		if err == nil {
			return nil
		}

		// 检查是否可重试
		if !IsRetryableError(err) {
			return err
		}

		// 保存最后一个错误
		lastErr = err

		// 如果是最后一次尝试，直接返回错误
		if i == config.MaxRetries {
			return fmt.Errorf("达到最大重试次数: %v", lastErr)
		}

		// 计算延迟时间
		delay := calculateDelay(i, config)

		// 等待一段时间
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	return lastErr
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	retryErr, ok := err.(*RetryableError)
	return ok && retryErr.Retry
}

// IsRetryNowError 检查错误是否需要立即重试
func IsRetryNowError(err error) bool {
	retryErr, ok := err.(*RetryableError)
	return ok && retryErr.RetryNow
}

// WrapError 包装错误
func WrapError(err error, retry bool) error {
	if err == nil {
		return nil
	}
	return NewRetryableError(err, retry)
}

// WrapRetryNowError 包装需要立即重试的错误
func WrapRetryNowError(err error) error {
	if err == nil {
		return nil
	}
	return NewRetryNowError(err)
}

func calculateDelay(attempt int, config RetryConfig) time.Duration {
	delay := config.InitialDelay
	for i := 0; i < attempt; i++ {
		delay = time.Duration(float64(delay) * config.BackoffFactor)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		jitter := time.Duration(float64(delay) * config.JitterFactor)
		delay += time.Duration(math.Floor(float64(jitter) * (math.Floor(rand.Float64()*2) - 1)))
	}
	return delay
}
