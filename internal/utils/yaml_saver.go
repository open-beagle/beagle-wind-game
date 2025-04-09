package utils

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLSaver YAML文件保存器，提供延迟保存功能
type YAMLSaver struct {
	// 数据文件路径
	dataFile string
	// 数据提供函数，返回需要序列化和保存的数据
	dataProviderFn func() interface{}
	// 延迟保存器
	delayedSaver *DelayedSaver
	// 互斥锁，保护数据访问
	mu sync.RWMutex
	// 日志器
	logger Logger
	// 文件模式
	fileMode os.FileMode
}

// YAMLSaverOption 定义YAMLSaver的选项
type YAMLSaverOption func(*YAMLSaver)

// WithDelay 设置延迟保存的时间
func WithDelay(delay time.Duration) YAMLSaverOption {
	return func(s *YAMLSaver) {
		s.delayedSaver = NewDelayedSaver(s.doSave, delay, s.logger)
	}
}

// WithFileMode 设置保存文件的权限模式
func WithFileMode(mode os.FileMode) YAMLSaverOption {
	return func(s *YAMLSaver) {
		s.fileMode = mode
	}
}

// NewYAMLSaver 创建一个新的YAML保存器
// dataFile: 保存的文件路径
// dataProviderFn: 提供需要保存的数据的函数
// logger: 日志器
// options: 可选配置项
func NewYAMLSaver(dataFile string, dataProviderFn func() interface{}, logger Logger, options ...YAMLSaverOption) *YAMLSaver {
	saver := &YAMLSaver{
		dataFile:       dataFile,
		dataProviderFn: dataProviderFn,
		logger:         logger,
		fileMode:       0644, // 默认文件权限
	}

	// 默认使用1秒的延迟
	saver.delayedSaver = NewDelayedSaver(saver.doSave, time.Second, logger)

	// 应用选项
	for _, option := range options {
		option(saver)
	}

	return saver
}

// Save 触发保存操作，使用延迟机制
func (s *YAMLSaver) Save(ctx context.Context) error {
	return s.delayedSaver.Save(ctx)
}

// SaveNow 立即保存，不使用延迟
func (s *YAMLSaver) SaveNow(ctx context.Context) error {
	return s.delayedSaver.Flush(ctx)
}

// doSave 实际执行保存操作
func (s *YAMLSaver) doSave(ctx context.Context) error {
	// 检查上下文是否已取消
	if err := ctx.Err(); err != nil {
		return err
	}

	s.logger.Debug("正在保存数据到文件: %s", s.dataFile)

	// 获取数据
	data := s.dataProviderFn()

	// 序列化为YAML
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		s.logger.Error("序列化YAML失败: %v", err)
		return fmt.Errorf("序列化YAML失败: %w", err)
	}

	// 写入临时文件，然后重命名，确保原子性
	tempFile := s.dataFile + ".tmp"

	// 写入临时文件
	if err := os.WriteFile(tempFile, yamlData, s.fileMode); err != nil {
		s.logger.Error("写入临时文件失败: %v", err)
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	// 重命名临时文件，确保原子性写入
	if err := os.Rename(tempFile, s.dataFile); err != nil {
		s.logger.Error("重命名文件失败: %v", err)
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	s.logger.Debug("成功保存数据到文件: %s", s.dataFile)
	return nil
}

// Close 关闭保存器，确保所有待处理的保存操作完成
func (s *YAMLSaver) Close() {
	s.delayedSaver.Close()
}
