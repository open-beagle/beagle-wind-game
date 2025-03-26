package service

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
	"github.com/stretchr/testify/assert"
)

// 测试数据
var testInstance = models.GameInstance{
	ID:         "test-instance-1",
	NodeID:     "test-node-1",
	CardID:     "test-card-1",
	PlatformID: "test-platform-1",
	Status:     "stopped",
	CreatedAt:  time.Now(),
	UpdatedAt:  time.Now(),
}

func TestListInstances(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	instanceStore, err := store.NewInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer instanceStore.Cleanup()

	service := NewInstanceService(instanceStore)

	tests := []struct {
		name           string
		params         InstanceListParams
		setup          func()
		expectedResult InstanceListResult
		expectedError  error
	}{
		{
			name: "成功获取实例列表",
			params: InstanceListParams{
				Page:     1,
				PageSize: 20,
			},
			setup: func() {
				err := instanceStore.Add(testInstance)
				assert.NoError(t, err)
			},
			expectedResult: InstanceListResult{
				Total: 1,
				Items: []models.GameInstance{testInstance},
			},
			expectedError: nil,
		},
		{
			name: "存储层返回错误",
			params: InstanceListParams{
				Page:     1,
				PageSize: 20,
			},
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
				// 创建一个目录来替代文件，这样会导致读取错误
				err := os.MkdirAll(tmpFile, 0755)
				assert.NoError(t, err)
			},
			expectedResult: InstanceListResult{},
			expectedError:  fmt.Errorf("存储层错误"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result, err := service.ListInstances(tt.params)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Equal(t, tt.expectedResult, result)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult.Total, result.Total)
			assert.Equal(t, len(tt.expectedResult.Items), len(result.Items))
			for i, expected := range tt.expectedResult.Items {
				assert.Equal(t, expected.ID, result.Items[i].ID)
				assert.Equal(t, expected.NodeID, result.Items[i].NodeID)
				assert.Equal(t, expected.CardID, result.Items[i].CardID)
				assert.Equal(t, expected.Status, result.Items[i].Status)
			}
		})
	}
}

func TestGetInstance(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	instanceStore, err := store.NewInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer instanceStore.Cleanup()

	service := NewInstanceService(instanceStore)

	tests := []struct {
		name           string
		instanceID     string
		setup          func()
		expectedResult models.GameInstance
		expectedError  error
	}{
		{
			name:       "成功获取实例",
			instanceID: "test-instance-1",
			setup: func() {
				err := instanceStore.Add(testInstance)
				assert.NoError(t, err)
			},
			expectedResult: testInstance,
			expectedError:  nil,
		},
		{
			name:       "实例不存在",
			instanceID: "non-existent-instance",
			setup: func() {
				err := instanceStore.Add(testInstance)
				assert.NoError(t, err)
			},
			expectedResult: models.GameInstance{},
			expectedError:  fmt.Errorf("实例不存在: non-existent-instance"),
		},
		{
			name:       "存储层返回错误",
			instanceID: "test-instance-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
				// 创建一个目录来替代文件，这样会导致读取错误
				err := os.MkdirAll(tmpFile, 0755)
				assert.NoError(t, err)
			},
			expectedResult: models.GameInstance{},
			expectedError:  fmt.Errorf("存储层错误"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result, err := service.GetInstance(tt.instanceID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Equal(t, tt.expectedResult, result)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult.ID, result.ID)
			assert.Equal(t, tt.expectedResult.NodeID, result.NodeID)
			assert.Equal(t, tt.expectedResult.CardID, result.CardID)
			assert.Equal(t, tt.expectedResult.Status, result.Status)
		})
	}
}

