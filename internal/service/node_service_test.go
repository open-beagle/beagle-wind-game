package service

import (
	"testing"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListNodes(t *testing.T) {
	mockNodeStore := new(store.MockNodeStore)
	mockInstanceStore := new(store.MockInstanceStore)
	service := NewNodeService(mockNodeStore, mockInstanceStore)

	tests := []struct {
		name           string
		params         NodeListParams
		mockSetup      func()
		expectedResult NodeListResult
		expectedError  error
	}{
		{
			name: "成功获取节点列表",
			params: NodeListParams{
				Page:     1,
				PageSize: 20,
			},
			mockSetup: func() {
				nodes := []models.GameNode{
					{
						ID:   "node-1",
						Name: "测试节点1",
						Status: models.GameNodeStatus{
							State:      models.GameNodeStateReady,
							Online:     true,
							LastOnline: time.Now(),
							UpdatedAt:  time.Now(),
							Resources:  make(map[string]string),
							Metrics:    make(map[string]interface{}),
						},
					},
				}
				mockNodeStore.On("List").Return(nodes, nil)
			},
			expectedResult: NodeListResult{
				Total: 1,
				Items: []models.GameNode{
					{
						ID:   "node-1",
						Name: "测试节点1",
						Status: models.GameNodeStatus{
							State:      models.GameNodeStateReady,
							Online:     true,
							LastOnline: time.Now(),
							UpdatedAt:  time.Now(),
							Resources:  make(map[string]string),
							Metrics:    make(map[string]interface{}),
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "存储层返回错误",
			params: NodeListParams{
				Page:     1,
				PageSize: 20,
			},
			mockSetup: func() {
				mockNodeStore.On("List").Return(nil, assert.AnError)
			},
			expectedResult: NodeListResult{},
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.ListNodes(tt.params)
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
				assert.Equal(t, expected.Name, result.Items[i].Name)
				assert.Equal(t, expected.Status.State, result.Items[i].Status.State)
				assert.Equal(t, expected.Status.Online, result.Items[i].Status.Online)
			}
		})
	}
}

func TestGetNode(t *testing.T) {
	mockNodeStore := new(store.MockNodeStore)
	mockInstanceStore := new(store.MockInstanceStore)
	service := NewNodeService(mockNodeStore, mockInstanceStore)

	tests := []struct {
		name           string
		nodeID         string
		mockSetup      func()
		expectedResult models.GameNode
		expectedError  error
	}{
		{
			name:   "成功获取节点",
			nodeID: "node-1",
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-1",
					Name: "测试节点1",
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
			},
			expectedResult: models.GameNode{
				ID:   "node-1",
				Name: "测试节点1",
				Status: models.GameNodeStatus{
					State:      models.GameNodeStateReady,
					Online:     true,
					LastOnline: time.Now(),
					UpdatedAt:  time.Now(),
					Resources:  make(map[string]string),
					Metrics:    make(map[string]interface{}),
				},
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "node-2",
			mockSetup: func() {
				mockNodeStore.On("Get", "node-2").Return(models.GameNode{}, ErrNodeNotFound)
			},
			expectedResult: models.GameNode{},
			expectedError:  ErrNodeNotFound,
		},
		{
			name:   "存储层返回错误",
			nodeID: "node-3",
			mockSetup: func() {
				mockNodeStore.On("Get", "node-3").Return(models.GameNode{}, assert.AnError)
			},
			expectedResult: models.GameNode{},
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.GetNode(tt.nodeID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult.ID, result.ID)
			assert.Equal(t, tt.expectedResult.Name, result.Name)
			assert.Equal(t, tt.expectedResult.Status.State, result.Status.State)
			assert.Equal(t, tt.expectedResult.Status.Online, result.Status.Online)
		})
	}
}

func TestCreateNode(t *testing.T) {
	mockNodeStore := new(store.MockNodeStore)
	mockInstanceStore := new(store.MockInstanceStore)
	service := NewNodeService(mockNodeStore, mockInstanceStore)

	tests := []struct {
		name           string
		params         CreateNodeParams
		mockSetup      func()
		expectedResult string
		expectedError  error
	}{
		{
			name: "成功创建节点",
			params: CreateNodeParams{
				Name: "测试节点1",
				Type: "game",
			},
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-1",
					Name: "测试节点1",
					Type: "game",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateOffline,
						Online:     false,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockNodeStore.On("Add", mock.AnythingOfType("models.GameNode")).Return(nil)
			},
			expectedResult: "node-1",
			expectedError:  nil,
		},
		{
			name: "存储层返回错误",
			params: CreateNodeParams{
				Name: "测试节点2",
				Type: "game",
			},
			mockSetup: func() {
				mockNodeStore.On("Add", mock.AnythingOfType("models.GameNode")).Return(assert.AnError)
			},
			expectedResult: "",
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.CreateNode(tt.params)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "node-")
		})
	}
}

