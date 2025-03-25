package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListInstances(t *testing.T) {
	mockInstanceStore := new(store.MockInstanceStore)
	mockNodeStore := new(store.MockNodeStore)
	mockCardStore := new(store.MockGameCardStore)
	mockPlatformStore := new(store.MockPlatformStore)
	service := NewInstanceService(mockInstanceStore, mockNodeStore, mockCardStore, mockPlatformStore)

	tests := []struct {
		name           string
		params         InstanceListParams
		mockSetup      func()
		expectedResult InstanceListResult
		expectedError  error
	}{
		{
			name: "成功获取实例列表",
			params: InstanceListParams{
				Page:     1,
				PageSize: 20,
			},
			mockSetup: func() {
				instances := []models.GameInstance{
					{
						ID:         "instance-1",
						NodeID:     "node-1",
						PlatformID: "platform-1",
						CardID:     "card-1",
						Status:     "running",
						Resources:  "{}",
						Config:     "{}",
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
						StartedAt:  time.Now(),
					},
				}
				mockInstanceStore.On("List").Return(instances, nil)
			},
			expectedResult: InstanceListResult{
				Total: 1,
				Items: []models.GameInstance{
					{
						ID:         "instance-1",
						NodeID:     "node-1",
						PlatformID: "platform-1",
						CardID:     "card-1",
						Status:     "running",
						Resources:  "{}",
						Config:     "{}",
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "存储层返回错误",
			params: InstanceListParams{
				Page:     1,
				PageSize: 20,
			},
			mockSetup: func() {
				mockInstanceStore.On("List").Return(nil, assert.AnError)
			},
			expectedResult: InstanceListResult{},
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.ListInstances(tt.params)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult.Total, result.Total)
			assert.Equal(t, len(tt.expectedResult.Items), len(result.Items))
			for i, expected := range tt.expectedResult.Items {
				assert.Equal(t, expected.ID, result.Items[i].ID)
				assert.Equal(t, expected.NodeID, result.Items[i].NodeID)
				assert.Equal(t, expected.PlatformID, result.Items[i].PlatformID)
				assert.Equal(t, expected.CardID, result.Items[i].CardID)
				assert.Equal(t, expected.Status, result.Items[i].Status)
				assert.Equal(t, expected.Resources, result.Items[i].Resources)
				assert.Equal(t, expected.Config, result.Items[i].Config)
			}
		})
	}
}

func TestGetInstance(t *testing.T) {
	mockInstanceStore := new(store.MockInstanceStore)
	mockNodeStore := new(store.MockNodeStore)
	mockCardStore := new(store.MockGameCardStore)
	mockPlatformStore := new(store.MockPlatformStore)
	service := NewInstanceService(mockInstanceStore, mockNodeStore, mockCardStore, mockPlatformStore)

	tests := []struct {
		name           string
		instanceID     string
		mockSetup      func()
		expectedResult models.GameInstance
		expectedError  error
	}{
		{
			name:       "成功获取实例",
			instanceID: "instance-1",
			mockSetup: func() {
				instance := models.GameInstance{
					ID:         "instance-1",
					NodeID:     "node-1",
					PlatformID: "platform-1",
					CardID:     "card-1",
					Status:     "running",
					Resources:  "{}",
					Config:     "{}",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
					StartedAt:  time.Now(),
				}
				mockInstanceStore.On("Get", "instance-1").Return(instance, nil)
			},
			expectedResult: models.GameInstance{
				ID:         "instance-1",
				NodeID:     "node-1",
				PlatformID: "platform-1",
				CardID:     "card-1",
				Status:     "running",
				Resources:  "{}",
				Config:     "{}",
			},
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "instance-2",
			mockSetup: func() {
				mockInstanceStore.On("Get", "instance-2").Return(models.GameInstance{}, nil)
			},
			expectedResult: models.GameInstance{},
			expectedError:  nil,
		},
		{
			name:       "存储层返回错误",
			instanceID: "instance-3",
			mockSetup: func() {
				mockInstanceStore.On("Get", "instance-3").Return(models.GameInstance{}, assert.AnError)
			},
			expectedResult: models.GameInstance{},
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.GetInstance(tt.instanceID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult.ID, result.ID)
			assert.Equal(t, tt.expectedResult.NodeID, result.NodeID)
			assert.Equal(t, tt.expectedResult.PlatformID, result.PlatformID)
			assert.Equal(t, tt.expectedResult.CardID, result.CardID)
			assert.Equal(t, tt.expectedResult.Status, result.Status)
			assert.Equal(t, tt.expectedResult.Resources, result.Resources)
			assert.Equal(t, tt.expectedResult.Config, result.Config)
		})
	}
}

