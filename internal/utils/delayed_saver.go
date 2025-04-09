package utils

import (
	"context"
	"sync"
	"time"
)

// SaveFn 定义一个保存函数的类型
type SaveFn func(ctx context.Context) error

// DelayedSaver 延迟保存器，用于合并短时间内的多次保存请求
type DelayedSaver struct {
	// 实际的保存函数
	saveFn SaveFn
	// 延迟时间
	delay time.Duration
	// 锁，保护内部状态
	mu sync.Mutex
	// 计时器，用于延迟保存
	timer *time.Timer
	// 是否有等待的保存请求
	pending bool
	// 日志器
	logger Logger
	// 上下文，用于取消保存
	ctx    context.Context
	cancel context.CancelFunc
}

// NewDelayedSaver 创建一个新的延迟保存器
// saveFn: 实际的保存函数
// delay: 延迟时间，在此时间内的多次调用会被合并
// logger: 日志器
func NewDelayedSaver(saveFn SaveFn, delay time.Duration, logger Logger) *DelayedSaver {
	ctx, cancel := context.WithCancel(context.Background())
	return &DelayedSaver{
		saveFn: saveFn,
		delay:  delay,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Save 触发一次保存请求
// ctx: 上下文，用于取消保存
func (s *DelayedSaver) Save(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	// 如果已经有等待的保存请求，只需设置pending为true
	if s.timer != nil {
		s.logger.Debug("已有等待的保存请求，将在%v后执行", s.delay)
		s.pending = true
		return nil
	}

	// 创建一个新的计时器，延迟执行保存
	s.timer = time.AfterFunc(s.delay, func() {
		s.executeDelayedSave(ctx)
	})

	s.logger.Debug("计划在%v后执行保存", s.delay)
	return nil
}

// 实际执行保存的内部方法
func (s *DelayedSaver) executeDelayedSave(ctx context.Context) {
	s.mu.Lock()

	// 重置计时器
	s.timer = nil

	// 检查是否有等待的保存请求
	pending := s.pending
	s.pending = false

	s.mu.Unlock()

	// 执行保存
	err := s.saveFn(ctx)
	if err != nil {
		s.logger.Error("延迟保存执行失败: %v", err)
	} else {
		s.logger.Debug("延迟保存执行成功")
	}

	// 如果在保存执行期间又有新的保存请求，则再次触发保存
	if pending {
		s.mu.Lock()
		if s.timer == nil {
			s.timer = time.AfterFunc(s.delay, func() {
				s.executeDelayedSave(ctx)
			})
			s.logger.Debug("检测到新的保存请求，计划在%v后再次执行", s.delay)
		}
		s.mu.Unlock()
	}
}

// Flush 立即执行保存，忽略延迟
func (s *DelayedSaver) Flush(ctx context.Context) error {
	s.mu.Lock()

	// 如果有计时器，取消它
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}

	// 重置pending状态
	s.pending = false

	s.mu.Unlock()

	// 执行实际的保存操作
	s.logger.Debug("正在执行立即保存")
	return s.saveFn(ctx)
}

// Close 关闭延迟保存器，取消所有未执行的保存
func (s *DelayedSaver) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 取消上下文
	s.cancel()

	// 停止计时器
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}

	s.logger.Debug("延迟保存器已关闭")
}