func TestStartInstance(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	instanceStore, err := store.NewInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer instanceStore.Cleanup()

	service := NewInstanceService(instanceStore)

	tests := []struct {
		name          string
		instanceID    string
		setup         func()
		expectedError error
	}{
		{
			name:       "成功启动实例",
			instanceID: "test-instance-1",
			setup: func() {
				err := instanceStore.Add(testInstance)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "non-existent-instance",
			setup: func() {
				err := instanceStore.Add(testInstance)
				assert.NoError(t, err)
			},
			expectedError: fmt.Errorf("实例不存在: non-existent-instance"),
		},
		{
			name:       "存储层返回错误",
			instanceID: "test-instance-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
				// 创建一个目录来替代文件，这样会导致读取错误
				err := os.MkdirAll(tmpFile, 0755)
				assert.NoError(t, err)
			},
			expectedError: fmt.Errorf("存储层错误"),
		},
		{
			name:       "实例已在运行中",
			instanceID: "test-instance-1",
			setup: func() {
				runningInstance := testInstance
				runningInstance.Status = "running"
				err := instanceStore.Add(runningInstance)
				assert.NoError(t, err)
			},
			expectedError: ErrInstanceAlreadyRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.StartInstance(tt.instanceID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestStopInstance(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	instanceStore, err := store.NewInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer instanceStore.Cleanup()

	service := NewInstanceService(instanceStore)

	tests := []struct {
		name          string
		instanceID    string
		setup         func()
		expectedError error
	}{
		{
			name:       "成功停止实例",
			instanceID: "test-instance-1",
			setup: func() {
				runningInstance := testInstance
				runningInstance.Status = "running"
				err := instanceStore.Add(runningInstance)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "non-existent-instance",
			setup: func() {
				err := instanceStore.Add(testInstance)
				assert.NoError(t, err)
			},
			expectedError: fmt.Errorf("实例不存在: non-existent-instance"),
		},
		{
			name:       "存储层返回错误",
			instanceID: "test-instance-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
				// 创建一个目录来替代文件，这样会导致读取错误
				err := os.MkdirAll(tmpFile, 0755)
				assert.NoError(t, err)
			},
			expectedError: fmt.Errorf("存储层错误"),
		},
		{
			name:       "实例未在运行中",
			instanceID: "test-instance-1",
			setup: func() {
				stoppedInstance := testInstance
				stoppedInstance.Status = "stopped"
				err := instanceStore.Add(stoppedInstance)
				assert.NoError(t, err)
			},
			expectedError: ErrInstanceNotRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.StopInstance(tt.instanceID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestCreateInstance(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	instanceStore, err := store.NewInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer instanceStore.Cleanup()

	service := NewInstanceService(instanceStore)

	tests := []struct {
		name          string
		params        CreateInstanceParams
		setup         func()
		expectedID    string
		expectedError error
	}{
		{
			name: "成功创建实例",
			params: CreateInstanceParams{
				NodeID:     "test-node-1",
				PlatformID: "test-platform-1",
				CardID:     "test-card-1",
				Config:     "{}",
			},
			setup: func() {
				// 不需要特殊设置
			},
			expectedID:    "inst-test-node-1-test-card-1-",
			expectedError: nil,
		},
		{
			name: "存储层返回错误",
			params: CreateInstanceParams{
				NodeID:     "test-node-1",
				PlatformID: "test-platform-1",
				CardID:     "test-card-1",
				Config:     "{}",
			},
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedID:    "",
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			id, err := service.CreateInstance(tt.params)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Empty(t, id)
				return
			}
			assert.NoError(t, err)
			assert.Contains(t, id, tt.expectedID)
		})
	}
}

func TestUpdateInstance(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	instanceStore, err := store.NewInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer instanceStore.Cleanup()

	service := NewInstanceService(instanceStore)

	tests := []struct {
		name          string
		instanceID    string
		instance      models.GameInstance
		setup         func()
		expectedError error
	}{
		{
			name:       "成功更新实例",
			instanceID: "test-instance-1",
			instance:   testInstance,
			setup: func() {
				err := instanceStore.Add(testInstance)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "non-existent-instance",
			instance:   testInstance,
			setup: func() {
				err := instanceStore.Add(testInstance)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:       "存储层返回错误",
			instanceID: "test-instance-1",
			instance:   testInstance,
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.UpdateInstance(tt.instanceID, tt.instance)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestDeleteInstance(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	instanceStore, err := store.NewInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer instanceStore.Cleanup()

	service := NewInstanceService(instanceStore)

	tests := []struct {
		name          string
		instanceID    string
		setup         func()
		expectedError error
	}{
		{
			name:       "成功删除实例",
			instanceID: "test-instance-1",
			setup: func() {
				err := instanceStore.Add(testInstance)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:       "存储层返回错误",
			instanceID: "test-instance-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.DeleteInstance(tt.instanceID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
