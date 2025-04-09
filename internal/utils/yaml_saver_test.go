package utils

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 模拟Logger接口，用于测试
type mockLogger struct{}

func (m *mockLogger) Debug(format string, args ...interface{})        {}
func (m *mockLogger) Info(format string, args ...interface{})         {}
func (m *mockLogger) Warn(format string, args ...interface{})         {}
func (m *mockLogger) Error(format string, args ...interface{})        {}
func (m *mockLogger) Fatal(format string, args ...interface{})        {}
func (m *mockLogger) Sync() error                                     { return nil }
func (m *mockLogger) WithField(key string, value interface{}) Logger  { return m }
func (m *mockLogger) WithFields(fields map[string]interface{}) Logger { return m }
func (m *mockLogger) WithRequestID(requestID string) Logger           { return m }

func TestYAMLSaver_BasicSave(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "yaml-saver-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// 测试数据
	testData := []string{"item1", "item2", "item3"}

	// 创建YAMLSaver，设置0延迟以立即保存
	saver := NewYAMLSaver(
		tempFile.Name(),
		func() interface{} {
			return testData
		},
		&mockLogger{},
		WithDelay(0),
	)

	// 保存数据
	err = saver.Save(context.Background())
	require.NoError(t, err)

	// 验证文件内容
	data, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)
	assert.Contains(t, string(data), "item1")
	assert.Contains(t, string(data), "item2")
	assert.Contains(t, string(data), "item3")
}

func TestYAMLSaver_DelayedSave(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "yaml-saver-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// 测试数据
	var data []string
	var dataLock sync.Mutex

	// 创建YAMLSaver，设置200毫秒延迟
	saver := NewYAMLSaver(
		tempFile.Name(),
		func() interface{} {
			dataLock.Lock()
			defer dataLock.Unlock()
			return data
		},
		&mockLogger{},
		WithDelay(200*time.Millisecond),
	)

	// 第一次保存，数据为空
	dataLock.Lock()
	data = []string{}
	dataLock.Unlock()
	err = saver.Save(context.Background())
	require.NoError(t, err)

	// 快速添加项目并保存多次
	for i := 0; i < 10; i++ {
		dataLock.Lock()
		data = append(data, "item")
		dataLock.Unlock()
		err = saver.Save(context.Background())
		require.NoError(t, err)
	}

	// 最终添加最重要的项目
	dataLock.Lock()
	data = append(data, "final-item")
	dataLock.Unlock()
	err = saver.Save(context.Background())
	require.NoError(t, err)

	// 等待保存完成
	time.Sleep(300 * time.Millisecond)

	// 验证文件内容，应该包含最终的数据
	fileData, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)
	assert.Contains(t, string(fileData), "final-item")

	// 验证保存次数（通过检查文件修改时间可以大致推断）
	fileInfo, err := os.Stat(tempFile.Name())
	require.NoError(t, err)
	modTime := fileInfo.ModTime()

	// 确保修改时间在合理范围内（保存应该在延迟时间后完成）
	timeSinceModification := time.Since(modTime)
	assert.Less(t, timeSinceModification, 300*time.Millisecond)
}

func TestYAMLSaver_Flush(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "yaml-saver-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// 测试数据
	testData := []string{"flush-test"}

	// 创建YAMLSaver，设置很长的延迟（10秒）
	saver := NewYAMLSaver(
		tempFile.Name(),
		func() interface{} {
			return testData
		},
		&mockLogger{},
		WithDelay(10*time.Second), // 设置很长的延迟
	)

	// 保存数据
	err = saver.Save(context.Background())
	require.NoError(t, err)

	// 检查文件内容，应该为空（因为延迟尚未到期）
	_, err = os.Stat(tempFile.Name())
	assert.True(t, os.IsNotExist(err), "文件不应该存在，因为延迟保存尚未执行")

	// 立即刷新
	err = saver.SaveNow(context.Background())
	require.NoError(t, err)

	// 验证文件内容，应该已保存
	data, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)
	assert.Contains(t, string(data), "flush-test")
}

func TestYAMLSaver_AtomicWrite(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "yaml-saver-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// 初始数据
	initialData := []string{"initial"}

	// 写入初始数据
	saver := NewYAMLSaver(
		tempFile.Name(),
		func() interface{} {
			return initialData
		},
		&mockLogger{},
	)
	err = saver.SaveNow(context.Background())
	require.NoError(t, err)

	// 验证初始数据
	data, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)
	assert.Contains(t, string(data), "initial")

	// 更新数据
	finalData := []string{"final"}
	saver = NewYAMLSaver(
		tempFile.Name(),
		func() interface{} {
			return finalData
		},
		&mockLogger{},
	)

	// 保存更新后的数据
	err = saver.SaveNow(context.Background())
	require.NoError(t, err)

	// 验证更新后的数据
	data, err = os.ReadFile(tempFile.Name())
	require.NoError(t, err)
	assert.Contains(t, string(data), "final")
	assert.NotContains(t, string(data), "initial")
}