func TestStartInstance(t *testing.T) {
	mockInstanceStore := new(store.MockInstanceStore)
	mockNodeStore := new(store.MockNodeStore)
	mockCardStore := new(store.MockGameCardStore)
	mockPlatformStore := new(store.MockPlatformStore)
	service := NewInstanceService(mockInstanceStore, mockNodeStore, mockCardStore, mockPlatformStore)

	tests := []struct {
		name          string
		instanceID    string
		mockSetup     func()
		expectedError error
	}{
		{
			name:       "成功启动实例",
			instanceID: "instance-1",
			mockSetup: func() {
				instance := models.GameInstance{
					ID:         "instance-1",
					NodeID:     "node-1",
					PlatformID: "platform-1",
					CardID:     "card-1",
					Status:     "created",
					Resources:  "{}",
					Config:     "{}",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				node := models.GameNode{
					ID: "node-1",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateReady,
						Online:     true,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockInstanceStore.On("Get", "instance-1").Return(instance, nil)
				mockNodeStore.On("Get", "node-1").Return(node, nil)
				mockPlatformStore.On("Get", "platform-1").Return(models.GamePlatform{}, nil)
				mockInstanceStore.On("Update", mock.AnythingOfType("models.GameInstance")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "instance-2",
			mockSetup: func() {
				mockInstanceStore.On("Get", "instance-2").Return(models.GameInstance{}, ErrInstanceNotFound)
			},
			expectedError: ErrInstanceNotFound,
		},
		{
			name:       "实例状态不允许启动",
			instanceID: "instance-3",
			mockSetup: func() {
				instance := models.GameInstance{
					ID:         "instance-3",
					NodeID:     "node-1",
					PlatformID: "platform-1",
					CardID:     "card-1",
					Status:     "running",
					Resources:  "{}",
					Config:     "{}",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
					StartedAt:  time.Now(),
				}
				mockInstanceStore.On("Get", "instance-3").Return(instance, nil)
			},
			expectedError: ErrInstanceAlreadyRunning,
		},
		{
			name:       "节点不存在",
			instanceID: "instance-4",
			mockSetup: func() {
				instance := models.GameInstance{
					ID:         "instance-4",
					NodeID:     "node-2",
					PlatformID: "platform-1",
					CardID:     "card-1",
					Status:     "created",
					Resources:  "{}",
					Config:     "{}",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				mockInstanceStore.On("Get", "instance-4").Return(instance, nil)
				mockNodeStore.On("Get", "node-2").Return(models.GameNode{}, ErrNodeNotFound)
			},
			expectedError: ErrNodeNotFound,
		},
		{
			name:       "节点状态不正确",
			instanceID: "instance-5",
			mockSetup: func() {
				instance := models.GameInstance{
					ID:         "instance-5",
					NodeID:     "node-3",
					PlatformID: "platform-1",
					CardID:     "card-1",
					Status:     "created",
					Resources:  "{}",
					Config:     "{}",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				node := models.GameNode{
					ID: "node-3",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateOffline,
						Online:     false,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockInstanceStore.On("Get", "instance-5").Return(instance, nil)
				mockNodeStore.On("Get", "node-3").Return(node, nil)
			},
			expectedError: ErrNodeNotReady,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.StartInstance(tt.instanceID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestStopInstance(t *testing.T) {
	mockInstanceStore := new(store.MockInstanceStore)
	mockNodeStore := new(store.MockNodeStore)
	mockCardStore := new(store.MockGameCardStore)
	mockPlatformStore := new(store.MockPlatformStore)
	service := NewInstanceService(mockInstanceStore, mockNodeStore, mockCardStore, mockPlatformStore)

	tests := []struct {
		name          string
		instanceID    string
		mockSetup     func()
		expectedError error
	}{
		{
			name:       "成功停止实例",
			instanceID: "instance-1",
			mockSetup: func() {
				instance := models.GameInstance{
					ID:         "instance-1",
					NodeID:     "node-1",
					PlatformID: "platform-1",
					CardID:     "card-1",
					Status:     "running",
					Resources:  "{}",
					Config:     "{}",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
					StartedAt:  time.Now(),
				}
				mockInstanceStore.On("Get", "instance-1").Return(instance, nil)
				mockInstanceStore.On("Update", mock.AnythingOfType("models.GameInstance")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "instance-2",
			mockSetup: func() {
				mockInstanceStore.On("Get", "instance-2").Return(models.GameInstance{}, ErrInstanceNotFound)
			},
			expectedError: fmt.Errorf("获取实例失败: %w", ErrInstanceNotFound),
		},
		{
			name:       "实例状态不允许停止",
			instanceID: "instance-3",
			mockSetup: func() {
				instance := models.GameInstance{
					ID:         "instance-3",
					NodeID:     "node-1",
					PlatformID: "platform-1",
					CardID:     "card-1",
					Status:     "stopped",
					Resources:  "{}",
					Config:     "{}",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
					StartedAt:  time.Now(),
				}
				mockInstanceStore.On("Get", "instance-3").Return(instance, nil)
			},
			expectedError: ErrInstanceNotRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
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
	mockInstanceStore := new(store.MockInstanceStore)
	mockNodeStore := new(store.MockNodeStore)
	mockCardStore := new(store.MockGameCardStore)
	mockPlatformStore := new(store.MockPlatformStore)
	service := NewInstanceService(mockInstanceStore, mockNodeStore, mockCardStore, mockPlatformStore)

	tests := []struct {
		name           string
		params         CreateInstanceParams
		mockSetup      func()
		expectedResult string
		expectedError  error
	}{
		{
			name: "成功创建实例",
			params: CreateInstanceParams{
				NodeID:     "node-1",
				PlatformID: "platform-1",
				CardID:     "card-1",
				Config:     "{}",
			},
			mockSetup: func() {
				node := models.GameNode{
					ID: "node-1",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateReady,
						Online:     true,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				platform := models.GamePlatform{
					ID:   "platform-1",
					Type: "game",
				}
				card := models.GameCard{
					ID:         "card-1",
					PlatformID: "platform-1",
				}

				mockNodeStore.On("Get", "node-1").Return(node, nil)
				mockPlatformStore.On("Get", "platform-1").Return(platform, nil)
				mockCardStore.On("Get", "card-1").Return(card, nil)
				mockInstanceStore.On("Add", mock.AnythingOfType("models.GameInstance")).Return(nil)
			},
			expectedResult: "inst-node-1-card-1-",
			expectedError:  nil,
		},
		{
			name: "节点不存在",
			params: CreateInstanceParams{
				NodeID:     "node-2",
				PlatformID: "platform-1",
				CardID:     "card-1",
				Config:     "{}",
			},
			mockSetup: func() {
				mockNodeStore.On("Get", "node-2").Return(models.GameNode{}, ErrNodeNotFound)
			},
			expectedResult: "",
			expectedError:  fmt.Errorf("节点不存在: %w", ErrNodeNotFound),
		},
		{
			name: "节点状态不正确",
			params: CreateInstanceParams{
				NodeID:     "node-3",
				PlatformID: "platform-1",
				CardID:     "card-1",
				Config:     "{}",
			},
			mockSetup: func() {
				node := models.GameNode{
					ID: "node-3",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateOffline,
						Online:     false,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockNodeStore.On("Get", "node-3").Return(node, nil)
			},
			expectedResult: "",
			expectedError:  fmt.Errorf("节点状态不正确: %s", models.GameNodeStateOffline),
		},
		{
			name: "平台不存在",
			params: CreateInstanceParams{
				NodeID:     "node-1",
				PlatformID: "platform-2",
				CardID:     "card-1",
				Config:     "{}",
			},
			mockSetup: func() {
				node := models.GameNode{
					ID: "node-1",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateReady,
						Online:     true,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockNodeStore.On("Get", "node-1").Return(node, nil)
				mockPlatformStore.On("Get", "platform-2").Return(models.GamePlatform{}, fmt.Errorf("平台不存在: platform-2"))
			},
			expectedResult: "",
			expectedError:  fmt.Errorf("平台不存在: 平台不存在: platform-2"),
		},
		{
			name: "游戏卡片不存在",
			params: CreateInstanceParams{
				NodeID:     "node-1",
				PlatformID: "platform-1",
				CardID:     "card-2",
				Config:     "{}",
			},
			mockSetup: func() {
				node := models.GameNode{
					ID: "node-1",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateReady,
						Online:     true,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				platform := models.GamePlatform{
					ID:   "platform-1",
					Type: "game",
				}
				mockNodeStore.On("Get", "node-1").Return(node, nil)
				mockPlatformStore.On("Get", "platform-1").Return(platform, nil)
				mockCardStore.On("Get", "card-2").Return(models.GameCard{}, fmt.Errorf("游戏卡片不存在: card-2"))
			},
			expectedResult: "",
			expectedError:  fmt.Errorf("游戏卡片不存在: 游戏卡片不存在: card-2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.CreateInstance(tt.params)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Contains(t, result, tt.expectedResult)
		})
	}
}

func TestUpdateInstance(t *testing.T) {
	mockInstanceStore := new(store.MockInstanceStore)
	mockNodeStore := new(store.MockNodeStore)
	mockCardStore := new(store.MockGameCardStore)
	mockPlatformStore := new(store.MockPlatformStore)
	service := NewInstanceService(mockInstanceStore, mockNodeStore, mockCardStore, mockPlatformStore)

	tests := []struct {
		name          string
		instanceID    string
		params        UpdateInstanceParams
		mockSetup     func()
		expectedError error
	}{
		{
			name:       "成功更新实例",
			instanceID: "instance-1",
			params: UpdateInstanceParams{
				Status:      "running",
				Resources:   "{}",
				Performance: "{}",
				SaveData:    "{}",
				Config:      "{}",
				Backup:      "{}",
			},
			mockSetup: func() {
				instance := models.GameInstance{
					ID:         "instance-1",
					NodeID:     "node-1",
					PlatformID: "platform-1",
					CardID:     "card-1",
					Status:     "starting",
					Resources:  "{}",
					Config:     "{}",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				mockInstanceStore.On("Get", "instance-1").Return(instance, nil)
				mockInstanceStore.On("Update", mock.AnythingOfType("models.GameInstance")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "instance-2",
			params: UpdateInstanceParams{
				Status: "running",
			},
			mockSetup: func() {
				mockInstanceStore.On("Get", "instance-2").Return(models.GameInstance{}, ErrInstanceNotFound)
			},
			expectedError: fmt.Errorf("获取实例失败: %w", ErrInstanceNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.UpdateInstance(tt.instanceID, tt.params)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestDeleteInstance(t *testing.T) {
	mockInstanceStore := new(store.MockInstanceStore)
	mockNodeStore := new(store.MockNodeStore)
	mockCardStore := new(store.MockGameCardStore)
	mockPlatformStore := new(store.MockPlatformStore)
	service := NewInstanceService(mockInstanceStore, mockNodeStore, mockCardStore, mockPlatformStore)

	tests := []struct {
		name          string
		instanceID    string
		mockSetup     func()
		expectedError error
	}{
		{
			name:       "成功删除实例",
			instanceID: "instance-1",
			mockSetup: func() {
				instance := models.GameInstance{
					ID:         "instance-1",
					NodeID:     "node-1",
					PlatformID: "platform-1",
					CardID:     "card-1",
					Status:     "stopped",
					Resources:  "{}",
					Config:     "{}",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				mockInstanceStore.On("Get", "instance-1").Return(instance, nil)
				mockInstanceStore.On("Delete", "instance-1").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "instance-2",
			mockSetup: func() {
				mockInstanceStore.On("Get", "instance-2").Return(models.GameInstance{}, ErrInstanceNotFound)
			},
			expectedError: fmt.Errorf("获取实例失败: %w", ErrInstanceNotFound),
		},
		{
			name:       "实例正在运行",
			instanceID: "instance-3",
			mockSetup: func() {
				instance := models.GameInstance{
					ID:         "instance-3",
					NodeID:     "node-1",
					PlatformID: "platform-1",
					CardID:     "card-1",
					Status:     "running",
					Resources:  "{}",
					Config:     "{}",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
					StartedAt:  time.Now(),
				}
				mockInstanceStore.On("Get", "instance-3").Return(instance, nil)
			},
			expectedError: fmt.Errorf("实例正在运行中，无法删除"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.DeleteInstance(tt.instanceID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
