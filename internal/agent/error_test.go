package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWrapError 测试错误包装
func TestWrapError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retriable bool
		wantErr   bool
	}{
		{
			name:      "可重试错误",
			err:       errors.New("网络错误"),
			retriable: true,
			wantErr:   true,
		},
		{
			name:      "不可重试错误",
			err:       errors.New("参数错误"),
			retriable: false,
			wantErr:   true,
		},
		{
			name:      "空错误",
			err:       nil,
			retriable: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrappedErr := WrapError(tt.err, tt.retriable)
			if tt.wantErr {
				assert.Error(t, wrappedErr)
				if tt.err != nil {
					assert.Contains(t, wrappedErr.Error(), tt.err.Error())
				}
				var retryErr *RetryableError
				assert.True(t, errors.As(wrappedErr, &retryErr))
				assert.Equal(t, tt.retriable, retryErr.Retry)
			} else {
				assert.NoError(t, wrappedErr)
			}
		})
	}
}

// TestRetry 测试重试机制
func TestRetry(t *testing.T) {
	tests := []struct {
		name      string
		fn        func() error
		config    RetryConfig
		wantErr   bool
		wantCount int
	}{
		{
			name: "成功不需要重试",
			fn: func() error {
				return nil
			},
			config:    DefaultRetryConfig,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "可重试错误最终成功",
			fn: func() error {
				static := 0
				return func() error {
					static++
					if static < 3 {
						return WrapError(errors.New("临时错误"), true)
					}
					return nil
				}()
			},
			config:    DefaultRetryConfig,
			wantErr:   false,
			wantCount: 3,
		},
		{
			name: "不可重试错误立即失败",
			fn: func() error {
				return WrapError(errors.New("永久错误"), false)
			},
			config:    DefaultRetryConfig,
			wantErr:   true,
			wantCount: 1,
		},
		{
			name: "可重试错误超过最大重试次数",
			fn: func() error {
				return WrapError(errors.New("持续错误"), true)
			},
			config: RetryConfig{
				MaxRetries:    3,
				InitialDelay:  10,
				MaxDelay:      100,
				BackoffFactor: 2,
				JitterFactor:  0.2,
			},
			wantErr:   true,
			wantCount: 4, // 初始尝试 + 3次重试
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			wrappedFn := func() error {
				count++
				return tt.fn()
			}

			err := Retry(context.Background(), wrappedFn, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantCount, count)
		})
	}
}