func TestUpdateNode(t *testing.T) {
	mockNodeStore := new(store.MockNodeStore)
	mockInstanceStore := new(store.MockInstanceStore)
	service := NewNodeService(mockNodeStore, mockInstanceStore)

	tests := []struct {
		name          string
		nodeID        string
		params        UpdateNodeParams
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "成功更新节点",
			nodeID: "node-1",
			params: UpdateNodeParams{
				Name: "更新后的节点",
				Type: "game",
			},
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-1",
					Name: "测试节点1",
					Type: "game",
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
				mockNodeStore.On("Update", mock.AnythingOfType("models.GameNode")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "node-2",
			params: UpdateNodeParams{
				Name: "更新后的节点",
				Type: "game",
			},
			mockSetup: func() {
				mockNodeStore.On("Get", "node-2").Return(models.GameNode{}, ErrNodeNotFound)
			},
			expectedError: ErrNodeNotFound,
		},
		{
			name:   "存储层返回错误",
			nodeID: "node-3",
			params: UpdateNodeParams{
				Name: "更新后的节点",
				Type: "game",
			},
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-3",
					Name: "测试节点3",
					Type: "game",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateReady,
						Online:     true,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockNodeStore.On("Get", "node-3").Return(node, nil)
				mockNodeStore.On("Update", mock.AnythingOfType("models.GameNode")).Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.UpdateNode(tt.nodeID, tt.params)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestDeleteNode(t *testing.T) {
	mockNodeStore := new(store.MockNodeStore)
	mockInstanceStore := new(store.MockInstanceStore)
	service := NewNodeService(mockNodeStore, mockInstanceStore)

	tests := []struct {
		name          string
		nodeID        string
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "成功删除节点",
			nodeID: "node-1",
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-1",
					Name: "测试节点1",
					Type: "game",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateOffline,
						Online:     false,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockNodeStore.On("Get", "node-1").Return(node, nil)
				mockNodeStore.On("Delete", "node-1").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "node-2",
			mockSetup: func() {
				mockNodeStore.On("Get", "node-2").Return(models.GameNode{}, ErrNodeNotFound)
			},
			expectedError: ErrNodeNotFound,
		},
		{
			name:   "节点正在运行",
			nodeID: "node-3",
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-3",
					Name: "测试节点3",
					Type: "game",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateReady,
						Online:     true,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockNodeStore.On("Get", "node-3").Return(node, nil)
			},
			expectedError: ErrNodeIsRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.DeleteNode(tt.nodeID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateNodeStatus(t *testing.T) {
	mockNodeStore := new(store.MockNodeStore)
	mockInstanceStore := new(store.MockInstanceStore)
	service := NewNodeService(mockNodeStore, mockInstanceStore)

	tests := []struct {
		name          string
		nodeID        string
		status        string
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "成功更新节点状态",
			nodeID: "node-1",
			status: string(models.GameNodeStateReady),
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-1",
					Name: "测试节点1",
					Type: "game",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateOffline,
						Online:     false,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockNodeStore.On("Get", "node-1").Return(node, nil)
				mockNodeStore.On("Update", mock.AnythingOfType("models.GameNode")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "node-2",
			status: string(models.GameNodeStateReady),
			mockSetup: func() {
				mockNodeStore.On("Get", "node-2").Return(models.GameNode{}, ErrNodeNotFound)
			},
			expectedError: ErrNodeNotFound,
		},
		{
			name:   "存储层返回错误",
			nodeID: "node-3",
			status: string(models.GameNodeStateReady),
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-3",
					Name: "测试节点3",
					Type: "game",
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
				mockNodeStore.On("Update", mock.AnythingOfType("models.GameNode")).Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.UpdateNodeStatus(tt.nodeID, tt.status)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateNodeMetrics(t *testing.T) {
	mockNodeStore := new(store.MockNodeStore)
	mockInstanceStore := new(store.MockInstanceStore)
	service := NewNodeService(mockNodeStore, mockInstanceStore)

	tests := []struct {
		name          string
		nodeID        string
		metrics       map[string]interface{}
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "成功更新节点指标",
			nodeID: "node-1",
			metrics: map[string]interface{}{
				"cpu_usage": 0.5,
				"mem_usage": 0.6,
			},
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-1",
					Name: "测试节点1",
					Type: "game",
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
				mockNodeStore.On("Update", mock.AnythingOfType("models.GameNode")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "node-2",
			metrics: map[string]interface{}{
				"cpu_usage": 0.5,
				"mem_usage": 0.6,
			},
			mockSetup: func() {
				mockNodeStore.On("Get", "node-2").Return(models.GameNode{}, ErrNodeNotFound)
			},
			expectedError: ErrNodeNotFound,
		},
		{
			name:   "存储层返回错误",
			nodeID: "node-3",
			metrics: map[string]interface{}{
				"cpu_usage": 0.5,
				"mem_usage": 0.6,
			},
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-3",
					Name: "测试节点3",
					Type: "game",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateReady,
						Online:     true,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockNodeStore.On("Get", "node-3").Return(node, nil)
				mockNodeStore.On("Update", mock.AnythingOfType("models.GameNode")).Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.UpdateNodeMetrics(tt.nodeID, tt.metrics)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateNodeResources(t *testing.T) {
	mockNodeStore := new(store.MockNodeStore)
	mockInstanceStore := new(store.MockInstanceStore)
	service := NewNodeService(mockNodeStore, mockInstanceStore)

	tests := []struct {
		name          string
		nodeID        string
		resources     map[string]string
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "成功更新节点资源",
			nodeID: "node-1",
			resources: map[string]string{
				"cpu": "4",
				"mem": "8G",
			},
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-1",
					Name: "测试节点1",
					Type: "game",
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
				mockNodeStore.On("Update", mock.AnythingOfType("models.GameNode")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "node-2",
			resources: map[string]string{
				"cpu": "4",
				"mem": "8G",
			},
			mockSetup: func() {
				mockNodeStore.On("Get", "node-2").Return(models.GameNode{}, ErrNodeNotFound)
			},
			expectedError: ErrNodeNotFound,
		},
		{
			name:   "存储层返回错误",
			nodeID: "node-3",
			resources: map[string]string{
				"cpu": "4",
				"mem": "8G",
			},
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-3",
					Name: "测试节点3",
					Type: "game",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateReady,
						Online:     true,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockNodeStore.On("Get", "node-3").Return(node, nil)
				mockNodeStore.On("Update", mock.AnythingOfType("models.GameNode")).Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.UpdateNodeResources(tt.nodeID, tt.resources)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateNodeOnlineStatus(t *testing.T) {
	mockNodeStore := new(store.MockNodeStore)
	mockInstanceStore := new(store.MockInstanceStore)
	service := NewNodeService(mockNodeStore, mockInstanceStore)

	tests := []struct {
		name          string
		nodeID        string
		online        bool
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "成功更新节点在线状态",
			nodeID: "node-1",
			online: true,
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-1",
					Name: "测试节点1",
					Type: "game",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateReady,
						Online:     false,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockNodeStore.On("Get", "node-1").Return(node, nil)
				mockNodeStore.On("Update", mock.AnythingOfType("models.GameNode")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "node-2",
			online: true,
			mockSetup: func() {
				mockNodeStore.On("Get", "node-2").Return(models.GameNode{}, ErrNodeNotFound)
			},
			expectedError: ErrNodeNotFound,
		},
		{
			name:   "存储层返回错误",
			nodeID: "node-3",
			online: true,
			mockSetup: func() {
				node := models.GameNode{
					ID:   "node-3",
					Name: "测试节点3",
					Type: "game",
					Status: models.GameNodeStatus{
						State:      models.GameNodeStateReady,
						Online:     false,
						LastOnline: time.Now(),
						UpdatedAt:  time.Now(),
						Resources:  make(map[string]string),
						Metrics:    make(map[string]interface{}),
					},
				}
				mockNodeStore.On("Get", "node-3").Return(node, nil)
				mockNodeStore.On("Update", mock.AnythingOfType("models.GameNode")).Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.UpdateNodeOnlineStatus(tt.nodeID, tt.online)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
